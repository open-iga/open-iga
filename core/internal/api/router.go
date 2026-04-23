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
	serverMux   *http.ServeMux
	logger      *slog.Logger
	application *contract.RuntimeApplication
	appConfig   *common.AppConfig
}

type Option func(*Router)

func WithApi(api huma.API) Option {
	return func(router *Router) {
		router.api = api
	}
}

func NewRouter(appConfig *common.AppConfig, application *contract.RuntimeApplication, logger *slog.Logger, opts ...Option) *Router {
	serveMux := http.NewServeMux()

	api := humago.New(serveMux, huma.DefaultConfig("OpenIGA API", "1.0.0"))
	router := &Router{serverMux: serveMux, api: api, logger: logger, application: application, appConfig: appConfig}

	for _, opt := range opts {
		opt(router)
	}

	logger.Debug("Setting up routes")
	router.setupRoutes()

	return router
}

func (r *Router) Start() {
	addr := r.appConfig.Port
	server := &http.Server{
		Addr:              addr,
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           r.serverMux,
	}

	r.logger.Info("Starting server", "address", addr)
	log.Fatal(server.ListenAndServe())
}
