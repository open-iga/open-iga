package wasm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/open-iga/core/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// emptyModule is the minimal valid wasm binary: the magic bytes "\0asm" + version 1,
// with no sections. It compiles but carries no custom section.
var emptyModule = []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}

// serve starts a test server returning body and yields its URL; stopped on cleanup.
func serve(t *testing.T, body []byte) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

func newRuntime() *ConnectorRuntime {
	return NewConnectorRuntime(testutil.NewTestLogger())
}

func TestConnectorRuntime_ValidateConnectorByURL(t *testing.T) {
	t.Run("errors when hash is empty", func(t *testing.T) {
		spec, err := newRuntime().ValidateConnectorByURL(context.TODO(), "http://example.com/aws.wasm", "")

		assert.Nil(t, spec)
		assert.ErrorContains(t, err, "hash required")
	})

	t.Run("errors when the wasm cannot be fetched", func(t *testing.T) {
		// start then immediately stop a server so the fetch hits a connection error
		srv := httptest.NewServer(http.NotFoundHandler())
		closedURL := srv.URL
		srv.Close()

		spec, err := newRuntime().ValidateConnectorByURL(context.TODO(), closedURL+"/aws.wasm", "somehash")

		assert.Nil(t, spec)
		assert.ErrorContains(t, err, "failed to fetch wasm data")
	})

	t.Run("errors when the hash does not match", func(t *testing.T) {
		url := serve(t, []byte("any bytes"))

		spec, err := newRuntime().ValidateConnectorByURL(context.TODO(), url, "deadbeef")

		assert.Nil(t, spec)
		assert.ErrorContains(t, err, "expected hash")
	})

	t.Run("errors when the bytes are not a valid wasm module", func(t *testing.T) {
		body := []byte("not a wasm module")
		url := serve(t, body)

		spec, err := newRuntime().ValidateConnectorByURL(context.TODO(), url, sha256Hex(body))

		assert.Nil(t, spec)
		assert.ErrorContains(t, err, "failed to compile wasm module")
	})

	t.Run("errors when the manifest custom section is missing", func(t *testing.T) {
		url := serve(t, emptyModule)

		spec, err := newRuntime().ValidateConnectorByURL(context.TODO(), url, sha256Hex(emptyModule))

		assert.Nil(t, spec)
		assert.ErrorContains(t, err, "failed to find custom section")
	})

	t.Run("returns the connector spec for a real connector wasm", func(t *testing.T) {
		// Fixture built once with the connector-sdk (bun compile) and committed here.
		// Skips until the fixture is added, so uleb128/custom-section encoding never
		// has to be re-implemented in Go.
		body, err := os.ReadFile("testdata/connector.wasm")
		if err != nil {
			t.Skip("add testdata/connector.wasm (built via connector-sdk) to run this case")
		}
		url := serve(t, body)

		spec, err := newRuntime().ValidateConnectorByURL(context.TODO(), url, sha256Hex(body))

		require.NoError(t, err)
		require.NotNil(t, spec)
		assert.NotEmpty(t, spec.Name)
	})
}
