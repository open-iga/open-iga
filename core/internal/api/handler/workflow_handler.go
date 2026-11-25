package handler

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/open-iga/core/internal/api/generated"
	"github.com/open-iga/core/internal/domain"
)

var (
	errIdentityRequired = errors.New("identityId is required")
	errIdentityInvalid  = errors.New("identityId is not a valid id")
)

// isInvalidWorkflow is true for errors that come from what the caller sent: a bad workflow, or a system, tenant or entitlement that does not exist.
func isInvalidWorkflow(err error) bool {
	return errors.Is(err, domain.ErrInvalidWorkflow) || errors.Is(err, domain.ErrManagedSystemNotFound) ||
		errors.Is(err, domain.ErrTenantNotFound) || errors.Is(err, domain.ErrEntitlementNotFound)
}

func (h *Handler) CreateWorkflow(ctx context.Context, request generated.CreateWorkflowRequestObject) (generated.CreateWorkflowResponseObject, error) {
	err := h.application.WorkflowService.CreateWorkflow(ctx, domain.WorkflowDefinition{Name: request.Body.Name, Type: request.Body.Type, Steps: request.Body.Steps})
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

	schedule, err := h.application.WorkflowService.GetWorkflowSchedule(ctx, request.Name)
	if err != nil {
		return generated.GetWorkflow500JSONResponse{Message: err.Error()}, nil
	}

	return generated.GetWorkflow200JSONResponse{Name: definition.Name, Type: definition.Type, Steps: definition.Steps, Schedule: schedule}, nil
}

func (h *Handler) StartWorkflow(ctx context.Context, request generated.StartWorkflowRequestObject) (generated.StartWorkflowResponseObject, error) {
	identityId, err := parseRunIdentity(request.Name, request.Body)
	if err != nil {
		return generated.StartWorkflow400JSONResponse{Message: err.Error()}, nil
	}

	runId, err := h.application.WorkflowService.StartWorkflow(ctx, request.Name, identityId)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrWorkflowNotFound) || errors.Is(err, domain.ErrNoIdentityFound):
			return generated.StartWorkflow404JSONResponse{Message: err.Error()}, nil
		case errors.Is(err, domain.ErrInvalidWorkflow):
			return generated.StartWorkflow400JSONResponse{Message: err.Error()}, nil
		}

		return generated.StartWorkflow500JSONResponse{Message: err.Error()}, nil
	}

	if runId == "" {
		return generated.StartWorkflow204Response{}, nil
	}

	return generated.StartWorkflow201JSONResponse{Id: runId}, nil
}

// parseRunIdentity reads the identity of a run. Discovery is about no identity, so it needs none.
func parseRunIdentity(workflowName string, body *generated.StartWorkflowJSONRequestBody) (uuid.UUID, error) {
	if body == nil || body.IdentityId == nil || *body.IdentityId == "" {
		if workflowName == domain.DiscoveryWorkflowName {
			return uuid.Nil, nil
		}

		return uuid.Nil, errIdentityRequired
	}

	identityId, err := uuid.Parse(*body.IdentityId)
	if err != nil {
		return uuid.Nil, errIdentityInvalid
	}

	return identityId, nil
}

func (h *Handler) GetWorkflowRun(ctx context.Context, request generated.GetWorkflowRunRequestObject) (generated.GetWorkflowRunResponseObject, error) {
	state, err := h.application.WorkflowService.GetRun(ctx, request.Name, request.Id)
	if err != nil {
		if errors.Is(err, domain.ErrWorkflowRunNotFound) {
			return generated.GetWorkflowRun404JSONResponse{Message: err.Error()}, nil
		}

		return generated.GetWorkflowRun500JSONResponse{Message: err.Error()}, nil
	}

	return generated.GetWorkflowRun200JSONResponse{Id: state.Id, Steps: state.Steps}, nil
}

func (h *Handler) ContinueWorkflowStep(ctx context.Context, request generated.ContinueWorkflowStepRequestObject) (generated.ContinueWorkflowStepResponseObject, error) {
	if err := h.application.WorkflowService.Continue(ctx, request.Name, request.Id, request.Step); err != nil {
		if errors.Is(err, domain.ErrWorkflowRunNotFound) {
			return generated.ContinueWorkflowStep404JSONResponse{Message: err.Error()}, nil
		}

		if errors.Is(err, domain.ErrWorkflowStepNotGated) {
			return generated.ContinueWorkflowStep409JSONResponse{Message: err.Error()}, nil
		}

		return generated.ContinueWorkflowStep500JSONResponse{Message: err.Error()}, nil
	}

	return generated.ContinueWorkflowStep204Response{}, nil
}

func (h *Handler) SetWorkflowSchedule(ctx context.Context, request generated.SetWorkflowScheduleRequestObject) (generated.SetWorkflowScheduleResponseObject, error) {
	if err := h.application.WorkflowService.SetWorkflowSchedule(ctx, request.Body.Name, request.Body.Schedule); err != nil {
		if errors.Is(err, domain.ErrInvalidWorkflow) {
			return generated.SetWorkflowSchedule400JSONResponse{Message: err.Error()}, nil
		}

		return generated.SetWorkflowSchedule500JSONResponse{Message: err.Error()}, nil
	}

	return generated.SetWorkflowSchedule204Response{}, nil
}
