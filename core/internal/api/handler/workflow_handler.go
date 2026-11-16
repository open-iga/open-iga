package handler

import (
	"context"
	"errors"

	"github.com/open-iga/core/internal/api/generated"
	"github.com/open-iga/core/internal/domain"
)

// isInvalidWorkflow is true for errors that come from what the caller sent: a bad workflow, or a system, tenant or entitlement that does not exist.
func isInvalidWorkflow(err error) bool {
	return errors.Is(err, domain.ErrInvalidWorkflow) || errors.Is(err, domain.ErrManagedSystemNotFound) ||
		errors.Is(err, domain.ErrTenantNotFound) || errors.Is(err, domain.ErrEntitlementNotFound)
}

func (h *Handler) CreateWorkflow(ctx context.Context, request generated.CreateWorkflowRequestObject) (generated.CreateWorkflowResponseObject, error) {
	err := h.application.WorkflowService.CreateWorkflow(ctx, domain.WorkflowDefinition{Name: request.Body.Name, Steps: request.Body.Steps})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrWorkflowAlreadyExists):
			return generated.CreateWorkflow409JSONResponse{Message: err.Error()}, nil
		case isInvalidWorkflow(err):
			return generated.CreateWorkflow400JSONResponse{Message: err.Error()}, nil
		}

		return generated.CreateWorkflow500JSONResponse{Message: err.Error()}, nil
	}

	return generated.CreateWorkflow201Response{}, nil
}

func (h *Handler) GetWorkflow(ctx context.Context, request generated.GetWorkflowRequestObject) (generated.GetWorkflowResponseObject, error) {
	definition, err := h.application.WorkflowService.GetWorkflow(ctx, request.Name)
	if err != nil {
		if errors.Is(err, domain.ErrWorkflowNotFound) {
			return generated.GetWorkflow404JSONResponse{Message: err.Error()}, nil
		}

		return generated.GetWorkflow500JSONResponse{Message: err.Error()}, nil
	}

	return generated.GetWorkflow200JSONResponse{Name: definition.Name, Steps: definition.Steps}, nil
}
