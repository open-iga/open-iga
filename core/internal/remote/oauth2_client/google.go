package oauth2_client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/contract"
	"github.com/open-iga/core/internal/domain"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type oauth2Config interface {
	AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	Client(ctx context.Context, t *oauth2.Token) *http.Client
}

type GoogleOauth2Client struct {
	config oauth2Config
	logger *slog.Logger
}

type googleOauth2UserinfoDto struct {
	Email      string `json:"email"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
}

// for type checking at compile time
var _ contract.Oauth2ClientAdapter = &GoogleOauth2Client{}

const (
	googleOauth2UserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
)

// NewGoogleOauth2Client Creates a new GoogleOauth2Client with the given client ID and secret
func NewGoogleOauth2Client(appConfig *common.AppConfig, logger *slog.Logger) *GoogleOauth2Client {
	config := &oauth2.Config{
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
		logger: logger,
	}
}

func (g *GoogleOauth2Client) GetConsentDetails(_ context.Context) (*domain.ConsentDetails, error) {
	state, err := common.GenerateHighEntropyID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state for consent page: %w", err)
	}

	return domain.NewConsentDetails(g.config.AuthCodeURL(state), state), nil
}

func (g *GoogleOauth2Client) FetchOauthUser(ctx context.Context, code string) (*domain.OauthUser, error) {
	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange auth code for token: %w", err)
	}

	resp, err := g.config.Client(ctx, token).Get(googleOauth2UserInfoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user info from google: %w", err)
	}

	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			g.logger.Error("failed to close response body", "error", err)
		}
	}(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode > 300 {
		return nil, fmt.Errorf("unexpected status code from google for user info: %d", resp.StatusCode)
	}

	var userinfo googleOauth2UserinfoDto
	if err := json.NewDecoder(resp.Body).Decode(&userinfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info response: %w", err)
	}

	return userinfo.toOauthUser(), nil
}

func (g *googleOauth2UserinfoDto) toOauthUser() *domain.OauthUser {
	return domain.NewOauthUser(g.GivenName, g.FamilyName, g.Email)
}
