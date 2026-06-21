package middleware

import (
	"log/slog"

	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/contract"
)

type Middleware struct {
	logger      *slog.Logger
	application *contract.RuntimeApplication
	appConfig   *common.AppConfig
}

func NewMiddleware(appConfig *common.AppConfig, logger *slog.Logger, application *contract.RuntimeApplication) *Middleware {
	return &Middleware{logger: logger, application: application, appConfig: appConfig}
}
