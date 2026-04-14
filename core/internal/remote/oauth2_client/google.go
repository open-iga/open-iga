package oauth2_client

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/open-iga/core/internal/application/adapter"
	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/domain"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type (
	GoogleOauth2Client struct {
		config oauth2.Config
		logger *slog.Logger
	}

	GoogleOauth2UserinfoDto struct {
		Email      string `json:"email"`
		GivenName  string `json:"given_name"`
		FamilyName string `json:"family_name"`
	}
)

// for type checking at compile time
var _ adapter.Client = &GoogleOauth2Client{}

const (
	GoogleOauth2UserInfoUrl = "https://www.googleapis.com/oauth2/v2/userinfo"
)

// NewGoogleOauth2Client Creates a new GoogleOauth2Client with the given client ID and secret
func NewGoogleOauth2Client(appConfig *common.AppConfig, logger *slog.Logger) *GoogleOauth2Client {
	config := oauth2.Config{
		ClientID:     appConfig.Oauth.Google.ClientId,
		ClientSecret: appConfig.Oauth.Google.ClientSecret,
		RedirectURL:  fmt.Sprintf("%s/login/google/callback", appConfig.HostUrl),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}

	return &GoogleOauth2Client{
		config: config,
	}
}

func (g *GoogleOauth2Client) GetConsentPageDetails(_ context.Context) *adapter.ConsentPageDetails {
	state := common.GenerateHighEntropyID()

	return &adapter.ConsentPageDetails{
		AuthCodeURL: g.config.AuthCodeURL(state),
		State:       state,
	}
}

func (g *GoogleOauth2Client) FetchOauthUser(ctx context.Context, code string) (*domain.OauthUser, error) {
	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		g.logger.Error("failed to exchange auth code for token", "error", err)
		return nil, err
	}

	resp, err := g.config.Client(ctx, token).Get(GoogleOauth2UserInfoUrl)
	if err != nil {
		g.logger.Error("failed to fetch user info from google", "error", err)
		return nil, err
	}

	defer resp.Body.Close()

	var userinfo GoogleOauth2UserinfoDto
	if err := json.NewDecoder(resp.Body).Decode(&userinfo); err != nil {
		g.logger.Error("failed to decode user info response", "error", err)
		return nil, err
	}

	return userinfo.toOauthUser(), nil
}

func (g *GoogleOauth2UserinfoDto) toOauthUser() *domain.OauthUser {
	return domain.NewOauthUser(g.GivenName, g.FamilyName, g.Email)
}
