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
		deleteRows(t, "DELETE FROM entitlement_catalog WHERE managed_system_id = $1", managedSystem.Id)
		deleteRows(t, "DELETE FROM managed_system_tenant WHERE managed_system_id = $1", managedSystem.Id)
		deleteRows(t, "DELETE FROM managed_system WHERE id = $1", managedSystem.Id)
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

func TestEntitlementRepository_SearchEntitlement(t *testing.T) {
	ctx := t.Context()
	system, err := repository.ManagedSystemRepository.CreateManagedSystem(ctx, "entitlement-search-system", "https://example.com/c.wasm", "hash", &domain.ConnectorSpec{})
	require.NoError(t, err)
	t.Cleanup(func() {
		deleteRows(t, "DELETE FROM entitlement_catalog WHERE managed_system_id = $1", system.Id)
		deleteRows(t, "DELETE FROM managed_system WHERE id = $1", system.Id)
	})
	for resource, entitlements := range map[string][]domain.DiscoveredEntitlement{
		"role":  {{EntitlementID: "role-admin", Name: "Administrator"}, {EntitlementID: "role-reader", Name: "Reader"}, {EntitlementID: "role-100%", Name: "Full access"}},
		"group": {{EntitlementID: "grp-1", Name: "Admins"}, {EntitlementID: "grp-2", Name: "Super reader"}},
	} {
		for _, entitlement := range entitlements {
			_, err := repository.EntitlementCatalogRepository.UpsertEntitlementCatalog(ctx, system.Id, uuid.Nil, resource, entitlement)
			require.NoError(t, err)
		}
	}
	systemId := uuid.MustParse(system.Id)
	search := func(search domain.EntitlementSearch) ([]string, int64) {
		search.ManagedSystemId = systemId
		if search.Limit == 0 {
			search.Limit = 10
		}
		found, total, err := repository.EntitlementCatalogRepository.SearchEntitlementCatalog(ctx, search)
		require.NoError(t, err)
		externalIds := make([]string, 0, len(found))
		for _, entitlement := range found {
			externalIds = append(externalIds, entitlement.ExternalId)
		}

		return externalIds, total
	}

	t.Run("finds part of the name in any case", func(t *testing.T) {
		found, total := search(domain.EntitlementSearch{Query: "ADMIN"})
		assert.ElementsMatch(t, []string{"role-admin", "grp-1"}, found)
		assert.EqualValues(t, 2, total)
	})

	t.Run("finds part of the external id", func(t *testing.T) {
		found, _ := search(domain.EntitlementSearch{Query: "role-read"})
		assert.Equal(t, []string{"role-reader"}, found)
	})

	t.Run("a match that starts with the query comes before one that only contains it", func(t *testing.T) {
		found, _ := search(domain.EntitlementSearch{Query: "read"})
		assert.Equal(t, []string{"role-reader", "grp-2"}, found, "Reader starts with it, Super reader does not")
	})

	t.Run("a % in the query is not a wildcard", func(t *testing.T) {
		found, _ := search(domain.EntitlementSearch{Query: "%"})
		assert.Equal(t, []string{"role-100%"}, found)
	})

	t.Run("a resource filter narrows the result", func(t *testing.T) {
		found, _ := search(domain.EntitlementSearch{Query: "admin", Resource: "group"})
		assert.Equal(t, []string{"grp-1"}, found)
	})

	t.Run("a page gives its rows and the total of all matches", func(t *testing.T) {
		found, total := search(domain.EntitlementSearch{Limit: 2, Offset: 2})
		assert.Len(t, found, 2)
		assert.EqualValues(t, 5, total)
	})
}
