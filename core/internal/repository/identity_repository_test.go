package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/open-iga/core/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestIdentityRepository_FindOrCreate(t *testing.T) {
	t.Run("returns error if user is nil", func(t *testing.T) {
		identity, err := repository.IdentityRepository.FindOrCreateWithDefaultRole(context.TODO(), nil)

		assert.Nil(t, identity)
		assert.EqualError(t, err, "user is nil")
	})

	t.Run("returns error if email is empty", func(t *testing.T) {
		emptyUser := testutil.NewOauthUser()
		emptyUser.Email = ""

		identity, err := repository.IdentityRepository.FindOrCreateWithDefaultRole(context.TODO(), &emptyUser)

		assert.Nil(t, identity)
		assert.EqualError(t, err, "email is empty")
	})

	t.Run("should insert identity if email already exists", func(t *testing.T) {
		mockOauthUser := testutil.NewOauthUser()

		identity, err := repository.IdentityRepository.FindOrCreateWithDefaultRole(context.TODO(), &mockOauthUser)

		assert.Nil(t, err)

		assert.NotNil(t, identity)
		t.Cleanup(func() {
			_, _ = conn.Exec(context.Background(), "DELETE FROM identity WHERE email = $1", identity.Email)
		})
		assert.IsType(t, uuid.UUID{}, identity.Id)
		assert.Equal(t, mockOauthUser.Email, identity.Email)
		assert.IsType(t, mockOauthUser.FirstName, identity.FirstName)
		assert.IsType(t, mockOauthUser.LastName, identity.LastName)
	})

	t.Run("should not insert identity if email already exists", func(t *testing.T) {
		mockOauthUser := testutil.NewOauthUser()

		identity, err := repository.IdentityRepository.FindOrCreateWithDefaultRole(context.TODO(), &mockOauthUser)

		assert.Nil(t, err)
		assert.NotNil(t, identity)
		t.Cleanup(func() {
			_, _ = conn.Exec(context.Background(), "DELETE FROM identity WHERE email = $1", identity.Email)
		})

		upsertedIdentity, upsertingErr := repository.IdentityRepository.FindOrCreateWithDefaultRole(context.TODO(), &mockOauthUser)

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
}
