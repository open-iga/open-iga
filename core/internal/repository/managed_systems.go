package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/open-iga/core/internal/domain"
	"github.com/open-iga/core/internal/repository/db"
)

// Postgres SQLSTATE codes. https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"
)

type ManagedSystemRepository struct {
	queries *db.Queries
	logger  *slog.Logger
}

func NewManagedSystemRepository(queries *db.Queries, logger *slog.Logger) *ManagedSystemRepository {
	return &ManagedSystemRepository{queries: queries, logger: logger}
}

func (m *ManagedSystemRepository) CreateManagedSystem(ctx context.Context, name string, connectorUrl string, connectorHash string, spec *domain.ConnectorSpec) (*domain.ManagedSystem, error) {
	specJSON, err := json.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal connector spec: %w", err)
	}

	managedSystem, err := m.queries.CreateManagedSystem(ctx, db.CreateManagedSystemParams{
		Name:          name,
		ConnectorUrl:  connectorUrl,
		ConnectorHash: connectorHash,
		ConnectorSpec: specJSON,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return nil, fmt.Errorf("%w: %s", domain.ErrManagedSystemNameExists, name)
		}
		return nil, fmt.Errorf("failed to create managed system: %w", err)
	}

	return managedSystem.ToDomain(m.logger), nil
}

// CreateManagedSystemTenant is idempotent: if a tenant with the same external id
// already exists for the managed system, the existing row is returned with created=false.
func (m *ManagedSystemRepository) CreateManagedSystemTenant(ctx context.Context, managedSystemID string, externalID string, name string) (tenant *domain.ManagedSystemTenant, created bool, err error) {
	id, err := uuid.Parse(managedSystemID)
	if err != nil {
		return nil, false, fmt.Errorf("%w: %s", domain.ErrManagedSystemNotFound, managedSystemID)
	}

	row, err := m.queries.CreateManagedSystemTenant(ctx, db.CreateManagedSystemTenantParams{
		ManagedSystemID: id,
		ExternalID:      externalID,
		Name:            name,
	})
	if err == nil {
		return row.ToDomain(), true, nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgForeignKeyViolation:
			return nil, false, fmt.Errorf("%w: %s", domain.ErrManagedSystemNotFound, managedSystemID)
		case pgUniqueViolation:
			// A clashing name (distinct external id) is a real conflict, not an idempotent retry.
			if pgErr.ConstraintName == "idx_unique_managed_system_tenant_name" {
				return nil, false, fmt.Errorf("%w: %s", domain.ErrTenantNameExists, name)
			}
			// Same external id in the same system: idempotent, return the existing row.
			existing, getErr := m.queries.GetManagedSystemTenantByExternalId(ctx, db.GetManagedSystemTenantByExternalIdParams{
				ManagedSystemID: id,
				ExternalID:      externalID,
			})
			if getErr != nil {
				m.logger.Error("failed to get existing managed system tenant by external ID", "error", getErr)
				return nil, false, fmt.Errorf("failed to fetch existing managed system tenant: %w", getErr)
			}
			return existing.ToDomain(), false, nil
		}
	}

	return nil, false, fmt.Errorf("failed to create managed system tenant: %w", err)
}

func (m *ManagedSystemRepository) GetManagedSystem(ctx context.Context, managedSystemID string) (*domain.ManagedSystem, error) {
	id, err := uuid.Parse(managedSystemID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", domain.ErrManagedSystemNotFound, managedSystemID)
	}

	row, err := m.queries.GetManagedSystem(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %s", domain.ErrManagedSystemNotFound, managedSystemID)
		}
		return nil, fmt.Errorf("failed to get managed system: %w", err)
	}

	return row.ToDomain(m.logger), nil
}

func (m *ManagedSystemRepository) ListManagedSystemTenants(ctx context.Context, managedSystemID string) ([]*domain.ManagedSystemTenant, error) {
	id, err := uuid.Parse(managedSystemID)
	if err != nil {
		// Invalid UUID = not a valid managed system id
		return nil, fmt.Errorf("%w: %s", domain.ErrManagedSystemNotFound, managedSystemID)
	}

	managedSystemTenants, err := m.queries.ListManagedSystemTenants(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to list managed system tenants: %w", err)
	}

	tenants := make([]*domain.ManagedSystemTenant, len(managedSystemTenants))
	for i := range managedSystemTenants {
		tenants[i] = managedSystemTenants[i].ToDomain()
	}

	return tenants, nil
}

func (m *ManagedSystemRepository) ListManagedSystems(ctx context.Context) ([]*domain.ManagedSystem, error) {
	listManagedSystems, err := m.queries.ListManagedSystems(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list managed systems: %w", err)
	}

	managedSystems := make([]*domain.ManagedSystem, len(listManagedSystems))
	for i := range listManagedSystems {
		managedSystems[i] = listManagedSystems[i].ToDomain(m.logger)
	}

	return managedSystems, nil
}
