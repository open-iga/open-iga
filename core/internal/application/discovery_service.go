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

// discoveryTenant is a tenant of a system as discovery needs it; the zero value stands for the one tenant of a single-tenant system.
type discoveryTenant struct {
	id   uuid.UUID
	name string
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

		tenants, err := d.listDiscoveryTenants(ctx, managedSystem.Id)
		if err != nil {
			return nil, fmt.Errorf("list discovery targets: %w", err)
		}

		// Sorted, so the same systems give the same list on every run.
		for _, resource := range slices.Sorted(maps.Keys(managedSystem.ConnectorSpec.Entitlements)) {
			for _, tenant := range tenants {
				targets = append(targets, domain.DiscoveryTarget{SystemId: managedSystem.Id, SystemName: managedSystem.Name, TenantId: tenant.id, TenantName: tenant.name, Resource: resource})
			}
		}
	}

	return targets, nil
}

func (d *DiscoveryService) listDiscoveryTenants(ctx context.Context, managedSystemId string) ([]discoveryTenant, error) {
	tenants, err := d.managedSystemRepository.ListManagedSystemTenants(ctx, managedSystemId)
	if err != nil {
		return nil, fmt.Errorf("tenants of system %s: %w", managedSystemId, err)
	}

	if len(tenants) == 0 {
		return []discoveryTenant{{id: uuid.Nil, name: ""}}, nil
	}

	discoveryTenants := make([]discoveryTenant, 0, len(tenants))
	for _, tenant := range tenants {
		tenantId, err := uuid.Parse(tenant.Id)
		if err != nil {
			return nil, fmt.Errorf("tenant id %q: %w", tenant.Id, err)
		}

		discoveryTenants = append(discoveryTenants, discoveryTenant{id: tenantId, name: tenant.Name})
	}

	return discoveryTenants, nil
}
