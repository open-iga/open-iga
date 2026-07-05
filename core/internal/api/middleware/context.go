package middleware

import (
	"context"
	"errors"

	"github.com/open-iga/core/internal/domain"
)

type contextKey string

const (
	IdentityContextKey contextKey = "identity"
	SessionContextKey  contextKey = "session"
	RolesContextKey    contextKey = "roles"
)

func WithIdentity(ctx context.Context, identity *domain.Identity) context.Context {
	return context.WithValue(ctx, IdentityContextKey, identity)
}

func GetIdentity(ctx context.Context) (*domain.Identity, error) {
	identity, ok := ctx.Value(IdentityContextKey).(*domain.Identity)

	if !ok {
		return nil, errors.New("identity missing in context")
	}
	return identity, nil
}

func WithSession(ctx context.Context, session *domain.Session) context.Context {
	return context.WithValue(ctx, SessionContextKey, session)
}

func GetSession(ctx context.Context) (*domain.Session, error) {
	session, ok := ctx.Value(SessionContextKey).(*domain.Session)
	if !ok {
		return nil, errors.New("session missing in context")
	}

	return session, nil
}

func WithRoles(ctx context.Context, roles []string) context.Context {
	return context.WithValue(ctx, RolesContextKey, roles)
}

func GetRoles(ctx context.Context) ([]string, error) {
	roles, ok := ctx.Value(RolesContextKey).([]string)
	if !ok {
		return nil, errors.New("roles missing in context")
	}
	return roles, nil
}
