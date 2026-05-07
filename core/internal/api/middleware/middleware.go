package middleware

import (
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/contract"
)

type Middleware struct {
	logger      *slog.Logger
	application *contract.RuntimeApplication
	appConfig   *common.AppConfig
	api         huma.API
}

func NewMiddleware(appConfig *common.AppConfig, logger *slog.Logger, application *contract.RuntimeApplication, api huma.API) *Middleware {
	return &Middleware{logger, application, appConfig, api}
}
