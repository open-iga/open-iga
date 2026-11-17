package domain

import "github.com/google/uuid"

// DiscoveryTarget is one place discovery reads: a connector resource of a managed system, in one tenant.
// It is plain data, so a workflow engine can store it. TenantId is uuid.Nil for a single-tenant system.
type DiscoveryTarget struct {
	SystemId   string    `json:"systemId"`
	SystemName string    `json:"systemName"`
	TenantId   uuid.UUID `json:"tenantId"`
	Resource   string    `json:"resource"`
}
