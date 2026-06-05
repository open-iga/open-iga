package repository

import (
	"context"
	"testing"

	"github.com/open-iga/core/internal/domain"
	"github.com/open-iga/core/internal/testutil"
	"github.com/stretchr/testify/assert"
)

var (
	mockOauthUser = testutil.NewOauthUser()
	identity      *domain.Identity
)

func SetupIdentity(t *testing.T) {
	t.Helper()
	identity, _ = repository.IdentityRepository.FindOrCreate(t.Context(), new(mockOauthUser))
	t.Cleanup(func() {
		conn.Exec(t.Context(), "DELETE FROM identity where email = $1", mockOauthUser.Email)
	})
}

func TestSessionRepository_Create(t *testing.T) {
	t.Run("generates session from identity and save it in ", func(t *testing.T) {
		SetupIdentity(t)
		session, err := repository.SessionRepository.Create(context.Background(), identity)

		assert.Nil(t, err)
		assert.NotNil(t, session)
		t.Cleanup(func() {
			_, _ = conn.Exec(context.Background(), "DELETE FROM session WHERE session_id = $1", session.SessionId)
		})
		assert.Equal(t, identity.Id.String(), session.IdentityId.String())
	})

	t.Run("throws error if session already exists", func(t *testing.T) {
		SetupIdentity(t)

		session, err := repository.SessionRepository.Create(context.Background(), identity)

		assert.Nil(t, err)
		assert.NotNil(t, session)
		t.Cleanup(func() {
			_, _ = conn.Exec(context.Background(), "DELETE FROM session WHERE session_id = $1", session.SessionId)
		})

		invalidSession, err := repository.SessionRepository.Create(context.Background(), identity)
		assert.NotNil(t, err)
		assert.Nil(t, invalidSession)
	})
}

func TestSessionRepository_FindBySessionId(t *testing.T) {
	t.Run("returns no session found when there are no sessions for a session ID", func(t *testing.T) {
		SetupIdentity(t)
		session, _ := repository.SessionRepository.Create(context.Background(), identity)
		assert.NotNil(t, session)
		t.Cleanup(func() {
			_, _ = conn.Exec(context.Background(), "DELETE FROM session WHERE session_id = $1", session.SessionId)
		})

		i, s, err := repository.SessionRepository.FindBySessionId(context.Background(), "unknown-id")

		assert.Nil(t, s)
		assert.Nil(t, i)
		assert.Equal(t, err, domain.SessionNotFound)
	})

	t.Run("finds session by id", func(t *testing.T) {
		SetupIdentity(t)
		session, _ := repository.SessionRepository.Create(context.Background(), identity)
		assert.NotNil(t, session)
		t.Cleanup(func() {
			_, _ = conn.Exec(context.Background(), "DELETE FROM session WHERE session_id = $1", session.SessionId)
		})

		i, s, err := repository.SessionRepository.FindBySessionId(context.Background(), session.SessionId)

		assert.Nil(t, err)
		assert.NotNil(t, i)
		assert.NotNil(t, s)
		assert.Equal(t, identity, i)
		assert.Equal(t, session, s)
	})
}

func TestSessionRepository_DeactivateByIdentityId(t *testing.T) {
	t.Run("deactivates session by identity id", func(t *testing.T) {
		SetupIdentity(t)
		session, err := repository.SessionRepository.Create(context.Background(), identity)

		assert.Nil(t, err)
		assert.NotNil(t, session)
		t.Cleanup(func() {
			_, _ = conn.Exec(context.Background(), "DELETE FROM session WHERE session_id = $1", session.SessionId)
		})

		deactivatedSession, err := repository.SessionRepository.DeactivateByIdentityId(context.Background(), identity.Id)

		assert.Nil(t, err)
		assert.NotNil(t, deactivatedSession)
		assert.Equal(t, session.Id, deactivatedSession.Id)
		assert.Equal(t, session.SessionId, deactivatedSession.SessionId)
		assert.Equal(t, identity.Id, deactivatedSession.IdentityId)
		assert.False(t, deactivatedSession.Active)
	})

	t.Run("returns error when failed to find the session by identity id", func(t *testing.T) {
		SetupIdentity(t)

		deactivatedSession, err := repository.SessionRepository.DeactivateByIdentityId(context.Background(), identity.Id)

		assert.NotNil(t, err)
		assert.Nil(t, deactivatedSession)
	})
}

func TestSessionRepository_DeactivateBySessionId(t *testing.T) {
	t.Run("deactivates session by session id", func(t *testing.T) {
		SetupIdentity(t)
		session, err := repository.SessionRepository.Create(context.Background(), identity)

		assert.Nil(t, err)
		assert.NotNil(t, session)
		t.Cleanup(func() {
			_, _ = conn.Exec(context.Background(), "DELETE FROM session WHERE session_id = $1", session.SessionId)
		})

		deactivatedSession, err := repository.SessionRepository.DeactivateBySessionId(context.Background(), session.SessionId)

		assert.Nil(t, err)
		assert.NotNil(t, deactivatedSession)
		assert.Equal(t, session.Id, deactivatedSession.Id)
		assert.Equal(t, session.SessionId, deactivatedSession.SessionId)
		assert.Equal(t, identity.Id, deactivatedSession.IdentityId)
		assert.False(t, deactivatedSession.Active)
	})

	t.Run("returns error when failed to find the session by session id", func(t *testing.T) {
		SetupIdentity(t)

		deactivatedSession, err := repository.SessionRepository.DeactivateBySessionId(context.Background(), "unknown-id")

		assert.NotNil(t, err)
		assert.Nil(t, deactivatedSession)
	})
}

func TestSessionRepository_FindActiveSessionByIdentityId(t *testing.T) {
	t.Run("returns no session found when there are no active sessions for a session ID", func(t *testing.T) {
		SetupIdentity(t)

		session, err := repository.SessionRepository.FindActiveSessionByIdentityId(context.Background(), identity.Id)

		assert.Nil(t, session)
		assert.Equal(t, domain.NoActiveSession, err)
	})

	t.Run("finds session by id", func(t *testing.T) {
		SetupIdentity(t)
		session, err := repository.SessionRepository.Create(context.Background(), identity)

		assert.Nil(t, err)
		assert.NotNil(t, session)
		t.Cleanup(func() {
			_, _ = conn.Exec(context.Background(), "DELETE FROM session WHERE session_id = $1", session.SessionId)
		})

		activeSession, err := repository.SessionRepository.FindActiveSessionByIdentityId(context.Background(), identity.Id)

		assert.Nil(t, err)
		assert.NotNil(t, activeSession)
		assert.Equal(t, session.Id, activeSession.Id)
		assert.Equal(t, session.SessionId, activeSession.SessionId)
		assert.Equal(t, identity.Id, activeSession.IdentityId)
		assert.True(t, activeSession.Active)
	})
}
