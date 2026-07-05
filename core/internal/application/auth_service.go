package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/open-iga/core/internal/contract"
	"github.com/open-iga/core/internal/domain"
)

type AuthService struct {
	oauth2Clients      contract.Oauth2Clients
	logger             *slog.Logger
	sessionRepository  contract.SessionRepository
	identityRepository contract.IdentityRepository
}

var _ contract.AuthService = (*AuthService)(nil)

func NewAuthService(oauth2Clients contract.Oauth2Clients, logger *slog.Logger, sessionRepository contract.SessionRepository, identityRepository contract.IdentityRepository) *AuthService {
	return &AuthService{oauth2Clients, logger, sessionRepository, identityRepository}
}

func (a *AuthService) GetConsentPageDetails(ctx context.Context, provider string) (*domain.ConsentDetails, error) {
	client, ok := a.oauth2Clients[contract.Provider(provider)]

	if !ok {
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	consentPageDetails, err := client.GetConsentDetails(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get consent page details: %w", err)
	}

	return consentPageDetails, nil
}

// GenerateSession generates a session for the user based on the provider and auth code
// 1. Fetch the user details from the provider using the auth code
// 2. Find or create the identity in the database based on the user details
// 3. Check if there is an active session for the identity, if yes return the session
// 4. If there is an active session but it is expired, deactivate the session and create a new session and return it
// 5. If there is no active session, create a new session and return it
// AuthMiddleware handles the duplicate login; this is to guard generating duplicate session as this will cause DB error due to unique constraint
func (a *AuthService) GenerateSession(ctx context.Context, provider string, authCode string) (*domain.Session, error) {
	client, ok := a.oauth2Clients[contract.Provider(provider)]

	if !ok {
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	oauthUser, err := client.FetchOauthUser(ctx, authCode)
	if err != nil {
		return nil, fmt.Errorf("generate session: %w", err)
	}

	identity, err := a.identityRepository.FindOrCreateWithDefaultRole(ctx, oauthUser)
	if err != nil {
		return nil, fmt.Errorf("generate session: %w", err)
	}

	return a.createSession(ctx, identity)
}

func (a *AuthService) createSession(ctx context.Context, identity *domain.Identity) (*domain.Session, error) {
	session, err := a.sessionRepository.FindActiveSessionByIdentityId(ctx, identity.Id)
	if err != nil && !errors.Is(err, domain.ErrNoActiveSession) {
		return nil, fmt.Errorf("generate session: %w", err)
	}

	if session != nil && !session.IsExpired() {
		return session, nil
	}

	if session != nil && session.IsExpired() {
		_, err := a.sessionRepository.DeactivateByIdentityId(ctx, identity.Id)
		if err != nil {
			return nil, fmt.Errorf("generate session: %w", err)
		}
	}

	session, err = a.sessionRepository.Create(ctx, identity)
	if err != nil {
		return nil, fmt.Errorf("generate session: %w", err)
	}

	return session, nil
}

// ValidateSession Validates and returns the identity associated with the session id
// error can be
// 1. SessionNotFound: when there is no session associated with the session id
// 2. InactiveSession: when the session is not active
// 3. ExpiredSession: when the session is expired, in this case the session will be deactivated and user needs to login again to get a new session
// 4. FailedToExpireSession: when there is an error while expiring the session, in this case user needs to login again to get a new session
func (a *AuthService) ValidateSession(ctx context.Context, sessionId string) (*domain.Identity, *domain.Session, error) {
	identity, session, err := a.sessionRepository.FindBySessionId(ctx, sessionId)

	if err != nil && errors.Is(err, domain.ErrSessionNotFound) {
		return nil, nil, domain.ErrSessionNotFound
	}

	if err != nil {
		return nil, nil, fmt.Errorf("validate session: %w", err)
	}

	if !session.Active {
		return nil, nil, domain.ErrInactiveSession
	}

	if !session.IsExpired() {
		return identity, session, nil
	}

	a.logger.Debug("expired session", "sessionId", sessionId)
	_, err = a.sessionRepository.DeactivateBySessionId(ctx, sessionId)
	if err != nil {
		a.logger.Error("failed to deactivate session", "sessionIdPrefix", sessionId[:8], "error", err.Error())
		return nil, nil, domain.ErrFailedToExpireSession
	}

	return nil, nil, domain.ErrExpiredSession
}

func (a *AuthService) DeactivateSession(ctx context.Context, sessionId string) error {
	_, err := a.sessionRepository.DeactivateBySessionId(ctx, sessionId)
	if err != nil && errors.Is(err, domain.ErrSessionNotFound) {
		return domain.ErrSessionNotFound
	}
	if err != nil {
		return fmt.Errorf("deactivate session: %w", err)
	}

	return nil
}

func (a *AuthService) GetRoles(ctx context.Context, identityId uuid.UUID) []string {
	identityRole, err := a.identityRepository.GetRolesByIdentityId(ctx, identityId)

	if err != nil {
		a.logger.Error("failed to get roles for identity", "identity", identityId, "error", err.Error(), "default role is returned", domain.DefaultIdentityRole)
		return []string{domain.DefaultIdentityRole}
	}

	return identityRole.Roles
}
