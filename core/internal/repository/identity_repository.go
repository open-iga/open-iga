package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/open-iga/core/internal/contract"
	"github.com/open-iga/core/internal/domain"
	"github.com/open-iga/core/internal/repository/db"
)

type IdentityRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
	logger  *slog.Logger
}

var _ contract.IdentityRepository = (*IdentityRepository)(nil)

func NewIdentityRepository(pool *pgxpool.Pool, queries *db.Queries, logger *slog.Logger) *IdentityRepository {
	return &IdentityRepository{pool, queries, logger}
}

func (i *IdentityRepository) FindOrCreateWithDefaultRole(ctx context.Context, user *domain.OauthUser) (*domain.Identity, error) {
	if user == nil {
		return nil, errors.New("user is nil")
	}

	if user.Email == "" {
		return nil, errors.New("email is empty")
	}

	tx, err := i.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction during identity insert / update: %w", err)
	}
	defer tx.Rollback(ctx)

	queries := i.queries.WithTx(tx)

	identity, err := queries.UpsertIdentity(ctx, db.UpsertIdentityParams{
		// accept empty string as oauth server can return empty string based on the privacy policy
		FirstName: pgtype.Text{String: user.FirstName, Valid: true},
		LastName:  pgtype.Text{String: user.LastName, Valid: true},
		Type:      db.IdentityTypeUser,
		Email:     user.Email,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to insert / update the identity %w", err)
	}

	_, err = queries.UpsertRoleByIdentityId(ctx, db.UpsertRoleByIdentityIdParams{
		Name:       domain.DefaultIdentityRole, // Not the best; but fine for now
		IdentityID: identity.ID,
	})

	if err != nil && !errors.Is(err, pgx.ErrNoRows) { // To skip err where user already has a default role
		return nil, fmt.Errorf("failed to insert / update the role %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to commit identity during identity creation: %w", err)
	}

	return identity.ToDomain(), nil
}

func (i *IdentityRepository) GetRolesByIdentityId(ctx context.Context, identityId uuid.UUID) (*domain.IdentityRole, error) {
	roles, err := i.queries.GetRolesByIdentityId(ctx, identityId)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNoIdentityFound
	}

	if err != nil {
		return nil, err
	}

	return &domain.IdentityRole{
		IdentityId: identityId,
		Roles:      roles,
	}, nil
}

func (i *IdentityRepository) UpdateRoleByIdentityId(ctx context.Context, identityId uuid.UUID, role string) (*domain.IdentityRole, error) {
	insertedRole, err := i.queries.UpsertRoleByIdentityId(ctx, db.UpsertRoleByIdentityIdParams{Name: role, IdentityID: identityId})
	if err != nil {
		return nil, err
	}

	return &domain.IdentityRole{
		IdentityId: identityId,
		Roles:      []string{insertedRole},
	}, nil
}
