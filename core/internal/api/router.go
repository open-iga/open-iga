package api

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	middleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/open-iga/core/internal/api/generated"
	"github.com/open-iga/core/internal/api/handler"
	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/contract"
)

func NewRouter(appConfig *common.AppConfig, logger *slog.Logger, application *contract.RuntimeApplication) *chi.Mux {
	reqHandler := handler.NewHandler(appConfig, logger, application)

	spec, err := generated.GetSpec()

	if err != nil {
		panic(err)
	}

	router := chi.NewRouter()

	router.Use(middleware.OapiRequestValidator(spec))
	serverInterface := generated.NewStrictHandler(reqHandler, nil)
	generated.HandlerFromMux(serverInterface, router)

	return router
}
