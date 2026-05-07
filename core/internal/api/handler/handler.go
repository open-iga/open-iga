package handler

import (
	"log/slog"

	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/contract"
)

type Handler struct {
	logger      *slog.Logger
	application *contract.RuntimeApplication
	appConfig   *common.AppConfig
}

func NewHandler(appConfig *common.AppConfig, logger *slog.Logger, application *contract.RuntimeApplication) *Handler {
	return &Handler{logger, application, appConfig}
}
