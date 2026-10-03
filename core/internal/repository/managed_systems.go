package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/open-iga/core/internal/domain"
	"github.com/open-iga/core/internal/repository/db"
)

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
		return nil, fmt.Errorf("failed to create managed system: %w", err)
	}

	return managedSystem.ToDomain(), nil
}
