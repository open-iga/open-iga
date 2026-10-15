package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
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

	return managedSystem.ToDomain(), nil
}

// CreateManagedSystemTenant is idempotent: if a tenant with the same external id
// already exists for the managed system, the existing row is returned with created=false.
func (m *ManagedSystemRepository) CreateManagedSystemTenant(ctx context.Context, managedSystemID string, externalID string, name string, metadata []byte) (tenant *domain.ManagedSystemTenant, created bool, err error) {
	id, err := uuid.Parse(managedSystemID)
	if err != nil {
		return nil, false, fmt.Errorf("%w: %s", domain.ErrManagedSystemNotFound, managedSystemID)
	}

	row, err := m.queries.CreateManagedSystemTenant(ctx, db.CreateManagedSystemTenantParams{
		ManagedSystemID: id,
		ExternalID:      externalID,
		Name:            name,
		Metadata:        metadata,
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

func (m *ManagedSystemRepository) ListManagedSystems(ctx context.Context) ([]*domain.ManagedSystem, error) {
	rows, err := m.queries.ListManagedSystems(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list managed systems: %w", err)
	}

	managedSystems := make([]*domain.ManagedSystem, len(rows))
	for i := range rows {
		managedSystems[i] = rows[i].ToDomain()
	}

	return managedSystems, nil
}
