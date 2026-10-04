package handler

import (
	"context"
	"fmt"

	"github.com/open-iga/core/internal/api/generated"
)

func (h *Handler) OnboardConnector(ctx context.Context, request generated.OnboardConnectorRequestObject) (generated.OnboardConnectorResponseObject, error) {
	onboardingId := h.application.ConnectorService.ValidateByUrl(ctx, request.Body.ConnectorUrl, request.Body.ConnectorSha)

	return generated.OnboardConnector201JSONResponse{OnboardingId: onboardingId}, nil
}

func (h *Handler) GetConnectorOnboardingRequestDetails(ctx context.Context, request generated.GetConnectorOnboardingRequestDetailsRequestObject) (generated.GetConnectorOnboardingRequestDetailsResponseObject, error) {
	connectorValidationResult, ok := h.application.ConnectorService.GetOnboardedConnectorDetails(ctx, request.OnboardingId)

	if !ok {
		return generated.GetConnectorOnboardingRequestDetails404JSONResponse{
			Message: fmt.Sprintf("No connector found for onboarding ID: %s", request.OnboardingId),
		}, nil
	}

	resp := generated.GetConnectorOnboardingRequestDetails200JSONResponse{
		Status:        generated.GetConnectorOnboardingRequestDetails200JSONResponseBodyStatus(connectorValidationResult.Status),
		Error:         nil,
		ConnectorSpec: nil,
	}

	if connectorValidationResult.Error != nil {
		resp.Error = new(connectorValidationResult.Error.Error())
	}
	resp.ConnectorSpec = connectorValidationResult.ConnectorSpec

	return resp, nil
}
