package oauth2_client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/contract"
	"github.com/open-iga/core/internal/domain"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type (
	GoogleOauth2Client struct {
		config oauth2.Config
		logger *slog.Logger
	}

	googleOauth2UserinfoDto struct {
		Email      string `json:"email"`
		GivenName  string `json:"given_name"`
		FamilyName string `json:"family_name"`
	}
)

// for type checking at compile time
var _ contract.Client = &GoogleOauth2Client{}

const (
	googleOauth2UserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
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
		logger: logger,
	}
}

// GetConsentPageDetails generates a random state and returns the URL for the Google consent page along with the state
func (g *GoogleOauth2Client) GetConsentPageDetails(_ context.Context) (*contract.ConsentPageDetails, error) {
	state, err := common.GenerateHighEntropyID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state for consent page: %w", err)
	}

	return &contract.ConsentPageDetails{
		AuthCodeURL: g.config.AuthCodeURL(state),
		State:       state,
	}, nil
}

// FetchOauthUser exchanges the authorization code for an access token and then uses that token to fetch the user's profile information from Google
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

	var userinfo googleOauth2UserinfoDto
	if err := json.NewDecoder(resp.Body).Decode(&userinfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info response: %w", err)
	}

	return userinfo.toOauthUser(), nil
}

func (g *googleOauth2UserinfoDto) toOauthUser() *domain.OauthUser {
	return domain.NewOauthUser(g.GivenName, g.FamilyName, g.Email)
}
