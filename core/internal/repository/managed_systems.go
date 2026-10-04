package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/open-iga/core/internal/domain"
	"github.com/open-iga/core/internal/repository/db"
)

// pgUniqueViolation is the Postgres SQLSTATE code for a unique constraint violation.
// https://www.postgresql.org/docs/current/errcodes-appendix.html
const pgUniqueViolation = "23505"

type ManagedSystemRepository struct {
	queries *db.Queries
	logger  *slog.Logger
}

func NewManagedSystemRepository(queries *db.Queries, logger *slog.Logger) *ManagedSystemRepository {
	return &ManagedSystemRepository{queries: queries, logger: logger}
}

func (m *ManagedSystemRepository) CreateManagedSystem(ctx context.Context, name string, connectorUrl string, connectorHash string) (*domain.ManagedSystem, error) {
	managedSystem, err := m.queries.CreateManagedSystem(ctx, db.CreateManagedSystemParams{
		Name:          name,
		ConnectorUrl:  connectorUrl,
		ConnectorHash: connectorHash,
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
