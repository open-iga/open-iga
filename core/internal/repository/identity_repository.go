package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/open-iga/core/internal/contract"
	"github.com/open-iga/core/internal/domain"
	"github.com/open-iga/core/internal/repository/db"
)

type IdentityRepository struct {
	queries *db.Queries
	logger  *slog.Logger
}

var _ contract.IdentityRepository = (*IdentityRepository)(nil)

func NewIdentityRepository(queries *db.Queries, logger *slog.Logger) *IdentityRepository {
	return &IdentityRepository{queries, logger}
}

func (i *IdentityRepository) FindOrCreate(ctx context.Context, user *domain.OauthUser) (*domain.Identity, error) {
	if user == nil {
		return nil, errors.New("user is nil")
	}

	if user.Email == "" {
		return nil, errors.New("email is empty")
	}

	identity, err := i.queries.UpsertIdentity(ctx, db.UpsertIdentityParams{
		// accept empty string as oauth server can return empty string based on the privacy policy
		FirstName: pgtype.Text{String: user.FirstName, Valid: true},
		LastName:  pgtype.Text{String: user.LastName, Valid: true},
		Type:      db.IdentityTypeUser,
		Email:     user.Email,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to insert / update the identity %w", err)
	}

	return identity.ToDomain(), nil
}
