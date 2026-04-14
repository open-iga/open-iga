package api

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
)

type Output struct {
	Body struct {
		Message string `json:"message" doc:"Health status of the service"`
	}
}

func (r *Router) createHealthRouter() {
	huma.Get(*r.api, "/health", func(_ context.Context, _ *struct{}) (*Output, error) {
		resp := &Output{}
		resp.Body.Message = "I'm Healthy!"
		return resp, nil
	})
}
