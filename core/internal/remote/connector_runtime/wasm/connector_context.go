package wasm

import (
	"context"

	"github.com/open-iga/core/internal/domain"
)

// allowedEndpointsKey carries the current operation's resolved endpoints on the call context, so the
// shared fetch host function can gate per invocation without per-connector mutable state.
type allowedEndpointsKey struct{}

// withAllowedEndpoints returns ctx carrying the endpoints the fetch host function may reach.
func withAllowedEndpoints(ctx context.Context, endpoints []domain.ResolvedEndpoint) context.Context {
	return context.WithValue(ctx, allowedEndpointsKey{}, endpoints)
}

// allowedEndpointsFrom returns the endpoints set by withAllowedEndpoints, or nil (egress blocked) if none were set
func allowedEndpointsFrom(ctx context.Context) []domain.ResolvedEndpoint {
	endpoints, _ := ctx.Value(allowedEndpointsKey{}).([]domain.ResolvedEndpoint)
	return endpoints
}
