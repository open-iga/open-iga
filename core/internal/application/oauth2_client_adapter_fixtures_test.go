package application

import (
	"context"

	"github.com/open-iga/core/internal/contract"
	"github.com/open-iga/core/internal/domain"
)

var _ contract.Oauth2ClientAdapter = &MockOauth2ClientAdapter{}

type MockOauth2ClientAdapter struct {
	consentDetails *domain.ConsentDetails
	error          error
	oauthUser      *domain.OauthUser
}

func (m *MockOauth2ClientAdapter) GetConsentDetails(_ context.Context) (*domain.ConsentDetails, error) {
	return m.consentDetails, m.error
}

func (m *MockOauth2ClientAdapter) FetchOauthUser(_ context.Context, _ string) (*domain.OauthUser, error) {
	return m.oauthUser, m.error
}
