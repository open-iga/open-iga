package contract

import (
	"context"

	"github.com/open-iga/core/internal/domain"
)

type LoginService interface {
	GetConsentPageDetails(ctx context.Context, provider string) (*domain.ConsentDetails, error)
	GenerateSession(ctx context.Context, provider string, authCode string) (*domain.Session, error)
	ValidateSession(ctx context.Context, sessionId string) (*domain.Identity, *domain.Session, error)
	DeactivateSession(ctx context.Context, sessionId string) error
}

type RuntimeApplication struct {
	LoginService LoginService
}
