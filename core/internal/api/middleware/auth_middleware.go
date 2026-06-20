package middleware

import (
	"net/http"
	"slices"
	"strings"

	"github.com/open-iga/core/internal/api/generated"
	"github.com/open-iga/core/internal/common"
)

var NoAuthPath = []string{"/api/health"}

const AuthEndpointPrefix = "/api/v1/auth"
const AuthCallbackEndpointPrefix = "/callback"

func isRequestToAuthEndpoint(r *http.Request) bool {
	if strings.HasPrefix(r.URL.Path, AuthEndpointPrefix) && r.Method == http.MethodGet {
		return true
	}

	if strings.HasPrefix(r.URL.Path, AuthEndpointPrefix) &&
		strings.HasSuffix(r.URL.Path, AuthCallbackEndpointPrefix) && r.Method == http.MethodPost {
		return true
	}

	return false
}

func (m *Middleware) redirectResponseToSignIn(w http.ResponseWriter) {
	response := generated.GetUserDetails401JSONResponse{ // user details is chose as this is one of the first protected resource
		Message:  "No session cookie found",
		Redirect: m.appConfig.Redirect.SignIn,
	}
	err := response.VisitGetUserDetailsResponse(w)
	if err != nil {
		m.logger.Error("failed to respond with 401 from auth middleware", "error", err)
	}
}

func (m *Middleware) redirectResponseToHomePage(w http.ResponseWriter) {
	response := generated.AuthCallback200JSONResponse{ // user details is chose as this is one of the first protected resource
		Redirect: m.appConfig.Redirect.Home,
	}
	err := response.VisitAuthCallbackResponse(w)
	if err != nil {
		m.logger.Error("failed to respond with 401 from auth middleware", "error", err)
	}
}

func (m *Middleware) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if slices.Contains(NoAuthPath, path) {
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie(common.SessionCookieName)
		// Redirect to login page when session is missing where the users can select the required oauth provider
		if err != nil && !isRequestToAuthEndpoint(r) {
			m.logger.Debug("unable to read session cookie from the request")
			m.redirectResponseToSignIn(w)
			return
		}

		// Login path would not have the session key during session creation
		if err != nil && isRequestToAuthEndpoint(r) {
			next.ServeHTTP(w, r)
			return
		}

		// If session cookie is available
		// 1. If invalid, redirect to login page
		// 2. if expired, redirect to login page
		// 3. If failed to expire, redirect to login again
		identity, session, err := m.application.LoginService.ValidateSession(r.Context(), cookie.Value)
		if err != nil && !isRequestToAuthEndpoint(r) { // user might have an expired session in login and login callback handler
			m.logger.Debug("unable to validate session", "error", err)
			m.redirectResponseToSignIn(w)
			return
		}

		// if the user already has a valid session and requests to login again; redirect to home page
		if err == nil && isRequestToAuthEndpoint(r) {
			m.redirectResponseToSignIn(w)
		}

		ctx := WithIdentity(r.Context(), identity)
		ctx = WithSession(ctx, session)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
