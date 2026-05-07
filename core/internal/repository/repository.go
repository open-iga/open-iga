package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/contract"
	"github.com/open-iga/core/internal/repository/db"
)

type Repository struct {
	conn   *pgxpool.Pool
	logger slog.Logger
}

const MaxConnections = 10
const MinConnections = 2

func NewRepository(appConfig *common.AppConfig, logger *slog.Logger) (*contract.Repository, error) {
	config, err := pgxpool.ParseConfig(appConfig.Database.URL)

	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string %w", err)
	}
	config.MaxConns = MaxConnections
	config.MinConns = MinConnections

	conn, err := pgxpool.NewWithConfig(context.Background(), config)

	if err != nil {
		return nil, fmt.Errorf("failed to created DB connection %w", err)
	}

	err = runMigration(conn)
	if err != nil {
		return nil, fmt.Errorf("error running migration %w", err)
	}

	queries := db.New(conn)

	return &contract.Repository{
		IdentityRepository: NewIdentityRepository(queries, logger),
		SessionRepository:  NewSessionRepository(queries, logger),
	}, nil
}

func runMigration(conn *pgxpool.Pool) error {
	dbInstance := stdlib.OpenDBFromPool(conn)
	driver, err := pgx.WithInstance(dbInstance, &pgx.Config{})

	if err != nil {
		return fmt.Errorf("failed to create migration driver %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://sql/migration",
		"pgx5",
		driver,
	)

	if err != nil {
		return fmt.Errorf("failed to create migrate instance %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migration %w", err)
	}

	return nil
}
