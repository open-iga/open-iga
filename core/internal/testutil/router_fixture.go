package testutil

import (
	"log/slog"
	"os"

	"github.com/go-chi/chi/v5"
	middleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/open-iga/core/internal/api/generated"
	"github.com/open-iga/core/internal/common"
)

type TestConfig struct {
	HostUrl                 string
	GoogleOauthClientId     string
	GoogleOauthClientSecret string
	DatabaseURL             string
}

func NewTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func WithDatabaseUrlOverride(value string) func(config *common.AppConfig) {
	return func(config *common.AppConfig) { config.Database.URL = value }
}

func NewTestAppConfig(overWriteFn ...func(config *common.AppConfig)) *common.AppConfig {
	config := &common.AppConfig{
		Port:    ":8080",
		HostUrl: "http://localhost:8080",
		Oauth: common.Oauth{Google: common.OauthConfig{
			ClientId:     "client-id",
			ClientSecret: "client-secret",
		}},
		Database: common.Database{URL: "postgres://test:test@localhost:5432/open_iga"},
		Redirect: common.Redirect{
			Home:               "/",
			SignIn:             "/auth/sign-in",
			GoogleAuthCallback: "/auth/google/callback",
		},
	}
	for _, fn := range overWriteFn {
		fn(config)
	}

	return config
}

func NewMockRouter(ssi generated.StrictServerInterface) *chi.Mux {
	spec, _ := generated.GetSpec()

	router := chi.NewRouter()
	router.Use(middleware.OapiRequestValidator(spec))
	serverInterface := generated.NewStrictHandler(ssi, nil)
	generated.HandlerFromMux(serverInterface, router)

	return router
}
