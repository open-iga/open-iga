package contract

import (
	"context"

	"github.com/google/uuid"
	"github.com/open-iga/core/internal/domain"
)

type AuthService interface {
	GetConsentPageDetails(ctx context.Context, provider string) (*domain.ConsentDetails, error)
	GenerateSession(ctx context.Context, provider string, authCode string) (*domain.Session, error)
	ValidateSession(ctx context.Context, sessionId string) (*domain.Identity, *domain.Session, error)
	DeactivateSession(ctx context.Context, sessionId string) error
	GetRoles(ctx context.Context, identityId uuid.UUID) []string
}

type RuntimeApplication struct {
	AuthService AuthService
}
