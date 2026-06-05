package handler

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/open-iga/core/internal/contract"
	"github.com/open-iga/core/internal/domain"
	"github.com/open-iga/core/internal/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func setupRouterWithMockLoginService(t *testing.T) (*chi.Mux, *testutil.MockLoginService) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	loginServiceMock := testutil.NewMockLoginService(ctrl)
	applicationMock := &contract.RuntimeApplication{LoginService: loginServiceMock}

	handler := NewHandler(testutil.NewTestAppConfig(), testutil.NewTestLogger(), applicationMock)

	router := testutil.NewMockRouter(handler)

	return router, loginServiceMock
}

func TestHandler_AuthDetails(t *testing.T) {
	t.Run("returns 400 when provider is invalid", func(t *testing.T) {
		router, _ := setupRouterWithMockLoginService(t)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/unknown-provider", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("returns 500 when the consent details failed to generate", func(t *testing.T) {
		router, loginServiceMock := setupRouterWithMockLoginService(t)
		loginServiceMock.EXPECT().GetConsentPageDetails(gomock.Any(), "google").Return(nil, errors.New("unsupported provider"))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/google", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("returns 200 with valid auth code URL and sets the state cookie", func(t *testing.T) {
		router, loginServiceMock := setupRouterWithMockLoginService(t)
		loginServiceMock.EXPECT().GetConsentPageDetails(gomock.Any(), "google").Return(
			&domain.ConsentDetails{
				AuthCodeURL: "https://auth-code-url.com",
				State:       "state-cookie",
			}, nil,
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/google", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		cookieInRec := rec.Header().Get("Set-Cookie")
		expectedCookie := http.Cookie{
			Name:     AuthStateCookieName,
			Value:    "state-cookie",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
			MaxAge:   20,
			Secure:   true,
		}

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, expectedCookie.String(), cookieInRec)
		assert.JSONEq(t, `{"authCodeUrl": "https://auth-code-url.com"}`, rec.Body.String())
	})
}

func TestHandler_AuthCallback(t *testing.T) {
	t.Run("returns 400 when provider is unknown", func(t *testing.T) {
		router, _ := setupRouterWithMockLoginService(t)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/google/callback", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("returns 400 when code param is missing", func(t *testing.T) {
		router, _ := setupRouterWithMockLoginService(t)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/google/callback?state=state", nil)
		req.AddCookie(&http.Cookie{
			Name:  AuthStateCookieName,
			Value: "",
		})

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("returns 400 when state param is missing", func(t *testing.T) {
		router, _ := setupRouterWithMockLoginService(t)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/google/callback?code=code", nil)
		req.AddCookie(&http.Cookie{
			Name:  AuthStateCookieName,
			Value: "",
		})

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("returns 422 when state cookie is missing", func(t *testing.T) {
		router, _ := setupRouterWithMockLoginService(t)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/google/callback?code=code&state=state", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.JSONEq(t, `{"message": "missing state cookie"}`, rec.Body.String())
	})

	t.Run("returns 422 when auth state cookie is empty string", func(t *testing.T) {
		router, _ := setupRouterWithMockLoginService(t)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/google/callback?code=code&state=state", nil)
		req.AddCookie(&http.Cookie{
			Name:  AuthStateCookieName,
			Value: "",
		})

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.JSONEq(t, `{"message": "empty state cookie"}`, rec.Body.String())
	})

	t.Run("returns 422 when state in query param and cookie doesn't match", func(t *testing.T) {
		router, _ := setupRouterWithMockLoginService(t)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/google/callback?code=code&state=state", nil)
		req.AddCookie(&http.Cookie{
			Name:  AuthStateCookieName,
			Value: "invalid-state",
		})

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.JSONEq(t, `{"message": "state mismatch"}`, rec.Body.String())
	})

	t.Run("returns 500 when generating session fails", func(t *testing.T) {
		router, loginServiceMock := setupRouterWithMockLoginService(t)
		loginServiceMock.EXPECT().GenerateSession(gomock.Any(), "google", "code").Return(nil, errors.New("failed to generate session"))

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/google/callback?code=code&state=state", nil)
		req.AddCookie(&http.Cookie{
			Name:  AuthStateCookieName,
			Value: "state",
		})

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.JSONEq(t, `{"message": "auth callback handler: failed to generate session"}`, rec.Body.String())
	})

	t.Run("returns 201 with redirect details and sets session id cookie", func(t *testing.T) {
		router, loginServiceMock := setupRouterWithMockLoginService(t)
		loginServiceMock.EXPECT().GenerateSession(gomock.Any(), "google", "code").Return(&domain.Session{
			SessionId: "session-id",
		}, nil)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/google/callback?code=code&state=state", nil)
		req.AddCookie(&http.Cookie{
			Name:  AuthStateCookieName,
			Value: "state",
		})

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		cookie := rec.Header().Get("Set-Cookie")

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.JSONEq(t, `{"redirect": "/"}`, rec.Body.String())
		fmt.Print(cookie)
		assert.Contains(t, cookie, "sid=session-id")
		assert.Contains(t, cookie, "Path=/")
		assert.Contains(t, cookie, "HttpOnly")
		assert.Contains(t, cookie, "Secure")
		assert.Contains(t, cookie, "SameSite=Lax")
	})
}
