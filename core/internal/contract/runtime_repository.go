package contract

import (
	"context"

	"github.com/google/uuid"
	"github.com/open-iga/core/internal/domain"
)

type IdentityRepository interface {
	FindOrCreateWithDefaultRole(ctx context.Context, user *domain.OauthUser) (*domain.Identity, error)
	GetRolesByIdentityId(ctx context.Context, identityId uuid.UUID) (*domain.IdentityRole, error)
	UpsertRoleByIdentityId(ctx context.Context, identityId uuid.UUID, role string) (*domain.IdentityRole, error)
}

type SessionRepository interface {
	Create(ctx context.Context, identity *domain.Identity) (*domain.Session, error)
	FindBySessionId(ctx context.Context, sessionId string) (*domain.Identity, *domain.Session, error)
	DeactivateBySessionId(ctx context.Context, sessionId string) (*domain.Session, error)
	DeactivateByIdentityId(ctx context.Context, identityId uuid.UUID) (*domain.Session, error)
	FindActiveSessionByIdentityId(ctx context.Context, identityId uuid.UUID) (*domain.Session, error)
}

type Repository struct {
	IdentityRepository IdentityRepository
	SessionRepository  SessionRepository
}
