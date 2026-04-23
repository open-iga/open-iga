package oauth

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/open-iga/core/internal/contract"
	"github.com/open-iga/core/internal/domain"
)

type LoginService struct {
	oauth2Clients contract.Oauth2Clients
	logger        *slog.Logger
}

var _ contract.LoginService = &LoginService{}

func NewLoginService(oauth2Clients contract.Oauth2Clients, logger *slog.Logger) *LoginService {
	return &LoginService{oauth2Clients, logger}
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

func (l *LoginService) GenerateSession(ctx context.Context, provider string, authCode string) (*domain.OauthUser, error) {
	client, ok := l.oauth2Clients[contract.Provider(provider)]

	if !ok {
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	oauthUser, err := client.FetchOauthUser(ctx, authCode)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch oauth user: %w", err)
	}

	return oauthUser, nil
}
