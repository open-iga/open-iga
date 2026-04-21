package remote

import (
	"log/slog"

	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/contract"
	"github.com/open-iga/core/internal/remote/oauth2_client"
)

// NewRemote creates runtime remotes that could be injected into the application layer
func NewRemote(appConfig *common.AppConfig, logger *slog.Logger) *contract.RuntimeRemote {
	return &contract.RuntimeRemote{
		Oauth2Clients: contract.Oauth2Clients{
			contract.Google: oauth2_client.NewGoogleOauth2Client(appConfig, logger),
		},
	}
}
