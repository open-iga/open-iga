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
	// CompileConnector resolves endpoints from the environment and caches a compiled connector for
	// url+hash. Called post-approval and at startup, once the env is guaranteed complete.
	CompileConnector(ctx context.Context, url string, hash string, spec *domain.ConnectorSpec)
	// Execute runs one invocation on the connector cached for url+hash, returning its raw JSON output.
	Execute(ctx context.Context, url string, hash string, invocation domain.ConnectorInvocation) ([]byte, error)
}

type RuntimeRemote struct {
	Oauth2Clients    Oauth2Clients
	ConnectorRuntime ConnectorRuntime
}
