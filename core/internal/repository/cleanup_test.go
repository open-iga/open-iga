package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// deleteRows runs a cleanup delete on a live context: t.Context() is already cancelled when cleanups run.
func deleteRows(t *testing.T, query string, args ...any) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := conn.Exec(ctx, query, args...)
	assert.NoError(t, err)
}
