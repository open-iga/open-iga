package api

import (
	"context"
	"log/slog"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/contract"
	"github.com/open-iga/core/internal/domain"
)

func CreateMockRouter(t *testing.T, application *contract.RuntimeApplication) *Router {
	humaConfig := huma.DefaultConfig("Test", "1.0.0")
	humaConfig.CreateHooks = nil

	_, testApi := humatest.New(t, humaConfig)

	router := &Router{
		api:         testApi,
		logger:      slog.Default(),
		appConfig:   &common.AppConfig{},
		application: application,
	}

	router.setupRoutes()

	return router
}

var _ contract.LoginService = &mockLoginService{}

type mockLoginService struct {
	consentDetails *contract.ConsentPageDetails
	error          error
	oauthUser      *domain.OauthUser
}

func (m *mockLoginService) GetConsentPageDetails(_ context.Context, _ string) (*contract.ConsentPageDetails, error) {
	return m.consentDetails, m.error
}

func (m *mockLoginService) GenerateSession(_ context.Context, _ string, _ string) (*domain.OauthUser, error) {
	return m.oauthUser, m.error
}
