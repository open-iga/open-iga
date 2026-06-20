package handler

import (
	"context"

	"github.com/open-iga/core/internal/api/generated"
	"github.com/open-iga/core/internal/api/middleware"
)

func (h *Handler) GetUserDetails(ctx context.Context, _ generated.GetUserDetailsRequestObject) (generated.GetUserDetailsResponseObject, error) {
	identity, err := middleware.GetIdentity(ctx)

	if err != nil {
		return generated.GetUserDetails500JSONResponse{Message: err.Error()}, nil
	}

	return generated.GetUserDetails200JSONResponse{
		Email:     identity.Email,
		FirstName: identity.FirstName,
		Id:        identity.Id.String(),
		LastName:  identity.LastName,
	}, nil
}
