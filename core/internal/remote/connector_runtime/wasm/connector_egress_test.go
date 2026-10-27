package wasm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/open-iga/core/internal/domain"
	"github.com/open-iga/core/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestMatchesAnyGlob(t *testing.T) {
	globs := compileGlobs([]string{"*.amazonaws.com"})

	assert.True(t, matchesAnyGlob(globs, "sts.us-east-1.amazonaws.com"))
	assert.True(t, matchesAnyGlob(globs, "iam.amazonaws.com"))
	assert.False(t, matchesAnyGlob(globs, "evil.com"))
	assert.False(t, matchesAnyGlob(globs, "amazonaws.com.evil.com"))
}

func TestAllowedByEndpoints(t *testing.T) {
	endpoints := []domain.ResolvedEndpoint{
		{Method: "POST", URL: "https://iam.amazonaws.com/"},
		{Method: "POST", URL: "https://sts.us-east-1.amazonaws.com/"},
	}

	t.Run("allows a declared host+method", func(t *testing.T) {
		assert.True(t, allowedByEndpoints(endpoints, "iam.amazonaws.com", "POST"))
	})

	t.Run("denies an undeclared host even within the domain", func(t *testing.T) {
		assert.False(t, allowedByEndpoints(endpoints, "ec2.amazonaws.com", "POST"))
	})

	t.Run("denies a mismatched method", func(t *testing.T) {
		assert.False(t, allowedByEndpoints(endpoints, "iam.amazonaws.com", "GET"))
	})

	t.Run("fails closed on empty endpoints", func(t *testing.T) {
		assert.False(t, allowedByEndpoints(nil, "iam.amazonaws.com", "POST"))
	})
}

func TestSend_DoesNotFollowRedirects(t *testing.T) {
	var followed bool
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/elsewhere" {
			followed = true
			return
		}
		http.Redirect(w, r, "/elsewhere", http.StatusFound)
	}))
	t.Cleanup(srv.Close)
	rt := NewConnectorRuntime(testutil.NewTestLogger())
	rt.httpClient = srv.Client()

	response := rt.send(context.Background(), egressRequest{URL: srv.URL}, http.MethodGet, "127.0.0.1")

	assert.Equal(t, http.StatusFound, response.Status)
	assert.False(t, followed)
}
