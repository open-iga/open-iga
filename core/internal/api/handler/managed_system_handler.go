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

func (h *Handler) OnboardManagedSystemTenant(ctx context.Context, request generated.OnboardManagedSystemTenantRequestObject) (generated.OnboardManagedSystemTenantResponseObject, error) {
	tenant, created, err := h.application.ManagedSystemService.OnboardTenant(ctx, request.Id, request.Body.ExternalId, request.Body.Name)
	if err != nil {
		if errors.Is(err, domain.ErrManagedSystemNotFound) {
			return generated.OnboardManagedSystemTenant404JSONResponse{Message: err.Error()}, nil
		}

		return generated.OnboardManagedSystemTenant500JSONResponse{Message: err.Error()}, nil
	}

	// 201 for a fresh tenant, 200 when it already existed (idempotent).
	if !created {
		return generated.OnboardManagedSystemTenant200JSONResponse{Id: tenant.Id}, nil
	}

	return generated.OnboardManagedSystemTenant201JSONResponse{Id: tenant.Id}, nil
}

func (h *Handler) GetManagedSystem(ctx context.Context, request generated.GetManagedSystemRequestObject) (generated.GetManagedSystemResponseObject, error) {
	managedSystem, err := h.application.ManagedSystemService.GetManagedSystem(ctx, request.Id)
	if err != nil {
		if errors.Is(err, domain.ErrManagedSystemNotFound) {
			return generated.GetManagedSystem404JSONResponse{Message: err.Error()}, nil
		}
		return generated.GetManagedSystem500JSONResponse{Message: err.Error()}, nil
	}

	var spec domain.ConnectorSpec
	if managedSystem.ConnectorSpec != nil {
		spec = *managedSystem.ConnectorSpec
	}

	return generated.GetManagedSystem200JSONResponse{
		Id:            managedSystem.Id,
		Name:          managedSystem.Name,
		ConnectorUrl:  managedSystem.ConnectorUrl,
		ConnectorHash: managedSystem.ConnectorHash,
		ConnectorSpec: spec,
		CreatedAt:     managedSystem.CreatedAt,
	}, nil
}

func (h *Handler) ListManagedSystemTenants(ctx context.Context, request generated.ListManagedSystemTenantsRequestObject) (generated.ListManagedSystemTenantsResponseObject, error) {
	tenants, err := h.application.ManagedSystemService.ListTenants(ctx, request.Id)
	if err != nil {
		return generated.ListManagedSystemTenants500JSONResponse{Message: err.Error()}, nil
	}

	response := make(generated.ListManagedSystemTenants200JSONResponse, len(tenants))
	for i, tenant := range tenants {
		response[i].Id = tenant.Id
		response[i].ExternalId = tenant.ExternalId
		response[i].Name = tenant.Name
		response[i].CreatedAt = tenant.CreatedAt
	}

	return response, nil
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
