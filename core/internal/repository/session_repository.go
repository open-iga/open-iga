package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/open-iga/core/internal/contract"
	"github.com/open-iga/core/internal/domain"
	"github.com/open-iga/core/internal/repository/db"
)

type SessionRepository struct {
	queries *db.Queries
	logger  *slog.Logger
}

var _ contract.SessionRepository = (*SessionRepository)(nil)

func NewSessionRepository(queries *db.Queries, logger *slog.Logger) *SessionRepository {
	return &SessionRepository{queries, logger}
}

func (s *SessionRepository) Create(ctx context.Context, identity *domain.Identity) (*domain.Session, error) {
	if identity == nil {
		return nil, errors.New("identity is nil")
	}

	identitySession, err := identity.GenerateSession()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session %w", err)
	}

	session, err := s.queries.CreateSession(ctx, db.CreateSessionParams{
		SessionID:  identitySession.SessionId,
		IdentityID: identity.Id,
		ExpiresAt:  pgtype.Timestamptz{Time: identitySession.ExpiresAt, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create session %w", err)
	}

	return session.ToDomain(), nil
}

func (s *SessionRepository) FindBySessionId(ctx context.Context, sessionId string) (*domain.Identity, *domain.Session, error) {
	sessionDetails, err := s.queries.FindBySessionId(ctx, sessionId)

	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, domain.ErrSessionNotFound
	}

	if err != nil {
		return nil, nil, fmt.Errorf("failed to get active session %w", err)
	}

	return sessionDetails.Identity.ToDomain(), sessionDetails.Session.ToDomain(), nil
}

func (s *SessionRepository) DeactivateBySessionId(ctx context.Context, sessionId string) (*domain.Session, error) {
	session, err := s.queries.DeactivateBySessionId(ctx, sessionId)

	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrSessionNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to deactivate session %w", err)
	}

	return session.ToDomain(), nil
}

func (s *SessionRepository) DeactivateByIdentityId(ctx context.Context, identityId uuid.UUID) (*domain.Session, error) {
	session, err := s.queries.DeactivateByIdentityId(ctx, identityId)

	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrSessionNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get deactivate session %w", err)
	}

	return session.ToDomain(), nil
}

func (s *SessionRepository) FindActiveSessionByIdentityId(ctx context.Context, identityId uuid.UUID) (*domain.Session, error) {
	session, err := s.queries.FindActiveSessionByIdentityId(ctx, identityId)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNoActiveSession
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get deactivate session %w", err)
	}

	return session.ToDomain(), nil
}
