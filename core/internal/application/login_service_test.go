package application

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/open-iga/core/internal/contract"
	"github.com/open-iga/core/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestLoginService_GetConsentPageDetails(t *testing.T) {
	t.Run("returns error when the provider is unsupported", func(t *testing.T) {
		l := NewLoginService(contract.Oauth2Clients{
			contract.Google: &MockOauth2ClientAdapter{},
		}, slog.Default())

		consentDetails, err := l.GetConsentPageDetails(context.TODO(), "some-provider")

		assert.EqualError(t, err, "unsupported provider: some-provider")
		assert.Nil(t, consentDetails)
	})

	t.Run("returns error when the provider returns an error", func(t *testing.T) {
		l := NewLoginService(contract.Oauth2Clients{
			contract.Google: &MockOauth2ClientAdapter{error: errors.New("failed to generate state")},
		}, slog.Default())

		consentDetails, err := l.GetConsentPageDetails(context.TODO(), "google")

		assert.EqualError(t, err, "failed to get consent page details: failed to generate state")
		assert.Nil(t, consentDetails)
	})

	t.Run("returns consent page details for a supported provider", func(t *testing.T) {
		mockConsentDetails := &domain.ConsentDetails{AuthCodeURL: "auth-code-url", State: "state"}
		l := NewLoginService(contract.Oauth2Clients{
			contract.Google: &MockOauth2ClientAdapter{consentDetails: mockConsentDetails},
		}, slog.Default())

		consentDetails, err := l.GetConsentPageDetails(context.TODO(), "google")

		assert.Nil(t, err)
		assert.Equal(t, mockConsentDetails, consentDetails)
	})
}

func TestLoginService_GenerateSession(t *testing.T) {
	t.Run("returns error when the provider is unsupported", func(t *testing.T) {
		l := NewLoginService(contract.Oauth2Clients{
			contract.Google: &MockOauth2ClientAdapter{},
		}, slog.Default())

		oauthUser, err := l.GenerateSession(context.TODO(), "some-provider", "")

		assert.EqualError(t, err, "unsupported provider: some-provider")
		assert.Nil(t, oauthUser)
	})

	t.Run("returns error when the provider returns an error", func(t *testing.T) {
		l := NewLoginService(contract.Oauth2Clients{
			contract.Google: &MockOauth2ClientAdapter{error: errors.New("request failed")},
		}, slog.Default())

		consentDetails, err := l.GenerateSession(context.TODO(), "google", "auth-code")

		assert.EqualError(t, err, "failed to fetch oauth user: request failed")
		assert.Nil(t, consentDetails)
	})

	t.Run("returns session details for a supported provider", func(t *testing.T) {
		mockOauthUser := domain.NewOauthUser("firstname", "lastname", "name@email.com")
		l := NewLoginService(contract.Oauth2Clients{
			contract.Google: &MockOauth2ClientAdapter{oauthUser: mockOauthUser},
		}, slog.Default())

		consentDetails, err := l.GenerateSession(context.TODO(), "google", "auth-code")

		assert.Nil(t, err)
		assert.Equal(t, mockOauthUser, consentDetails)
	})
}
