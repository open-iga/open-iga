package contract

import (
	"context"

	"github.com/open-iga/core/internal/domain"
)

type Provider string

const (
	Google Provider = "google"
)

type Oauth2ClientAdapter interface {
	// GetConsentDetails generates a random state and returns the URL for the Oauth2 consent page along with the state
	GetConsentDetails(ctx context.Context) (*domain.ConsentDetails, error)

	// FetchOauthUser exchanges the authorization code for an access token and then uses that token to fetch the user's profile information
	FetchOauthUser(ctx context.Context, code string) (*domain.OauthUser, error)
}

type Oauth2Clients map[Provider]Oauth2ClientAdapter

type ConnectorRuntime interface {
	ValidateConnectorByURL(ctx context.Context, url string, hash string) (*domain.ConnectorSpec, error)
	// FetchAndVerify downloads the connector and returns its wasm bytes only if the
	// sha256 of the download matches hash.
	FetchAndVerify(ctx context.Context, url string, hash string) ([]byte, error)
}

type RuntimeRemote struct {
	Oauth2Clients    Oauth2Clients
	ConnectorRuntime ConnectorRuntime
}
