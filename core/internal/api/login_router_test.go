package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/open-iga/core/internal/contract"
	"github.com/open-iga/core/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestLoginRouter(t *testing.T) {
	t.Run("returns 500 when consent page details results in error", func(t *testing.T) {
		mockRouter := CreateMockRouter(t,
			&contract.RuntimeApplication{
				LoginService: &mockLoginService{nil, errors.New("consent error"), nil},
			})

		req := httptest.NewRequest(http.MethodGet, "/login/google", nil)
		w := httptest.NewRecorder()
		mockRouter.api.Adapter().ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("returns 422 when provider is unknown", func(t *testing.T) {
		mockRouter := CreateMockRouter(t,
			&contract.RuntimeApplication{
				LoginService: &mockLoginService{nil, errors.New("consent error"), nil},
			})

		req := httptest.NewRequest(http.MethodGet, "/login/some-provider", nil)
		w := httptest.NewRecorder()
		mockRouter.api.Adapter().ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("returns 302 with correct Location and Set-Cookie headers when consent page details are returned successfully", func(t *testing.T) {
		mockRouter := CreateMockRouter(t,
			&contract.RuntimeApplication{
				LoginService: &mockLoginService{&contract.ConsentPageDetails{
					AuthCodeURL: "https://consent-page.com",
					State:       "mock-state",
				}, nil, nil},
			})

		req := httptest.NewRequest(http.MethodGet, "/login/google", nil)
		w := httptest.NewRecorder()
		mockRouter.api.Adapter().ServeHTTP(w, req)

		assert.Equal(t, http.StatusFound, w.Code)
		assert.Equal(t, "https://consent-page.com", w.Header().Get("Location"))

		cookies := w.Result().Cookies()
		assert.Len(t, cookies, 1)
		cookie := cookies[0]
		assert.Equal(t, AuthStateCookieName, cookie.Name)
		assert.Equal(t, "mock-state", cookie.Value)
		assert.True(t, cookie.HttpOnly)
		assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
		assert.Equal(t, 300, cookie.MaxAge)
	})
}

func TestLoginCallbackRouter(t *testing.T) {
	t.Run("returns 422 when provider is unknown", func(t *testing.T) {
		mockRouter := CreateMockRouter(t, &contract.RuntimeApplication{LoginService: &mockLoginService{
			consentDetails: nil,
			error:          nil,
			oauthUser:      nil,
		}})

		req := httptest.NewRequest(http.MethodGet, "/login/some-provider/callback?code=auth-code&state=mock-state", nil)
		req.AddCookie(&http.Cookie{Name: AuthStateCookieName, Value: "mock-state"})
		w := httptest.NewRecorder()
		mockRouter.api.Adapter().ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("returns 422 when code is missing in query param", func(t *testing.T) {
		mockRouter := CreateMockRouter(t, &contract.RuntimeApplication{LoginService: &mockLoginService{
			consentDetails: nil,
			error:          nil,
			oauthUser:      nil,
		}})

		req := httptest.NewRequest(http.MethodGet, "/login/google/callback?state=mock-state", nil)
		req.AddCookie(&http.Cookie{Name: AuthStateCookieName, Value: "mock-state"})
		w := httptest.NewRecorder()
		mockRouter.api.Adapter().ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("returns 422 when state is missing in query param", func(t *testing.T) {
		mockRouter := CreateMockRouter(t, &contract.RuntimeApplication{LoginService: &mockLoginService{
			consentDetails: nil,
			error:          nil,
			oauthUser:      nil,
		}})

		req := httptest.NewRequest(http.MethodGet, "/login/google/callback?code=auth-code", nil)
		req.AddCookie(&http.Cookie{Name: AuthStateCookieName, Value: "mock-state"})
		w := httptest.NewRecorder()
		mockRouter.api.Adapter().ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("return 422 when state cookie does not match with state query", func(t *testing.T) {
		mockRouter := CreateMockRouter(t, &contract.RuntimeApplication{LoginService: &mockLoginService{
			consentDetails: nil,
			error:          nil,
			oauthUser:      nil,
		}})

		req := httptest.NewRequest(http.MethodGet, "/login/google/callback?code=auth-code&state=mock-state", nil)
		req.AddCookie(&http.Cookie{Name: AuthStateCookieName, Value: "some-other-state"})
		w := httptest.NewRecorder()
		mockRouter.api.Adapter().ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("returns 500 when login service returns an error", func(t *testing.T) {
		mockRouter := CreateMockRouter(t, &contract.RuntimeApplication{LoginService: &mockLoginService{
			consentDetails: nil,
			error:          errors.New("failed to generate session"),
			oauthUser:      nil,
		}})

		req := httptest.NewRequest(http.MethodGet, "/login/google/callback?code=auth-code&state=mock-state", nil)
		req.AddCookie(&http.Cookie{Name: AuthStateCookieName, Value: "mock-state"})
		w := httptest.NewRecorder()
		mockRouter.api.Adapter().ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("returns 200 with user details when login is successful", func(t *testing.T) {
		mockRouter := CreateMockRouter(t, &contract.RuntimeApplication{LoginService: &mockLoginService{
			consentDetails: nil,
			error:          nil,
			oauthUser: &domain.OauthUser{
				Email:     "user@email.com",
				FirstName: "firstname",
				LastName:  "lastname",
			},
		}})

		req := httptest.NewRequest(http.MethodGet, "/login/google/callback?code=auth-code&state=mock-state", nil)
		req.AddCookie(&http.Cookie{Name: AuthStateCookieName, Value: "mock-state"})
		w := httptest.NewRecorder()
		mockRouter.api.Adapter().ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.JSONEq(t, `{"firstName":"firstname","lastName":"lastname","email":"user@email.com"}`, w.Body.String())
	})
}
