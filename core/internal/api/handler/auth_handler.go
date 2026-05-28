package handler

import (
	"context"
	"net/http"

	"github.com/open-iga/core/internal/api/generated"
)

const AuthStateCookieName = "authState"
const SessionCookieName = "sid"

func (h *Handler) AuthDetails(ctx context.Context, request generated.AuthDetailsRequestObject) (generated.AuthDetailsResponseObject, error) {
	provider := string(request.Provider)
	consentPageDetails, err := h.application.LoginService.GetConsentPageDetails(ctx, provider)
	if err != nil {
		h.logger.Error("Failed to get consent page details", "error", err, "provider", provider)
		return generated.AuthDetails500JSONResponse{Message: "failed to get consent page details"}, nil
	}

	authStateCookie := http.Cookie{
		Name:     AuthStateCookieName,
		Value:    consentPageDetails.State,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   20, // Valid for only 20s
		Secure:   true,
	}
	cookieStr := authStateCookie.String()

	return generated.AuthDetails200JSONResponse{
		Body: struct {
			AuthCodeUrl string `json:"authCodeUrl"`
		}{
			AuthCodeUrl: consentPageDetails.AuthCodeURL,
		},
		Headers: generated.AuthDetails200ResponseHeaders{
			SetCookie: &cookieStr,
		},
	}, nil
}

func (h *Handler) AuthCallback(ctx context.Context, request generated.AuthCallbackRequestObject) (generated.AuthCallbackResponseObject, error) {
	authState := *request.Params.AuthState
	if authState == "" {
		h.logger.Error("Received empty string for auth code in login callback")
		return generated.AuthCallback500JSONResponse{Message: "received empty auth code"}, nil
	}

	if request.Params.State != authState {
		h.logger.Error("State mismatch in login callback")
		return generated.AuthCallback422JSONResponse{Message: "state mismatch in login callback"}, nil
	}

	session, err := h.application.LoginService.GenerateSession(ctx, string(request.Provider), request.Params.Code)
	if err != nil {
		h.logger.Error("Failed to generate session", "error", err)
		return generated.AuthCallback500JSONResponse{Message: "failed to generate session"}, nil
	}

	cookie := http.Cookie{
		Name:     SessionCookieName,
		Value:    session.SessionId,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   session.ValidityInSeconds(),
		Secure:   true,
	}
	cookieStr := cookie.String()

	return generated.AuthCallback201JSONResponse{
		Body: struct {
			Redirect string `json:"redirect"`
		}{
			Redirect: h.appConfig.Redirect.Home,
		},
		Headers: generated.AuthCallback201ResponseHeaders{
			SetCookie: &cookieStr,
		},
	}, nil
}
