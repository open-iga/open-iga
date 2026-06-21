package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/open-iga/core/internal/contract"
	"github.com/open-iga/core/internal/domain"
)

type LoginService struct {
	oauth2Clients      contract.Oauth2Clients
	logger             *slog.Logger
	sessionRepository  contract.SessionRepository
	identityRepository contract.IdentityRepository
}

var _ contract.LoginService = (*LoginService)(nil)

func NewLoginService(oauth2Clients contract.Oauth2Clients, logger *slog.Logger, sessionRepository contract.SessionRepository, identityRepository contract.IdentityRepository) *LoginService {
	return &LoginService{oauth2Clients, logger, sessionRepository, identityRepository}
}

func (l *LoginService) GetConsentPageDetails(ctx context.Context, provider string) (*domain.ConsentDetails, error) {
	client, ok := l.oauth2Clients[contract.Provider(provider)]

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
func (l *LoginService) GenerateSession(ctx context.Context, provider string, authCode string) (*domain.Session, error) {
	client, ok := l.oauth2Clients[contract.Provider(provider)]

	if !ok {
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	oauthUser, err := client.FetchOauthUser(ctx, authCode)
	if err != nil {
		return nil, fmt.Errorf("generate session: %w", err)
	}

	identity, err := l.identityRepository.FindOrCreate(ctx, oauthUser)
	if err != nil {
		return nil, fmt.Errorf("generate session: %w", err)
	}

	return l.createSession(ctx, identity)
}

func (l *LoginService) createSession(ctx context.Context, identity *domain.Identity) (*domain.Session, error) {
	session, err := l.sessionRepository.FindActiveSessionByIdentityId(ctx, identity.Id)
	if err != nil && !errors.Is(err, domain.ErrNoActiveSession) {
		return nil, fmt.Errorf("generate session: %w", err)
	}

	if session != nil && !session.IsExpired() {
		return session, nil
	}

	if session != nil && session.IsExpired() {
		_, err := l.sessionRepository.DeactivateByIdentityId(ctx, identity.Id)
		if err != nil {
			return nil, fmt.Errorf("generate session: %w", err)
		}
	}

	session, err = l.sessionRepository.Create(ctx, identity)
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
func (l *LoginService) ValidateSession(ctx context.Context, sessionId string) (*domain.Identity, *domain.Session, error) {
	identity, session, err := l.sessionRepository.FindBySessionId(ctx, sessionId)

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

	l.logger.Debug("expired session", "sessionId", sessionId)
	_, err = l.sessionRepository.DeactivateBySessionId(ctx, sessionId)
	if err != nil {
		l.logger.Error("failed to deactivate session", "sessionIdPrefix", sessionId[:8], "error", err.Error())
		return nil, nil, domain.ErrFailedToExpireSession
	}

	return nil, nil, domain.ErrExpiredSession
}

func (l *LoginService) DeactivateSession(ctx context.Context, sessionId string) error {
	_, err := l.sessionRepository.DeactivateBySessionId(ctx, sessionId)
	if err != nil && !errors.Is(err, domain.ErrSessionNotFound) {
		return domain.ErrSessionNotFound
	}
	if err != nil {
		return fmt.Errorf("deactivate session: %w", err)
	}

	return nil
}
