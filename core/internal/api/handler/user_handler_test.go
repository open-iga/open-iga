package handler

import (
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

func setupRouterWithIdentity(t *testing.T, identity *domain.Identity) (*chi.Mux, *testutil.MockLoginService) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	loginServiceMock := testutil.NewMockLoginService(ctrl)
	applicationMock := &contract.RuntimeApplication{LoginService: loginServiceMock}
	handler := NewHandler(testutil.NewTestAppConfig(), testutil.NewTestLogger(), applicationMock)

	router := testutil.NewMockRouter(handler, testutil.WithIdentitySetterMiddleware(identity))

	return router, loginServiceMock
}

func TestHandler_GetUserDetails(t *testing.T) {
	t.Run("return internal server error when identity is missing in context", func(t *testing.T) {
		router, _ := setupRouterWithMockLoginService(t)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.JSONEq(t, `{"message": "identity missing in context"}`, rec.Body.String())
	})

	t.Run("return user details when context contains identity", func(t *testing.T) {
		identity := testutil.NewIdentity()
		router, _ := setupRouterWithIdentity(t, &identity)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t,
			fmt.Sprintf(
				`{"email": "%s", "firstName": "%s", "lastName": "%s", "id": "%s"}`,
				identity.Email, identity.FirstName, identity.LastName, identity.Id),
			rec.Body.String(),
		)
	})
}
