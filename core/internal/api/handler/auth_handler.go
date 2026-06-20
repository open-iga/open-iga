package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/open-iga/core/internal/api/generated"
	"github.com/open-iga/core/internal/api/middleware"
	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/domain"
)

func (h *Handler) AuthDetails(ctx context.Context, request generated.AuthDetailsRequestObject) (generated.AuthDetailsResponseObject, error) {
	provider := string(request.Provider)
	consentPageDetails, err := h.application.LoginService.GetConsentPageDetails(ctx, provider)
	if err != nil {
		errDetails := fmt.Errorf("auth handler: %w", err)

		h.logger.Error(errDetails.Error())
		return generated.AuthDetails500JSONResponse{Message: errDetails.Error()}, nil
	}

	authStateCookie := http.Cookie{
		Name:     common.AuthStateCookieName,
		Value:    consentPageDetails.State,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   20, // Valid for only 20s
		Secure:   true,
	}
	return generated.AuthDetails201JSONResponse{
		Body: struct {
			AuthCodeUrl string `json:"authCodeUrl"`
		}{
			AuthCodeUrl: consentPageDetails.AuthCodeURL,
		},
		Headers: generated.AuthDetails201ResponseHeaders{
			SetCookie: new(authStateCookie.String()),
		},
	}, nil
}

func (h *Handler) AuthCallback(ctx context.Context, request generated.AuthCallbackRequestObject) (generated.AuthCallbackResponseObject, error) {
	if request.Params.AuthState == nil {
		return generated.AuthCallback422JSONResponse{Message: "missing state cookie"}, nil
	}

	authState := *request.Params.AuthState
	if authState == "" {
		h.logger.Error("Received empty string for auth code in auth callback")
		return generated.AuthCallback422JSONResponse{Message: "empty state cookie"}, nil
	}

	if request.Params.State != authState {
		h.logger.Error("State mismatch in auth callback", "provider", request.Provider)
		return generated.AuthCallback422JSONResponse{Message: "state mismatch"}, nil
	}

	session, err := h.application.LoginService.GenerateSession(ctx, string(request.Provider), request.Params.Code)
	if err != nil {
		errDetails := fmt.Errorf("auth callback handler: %w", err)

		h.logger.Error(errDetails.Error())
		return generated.AuthCallback500JSONResponse{Message: errDetails.Error()}, nil
	}

	cookie := http.Cookie{
		Name:     common.SessionCookieName,
		Value:    session.SessionId,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   session.ValidityInSeconds(),
		Secure:   true,
	}
	return generated.AuthCallback201JSONResponse{
		Body: struct {
			Redirect string `json:"redirect"`
		}{
			Redirect: h.appConfig.Redirect.Home,
		},
		Headers: generated.AuthCallback201ResponseHeaders{
			SetCookie: new(cookie.String()),
		},
	}, nil
}

func (h *Handler) Logout(ctx context.Context, _ generated.LogoutRequestObject) (generated.LogoutResponseObject, error) {
	session, err := middleware.GetSession(ctx)

	if err != nil {
		return generated.Logout500JSONResponse{Message: err.Error()}, nil
	}

	err = h.application.LoginService.DeactivateSession(ctx, session.SessionId)
	if err != nil && errors.Is(err, domain.ErrSessionNotFound) {
		return generated.Logout400JSONResponse{Message: "session not found"}, nil
	}

	if err != nil {
		return generated.Logout500JSONResponse{Message: err.Error()}, nil
	}

	return generated.Logout200JSONResponse{Message: "Session deactivated"}, nil
}
