package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/open-iga/core/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntitlementRepository_UpsertEntitlement(t *testing.T) {
	managedSystem, err := repository.ManagedSystemRepository.CreateManagedSystem(t.Context(), "entitlement-test-system", "https://example.com/c.wasm", "hash", &domain.ConnectorSpec{})
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = conn.Exec(t.Context(), "DELETE FROM entitlement_catalog WHERE managed_system_id = $1", managedSystem.Id)
		_, _ = conn.Exec(t.Context(), "DELETE FROM managed_system_tenant WHERE managed_system_id = $1", managedSystem.Id)
		_, _ = conn.Exec(t.Context(), "DELETE FROM managed_system WHERE id = $1", managedSystem.Id)
	})
	tenant, _, err := repository.ManagedSystemRepository.CreateManagedSystemTenant(t.Context(), managedSystem.Id, "tenant-1", "Tenant 1")
	require.NoError(t, err)
	tenantId := uuid.MustParse(tenant.Id)

	description := "can do everything"
	first, err := repository.EntitlementCatalogRepository.UpsertEntitlementCatalog(t.Context(), managedSystem.Id, tenantId, "iam-user", domain.DiscoveredEntitlement{
		EntitlementID: "role-admin", Name: "Admin", Description: &description, Metadata: domain.Metadata{"risk": "high"},
	})
	require.NoError(t, err)
	assert.Equal(t, "role-admin", first.ExternalId)
	assert.Equal(t, "Admin", first.DisplayName)
	assert.Equal(t, description, first.Description)
	assert.Equal(t, domain.Metadata{"risk": "high"}, first.Metadata)

	second, err := repository.EntitlementCatalogRepository.UpsertEntitlementCatalog(t.Context(), managedSystem.Id, tenantId, "iam-user", domain.DiscoveredEntitlement{EntitlementID: "role-admin", Name: "Administrator"})
	require.NoError(t, err)
	assert.Equal(t, first.Id, second.Id, "discovery finding it again keeps one row")
	assert.Equal(t, "Administrator", second.DisplayName)

	other, err := repository.EntitlementCatalogRepository.UpsertEntitlementCatalog(t.Context(), managedSystem.Id, tenantId, "iam-group", domain.DiscoveredEntitlement{EntitlementID: "role-admin", Name: "Admin group"})
	require.NoError(t, err)
	assert.NotEqual(t, first.Id, other.Id, "the same external id under another resource is another entitlement")

	loaded, err := repository.EntitlementCatalogRepository.GetEntitlementCatalog(t.Context(), uuid.MustParse(first.Id), managedSystem.Id, "iam-user")
	require.NoError(t, err)
	assert.Equal(t, "iam-user", loaded.Resource)

	_, err = repository.EntitlementCatalogRepository.GetEntitlementCatalog(t.Context(), uuid.MustParse(first.Id), managedSystem.Id, "iam-group")
	assert.ErrorIs(t, err, domain.ErrEntitlementNotFound, "the same entitlement under another resource is not found")

	_, err = repository.EntitlementCatalogRepository.GetEntitlementCatalog(t.Context(), uuid.MustParse(first.Id), uuid.NewString(), "iam-user")
	assert.ErrorIs(t, err, domain.ErrEntitlementNotFound, "the same entitlement under another system is not found")

	// A single-tenant system has no tenant: the entitlement hangs on the system alone.
	single, err := repository.EntitlementCatalogRepository.UpsertEntitlementCatalog(t.Context(), managedSystem.Id, uuid.Nil, "iam-user", domain.DiscoveredEntitlement{EntitlementID: "role-admin", Name: "Admin"})
	require.NoError(t, err)
	assert.Empty(t, single.TenantId)
	again, err := repository.EntitlementCatalogRepository.UpsertEntitlementCatalog(t.Context(), managedSystem.Id, uuid.Nil, "iam-user", domain.DiscoveredEntitlement{EntitlementID: "role-admin", Name: "Administrator"})
	require.NoError(t, err)
	assert.Equal(t, single.Id, again.Id, "discovery finding it again keeps one row, even without a tenant")
}
