package api

import (
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/contract"
)

type Router struct {
	api         huma.API
	router      *http.ServeMux
	logger      *slog.Logger
	application *contract.RuntimeApplication
	appConfig   *common.AppConfig
}

func NewRouter(appConfig *common.AppConfig, application *contract.RuntimeApplication, logger *slog.Logger) *Router {
	router := http.NewServeMux()

	api := humago.New(router, huma.DefaultConfig("OpenIGA API", "1.0.0"))
	return &Router{router: router, api: api, logger: logger, application: application, appConfig: appConfig}
}

func (r *Router) Start() {
	r.logger.Debug("Setting up routes")
	r.setupRoutes()

	addr := r.appConfig.Port
	server := &http.Server{
		Addr:              addr,
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           r.router,
	}

	r.logger.Info("Starting server", "address", addr)
	log.Fatal(server.ListenAndServe())
}
