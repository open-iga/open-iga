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
	// GetWorkflow also finds the built-in discovery workflow, which has no steps.
	GetWorkflow(ctx context.Context, name string) (domain.WorkflowDefinition, error)
	// GetWorkflowSchedule is nil for a workflow that has no schedule.
	GetWorkflowSchedule(ctx context.Context, name string) (*domain.WorkflowSchedule, error)
	// SetWorkflowSchedule saves the schedule of the workflow. It returns domain.ErrInvalidWorkflow when the workflow has no schedule or the cron is not valid.
	SetWorkflowSchedule(ctx context.Context, name string, schedule domain.WorkflowSchedule) error
	// StartWorkflow runs the saved workflow for the identity (empty for workflows that touch none) and returns the run id.
	// For discovery it starts a run with one step per target, and returns an empty id when there is no target.
	StartWorkflow(ctx context.Context, name string, identityId uuid.UUID) (runId string, err error)
	Continue(ctx context.Context, workflowName string, runId string, stepName string) error
	GetRun(ctx context.Context, workflowName string, runId string) (domain.WorkflowRunState, error)
}

// DiscoveryService finds the entitlements of the managed systems and saves them in the catalog.
type DiscoveryService interface {
	// ListDiscoveryTargets lists every entitlement resource of every managed system, once per tenant.
	ListDiscoveryTargets(ctx context.Context) ([]domain.DiscoveryTarget, error)
	// SearchEntitlements returns one page of the catalog and the number of entitlements that match in all.
	SearchEntitlements(ctx context.Context, search domain.EntitlementSearch) (entitlements []*domain.EntitlementCatalog, total int64, err error)
}

type RuntimeApplication struct {
	AuthService          AuthService
	ConnectorService     ConnectorService
	DiscoveryService     DiscoveryService
	ManagedSystemService ManagedSystemService
	WorkflowService      WorkflowService
}
