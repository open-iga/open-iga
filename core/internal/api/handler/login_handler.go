package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

const AuthStateCookieName = "authState"
const SessionCookieName = "sid"

type LoginOutput struct {
	Status    int
	Location  string       `header:"Location"`
	SetCookie *http.Cookie `header:"Set-Cookie"`
}

type loginInput struct {
	Provider string `path:"provider" required:"true" enum:"google" doc:"Oauth provider to use for login"`
}

type loginCallbackInput struct {
	Provider      string      `path:"provider" required:"true" enum:"google" doc:"Authorization code from Google"`
	Code          string      `query:"code" required:"true" doc:"Authorization code"`
	ActualState   string      `query:"state" required:"true" doc:"State parameter for CSRF protection"`
	ExpectedState http.Cookie `cookie:"authState" doc:"State cookie for CSRF protection"`
}

type LoginCallbackOutput struct {
	Location  string       `header:"Location"`
	SetCookie *http.Cookie `header:"Set-Cookie"`
	Status    int
}

func (h *Handler) LoginHandler(ctx context.Context, l *loginInput) (*LoginOutput, error) {
	consentPageDetails, err := h.application.LoginService.GetConsentPageDetails(ctx, l.Provider)
	if err != nil {
		h.logger.Error("Failed to get consent page details", "error", err, "provider", l.Provider)
		return nil, huma.Error500InternalServerError("failed to get consent page details", errors.New("failed to get consent page details"))
	}

	resp := &LoginOutput{
		Status:   http.StatusFound,
		Location: consentPageDetails.AuthCodeURL,
	}
	resp.SetCookie = &http.Cookie{
		Name:     AuthStateCookieName,
		Value:    consentPageDetails.State,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   20, // Valid for 20s
		Secure:   true,
	}

	return resp, nil
}

func (h *Handler) LoginCallBackHandler(ctx context.Context, i *loginCallbackInput) (*LoginCallbackOutput, error) {
	if i.ExpectedState.Value == "" || i.ExpectedState.Value != i.ActualState {
		h.logger.Error("State mismatch in login callback")
		return nil, huma.Error422UnprocessableEntity("state mismatch", errors.New("state mismatch"))
	}

	if i.Code == "" {
		h.logger.Error("Received empty string for auth code in login callback")
		return nil, huma.Error422UnprocessableEntity("missing authcode", errors.New("missing authcode"))
	}

	session, err := h.application.LoginService.GenerateSession(ctx, i.Provider, i.Code)
	if err != nil {
		h.logger.Error("Failed to generate session", "error", err)
		return nil, huma.Error500InternalServerError("failed to generate session", errors.New("failed to generate session"))
	}

	resp := &LoginCallbackOutput{
		Location: "/",
		SetCookie: &http.Cookie{
			Name:     SessionCookieName,
			Value:    session.SessionId,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
			MaxAge:   session.ValidityInSeconds(),
			Secure:   true,
		},
		Status: http.StatusFound,
	}

	return resp, nil
}
