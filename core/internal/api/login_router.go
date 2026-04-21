package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

const AuthStateCookieName = "authState"

type loginOutput struct {
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

type loginCallbackOutputBody struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

type loginCallbackOutput struct {
	Body   loginCallbackOutputBody
	Status int
}

func (r *Router) addLoginRoute() {
	huma.Get(r.api, "/login/{provider}", func(ctx context.Context, l *loginInput) (*loginOutput, error) {
		consentPageDetails, err := r.application.LoginService.GetConsentPageDetails(ctx, l.Provider)
		if err != nil {
			r.logger.Error("Failed to get consent page details", "error", err, "provider", l.Provider)
			return nil, huma.Error500InternalServerError("failed to get consent page details", errors.New("failed to get consent page details"))
		}

		resp := &loginOutput{
			Status:   http.StatusFound,
			Location: consentPageDetails.AuthCodeURL,
		}
		resp.SetCookie = &http.Cookie{
			Name:     AuthStateCookieName,
			Value:    consentPageDetails.State,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   300,
			Secure:   true,
		}

		return resp, nil
	})
}

func (r *Router) addLoginCallbackRoute() {
	huma.Get(r.api, "/login/{provider}/callback", func(ctx context.Context, i *loginCallbackInput) (*loginCallbackOutput, error) {
		if i.ExpectedState.Value != i.ActualState {
			r.logger.Error("State mismatch in login callback", "expected", i.ExpectedState.Value, "actual", i.ActualState)
			return nil, huma.Error422UnprocessableEntity("state mismatch", errors.New("state mismatch"))
		}

		session, err := r.application.LoginService.GenerateSession(ctx, i.Provider, i.Code)
		if err != nil {
			r.logger.Error("Failed to generate session", "error", err)
			return nil, huma.Error500InternalServerError("failed to generate session", errors.New("failed to generate session"))
		}

		resp := &loginCallbackOutput{
			Body: loginCallbackOutputBody{
				FirstName: session.FirstName,
				LastName:  session.LastName,
				Email:     session.Email,
			},
		}
		resp.Status = http.StatusCreated

		return resp, nil
	})
}
