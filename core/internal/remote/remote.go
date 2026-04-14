package remote

import (
	"log/slog"

	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/contract/remote"
	"github.com/open-iga/core/internal/remote/oauth2_client"
)

// NewRemote creates runtime remotes that could be injected into the application layer
func NewRemote(appConfig *common.AppConfig, logger *slog.Logger) *remote.RuntimeRemote {
	return &remote.RuntimeRemote{
		Oauth2Clients: remote.Oauth2Clients{
			remote.Google: oauth2_client.NewGoogleOauth2Client(appConfig, logger),
		},
	}
}
