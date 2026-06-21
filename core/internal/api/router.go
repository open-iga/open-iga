package api

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	codegenMiddleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/open-iga/core/internal/api/generated"
	"github.com/open-iga/core/internal/api/handler"
	"github.com/open-iga/core/internal/api/middleware"
	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/contract"
)

func NewRouter(appConfig *common.AppConfig, logger *slog.Logger, application *contract.RuntimeApplication) *chi.Mux {
	reqMiddleware := middleware.NewMiddleware(appConfig, logger, application)
	reqHandler := handler.NewHandler(appConfig, logger, application)

	spec, err := generated.GetSpec()
	if err != nil {
		panic(err)
	}

	router := chi.NewRouter()

	router.Use(codegenMiddleware.OapiRequestValidator(spec))
	router.Use(reqMiddleware.AuthMiddleware)
	serverInterface := generated.NewStrictHandler(reqHandler, nil)
	generated.HandlerFromMux(serverInterface, router)

	return router
}
