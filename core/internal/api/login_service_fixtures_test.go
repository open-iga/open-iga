package api

import (
	"context"

	"github.com/open-iga/core/internal/contract"
	"github.com/open-iga/core/internal/domain"
)

var _ contract.LoginService = &mockLoginService{}

type mockLoginService struct {
	consentDetails *domain.ConsentDetails
	err            error
	oauthUser      *domain.OauthUser
}

func (m *mockLoginService) GetConsentPageDetails(_ context.Context, _ string) (*domain.ConsentDetails, error) {
	return m.consentDetails, m.err
}

func (m *mockLoginService) GenerateSession(_ context.Context, _ string, _ string) (*domain.OauthUser, error) {
	return m.oauthUser, m.err
}
