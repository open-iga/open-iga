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
	managedSystemRepository contract.ManagedSystemRepository
}

var _ contract.DiscoveryService = (*DiscoveryService)(nil)

func NewDiscoveryService(managedSystemRepository contract.ManagedSystemRepository) *DiscoveryService {
	return &DiscoveryService{managedSystemRepository: managedSystemRepository}
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
