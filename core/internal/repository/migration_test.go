package repository

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
)

// fetchPGDump runs pg_dump inside the Postgres container so the client always
// matches the server version; both dbs live on the same container server.
func fetchPGDump(ctx context.Context, dbName string, t *testing.T) string {
	t.Helper()
	code, reader, err := pgContainer.Exec(ctx,
		[]string{"pg_dump", "--schema-only", "--no-owner", "--no-acl", "--schema=public", "-T", "schema_migrations", "-U", "test", dbName},
		tcexec.Multiplexed(),
	)
	if err != nil || code != 0 {
		t.Fatalf("pg_dump failed (exit %d): %v", code, err)
	}
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read pg_dump output: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	var actual []string
	for _, line := range lines {
		if strings.HasPrefix(line, "\\restrict") || strings.HasPrefix(line, "\\unrestrict") {
			continue
		}
		actual = append(actual, line)
	}

	return strings.TrimSpace(strings.Join(actual, "\n"))
}

func TestMigration(t *testing.T) {
	t.Run("DB is in expected state after migration", func(t *testing.T) {
		ctx := context.Background()

		// apply the expected schema to a temp db so both sides are canonicalized by pg_dump
		expectedSQL, _ := os.ReadFile("migration_test.expected_db_state.sql")
		assert.NotNil(t, expectedSQL, fmt.Errorf("expected db state file to exist"))

		_, err := conn.Exec(ctx, "CREATE DATABASE migration_expected")
		assert.Nil(t, err, fmt.Errorf("failed to create database migration_expected table: %w", err))

		defer conn.Exec(context.Background(), "DROP DATABASE IF EXISTS migration_expected WITH (FORCE)")

		expectedConnString := strings.Replace(pgConnString, "/open_iga", "/migration_expected", 1)
		cfg, _ := pgxpool.ParseConfig(expectedConnString)
		expectedConn, err := pgxpool.NewWithConfig(ctx, cfg)
		if err != nil {
			t.Fatalf("failed to connect to temp db: %v", err)
		}
		defer expectedConn.Close()
		if _, err := expectedConn.Exec(ctx, string(expectedSQL)); err != nil {
			t.Fatalf("failed to apply expected schema: %v", err)
		}

		actual, expected := fetchPGDump(ctx, "open_iga", t), fetchPGDump(ctx, "migration_expected", t)
		assert.Equal(t, expected, actual)
	})
}
