package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/open-iga/core/internal/contract"
	"github.com/open-iga/core/internal/domain"
	"github.com/open-iga/core/internal/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func setupRouterWithMockManagedSystemService(t *testing.T) (*chi.Mux, *testutil.MockManagedSystemService) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	managedSystemServiceMock := testutil.NewMockManagedSystemService(ctrl)
	applicationMock := &contract.RuntimeApplication{ManagedSystemService: managedSystemServiceMock}

	handler := NewHandler(testutil.NewTestAppConfig(), testutil.NewTestLogger(), applicationMock)

	router := testutil.NewMockRouter(handler)

	return router, managedSystemServiceMock
}

func TestHandler_OnboardManagedSystem(t *testing.T) {
	t.Run("returns 404 when connector is not found", func(t *testing.T) {
		router, managedSystemServiceMock := setupRouterWithMockManagedSystemService(t)
		managedSystemServiceMock.EXPECT().Onboard(gomock.Any(), "onboard-id").Return(nil, domain.ErrConnectorNotFound)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/managed-systems", strings.NewReader(`{"connectorOnboardingId": "onboard-id"}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.JSONEq(t, `{"message": "connector not found"}`, rec.Body.String())
	})

	t.Run("returns 404 when connector is not onboardable", func(t *testing.T) {
		router, managedSystemServiceMock := setupRouterWithMockManagedSystemService(t)
		managedSystemServiceMock.EXPECT().Onboard(gomock.Any(), "onboard-id").Return(nil, domain.ErrConnectorNotOnboardable)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/managed-systems", strings.NewReader(`{"connectorOnboardingId": "onboard-id"}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.JSONEq(t, `{"message": "connector cannot be onboarded"}`, rec.Body.String())
	})

	t.Run("returns 500 when onboarding fails", func(t *testing.T) {
		router, managedSystemServiceMock := setupRouterWithMockManagedSystemService(t)
		managedSystemServiceMock.EXPECT().Onboard(gomock.Any(), "onboard-id").Return(nil, errors.New("failed to onboard"))

		req := httptest.NewRequest(http.MethodPost, "/api/v1/managed-systems", strings.NewReader(`{"connectorOnboardingId": "onboard-id"}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.JSONEq(t, `{"message": "failed to onboard"}`, rec.Body.String())
	})

	t.Run("returns 200 with the managed system id when onboarding succeeds", func(t *testing.T) {
		router, managedSystemServiceMock := setupRouterWithMockManagedSystemService(t)
		managedSystemServiceMock.EXPECT().Onboard(gomock.Any(), "onboard-id").Return(&domain.ManagedSystem{Id: "system-id"}, nil)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/managed-systems", strings.NewReader(`{"connectorOnboardingId": "onboard-id"}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"id": "system-id"}`, rec.Body.String())
	})
}
