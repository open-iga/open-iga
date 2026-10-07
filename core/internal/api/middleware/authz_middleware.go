package middleware

import (
	"context"
	"net/http"
	"slices"

	"github.com/open-iga/core/internal/api/generated"
	"github.com/open-iga/core/internal/domain"
)

// operationRoles maps operationId to allowed roles. nil = public; absent = denied.
var operationRoles = map[string][]string{
	"Health":       nil,
	"AuthDetails":  nil,
	"AuthCallback": nil,
	"Logout":       nil,

	"GetUserDetails":                       {domain.DefaultIdentityRole, domain.AdminRole},
	"OnboardConnector":                     {domain.AdminRole},
	"GetConnectorOnboardingRequestDetails": {domain.AdminRole, domain.DefaultIdentityRole},
	"OnboardManagedSystem":                 {domain.AdminRole},
	"ListManagedSystems":                   {domain.AdminRole, domain.DefaultIdentityRole},
}

// AuthzStrictMiddleware enforces per-operation roles at the strict-handler layer.
// On deny it writes the response and returns (nil, nil) to short-circuit.
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
