package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrConnectorNotFound       = errors.New("connector not found")
	ErrConnectorNotOnboardable = errors.New("connector cannot be onboarded")
	ErrMissingRequiredEnvVars  = errors.New("required env variables are not set on the core")
	ErrManagedSystemNameExists = errors.New("managed system name already exists")
	ErrManagedSystemNotFound   = errors.New("managed system not found")
	ErrTenantNotFound          = errors.New("tenant not found")
	ErrAccountNotFound         = errors.New("account not found")
	ErrEntitlementNotFound     = errors.New("entitlement not found")
	ErrEmptyOutput             = errors.New("connector returned an empty output")
	ErrUnknownOperation        = errors.New("unknown connector operation")
	ErrMissingOperationData    = errors.New("missing data for connector operation")
	ErrTenantNameExists        = errors.New("tenant name already exists in this managed system")
)

type ManagedSystem struct {
	Id            string
	Name          string
	ConnectorUrl  string
	ConnectorHash string
	ConnectorSpec *ConnectorSpec
	CreatedAt     time.Time
}

type ManagedSystemTenant struct {
	Id              string
	ManagedSystemId string
	ExternalId      string
	Name            string
	CreatedAt       time.Time
}

// Account is an identity's managed account on a tenant. ExternalId is empty until it is provisioned.
type Account struct {
	Id         string
	TenantId   string
	IdentityId string
	ExternalId string
	Enabled    bool
}

// Entitlement is something a connector can grant; Name is its id in the target system.
type Entitlement struct {
	Id       string
	TenantId string
	Name     string
}

// OperationRef is all a caller says about a call: which rows to use. Everything the connector
// receives is loaded from these rows.
type OperationRef struct {
	TenantId      uuid.UUID
	IdentityId    uuid.UUID
	EntitlementId uuid.UUID
}

// needsExistingAccount is false for the operations does not require account level permissions
func needsExistingAccount(kind ConnectorKind, action string) bool {
	switch kind {
	case AccountActionKind:
		return action != "create" && action != "list"
	case EntitlementKind:
		return action != "discover"
	}
	return true
}

// NeedsAccount is true when the ref has a identity id and the operation require account level info;
// both account and entitlement operation requires account level info, excpet a few in both
func (r OperationRef) NeedsAccount(kind ConnectorKind, action string) bool {
	return r.IdentityId != uuid.Nil && needsExistingAccount(kind, action)
}

// NeedsEntitlement is true when the ref names an entitlement and the operation grants or revokes it.
func (r OperationRef) NeedsEntitlement(kind ConnectorKind, action string) bool {
	return r.EntitlementId != uuid.Nil && kind == EntitlementKind && (action == "grant" || action == "revoke")
}
