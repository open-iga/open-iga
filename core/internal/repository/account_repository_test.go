package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/open-iga/core/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountRepository_UpsertAccount(t *testing.T) {
	SetupIdentity(t)
	managedSystem, err := repository.ManagedSystemRepository.CreateManagedSystem(t.Context(), "account-test-system", "https://example.com/c.wasm", "hash", &domain.ConnectorSpec{})
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = conn.Exec(t.Context(), "DELETE FROM account WHERE identity_id = $1", identity.Id)
		_, _ = conn.Exec(t.Context(), "DELETE FROM managed_system_tenant WHERE managed_system_id = $1", managedSystem.Id)
		_, _ = conn.Exec(t.Context(), "DELETE FROM managed_system WHERE id = $1", managedSystem.Id)
	})
	tenant, _, err := repository.ManagedSystemRepository.CreateManagedSystemTenant(t.Context(), managedSystem.Id, "tenant-1", "Tenant 1")
	require.NoError(t, err)
	tenantId := uuid.MustParse(tenant.Id)

	first, err := repository.AccountRepository.UpsertAccount(t.Context(), managedSystem.Id, tenantId, identity.Id, "arn-1")
	require.NoError(t, err)
	assert.Equal(t, "arn-1", first.ExternalId)

	second, err := repository.AccountRepository.UpsertAccount(t.Context(), managedSystem.Id, tenantId, identity.Id, "arn-2")
	require.NoError(t, err)
	assert.Equal(t, first.Id, second.Id, "the same identity on the same tenant keeps one account")

	loaded, err := repository.AccountRepository.GetAccount(t.Context(), managedSystem.Id, tenantId, identity.Id)
	require.NoError(t, err)
	assert.Equal(t, "arn-2", loaded.ExternalId)

	// A single-tenant system has no tenant: the account hangs on the system alone.
	single, err := repository.AccountRepository.UpsertAccount(t.Context(), managedSystem.Id, uuid.Nil, identity.Id, "arn-single")
	require.NoError(t, err)
	assert.Empty(t, single.TenantId)
	assert.NotEqual(t, first.Id, single.Id, "no tenant is not the same as a tenant")

	again, err := repository.AccountRepository.UpsertAccount(t.Context(), managedSystem.Id, uuid.Nil, identity.Id, "arn-single-2")
	require.NoError(t, err)
	assert.Equal(t, single.Id, again.Id, "the same identity on the same tenantless system keeps one account")

	loadedSingle, err := repository.AccountRepository.GetAccount(t.Context(), managedSystem.Id, uuid.Nil, identity.Id)
	require.NoError(t, err)
	assert.Equal(t, "arn-single-2", loadedSingle.ExternalId)
}
