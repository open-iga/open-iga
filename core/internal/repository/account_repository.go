package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/open-iga/core/internal/domain"
	"github.com/open-iga/core/internal/repository/db"
)

type AccountRepository struct {
	queries *db.Queries
}

func NewAccountRepository(queries *db.Queries) *AccountRepository {
	return &AccountRepository{queries: queries}
}

func (a *AccountRepository) GetAccount(ctx context.Context, managedSystemId string, tenantId uuid.UUID, identityId uuid.UUID) (*domain.Account, error) {
	systemId, err := parseManagedSystemId(managedSystemId)
	if err != nil {
		return nil, err
	}

	row, err := a.queries.GetAccount(ctx, db.GetAccountParams{ManagedSystemID: systemId, TenantID: toNullableTenantId(tenantId), IdentityID: identityId})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: identity %s on system %s", domain.ErrAccountNotFound, identityId, managedSystemId)
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	return row.ToDomain(), nil
}

func (a *AccountRepository) UpsertAccount(ctx context.Context, managedSystemId string, tenantId uuid.UUID, identityId uuid.UUID, externalId string) (*domain.Account, error) {
	systemId, err := parseManagedSystemId(managedSystemId)
	if err != nil {
		return nil, err
	}

	row, err := a.queries.UpsertAccount(ctx, db.UpsertAccountParams{
		ManagedSystemID: systemId,
		TenantID:        toNullableTenantId(tenantId),
		IdentityID:      identityId,
		ExternalID:      pgtype.Text{String: externalId, Valid: true},
	})
	if err != nil {
		// No row comes back when the identity already has an account with another id: it is never overwritten.
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: system %s, identity %s", domain.ErrAccountConflict, managedSystemId, identityId)
		}

		return nil, fmt.Errorf("failed to save account: %w", err)
	}
	return row.ToDomain(), nil
}

func parseManagedSystemId(managedSystemId string) (uuid.UUID, error) {
	id, err := uuid.Parse(managedSystemId)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: %s", domain.ErrManagedSystemNotFound, managedSystemId)
	}
	return id, nil
}

// toNullableTenantId stores uuid.Nil, the "no tenant" of a single-tenant system, as NULL.
func toNullableTenantId(tenantId uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: tenantId, Valid: tenantId != uuid.Nil}
}
