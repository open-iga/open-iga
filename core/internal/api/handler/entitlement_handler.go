package handler

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/open-iga/core/internal/api/generated"
	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/domain"
)

const (
	defaultEntitlementPageSize = 20
	maxEntitlementPageSize     = 100
)

func (h *Handler) SearchEntitlements(ctx context.Context, request generated.SearchEntitlementsRequestObject) (generated.SearchEntitlementsResponseObject, error) {
	search, err := toEntitlementSearch(request.Params)
	if err != nil {
		return generated.SearchEntitlements400JSONResponse{Message: err.Error()}, nil
	}

	entitlements, total, err := h.application.DiscoveryService.SearchEntitlements(ctx, search)
	if err != nil {
		return generated.SearchEntitlements500JSONResponse{Message: err.Error()}, nil
	}

	var response generated.SearchEntitlements200JSONResponse
	response.Total = total
	response.Entitlements = common.MakeSliceLike(response.Entitlements, len(entitlements))
	for i, entitlement := range entitlements {
		response.Entitlements[i].Id = entitlement.Id
		response.Entitlements[i].ManagedSystemId = entitlement.ManagedSystemId
		response.Entitlements[i].TenantId = toOptionalString(entitlement.TenantId)
		response.Entitlements[i].Resource = entitlement.Resource
		response.Entitlements[i].ExternalId = entitlement.ExternalId
		response.Entitlements[i].DisplayName = entitlement.DisplayName
		response.Entitlements[i].Description = toOptionalString(entitlement.Description)
	}

	return response, nil
}

// toEntitlementSearch reads the query parameters; the error says which one is wrong.
func toEntitlementSearch(params generated.SearchEntitlementsParams) (search domain.EntitlementSearch, err error) {
	if params.Q != nil {
		search.Query = *params.Q
	}

	if params.Resource != nil {
		search.Resource = *params.Resource
	}

	if search.ManagedSystemId, err = parseOptionalId(params.SystemId); err != nil {
		return search, errors.New("systemId is not a valid id")
	}

	if search.TenantId, err = parseOptionalId(params.TenantId); err != nil {
		return search, errors.New("tenantId is not a valid id")
	}

	search.Limit, search.Offset, err = toPage(params.Limit, params.Offset)

	return search, err
}

// toPage gives the page size and the offset, with the default size when none is set.
func toPage(limit *int, offset *int) (pageSize int32, pageOffset int32, err error) {
	pageSize = defaultEntitlementPageSize
	if limit != nil {
		if *limit < 1 || *limit > maxEntitlementPageSize {
			return 0, 0, errors.New("limit must be from 1 to 100")
		}
		pageSize = int32(*limit)
	}

	if offset != nil {
		if *offset < 0 || *offset > math.MaxInt32 {
			return 0, 0, errors.New("offset must be from 0 to 2147483647")
		}
		pageOffset = int32(*offset)
	}

	return pageSize, pageOffset, nil
}

// parseOptionalId gives uuid.Nil, which means "any", when the parameter is not set.
func parseOptionalId(id *string) (uuid.UUID, error) {
	if id == nil {
		return uuid.Nil, nil
	}

	parsed, err := uuid.Parse(*id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parse id: %w", err)
	}

	return parsed, nil
}

func toOptionalString(text string) *string {
	if text == "" {
		return nil
	}

	return &text
}
