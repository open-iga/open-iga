package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/open-iga/core/internal/contract"
	"github.com/open-iga/core/internal/domain"
	"github.com/open-iga/core/internal/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func setupAuthServiceWithMocks(t *testing.T) (*AuthService, *testutil.MockOauth2ClientAdapter, *testutil.MockSessionRepository, *testutil.MockIdentityRepository) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	sessionRepoMock := testutil.NewMockSessionRepository(ctrl)
	identityRepoMock := testutil.NewMockIdentityRepository(ctrl)
	oauth2Client := testutil.NewMockOauth2ClientAdapter(ctrl)

	authService := NewAuthService(
		contract.Oauth2Clients{contract.Google: oauth2Client},
		testutil.NewTestLogger(),
		sessionRepoMock, identityRepoMock,
	)

	return authService, oauth2Client, sessionRepoMock, identityRepoMock
}

func TestAuthService_GetConsentPageDetails(t *testing.T) {
	t.Run("returns error when the provider is unsupported", func(t *testing.T) {
		authService, _, _, _ := setupAuthServiceWithMocks(t)

		consentDetails, err := authService.GetConsentPageDetails(context.TODO(), "unknown-provider")

		assert.Nil(t, consentDetails)
		assert.ErrorContains(t, err, "unsupported provider")
	})

	t.Run("returns error when client fails to generate consent page details", func(t *testing.T) {
		authService, oauth2ClientMock, _, _ := setupAuthServiceWithMocks(t)
		oauth2ClientMock.EXPECT().GetConsentDetails(gomock.Any()).Return(nil, errors.New("failed to generate"))

		consentDetails, err := authService.GetConsentPageDetails(context.TODO(), "google")

		assert.Nil(t, consentDetails)
		assert.ErrorContains(t, err, "failed to generate")
	})

	t.Run("returns consent page details", func(t *testing.T) {
		authService, oauth2ClientMock, _, _ := setupAuthServiceWithMocks(t)
		oauth2ClientMock.EXPECT().GetConsentDetails(gomock.Any()).Return(&domain.ConsentDetails{
			AuthCodeURL: "auth-code-url",
			State:       "state",
		}, nil)

		consentDetails, err := authService.GetConsentPageDetails(context.TODO(), "google")

		assert.Nil(t, err)
		assert.Equal(t, "auth-code-url", consentDetails.AuthCodeURL)
		assert.Equal(t, "state", consentDetails.State)
	})
}

func TestAuthService_GenerateSession(t *testing.T) {
	t.Run("returns error when the provider is unsupported", func(t *testing.T) {
		authService, _, _, _ := setupAuthServiceWithMocks(t)

		consentDetails, err := authService.GenerateSession(context.TODO(), "unknown-provider", "code")

		assert.Nil(t, consentDetails)
		assert.ErrorContains(t, err, "unsupported provider")
	})

	t.Run("returns error when provider client returns an error during user details fetch", func(t *testing.T) {
		authService, oauth2ClientMock, _, _ := setupAuthServiceWithMocks(t)
		oauth2ClientMock.EXPECT().FetchOauthUser(gomock.Any(), "code").Return(nil, errors.New("failed to fetch user"))

		consentDetails, err := authService.GenerateSession(context.TODO(), "google", "code")

		assert.Nil(t, consentDetails)
		assert.ErrorContains(t, err, "failed to fetch user")
	})

	t.Run("returns error when persisting/fetching user details fail", func(t *testing.T) {
		authService, oauth2ClientMock, _, identityRepoMock := setupAuthServiceWithMocks(t)
		oauth2ClientMock.EXPECT().FetchOauthUser(gomock.Any(), "code").Return(
			&domain.OauthUser{
				Email:     "user@gmail.com",
				FirstName: "firstname",
				LastName:  "lastname",
			},
			nil,
		)
		identityRepoMock.EXPECT().FindOrCreateWithDefaultRole(gomock.Any(), gomock.Any()).Return(nil, errors.New("failed to fetch user"))

		consentDetails, err := authService.GenerateSession(context.TODO(), "google", "code")

		assert.Nil(t, consentDetails)
		assert.ErrorContains(t, err, "failed to fetch user")
	})

	t.Run("returns error when the session finder results in non-active session error", func(t *testing.T) {
		authService, oauth2ClientMock, sessionRepoMock, identityRepoMock := setupAuthServiceWithMocks(t)
		oauth2ClientMock.EXPECT().FetchOauthUser(gomock.Any(), "code").Return(
			&domain.OauthUser{Email: "user@gmail.com", FirstName: "firstname", LastName: "lastname"},
			nil,
		)
		identity := &domain.Identity{Id: uuid.New()}
		identityRepoMock.EXPECT().FindOrCreateWithDefaultRole(gomock.Any(), gomock.Any()).Return(identity, nil)
		sessionRepoMock.EXPECT().FindActiveSessionByIdentityId(gomock.Any(), identity.Id).Return(nil, errors.New("db down"))

		session, err := authService.GenerateSession(context.TODO(), "google", "code")

		assert.Nil(t, session)
		assert.ErrorContains(t, err, "db down")
	})

	t.Run("returns the active session if it is not expired", func(t *testing.T) {
		authService, oauth2ClientMock, sessionRepoMock, identityRepoMock := setupAuthServiceWithMocks(t)
		oauth2ClientMock.EXPECT().FetchOauthUser(gomock.Any(), "code").Return(
			&domain.OauthUser{Email: "user@gmail.com", FirstName: "firstname", LastName: "lastname"},
			nil,
		)
		identity := &domain.Identity{Id: uuid.New()}
		identityRepoMock.EXPECT().FindOrCreateWithDefaultRole(gomock.Any(), gomock.Any()).Return(identity, nil)
		existingSession := &domain.Session{SessionId: "sid-1", ExpiresAt: time.Now().UTC().Add(time.Hour)}
		sessionRepoMock.EXPECT().FindActiveSessionByIdentityId(gomock.Any(), identity.Id).Return(existingSession, nil)

		session, err := authService.GenerateSession(context.TODO(), "google", "code")

		assert.NoError(t, err)
		assert.Equal(t, existingSession, session)
	})

	t.Run("deactivates and returns a new session if the existing session is expired", func(t *testing.T) {
		authService, oauth2ClientMock, sessionRepoMock, identityRepoMock := setupAuthServiceWithMocks(t)
		oauth2ClientMock.EXPECT().FetchOauthUser(gomock.Any(), "code").Return(
			&domain.OauthUser{Email: "user@gmail.com", FirstName: "firstname", LastName: "lastname"},
			nil,
		)
		identity := &domain.Identity{Id: uuid.New()}
		identityRepoMock.EXPECT().FindOrCreateWithDefaultRole(gomock.Any(), gomock.Any()).Return(identity, nil)
		existingSession := &domain.Session{SessionId: "sid-1", ExpiresAt: time.Now().UTC().Add(-time.Hour)}
		sessionRepoMock.EXPECT().FindActiveSessionByIdentityId(gomock.Any(), identity.Id).Return(existingSession, nil)
		sessionRepoMock.EXPECT().DeactivateByIdentityId(gomock.Any(), identity.Id).Return(existingSession, nil)
		newSession := &domain.Session{SessionId: "sid-2", ExpiresAt: time.Now().UTC().Add(time.Hour)}
		sessionRepoMock.EXPECT().Create(gomock.Any(), identity).Return(newSession, nil)

		session, err := authService.GenerateSession(context.TODO(), "google", "code")

		assert.NoError(t, err)
		assert.Equal(t, newSession, session)
	})

	t.Run("returns a new session when there are no active sessions yet", func(t *testing.T) {
		authService, oauth2ClientMock, sessionRepoMock, identityRepoMock := setupAuthServiceWithMocks(t)
		oauth2ClientMock.EXPECT().FetchOauthUser(gomock.Any(), "code").Return(
			&domain.OauthUser{Email: "user@gmail.com", FirstName: "firstname", LastName: "lastname"},
			nil,
		)
		identity := &domain.Identity{Id: uuid.New()}
		identityRepoMock.EXPECT().FindOrCreateWithDefaultRole(gomock.Any(), gomock.Any()).Return(identity, nil)
		sessionRepoMock.EXPECT().FindActiveSessionByIdentityId(gomock.Any(), identity.Id).Return(nil, domain.ErrNoActiveSession)
		newSession := &domain.Session{SessionId: "sid-2", ExpiresAt: time.Now().UTC().Add(time.Hour)}
		sessionRepoMock.EXPECT().Create(gomock.Any(), identity).Return(newSession, nil)

		session, err := authService.GenerateSession(context.TODO(), "google", "code")

		assert.NoError(t, err)
		assert.Equal(t, newSession, session)
	})
}

func TestAuthService_ValidateSession(t *testing.T) {
	t.Run("returns session not found if there are not session for a session ID", func(t *testing.T) {
		authService, _, sessionRepoMock, _ := setupAuthServiceWithMocks(t)
		sessionRepoMock.EXPECT().FindBySessionId(gomock.Any(), "sid-1").Return(nil, nil, domain.ErrSessionNotFound)

		identity, session, err := authService.ValidateSession(context.TODO(), "sid-1")

		assert.Nil(t, identity)
		assert.Nil(t, session)
		assert.ErrorIs(t, err, domain.ErrSessionNotFound)
	})
	t.Run("returns inactive session when an old session id is used", func(t *testing.T) {
		authService, _, sessionRepoMock, _ := setupAuthServiceWithMocks(t)
		identity := &domain.Identity{}
		session := &domain.Session{Active: false}
		sessionRepoMock.EXPECT().FindBySessionId(gomock.Any(), "sid-1").Return(identity, session, nil)

		gotIdentity, gotSession, err := authService.ValidateSession(context.TODO(), "sid-1")

		assert.Nil(t, gotIdentity)
		assert.Nil(t, gotSession)
		assert.ErrorIs(t, err, domain.ErrInactiveSession)
	})
	t.Run("returns identity and session if the session is not expired", func(t *testing.T) {
		authService, _, sessionRepoMock, _ := setupAuthServiceWithMocks(t)
		identity := &domain.Identity{}
		session := &domain.Session{Active: true, ExpiresAt: time.Now().UTC().Add(time.Hour)}
		sessionRepoMock.EXPECT().FindBySessionId(gomock.Any(), "sid-1").Return(identity, session, nil)

		gotIdentity, gotSession, err := authService.ValidateSession(context.TODO(), "sid-1")

		assert.NoError(t, err)
		assert.Equal(t, identity, gotIdentity)
		assert.Equal(t, session, gotSession)
	})
	t.Run("expires the session and returns the expired session error when an active session is expired", func(t *testing.T) {
		authService, _, sessionRepoMock, _ := setupAuthServiceWithMocks(t)
		identity := &domain.Identity{}
		session := &domain.Session{Active: true, ExpiresAt: time.Now().UTC().Add(-time.Hour)}
		sessionRepoMock.EXPECT().FindBySessionId(gomock.Any(), "sid-1").Return(identity, session, nil)
		sessionRepoMock.EXPECT().DeactivateBySessionId(gomock.Any(), "sid-1").Return(session, nil)

		gotIdentity, gotSession, err := authService.ValidateSession(context.TODO(), "sid-1")

		assert.Nil(t, gotIdentity)
		assert.Nil(t, gotSession)
		assert.ErrorIs(t, err, domain.ErrExpiredSession)
	})
}

func TestAuthService_DeactivateSession(t *testing.T) {
	t.Run("returns session not found if the session id is missing", func(t *testing.T) {
		authService, _, sessionRepoMock, _ := setupAuthServiceWithMocks(t)
		sessionRepoMock.EXPECT().DeactivateBySessionId(gomock.Any(), "sid-1").Return(nil, domain.ErrSessionNotFound)

		err := authService.DeactivateSession(context.TODO(), "sid-1")

		assert.ErrorIs(t, err, domain.ErrSessionNotFound)
	})

	t.Run("deactivates the session if session id exists", func(t *testing.T) {
		authService, _, sessionRepoMock, _ := setupAuthServiceWithMocks(t)
		session := &domain.Session{SessionId: "sid-1"}
		sessionRepoMock.EXPECT().DeactivateBySessionId(gomock.Any(), "sid-1").Return(session, nil)

		err := authService.DeactivateSession(context.TODO(), "sid-1")

		assert.NoError(t, err)
	})
}
