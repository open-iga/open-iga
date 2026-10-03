package middleware

import (
	"context"
	"net/http"
	"slices"

	"github.com/open-iga/core/internal/api/generated"
	"github.com/open-iga/core/internal/domain"
)

// operationRoles maps an oapi-codegen operationId to the roles allowed to call it.
// nil/empty means the operation is public (no role required). An operation absent
// from this map is denied (fail closed) until a policy is added for it.
var operationRoles = map[string][]string{
	"Health":       nil,
	"AuthDetails":  nil,
	"AuthCallback": nil,
	"Logout":       nil,

	"GetUserDetails":                       {domain.DefaultIdentityRole, domain.AdminRole},
	"OnboardConnector":                     {domain.AdminRole},
	"GetConnectorOnboardingRequestDetails": {domain.AdminRole, domain.DefaultIdentityRole},
	"OnboardManagedSystem":                 {domain.AdminRole},
}

// AuthzStrictMiddleware enforces role-based access per operation. It runs at the
// oapi-codegen strict-handler layer (after routing), so it receives the operationId
// directly and needs no route-pattern matching. AuthnMiddleware runs earlier in the
// chi chain and populates roles in the context.
//
// On deny it writes the response and returns (nil, nil); the strict wrapper treats a
// nil response with no error as "already handled" and writes nothing further.
func (m *Middleware) AuthzStrictMiddleware(f generated.StrictHandlerFunc, operationID string) generated.StrictHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		requiredRoles, known := operationRoles[operationID]
		if !known {
			m.logger.Error("no authz policy for operation", "operationID", operationID)
			http.Error(w, "forbidden", http.StatusForbidden)
			return nil, nil
		}

		// public operation
		if len(requiredRoles) == 0 {
			return f(ctx, w, r, request)
		}

		roles, err := GetRoles(ctx)
		if err != nil {
			m.logger.Error("failed to get roles from context", "operationID", operationID, "err", err)
			http.Error(w, "forbidden", http.StatusForbidden)
			return nil, nil
		}

		for _, role := range roles {
			if slices.Contains(requiredRoles, role) {
				return f(ctx, w, r, request)
			}
		}

		http.Error(w, "forbidden", http.StatusForbidden)
		return nil, nil
	}
}
