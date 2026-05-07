package api

import (
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/open-iga/core/internal/api/handler"
	"github.com/open-iga/core/internal/api/middleware"
	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/contract"
)

type Router struct {
	api        huma.API
	serverMux  *http.ServeMux
	logger     *slog.Logger
	appConfig  *common.AppConfig
	handler    *handler.Handler
	middleware *middleware.Middleware
}

type Option func(*Router)

func WithApi(api huma.API) Option {
	return func(router *Router) {
		router.api = api
	}
}

func NewRouter(appConfig *common.AppConfig, logger *slog.Logger, application *contract.RuntimeApplication, opts ...Option) *Router {
	serveMux := http.NewServeMux()

	api := humago.New(serveMux, huma.DefaultConfig("OpenIGA API", "1.0.0"))
	reqHandler := handler.NewHandler(appConfig, logger, application)
	reqMiddleware := middleware.NewMiddleware(appConfig, logger, application, api)
	router := &Router{
		serverMux:  serveMux,
		api:        api,
		logger:     logger,
		appConfig:  appConfig,
		handler:    reqHandler,
		middleware: reqMiddleware,
	}

	for _, opt := range opts {
		opt(router)
	}

	logger.Debug("Setting up routes...")
	router.setupRoutes()

	return router
}

func (r *Router) setupRoutes() {

	r.api.UseMiddleware(r.middleware.AuthMiddleware)

	huma.Get(r.api, "/api/health", r.handler.HealthHandler)
	huma.Get(r.api, "/api/login/{provider}", r.handler.LoginHandler)
	huma.Get(r.api, "/api/login/{provider}/callback", r.handler.LoginCallBackHandler)
	huma.Get(r.api, "/", r.handler.HealthHandler)
	huma.Get(r.api, "/login", r.handler.HealthHandler)

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
