package middleware

import (
	"net/http"
	"slices"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/open-iga/core/internal/api/handler"
)

var NoAuthPath = []string{"/api/health"}

// AuthMiddleware middleware that protects all the routes
func (m *Middleware) AuthMiddleware(ctx huma.Context, next func(huma.Context)) {
	path := ctx.URL().Path
	defaultContext := ctx.Context()
	if slices.Contains(NoAuthPath, path) {
		next(ctx)
		return
	}

	// route all the FE requests to the asset; requests that doesn't start with /api
	if !strings.HasPrefix(path, "/api") {
		next(ctx)
		return
	}

	// Redirect to login page when session is missing where the users can select the required oauth provider
	cookie, err := huma.ReadCookie(ctx, handler.SessionCookieName)
	if err != nil {
		m.logger.Debug("unable to read session cookie from the request")
		huma.Context.AppendHeader(ctx, "Location", "/login")
		huma.Context.SetStatus(ctx, http.StatusFound)
		return
	}

	// If session cookie is available
	// 1. If invalid, redirect to login page
	// 2. if expired, redirect to login page
	// 3. If failed to expire, redirect to login again
	//FIXME: There is a potential infinite loop here with FailedToExpireSession. Let the users use first and run into this scenario
	identity, session, err := m.application.LoginService.ValidateSession(defaultContext, cookie.Value)
	if err != nil {
		m.logger.Debug(err.Error())
		huma.Context.AppendHeader(ctx, "Location", "/login")
		huma.Context.SetStatus(ctx, http.StatusFound)
		return
	}

	// if there is already a session and user clicks on the login page, redirect to home page
	if strings.HasPrefix(path, "/api/login") {
		huma.Context.AppendHeader(ctx, "Location", "/")
		huma.Context.SetStatus(ctx, http.StatusFound)
		return
	}

	huma.WithContext(ctx, SetIdentityInContent(ctx.Context(), identity))
	next(huma.WithContext(ctx, SetSessionInContent(ctx.Context(), session)))
}
