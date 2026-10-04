package wasm

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/open-iga/core/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// emptyModule is the minimal valid wasm binary: magic "\0asm" + version 1, no sections.
var emptyModule = []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}

// tlsServe starts an HTTPS test server and returns a runtime whose client trusts it and
// reaches loopback, so the pipeline past fetch can be exercised. The SSRF/scheme guards
// are covered separately against the real client.
func tlsServe(t *testing.T, body []byte) (*ConnectorRuntime, string) {
	t.Helper()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)

	rt := NewConnectorRuntime(testutil.NewTestLogger())
	rt.httpClient = srv.Client()
	return rt, srv.URL
}

func TestConnectorRuntime_ValidateConnectorByURL(t *testing.T) {
	t.Run("errors when hash is empty", func(t *testing.T) {
		rt := NewConnectorRuntime(testutil.NewTestLogger())

		spec, err := rt.ValidateConnectorByURL(context.TODO(), "https://example.com/aws.wasm", "")

		assert.Nil(t, spec)
		assert.ErrorContains(t, err, "hash required")
	})

	t.Run("rejects non-https urls", func(t *testing.T) {
		rt := NewConnectorRuntime(testutil.NewTestLogger())

		spec, err := rt.ValidateConnectorByURL(context.TODO(), "http://example.com/aws.wasm", "somehash")

		assert.Nil(t, spec)
		assert.ErrorContains(t, err, "must use https")
	})

	t.Run("blocks non-public addresses", func(t *testing.T) {
		rt := NewConnectorRuntime(testutil.NewTestLogger()) // the real, secure client

		spec, err := rt.ValidateConnectorByURL(context.TODO(), "https://127.0.0.1:9/aws.wasm", "somehash")

		assert.Nil(t, spec)
		assert.ErrorContains(t, err, "failed to fetch connector")
	})

	t.Run("errors when the hash does not match", func(t *testing.T) {
		rt, url := tlsServe(t, []byte("any bytes"))

		spec, err := rt.ValidateConnectorByURL(context.TODO(), url, "deadbeef")

		assert.Nil(t, spec)
		assert.ErrorContains(t, err, "hash mismatch")
	})

	t.Run("errors when the bytes are not a valid wasm module", func(t *testing.T) {
		body := []byte("not a wasm module")
		rt, url := tlsServe(t, body)

		spec, err := rt.ValidateConnectorByURL(context.TODO(), url, sha256Hex(body))

		assert.Nil(t, spec)
		assert.ErrorContains(t, err, "failed to compile wasm module")
	})

	t.Run("errors when the manifest custom section is missing", func(t *testing.T) {
		rt, url := tlsServe(t, emptyModule)

		spec, err := rt.ValidateConnectorByURL(context.TODO(), url, sha256Hex(emptyModule))

		assert.Nil(t, spec)
		assert.ErrorContains(t, err, "failed to find custom section")
	})

	t.Run("returns the connector spec for a real connector wasm", func(t *testing.T) {
		// real connector built via connector-sdk; skips until the fixture is committed.
		body, err := os.ReadFile("testdata/connector.wasm")
		if err != nil {
			t.Skip("add testdata/connector.wasm (built via connector-sdk) to run this case")
		}
		rt, url := tlsServe(t, body)

		spec, err := rt.ValidateConnectorByURL(context.TODO(), url, sha256Hex(body))

		require.NoError(t, err)
		require.NotNil(t, spec)
		assert.NotEmpty(t, spec.Name)
	})
}

func TestIsPublicIP(t *testing.T) {
	cases := map[string]bool{
		"8.8.8.8":     true,
		"1.1.1.1":     true,
		"2606:4700::": true,  // public v6
		"127.0.0.1":   false, // loopback
		"::1":         false, // loopback v6
		"10.0.0.1":    false, // private
		"192.168.1.1": false, // private
		"172.16.0.1":  false, // private
		"fd00::1":     false, // unique local v6
		"169.254.0.1": false, // link-local
		"fe80::1":     false, // link-local v6
		"0.0.0.0":     false, // unspecified
		"224.0.0.1":   false, // multicast
	}
	for ipStr, want := range cases {
		ip := net.ParseIP(ipStr)
		require.NotNil(t, ip, ipStr)
		assert.Equal(t, want, isPublicIP(ip), ipStr)
	}
}
