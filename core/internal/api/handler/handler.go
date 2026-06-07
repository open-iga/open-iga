package handler

import (
	"context"
	"log/slog"

	"github.com/open-iga/core/internal/api/generated"
	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/contract"
)

type Handler struct {
	logger      *slog.Logger
	application *contract.RuntimeApplication
	appConfig   *common.AppConfig
}

var _ generated.StrictServerInterface = (*Handler)(nil)

func NewHandler(appConfig *common.AppConfig, logger *slog.Logger, application *contract.RuntimeApplication) *Handler {
	return &Handler{logger, application, appConfig}
}

func (h *Handler) Health(_ context.Context, _ generated.HealthRequestObject) (generated.HealthResponseObject, error) {
	return generated.Health200TextResponse("I'm healthy"), nil
}
