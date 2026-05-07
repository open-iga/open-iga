package middleware

import (
	"context"

	"github.com/open-iga/core/internal/domain"
)

const (
	IdentityContextKey = "identity"
	SessionContextKey  = "session"
)

func SetIdentityInContent(ctx context.Context, identity *domain.Identity) context.Context {
	return context.WithValue(ctx, IdentityContextKey, identity)
}

func SetSessionInContent(ctx context.Context, session *domain.Session) context.Context {
	return context.WithValue(ctx, SessionContextKey, session)
}
