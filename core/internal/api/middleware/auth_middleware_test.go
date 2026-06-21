package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/open-iga/core/internal/api/middleware"
	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/contract"
	"github.com/open-iga/core/internal/domain"
	"github.com/open-iga/core/internal/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func setupMiddlewareWithMocks(t *testing.T) (*middleware.Middleware, *testutil.MockLoginService) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	loginServiceMock := testutil.NewMockLoginService(ctrl)
	applicationMock := &contract.RuntimeApplication{LoginService: loginServiceMock}

	m := middleware.NewMiddleware(testutil.NewTestAppConfig(), testutil.NewTestLogger(), applicationMock)

	return m, loginServiceMock
}

func mockHandlerWithOkResponse() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestMiddleware_AuthMiddleware(t *testing.T) {
	t.Run("redirects to signin page if the cookie is missing on non-auth routes", func(t *testing.T) {
		m, _ := setupMiddlewareWithMocks(t)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
		rec := httptest.NewRecorder()
		m.AuthMiddleware(mockHandlerWithOkResponse()).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.JSONEq(t, `{"message": "No session cookie found", "redirect": "/auth/sign-in"}`, rec.Body.String())
	})

	t.Run("redirects to signin page if the session is invalid on non-auth routes", func(t *testing.T) {
		m, loginServiceMock := setupMiddlewareWithMocks(t)
		loginServiceMock.EXPECT().ValidateSession(gomock.Any(), "sid-1").Return(nil, nil, errors.New("invalid session"))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
		req.AddCookie(&http.Cookie{Name: common.SessionCookieName, Value: "sid-1"})
		rec := httptest.NewRecorder()
		m.AuthMiddleware(mockHandlerWithOkResponse()).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.JSONEq(t, `{"message": "No session cookie found", "redirect": "/auth/sign-in"}`, rec.Body.String())
	})

	t.Run("allows the request if the session is missing on auth routes", func(t *testing.T) {
		m, _ := setupMiddlewareWithMocks(t)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/google", nil)
		rec := httptest.NewRecorder()
		m.AuthMiddleware(mockHandlerWithOkResponse()).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("allows the request if the session is invalid on auth routes", func(t *testing.T) {
		m, loginServiceMock := setupMiddlewareWithMocks(t)
		loginServiceMock.EXPECT().ValidateSession(gomock.Any(), "sid-1").Return(nil, nil, errors.New("invalid session"))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/google", nil)
		req.AddCookie(&http.Cookie{Name: common.SessionCookieName, Value: "sid-1"})
		rec := httptest.NewRecorder()
		m.AuthMiddleware(mockHandlerWithOkResponse()).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("redirects to home when a request with valid session is requested to auth routes", func(t *testing.T) {
		m, loginServiceMock := setupMiddlewareWithMocks(t)
		identity := testutil.NewIdentity()
		session := &domain.Session{SessionId: "sid-1"}
		loginServiceMock.EXPECT().ValidateSession(gomock.Any(), "sid-1").Return(&identity, session, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/google", nil)
		req.AddCookie(&http.Cookie{Name: common.SessionCookieName, Value: "sid-1"})
		rec := httptest.NewRecorder()
		m.AuthMiddleware(mockHandlerWithOkResponse()).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"redirect": "/"}`, rec.Body.String())
	})
}
