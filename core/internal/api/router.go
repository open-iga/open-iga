package api

import (
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/contract/application"
)

type Router struct {
	api         *huma.API
	router      *http.ServeMux
	logger      *slog.Logger
	application *application.RuntimeApplication
	appConfig   *common.AppConfig
}

func NewRouter(appConfig *common.AppConfig, application *application.RuntimeApplication, logger *slog.Logger) *Router {
	router := http.NewServeMux()

	api := humago.New(router, huma.DefaultConfig("Open IGA API", "1.0.0"))

	return &Router{router: router, api: &api, logger: logger, application: application, appConfig: appConfig}
}

func (r *Router) Start() error {
	r.logger.Debug("Setting up routes")
	r.setupRoutes()

	addr := r.appConfig.Port
	r.logger.Info("Starting API server", "address", addr)
	return http.ListenAndServe(addr, r.router)
}
