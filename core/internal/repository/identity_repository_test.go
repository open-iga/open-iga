package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/open-iga/core/internal/domain"
	"github.com/open-iga/core/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cleanupDbOnExit(t *testing.T, identity *domain.Identity) {
	t.Helper()

	t.Cleanup(func() {
		_, _ = conn.Exec(context.Background(), "DELETE FROM identity_role WHERE identity_id = $1", identity.Id)
		_, _ = conn.Exec(context.Background(), "DELETE FROM identity WHERE email = $1", identity.Email)
	})
}

func TestIdentityRepository_FindOrCreate(t *testing.T) {
	t.Run("returns error if user is nil", func(t *testing.T) {
		identity, err := repository.IdentityRepository.FindOrCreateWithRole(context.TODO(), nil, domain.DefaultIdentityRole)

		assert.Nil(t, identity)
		assert.EqualError(t, err, "user is nil")
	})

	t.Run("returns error if email is empty", func(t *testing.T) {
		emptyUser := testutil.NewOauthUser()
		emptyUser.Email = ""

		identity, err := repository.IdentityRepository.FindOrCreateWithRole(context.TODO(), &emptyUser, domain.DefaultIdentityRole)

		assert.Nil(t, identity)
		assert.EqualError(t, err, "email is empty")
	})

	t.Run("should insert identity if email already exists", func(t *testing.T) {
		mockOauthUser := testutil.NewOauthUser()

		identity, err := repository.IdentityRepository.FindOrCreateWithRole(context.TODO(), &mockOauthUser, domain.DefaultIdentityRole)

		assert.Nil(t, err)

		assert.NotNil(t, identity)
		cleanupDbOnExit(t, identity)

		assert.IsType(t, uuid.UUID{}, identity.Id)
		assert.Equal(t, mockOauthUser.Email, identity.Email)
		assert.IsType(t, mockOauthUser.FirstName, identity.FirstName)
		assert.IsType(t, mockOauthUser.LastName, identity.LastName)
	})

	t.Run("should not insert identity if email already exists", func(t *testing.T) {
		mockOauthUser := testutil.NewOauthUser()

		identity, err := repository.IdentityRepository.FindOrCreateWithRole(context.TODO(), &mockOauthUser, domain.DefaultIdentityRole)

		assert.Nil(t, err)
		assert.NotNil(t, identity)
		cleanupDbOnExit(t, identity)

		upsertedIdentity, upsertingErr := repository.IdentityRepository.FindOrCreateWithRole(context.TODO(), &mockOauthUser, domain.DefaultIdentityRole)

		assert.Nil(t, upsertingErr)
		assert.Equal(t, identity, upsertedIdentity)

		rows, err := conn.Query(context.Background(), "SELECT * FROM identity WHERE email = $1", identity.Email)

		assert.Nil(t, err)
		defer rows.Close()

		count := 0
		for rows.Next() {
			count++
		}
		assert.Equal(t, count, 1)
	})

	t.Run("should upsert role when upserting identity details", func(t *testing.T) {
		mockOauthUser := testutil.NewOauthUser()

		identity, err := repository.IdentityRepository.FindOrCreateWithRole(context.TODO(), &mockOauthUser, domain.DefaultIdentityRole)

		assert.Nil(t, err)
		assert.NotNil(t, identity)
		cleanupDbOnExit(t, identity)

		identityRole, err := repository.IdentityRepository.GetRolesByIdentityId(context.TODO(), identity.Id)

		assert.Nil(t, err)
		assert.NotNil(t, identityRole)
		assert.Contains(t, identityRole.Roles, domain.DefaultIdentityRole)
	})
}

func TestIdentityRepository_GetRolesByIdentityId(t *testing.T) {
	t.Run("returns role for a give", func(t *testing.T) {
		mockOauthUser := testutil.NewOauthUser()

		identity, err := repository.IdentityRepository.FindOrCreateWithRole(context.TODO(), &mockOauthUser, domain.DefaultIdentityRole)
		require.NoError(t, err)
		cleanupDbOnExit(t, identity)

		identityRole, err := repository.IdentityRepository.GetRolesByIdentityId(context.TODO(), identity.Id)

		assert.Nil(t, err)
		assert.NotNil(t, identityRole)
		assert.Contains(t, identityRole.Roles, domain.DefaultIdentityRole)
	})

	t.Run("returns no identity found when no identity exists a given ID", func(t *testing.T) {
		identityRole, err := repository.IdentityRepository.GetRolesByIdentityId(context.TODO(), uuid.New())

		assert.Nil(t, identityRole)
		assert.ErrorIs(t, err, domain.ErrNoIdentityFound)
	})
}

func TestIdentityRepository_UpsertRoleByIdentityId(t *testing.T) {

	t.Run("returns error if user is nil", func(t *testing.T) {
		identityRole, err := repository.IdentityRepository.UpsertRoleByIdentityId(context.TODO(), uuid.Nil, "admin")

		assert.Nil(t, identityRole)
		assert.Error(t, err)
	})

	t.Run("inserts a new role if the identity does not have the role", func(t *testing.T) {
		mockOauthUser := testutil.NewOauthUser()

		identity, err := repository.IdentityRepository.FindOrCreateWithRole(context.TODO(), &mockOauthUser, domain.DefaultIdentityRole)
		require.NoError(t, err)
		cleanupDbOnExit(t, identity)

		identityRole, err := repository.IdentityRepository.UpsertRoleByIdentityId(context.TODO(), identity.Id, "admin")

		assert.Nil(t, err)
		assert.NotNil(t, identityRole)
		assert.Equal(t, 2, len(identityRole.Roles))
		assert.Contains(t, identityRole.Roles, "admin")
		assert.Contains(t, identityRole.Roles, domain.DefaultIdentityRole)
	})

	t.Run("does nothing if the identity already has the role", func(t *testing.T) {
		mockOauthUser := testutil.NewOauthUser()

		identity, err := repository.IdentityRepository.FindOrCreateWithRole(context.TODO(), &mockOauthUser, domain.DefaultIdentityRole)
		require.NoError(t, err)
		cleanupDbOnExit(t, identity)

		identityRoles, err := repository.IdentityRepository.UpsertRoleByIdentityId(context.TODO(), identity.Id, domain.DefaultIdentityRole)

		assert.Nil(t, err)
		assert.Equal(t, 1, len(identityRoles.Roles))
		assert.Contains(t, identityRoles.Roles, domain.DefaultIdentityRole)
	})
}

func TestIdentityRepository_HasAdmin(t *testing.T) {
	t.Run("returns true if atleast one admin exists", func(t *testing.T) {
		mockOauthUser := testutil.NewOauthUser()
		mockOauthUser.Email = uuid.New().String() + "@test.com"

		identity, err := repository.IdentityRepository.FindOrCreateWithRole(context.TODO(), &mockOauthUser, domain.DefaultIdentityRole)
		require.NoError(t, err)
		require.NotNil(t, identity)
		cleanupDbOnExit(t, identity)

		_, err = repository.IdentityRepository.UpsertRoleByIdentityId(context.TODO(), identity.Id, "admin")
		require.NoError(t, err)

		hasAdmin, err := repository.IdentityRepository.HasAdmin(context.TODO())
		assert.NoError(t, err)
		assert.True(t, hasAdmin)
	})

	t.Run("returns false if no admin exists", func(t *testing.T) {
		hasAdmin, err := repository.IdentityRepository.HasAdmin(context.TODO())
		assert.NoError(t, err)
		assert.False(t, hasAdmin)
	})
}
