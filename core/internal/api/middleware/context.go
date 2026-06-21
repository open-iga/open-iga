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
