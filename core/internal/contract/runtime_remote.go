package contract

import (
	"context"

	"github.com/open-iga/core/internal/domain"
)

type Provider string

const (
	Google Provider = "google"
)

type Client interface {
	GetConsentPageDetails(ctx context.Context) (*ConsentPageDetails, error)
	FetchOauthUser(ctx context.Context, code string) (*domain.OauthUser, error)
}

type Oauth2Clients map[Provider]Client

type RuntimeRemote struct {
	Oauth2Clients
}
