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
	GetAccountByTenantAndIdentity(ctx context.Context, tenantId uuid.UUID, identityId uuid.UUID) (*domain.Account, error)
}

type EntitlementRepository interface {
	GetEntitlement(ctx context.Context, entitlementId uuid.UUID) (*domain.Entitlement, error)
}

type Repository struct {
	AccountRepository       AccountRepository
	EntitlementRepository   EntitlementRepository
	IdentityRepository      IdentityRepository
	SessionRepository       SessionRepository
	ManagedSystemRepository ManagedSystemRepository
}
