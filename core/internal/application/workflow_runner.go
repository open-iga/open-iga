package application

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/open-iga/core/internal/domain"
)

var errCursorRepeated = errors.New("connector returned a cursor it already returned")

// ExecuteConnectorOperation is the runner the workflow engine launches with: it runs the step's connector operation on the run's identity.
// It calls the connector again with the cursor it returns, until it returns none.
func (w *WorkflowService) ExecuteConnectorOperation(ctx context.Context, step domain.WorkflowStep, identityId uuid.UUID) error {
	managedSystem, ref, err := w.resolveOperation(ctx, step.Operation)
	if err != nil {
		return fmt.Errorf("run step %q: %w", step.Name, err)
	}
	ref.IdentityId = identityId
	
	// A buggy connector can return a cursor again; without this check the loop would never end.
	seenCursors := map[string]bool{}
	for {
		operation := step.Operation
		output, err := w.connectorService.Execute(ctx, operation.System, operation.Resource, operation.Kind, operation.Action, ref)
		if err != nil {
			return fmt.Errorf("run step %q: %w", step.Name, err)
		}

		nextCursor, err := w.saveResult(ctx, managedSystem, ref, operation, output)
		if err != nil {
			return fmt.Errorf("run step %q: %w", step.Name, err)
		}

		if nextCursor == "" {
			return nil
		}

		if seenCursors[nextCursor] {
			return fmt.Errorf("run step %q: %w: %q", step.Name, errCursorRepeated, nextCursor)
		}

		seenCursors[nextCursor] = true
		ref.Cursor = nextCursor
	}
}

// saveResult saves what the operation returned and gives the next cursor, empty when there is none.
// TODO: save the results of the other operations too, once a workflow needs them.
func (w *WorkflowService) saveResult(ctx context.Context, managedSystem *domain.ManagedSystem, ref domain.OperationRef, operation domain.WorkflowOperation, output domain.ConnectorOutput) (string, error) {
	switch {
	case operation.Kind == domain.AccountActionKind && operation.Action == "create":
		return "", w.saveCreatedAccount(ctx, managedSystem, ref, output)
	case operation.Kind == domain.EntitlementKind && operation.Action == "discover":
		return w.saveDiscoveredEntitlements(ctx, managedSystem, ref, operation, output)
	default:
		return "", nil
	}
}

// saveCreatedAccount saves the account: later steps load it from the DB.
func (w *WorkflowService) saveCreatedAccount(ctx context.Context, managedSystem *domain.ManagedSystem, ref domain.OperationRef, output domain.ConnectorOutput) error {
	created, ok := output.(domain.AccountCreateOutput)
	if !ok {
		return fmt.Errorf("unexpected output %T", output)
	}

	if _, err := w.accountRepository.UpsertAccount(ctx, managedSystem.Id, ref.TenantId, ref.IdentityId, created.AccountID); err != nil {
		return fmt.Errorf("save account: %w", err)
	}

	return nil
}

// saveDiscoveredEntitlements saves one page into the catalog. Saving a page again only updates its rows, so a replay is safe.
func (w *WorkflowService) saveDiscoveredEntitlements(ctx context.Context, managedSystem *domain.ManagedSystem, ref domain.OperationRef, operation domain.WorkflowOperation, output domain.ConnectorOutput) (string, error) {
	page, ok := output.(domain.EntitlementDiscoverOutput)
	if !ok {
		return "", fmt.Errorf("unexpected output %T", output)
	}

	for _, discovered := range page.Entitlements {
		if _, err := w.entitlementCatalogRepository.UpsertEntitlementCatalog(ctx, managedSystem.Id, ref.TenantId, operation.Resource, discovered); err != nil {
			return "", fmt.Errorf("save entitlement: %w", err)
		}
	}

	return page.NextCursor, nil
}

// resolveOperation checks the operation against the managed systems and finds the rows it works on: the tenant and the entitlement.
func (w *WorkflowService) resolveOperation(ctx context.Context, operation domain.WorkflowOperation) (*domain.ManagedSystem, domain.OperationRef, error) {
	managedSystem, err := w.managedSystemRepository.GetManagedSystemByName(ctx, operation.System)
	if err != nil {
		return nil, domain.OperationRef{}, fmt.Errorf("system %q: %w", operation.System, err)
	}

	if managedSystem.ConnectorSpec == nil || !managedSystem.ConnectorSpec.SupportsResource(operation.Kind, operation.Resource) {
		return nil, domain.OperationRef{}, fmt.Errorf("%w: system %q has no %s resource %q", domain.ErrInvalidWorkflow, operation.System, operation.Kind, operation.Resource)
	}

	if !managedSystem.ConnectorSpec.SupportsAction(operation.Kind, operation.Resource, operation.Action) {
		return nil, domain.OperationRef{}, fmt.Errorf("%w: system %q does not declare %s %q on resource %q", domain.ErrInvalidWorkflow, operation.System, operation.Kind, operation.Action, operation.Resource)
	}

	var entitlement *domain.EntitlementCatalog
	if operation.Entitlement != uuid.Nil {
		if entitlement, err = w.entitlementCatalogRepository.GetEntitlementCatalog(ctx, operation.Entitlement, managedSystem.Id, operation.Resource); err != nil {
			return nil, domain.OperationRef{}, fmt.Errorf("entitlement %s: %w", operation.Entitlement, err)
		}
	}

	tenantId, err := w.pickTenant(ctx, managedSystem, operation, entitlement)
	if err != nil {
		return nil, domain.OperationRef{}, err
	}

	return managedSystem, domain.OperationRef{TenantId: tenantId, IdentityId: uuid.Nil, EntitlementId: operation.Entitlement, Cursor: ""}, nil
}

// pickTenant is the tenant of the system that fits what the operation names (its tenant, the tenant of its entitlement)
// A system with a single tenant needs neither.
func (w *WorkflowService) pickTenant(ctx context.Context, managedSystem *domain.ManagedSystem, operation domain.WorkflowOperation, entitlement *domain.EntitlementCatalog) (uuid.UUID, error) {
	tenants, err := w.managedSystemRepository.ListManagedSystemTenants(ctx, managedSystem.Id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("tenants of system %q: %w", operation.System, err)
	}

	// A system without tenant rows is single-tenant.
	isSingleTenantSystem := len(tenants) == 0
	// The operation or its entitlement can name the tenant to use.
	entitlementNamesTenant := entitlement != nil && entitlement.TenantId != ""
	tenantIsNamed := operation.Tenant != "" || entitlementNamesTenant
	// An entitlement saved while the system had no tenants names none.
	entitlementIsTenantless := entitlement != nil && entitlement.TenantId == ""

	// A single-tenant system needs no tenant, and uuid.Nil stands for it.
	if isSingleTenantSystem && !tenantIsNamed {
		return uuid.Nil, nil
	}

	// A tenantless entitlement fits none of the tenants of a system that has some.
	if !isSingleTenantSystem && entitlementIsTenantless {
		return uuid.Nil, fmt.Errorf("%w: entitlement %s belongs to no tenant, but system %q has tenants", domain.ErrEntitlementNotFound, entitlement.Id, operation.System)
	}

	return selectTenant(tenants, operation, entitlement)
}

// selectTenant is the one tenant that fits; it is an error when none or several do.
func selectTenant(tenants []*domain.ManagedSystemTenant, operation domain.WorkflowOperation, entitlement *domain.EntitlementCatalog) (uuid.UUID, error) {
	tenants = filterFittingTenants(tenants, operation, entitlement)
	if len(tenants) == 0 {
		return uuid.Nil, fmt.Errorf("%w: system %q has no tenant that fits the operation", domain.ErrTenantNotFound, operation.System)
	}

	if len(tenants) > 1 {
		return uuid.Nil, fmt.Errorf("%w: system %q has several tenants, name one", domain.ErrInvalidWorkflow, operation.System)
	}

	tenantId, err := uuid.Parse(tenants[0].Id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("tenant id %q: %w", tenants[0].Id, err)
	}

	return tenantId, nil
}

// filterFittingTenants keeps the tenants that match the tenant the operation names and the tenant of its entitlement
func filterFittingTenants(tenants []*domain.ManagedSystemTenant, operation domain.WorkflowOperation, entitlement *domain.EntitlementCatalog) []*domain.ManagedSystemTenant {
	if entitlement != nil {
		tenants = slices.DeleteFunc(tenants, func(tenant *domain.ManagedSystemTenant) bool { return tenant.Id != entitlement.TenantId })
	}

	if operation.Tenant != "" {
		tenants = slices.DeleteFunc(tenants, func(tenant *domain.ManagedSystemTenant) bool { return tenant.Name != operation.Tenant })
	}

	return tenants
}
