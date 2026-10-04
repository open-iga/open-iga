package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/open-iga/core/internal/api/generated"
	"github.com/open-iga/core/internal/api/middleware"
	"github.com/open-iga/core/internal/domain"
	"github.com/stretchr/testify/assert"
)

// okStrictHandler writes 200 so a test can detect that the request was forwarded.
func okStrictHandler() generated.StrictHandlerFunc {
	return func(_ context.Context, w http.ResponseWriter, _ *http.Request, _ any) (any, error) {
		w.WriteHeader(http.StatusOK)
		return nil, nil
	}
}

func TestMiddleware_AuthzStrictMiddleware(t *testing.T) {
	newReq := func() *http.Request {
		return httptest.NewRequest(http.MethodPost, "/api/v1/connectors/onboarding-requests", nil)
	}

	t.Run("returns 403 for an operation with no policy", func(t *testing.T) {
		m, _ := setupMiddlewareWithMocks(t)
		rec := httptest.NewRecorder()

		_, _ = m.AuthzStrictMiddleware(okStrictHandler(), "UnknownOperation")(context.Background(), rec, newReq(), nil)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("forwards a public operation without roles", func(t *testing.T) {
		m, _ := setupMiddlewareWithMocks(t)
		rec := httptest.NewRecorder()

		_, _ = m.AuthzStrictMiddleware(okStrictHandler(), "Health")(context.Background(), rec, newReq(), nil)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("returns 403 when roles are missing from context", func(t *testing.T) {
		m, _ := setupMiddlewareWithMocks(t)
		rec := httptest.NewRecorder()

		_, _ = m.AuthzStrictMiddleware(okStrictHandler(), "OnboardConnector")(context.Background(), rec, newReq(), nil)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("returns 403 when identity lacks the required role", func(t *testing.T) {
		m, _ := setupMiddlewareWithMocks(t)
		ctx := middleware.WithRoles(context.Background(), []string{domain.DefaultIdentityRole})
		rec := httptest.NewRecorder()

		_, _ = m.AuthzStrictMiddleware(okStrictHandler(), "OnboardConnector")(ctx, rec, newReq(), nil)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("forwards when identity has the required role", func(t *testing.T) {
		m, _ := setupMiddlewareWithMocks(t)
		ctx := middleware.WithRoles(context.Background(), []string{domain.AdminRole})
		rec := httptest.NewRecorder()

		_, _ = m.AuthzStrictMiddleware(okStrictHandler(), "OnboardConnector")(ctx, rec, newReq(), nil)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}
