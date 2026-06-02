package testutil

import (
	"log/slog"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	middleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/open-iga/core/internal/api/generated"
	"github.com/open-iga/core/internal/common"
)

func NewTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func NewTestAppConfig(t *testing.T) *common.AppConfig {
	t.Setenv("HOST_URL", "http://localhost:8080")
	t.Setenv("GOOGLE_OAUTH_CLIENT_ID", "dummy-client-id")
	t.Setenv("GOOGLE_OAUTH_CLIENT_SECRET", "dummy-client-secret")
	t.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/open_iga")

	return common.NewAppConfig()
}

func NewMockRouter(ssi generated.StrictServerInterface) *chi.Mux {
	spec, _ := generated.GetSpec()

	router := chi.NewRouter()
	router.Use(middleware.OapiRequestValidator(spec))
	serverInterface := generated.NewStrictHandler(ssi, nil)
	generated.HandlerFromMux(serverInterface, router)

	return router
}
