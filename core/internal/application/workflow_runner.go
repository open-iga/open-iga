package application

import (
	"context"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/open-iga/core/internal/domain"
)

// ExecuteConnectorOperation is the runner the workflow engine launches with: it runs the step's connector operation on the run's identity.
func (w *WorkflowService) ExecuteConnectorOperation(ctx context.Context, step domain.WorkflowStep, identityId uuid.UUID) error {
	managedSystem, ref, err := w.resolveOperation(ctx, step.Operation)
	if err != nil {
		return fmt.Errorf("run step %q: %w", step.Name, err)
	}
	ref.IdentityId = identityId

	operation := step.Operation
	output, err := w.connectorService.Execute(ctx, operation.System, operation.Resource, operation.Kind, operation.Action, ref)
	if err != nil {
		return fmt.Errorf("run step %q: %w", step.Name, err)
	}

	// Later steps load the account from the DB, so a created account must be saved here.
	// TODO: save the results of the other operations too, once a workflow needs them.
	if created, ok := output.(domain.AccountCreateOutput); ok {
		if _, err := w.accountRepository.UpsertAccount(ctx, managedSystem.Id, ref.TenantId, identityId, created.AccountID); err != nil {
			return fmt.Errorf("run step %q: save account: %w", step.Name, err)
		}
	}

	return nil
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

	return managedSystem, domain.OperationRef{TenantId: tenantId, IdentityId: uuid.Nil, EntitlementId: operation.Entitlement}, nil
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
