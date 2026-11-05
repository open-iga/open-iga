package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/open-iga/core/internal/domain"
	"github.com/open-iga/core/internal/repository/db"
)

type EntitlementRepository struct {
	queries *db.Queries
}

func NewEntitlementRepository(queries *db.Queries) *EntitlementRepository {
	return &EntitlementRepository{queries: queries}
}

func (e *EntitlementRepository) GetEntitlement(ctx context.Context, entitlementId uuid.UUID) (*domain.Entitlement, error) {
	row, err := e.queries.GetEntitlement(ctx, entitlementId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %s", domain.ErrEntitlementNotFound, entitlementId)
		}
		return nil, fmt.Errorf("failed to get entitlement: %w", err)
	}
	return row.ToDomain(), nil
}
