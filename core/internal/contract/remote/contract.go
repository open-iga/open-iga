package remote

import (
	"github.com/open-iga/core/internal/application/adapter"
)

type Provider string

const (
	Google Provider = "google"
)

type Oauth2Clients map[Provider]adapter.Client

type RuntimeRemote struct {
	Oauth2Clients
}
