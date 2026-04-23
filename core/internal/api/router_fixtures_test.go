package api

import (
	"log/slog"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/contract"
)

func CreateMockRouter(t *testing.T, application *contract.RuntimeApplication) *Router {
	humaConfig := huma.DefaultConfig("Test", "1.0.0")
	humaConfig.CreateHooks = nil

	_, testApi := humatest.New(t, humaConfig)

	router := NewRouter(&common.AppConfig{}, application, slog.Default(), WithApi(testApi))

	return router
}
