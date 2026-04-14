package oauth

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/open-iga/core/internal/application/adapter"
	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/contract/remote"
	"github.com/open-iga/core/internal/domain"
)

type LoginService struct {
	appConfig *common.AppConfig
	remote    *remote.RuntimeRemote
	logger    *slog.Logger
}

func NewLoginService(appConfig *common.AppConfig, remotes *remote.RuntimeRemote, logger *slog.Logger) *LoginService {
	return &LoginService{appConfig, remotes, logger}
}

func (loginService *LoginService) GetConsentPageDetails(ctx context.Context, provider string) (*adapter.ConsentPageDetails, error) {
	client, ok := loginService.remote.Oauth2Clients[remote.Provider(provider)]

	if !ok {
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
	return client.GetConsentPageDetails(ctx), nil
}

func (loginService *LoginService) GenerateSession(ctx context.Context, provider string, authCode string) (*domain.OauthUser, error) {
	client, ok := loginService.remote.Oauth2Clients[remote.Provider(provider)]

	if !ok {
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	oauthUser, err := client.FetchOauthUser(ctx, authCode)
	if err != nil {
		return nil, err
	}

	return oauthUser, nil
}
