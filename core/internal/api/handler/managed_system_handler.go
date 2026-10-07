package handler

import (
	"context"
	"errors"

	"github.com/open-iga/core/internal/api/generated"
	"github.com/open-iga/core/internal/domain"
)

func (h *Handler) OnboardManagedSystem(ctx context.Context, request generated.OnboardManagedSystemRequestObject) (generated.OnboardManagedSystemResponseObject, error) {
	managedSystem, err := h.application.ManagedSystemService.Onboard(ctx, request.Body.ConnectorOnboardingId)
	if err != nil {

		if errors.Is(err, domain.ErrConnectorNotFound) || errors.Is(err, domain.ErrConnectorNotOnboardable) {
			return generated.OnboardManagedSystem404JSONResponse{
				Message: err.Error(),
			}, nil
		}

		if errors.Is(err, domain.ErrManagedSystemNameExists) {
			return generated.OnboardManagedSystem409JSONResponse{Message: err.Error()}, nil
		}

		return generated.OnboardManagedSystem500JSONResponse{Message: err.Error()}, nil
	}

	return generated.OnboardManagedSystem200JSONResponse{Id: managedSystem.Id}, nil
}

func (h *Handler) ListManagedSystems(ctx context.Context, _ generated.ListManagedSystemsRequestObject) (generated.ListManagedSystemsResponseObject, error) {
	managedSystems, err := h.application.ManagedSystemService.ListManagedSystems(ctx)
	if err != nil {
		return generated.ListManagedSystems500JSONResponse{Message: err.Error()}, nil
	}

	response := make(generated.ListManagedSystems200JSONResponse, len(managedSystems))
	for i, managedSystem := range managedSystems {
		response[i].Id = managedSystem.Id
		response[i].Name = managedSystem.Name
		response[i].ConnectorUrl = managedSystem.ConnectorUrl
		response[i].ConnectorHash = managedSystem.ConnectorHash
		response[i].CreatedAt = managedSystem.CreatedAt
	}

	return response, nil
}
