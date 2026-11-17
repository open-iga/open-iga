package application

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/google/uuid"
	"github.com/open-iga/core/internal/contract"
	"github.com/open-iga/core/internal/domain"
)

type DiscoveryService struct {
	connectorService             contract.ConnectorService
	managedSystemRepository      contract.ManagedSystemRepository
	entitlementCatalogRepository contract.EntitlementCatalogRepository
}

var _ contract.DiscoveryService = (*DiscoveryService)(nil)

func NewDiscoveryService(connectorService contract.ConnectorService, managedSystemRepository contract.ManagedSystemRepository, entitlementCatalogRepository contract.EntitlementCatalogRepository) *DiscoveryService {
	return &DiscoveryService{
		connectorService:             connectorService,
		managedSystemRepository:      managedSystemRepository,
		entitlementCatalogRepository: entitlementCatalogRepository,
	}
}

// ListDiscoveryTargets lists every entitlement resource of every managed system, once per tenant.
// A system without tenant rows is single-tenant and gives one target per resource.
func (d *DiscoveryService) ListDiscoveryTargets(ctx context.Context) ([]domain.DiscoveryTarget, error) {
	managedSystems, err := d.managedSystemRepository.ListManagedSystems(ctx)
	if err != nil {
		return nil, fmt.Errorf("list discovery targets: %w", err)
	}

	var targets []domain.DiscoveryTarget
	for _, managedSystem := range managedSystems {
		if managedSystem.ConnectorSpec == nil {
			continue
		}

		tenantIds, err := d.listTenantIds(ctx, managedSystem.Id)
		if err != nil {
			return nil, fmt.Errorf("list discovery targets: %w", err)
		}

		// Sorted, so the same systems give the same list on every run.
		for _, resource := range slices.Sorted(maps.Keys(managedSystem.ConnectorSpec.Entitlements)) {
			for _, tenantId := range tenantIds {
				targets = append(targets, domain.DiscoveryTarget{SystemId: managedSystem.Id, SystemName: managedSystem.Name, TenantId: tenantId, Resource: resource})
			}
		}
	}

	return targets, nil
}

// DiscoverPage reads one page of the target from the connector and saves its entitlements.
// An empty cursor reads the first page. It returns the cursor of the next page, empty after the last page.
// Saving the same page again only updates its rows, so a replay is safe.
func (d *DiscoveryService) DiscoverPage(ctx context.Context, target domain.DiscoveryTarget, cursor string) (string, error) {
	output, err := d.connectorService.Execute(ctx, target.SystemName, target.Resource, domain.EntitlementKind, "discover", domain.OperationRef{TenantId: target.TenantId, IdentityId: uuid.Nil, EntitlementId: uuid.Nil, Cursor: cursor})
	if err != nil {
		return "", fmt.Errorf("discover %s/%s: %w", target.SystemName, target.Resource, err)
	}

	page, ok := output.(domain.EntitlementDiscoverOutput)
	if !ok {
		return "", fmt.Errorf("discover %s/%s: unexpected output %T", target.SystemName, target.Resource, output)
	}

	for _, discovered := range page.Entitlements {
		if _, err := d.entitlementCatalogRepository.UpsertEntitlementCatalog(ctx, target.SystemId, target.TenantId, target.Resource, discovered); err != nil {
			return "", fmt.Errorf("discover %s/%s: %w", target.SystemName, target.Resource, err)
		}
	}

	return page.NextCursor, nil
}

// listTenantIds gives uuid.Nil alone for a single-tenant system.
func (d *DiscoveryService) listTenantIds(ctx context.Context, managedSystemId string) ([]uuid.UUID, error) {
	tenants, err := d.managedSystemRepository.ListManagedSystemTenants(ctx, managedSystemId)
	if err != nil {
		return nil, fmt.Errorf("tenants of system %s: %w", managedSystemId, err)
	}

	if len(tenants) == 0 {
		return []uuid.UUID{uuid.Nil}, nil
	}

	tenantIds := make([]uuid.UUID, 0, len(tenants))
	for _, tenant := range tenants {
		tenantId, err := uuid.Parse(tenant.Id)
		if err != nil {
			return nil, fmt.Errorf("tenant id %q: %w", tenant.Id, err)
		}

		tenantIds = append(tenantIds, tenantId)
	}

	return tenantIds, nil
}
