package contract

import (
	"context"

	"github.com/open-iga/core/internal/domain"
)

type ConsentPageDetails struct {
	AuthCodeURL string
	State       string
}

type LoginService interface {
	GetConsentPageDetails(ctx context.Context, provider string) (*ConsentPageDetails, error)
	GenerateSession(ctx context.Context, provider string, authCode string) (*domain.OauthUser, error)
}

type RuntimeApplication struct {
	LoginService LoginService
}
