package contract

import (
	"context"

	"github.com/google/uuid"
	"github.com/open-iga/core/internal/domain"
)

type AuthService interface {
	EnsureAdmin(ctx context.Context, adminEmail string) error
	GetConsentPageDetails(ctx context.Context, provider string) (*domain.ConsentDetails, error)
	GenerateSession(ctx context.Context, provider string, authCode string) (*domain.Session, error)
	ValidateSession(ctx context.Context, sessionId string) (*domain.Identity, *domain.Session, error)
	DeactivateSession(ctx context.Context, sessionId string) error
	GetRoles(ctx context.Context, identityId uuid.UUID) []string
}

type ConnectorService interface {
	ValidateByUrl(ctx context.Context, connectorUrl string, connectorHash string) string
	GetOnboardedConnectorDetails(ctx context.Context, onboardingId string) (*domain.ConnectorValidationResult, bool)
	LoadConnectorsOnStartup(ctx context.Context) error
	Execute(ctx context.Context, connectorName string, managedResource string, kind domain.ConnectorKind, action string, ref domain.OperationRef) (domain.ConnectorOutput, error)
}

type ManagedSystemService interface {
	Onboard(ctx context.Context, connectorOnboardId string) (*domain.ManagedSystem, error)
	GetManagedSystem(ctx context.Context, managedSystemId string) (*domain.ManagedSystem, error)
	ListManagedSystems(ctx context.Context) ([]*domain.ManagedSystem, error)
	OnboardTenant(ctx context.Context, managedSystemId string, externalId string, name string) (tenant *domain.ManagedSystemTenant, created bool, err error)
	ListTenants(ctx context.Context, managedSystemId string) ([]*domain.ManagedSystemTenant, error)
}

type WorkflowService interface {
	// CreateWorkflow checks every step against the managed systems, tenants and entitlements, then saves the definition under its name.
	// It returns domain.ErrWorkflowAlreadyExists when the name is taken.
	CreateWorkflow(ctx context.Context, definition domain.WorkflowDefinition) error
	GetWorkflow(ctx context.Context, name string) (domain.WorkflowDefinition, error)
	// StartWorkflow runs the saved workflow for the identity (empty for workflows that touch none) and returns the run id.
	StartWorkflow(ctx context.Context, name string, identityId uuid.UUID) (runId string, err error)
	Continue(ctx context.Context, workflowName string, runId string, stepName string) error
	GetRun(ctx context.Context, workflowName string, runId string) (domain.WorkflowRunState, error)
}

// DiscoveryService finds the entitlements of the managed systems and saves them in the catalog.
// The caller loops over the pages, so it can make each page durable.
type DiscoveryService interface {
	// ListDiscoveryTargets lists every entitlement resource of every managed system, once per tenant.
	ListDiscoveryTargets(ctx context.Context) ([]domain.DiscoveryTarget, error)
	// DiscoverPage saves one page of the target and returns the next cursor; empty cursor in reads the first page, empty cursor out means the last page.
	DiscoverPage(ctx context.Context, target domain.DiscoveryTarget, cursor string) (nextCursor string, err error)
}

type RuntimeApplication struct {
	AuthService          AuthService
	ConnectorService     ConnectorService
	DiscoveryService     DiscoveryService
	ManagedSystemService ManagedSystemService
	WorkflowService      WorkflowService
}
