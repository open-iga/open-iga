package contract

import (
	"context"
	"time"

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

// WorkflowEngine runs workflow runs durably: a run survives a restart. A step waits at its gate until it is continued, and fails after 30 days.
type WorkflowEngine interface {
	// Launch starts the engine with the runner that executes each step and the queue of each type that has one; it also picks up runs left from before a restart.
	Launch(ctx context.Context, runner domain.WorkflowStepRunner, queues map[domain.WorkflowType]domain.WorkflowQueue) error
	// Start begins a run under the given id. A run of a type with a queue waits for its turn. Steps without a dependency run in parallel.
	Start(ctx context.Context, runId string, run domain.WorkflowRun) error
	// Continue releases a step waiting at a gate.
	Continue(ctx context.Context, runId, stepName string) error
	Get(ctx context.Context, runId string) (domain.WorkflowRunState, error)
	Shutdown(timeout time.Duration)
}

type RuntimeRemote struct {
	Oauth2Clients    Oauth2Clients
	ConnectorRuntime ConnectorRuntime
	WorkflowEngine   WorkflowEngine
}
