package repository

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/open-iga/core/internal/contract"
	"github.com/open-iga/core/internal/testutil"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var (
	repository *contract.Repository
	conn       *pgxpool.Pool
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	var code int

	pgContainer, err := setupDB(ctx)
	defer func() {
		teardownDB(ctx, pgContainer)
		os.Exit(code)
	}()

	if err != nil {
		fmt.Printf("Error setting up postgres container: %v\n", err)
		code = 1
	}

	pgConnString, err := pgContainer.ConnectionString(ctx)
	if err != nil {
		fmt.Printf("Error getting connection string: %v\n", err)
		code = 1
	}

	repository, conn, err = NewRepository(
		testutil.NewTestAppConfig(testutil.WithDatabaseUrlOverride(pgConnString)),
		testutil.NewTestLogger(),
	)
	if err != nil {
		fmt.Printf("Error creating repository: %v\n", err)
		code = 1
	}

	code = m.Run()
}

func setupDB(ctx context.Context) (*postgres.PostgresContainer, error) {
	pgContainer, err := postgres.Run(
		ctx,
		"postgres:18.3-alpine3.23",
		postgres.WithDatabase("open_iga"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		return nil, err
	}

	return pgContainer, nil
}

func teardownDB(ctx context.Context, postgresContainer *postgres.PostgresContainer) {
	err := postgresContainer.Stop(ctx, nil)
	if err != nil {
		fmt.Printf("failed to stop postgres container: %v\n", err)
	}
}
