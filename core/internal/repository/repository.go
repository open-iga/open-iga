package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/contract"
	"github.com/open-iga/core/internal/repository/db"
	"github.com/open-iga/core/internal/repository/sql"
)

const (
	MaxConnections = 10
	MinConnections = 2
)

func NewRepository(appConfig *common.AppConfig, logger *slog.Logger) (*contract.Repository, *pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(appConfig.Database.URL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse connection string %w", err)
	}
	config.MaxConns = MaxConnections
	config.MinConns = MinConnections

	conn, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to created DB connection %w", err)
	}

	err = runMigration(conn)
	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("error running migration %w", err)
	}

	queries := db.New(conn)

	return &contract.Repository{
		IdentityRepository: NewIdentityRepository(queries, logger),
		SessionRepository:  NewSessionRepository(queries, logger),
	}, conn, nil
}

func runMigration(conn *pgxpool.Pool) error {
	dbInstance := stdlib.OpenDBFromPool(conn)
	driver, err := pgx.WithInstance(dbInstance, &pgx.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver %w", err)
	}
	defer func() {
		_ = driver.Close()
	}()

	src, err := iofs.New(sql.MigrationFiles, "migration")
	if err != nil {
		return fmt.Errorf("failed to create embedded migration source %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "pgx5", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migration %w", err)
	}

	return nil
}
