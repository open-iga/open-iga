package wasm

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/open-iga/core/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// compileRealConnector loads testdata/connector.wasm (built via connector-sdk) into the runtime
// and returns it with the connector's url and hash. It skips until the fixture is added.
func compileRealConnector(t *testing.T) (rt *ConnectorRuntime, url string, hash string) {
	t.Helper()
	body, err := os.ReadFile("testdata/connector.wasm")
	if err != nil {
		t.Skip("add testdata/connector.wasm (built via connector-sdk) to run this case")
	}
	rt, url = tlsServe(t, body)
	hash = sha256Hex(body)

	spec, err := rt.ValidateConnectorByURL(context.Background(), url, hash)
	require.NoError(t, err)
	rt.CompileConnector(context.Background(), url, hash, spec)
	return rt, url, hash
}

func awsInvocation(operation string, input string) domain.ConnectorInvocation {
	return domain.ConnectorInvocation{
		Kind:            domain.AccountActionKind,
		ManagedResource: "iam-user",
		Type:            operation,
		Input:           json.RawMessage(input),
	}
}

func TestConnectorRuntime_Execute_ReportsConnectorFailures(t *testing.T) {
	t.Setenv("AWS_REGION", "")
	rt, url, hash := compileRealConnector(t)

	t.Run("returns the connector's message when its config is missing", func(t *testing.T) {
		_, err := rt.Execute(context.Background(), url, hash,
			awsInvocation("read", `{"tenant":{"id":"000000000000"},"accountId":"arn:aws:iam::000000000000:user/ada"}`))

		require.Error(t, err)
		assert.ErrorContains(t, err, "connector dispatch failed")
		assert.ErrorContains(t, err, "is required but missing from the Host config")
		assert.NotContains(t, err.Error(), `{"error"`, "the message should not be raw JSON")
	})

	t.Run("returns the connector's message when the input does not match its schema", func(t *testing.T) {
		_, err := rt.Execute(context.Background(), url, hash, awsInvocation("read", `{"tenant":{"id":"000000000000"},"id":"old-field"}`))

		require.Error(t, err)
		assert.ErrorContains(t, err, "Host input validation error")
		assert.ErrorContains(t, err, "accountId")
	})

	t.Run("rejects an operation the connector does not declare", func(t *testing.T) {
		_, err := rt.Execute(context.Background(), url, hash, awsInvocation("list", `{"tenant":{"id":"000000000000"}}`))

		assert.ErrorContains(t, err, "declares no operation")
	})
}
