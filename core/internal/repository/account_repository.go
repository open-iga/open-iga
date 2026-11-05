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

type AccountRepository struct {
	queries *db.Queries
}

func NewAccountRepository(queries *db.Queries) *AccountRepository {
	return &AccountRepository{queries: queries}
}

func (a *AccountRepository) GetAccountByTenantAndIdentity(ctx context.Context, tenantId uuid.UUID, identityId uuid.UUID) (*domain.Account, error) {
	row, err := a.queries.GetAccountByTenantAndIdentity(ctx, db.GetAccountByTenantAndIdentityParams{TenantID: tenantId, IdentityID: identityId})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: identity %s on tenant %s", domain.ErrAccountNotFound, identityId, tenantId)
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	return row.ToDomain(), nil
}
