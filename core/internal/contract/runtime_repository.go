package contract

import (
	"context"

	"github.com/google/uuid"
	"github.com/open-iga/core/internal/domain"
)

type IdentityRepository interface {
	FindOrCreateWithRole(ctx context.Context, user *domain.OauthUser, role string) (*domain.Identity, error)
	GetRolesByIdentityId(ctx context.Context, identityId uuid.UUID) (*domain.IdentityRole, error)
	UpsertRoleByIdentityId(ctx context.Context, identityId uuid.UUID, role string) (*domain.IdentityRole, error)
	HasAdmin(ctx context.Context) (bool, error)
	GetIdentityById(ctx context.Context, identityId uuid.UUID) (*domain.Identity, error)
}

type SessionRepository interface {
	Create(ctx context.Context, identity *domain.Identity) (*domain.Session, error)
	FindBySessionId(ctx context.Context, sessionId string) (*domain.Identity, *domain.Session, error)
	DeactivateBySessionId(ctx context.Context, sessionId string) (*domain.Session, error)
	DeactivateByIdentityId(ctx context.Context, identityId uuid.UUID) (*domain.Session, error)
	FindActiveSessionByIdentityId(ctx context.Context, identityId uuid.UUID) (*domain.Session, error)
}

type ManagedSystemRepository interface {
	CreateManagedSystem(ctx context.Context, name string, connectorUrl string, connectorHash string, spec *domain.ConnectorSpec) (*domain.ManagedSystem, error)
	GetManagedSystem(ctx context.Context, managedSystemID string) (*domain.ManagedSystem, error)
	GetManagedSystemByName(ctx context.Context, name string) (*domain.ManagedSystem, error)
	ListManagedSystems(ctx context.Context) ([]*domain.ManagedSystem, error)
	CreateManagedSystemTenant(ctx context.Context, managedSystemID string, externalID string, name string) (tenant *domain.ManagedSystemTenant, created bool, err error)
	GetManagedSystemTenant(ctx context.Context, tenantId uuid.UUID) (*domain.ManagedSystemTenant, error)
	ListManagedSystemTenants(ctx context.Context, managedSystemID string) ([]*domain.ManagedSystemTenant, error)
}

type AccountRepository interface {
	// GetAccount finds the account of an identity on a system; tenantId is uuid.Nil for a single-tenant system.
	GetAccount(ctx context.Context, managedSystemId string, tenantId uuid.UUID, identityId uuid.UUID) (*domain.Account, error)
	// UpsertAccount saves the account an identity has on a system (and tenant, if it has one) with its id in the target system; repeating it updates the id.
	UpsertAccount(ctx context.Context, managedSystemId string, tenantId uuid.UUID, identityId uuid.UUID, externalId string) (*domain.Account, error)
}

// WorkflowRepository keeps the workflow definitions admins save, by name.
type WorkflowRepository interface {
	// CreateWorkflow returns domain.ErrWorkflowAlreadyExists when a definition has the name already.
	CreateWorkflow(ctx context.Context, definition domain.WorkflowDefinition) error
	// GetWorkflow returns domain.ErrWorkflowNotFound when no definition has the name.
	GetWorkflow(ctx context.Context, name string) (domain.WorkflowDefinition, error)
}

type EntitlementCatalogRepository interface {
	// GetEntitlementCatalog is not found when the entitlement belongs to another system or resource.
	GetEntitlementCatalog(ctx context.Context, entitlementId uuid.UUID, managedSystemId string, resource string) (*domain.EntitlementCatalog, error)
	// UpsertEntitlementCatalog saves what discovery found for a resource of a tenant; repeating it updates the row.
	UpsertEntitlementCatalog(ctx context.Context, managedSystemId string, tenantId uuid.UUID, resource string, discovered domain.DiscoveredEntitlement) (*domain.EntitlementCatalog, error)
}

type Repository struct {
	AccountRepository            AccountRepository
	EntitlementCatalogRepository EntitlementCatalogRepository
	IdentityRepository           IdentityRepository
	SessionRepository            SessionRepository
	ManagedSystemRepository      ManagedSystemRepository
	WorkflowRepository           WorkflowRepository
}
