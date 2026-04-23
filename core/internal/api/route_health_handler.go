package api

import (
	"context"
)

type Output struct {
	Body struct {
		Message string `json:"message" doc:"Health status of the service"`
	}
}

func (r *Router) healthHandler(_ context.Context, _ *struct{}) (*Output, error) {
	resp := &Output{}
	resp.Body.Message = "I'm Healthy!"
	return resp, nil
}
