package middleware

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"github.com/open-iga/core/internal/api/generated"
	"github.com/open-iga/core/internal/common"
)

var NoAuthPath = []string{"/api/health"}

const (
	AuthEndpointPrefix         = "/api/v1/auth"
	AuthCallbackEndpointPrefix = "/callback"
)

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

func (m *Middleware) AuthnMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if slices.Contains(NoAuthPath, r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// authenticate resolves the session; if it returns proceed=false it has already
		// written the response (redirect) or forwarded the request itself.
		if ctx, proceed := m.authenticate(w, r, next); proceed {
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

// authenticate validates the session cookie and, on success, returns a context carrying
// the identity, session and roles. It returns proceed=false when it has already handled
// the request (redirected to sign-in/home, or forwarded an unauthenticated auth request).
func (m *Middleware) authenticate(w http.ResponseWriter, r *http.Request, next http.Handler) (context.Context, bool) {
	cookie, err := r.Cookie(common.SessionCookieName)
	if err != nil && !isRequestToAuthEndpoint(r) {
		// no session on a protected resource: send the user to sign in
		m.logger.Debug("unable to read session cookie from the request")
		m.redirectResponseToSignIn(w)
		return nil, false
	}
	if err != nil {
		// auth endpoint without a session yet (e.g. login / callback): let it through
		next.ServeHTTP(w, r)
		return nil, false
	}

	// If session cookie is available
	// 1. If invalid, redirect to login page
	// 2. if expired, redirect to login page
	// 3. If failed to expire, redirect to login again
	identity, session, err := m.application.AuthService.ValidateSession(r.Context(), cookie.Value)
	if err != nil && !isRequestToAuthEndpoint(r) { // user might have an expired session in login and login callback handler
		m.logger.Debug("unable to validate session", "error", err)
		m.redirectResponseToSignIn(w)
		return nil, false
	}

	// if the user already has a valid session and requests to login again; redirect to home page
	if err == nil && isRequestToAuthEndpoint(r) {
		m.logger.Debug("user already validated. redirecting to home")
		m.redirectResponseToHomePage(w)
		return nil, false
	}

	var identityRoles []string
	// fetch roles only when identity is avail
	if identity != nil {
		identityRoles = m.application.AuthService.GetRoles(r.Context(), identity.Id)
	}

	ctx := WithRoles(r.Context(), identityRoles)
	ctx = WithIdentity(ctx, identity)
	ctx = WithSession(ctx, session)
	return ctx, true
}
