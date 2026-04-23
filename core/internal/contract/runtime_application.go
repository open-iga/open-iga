package contract

import (
	"context"

	"github.com/open-iga/core/internal/domain"
)

type LoginService interface {
	GetConsentPageDetails(ctx context.Context, provider string) (*domain.ConsentDetails, error)
	GenerateSession(ctx context.Context, provider string, authCode string) (*domain.OauthUser, error)
}

type RuntimeApplication struct {
	LoginService LoginService
}
