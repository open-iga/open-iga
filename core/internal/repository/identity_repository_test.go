package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/open-iga/core/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestIdentityRepository_FindOrCreate(t *testing.T) {
	t.Run("should insert identity if email already exists", func(t *testing.T) {
		mockOauthUser := testutil.NewOauthUser()

		identity, err := repository.IdentityRepository.FindOrCreate(context.TODO(), new(mockOauthUser))

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

		identity, err := repository.IdentityRepository.FindOrCreate(context.TODO(), new(mockOauthUser))

		assert.Nil(t, err)
		assert.NotNil(t, identity)
		t.Cleanup(func() {
			_, _ = conn.Exec(context.Background(), "DELETE FROM identity WHERE email = $1", identity.Email)
		})

		upsertedIdentity, upsertingErr := repository.IdentityRepository.FindOrCreate(context.TODO(), new(mockOauthUser))

		assert.Nil(t, upsertingErr)
		assert.Equal(t, identity, upsertedIdentity)

		rows, err := conn.Query(context.Background(), "SELECT * FROM identity WHERE email = $1", identity.Email)
		defer rows.Close()

		assert.Nil(t, err)
		count := 0
		for rows.Next() {
			count++
		}
		assert.Equal(t, count, 1)
	})
}
