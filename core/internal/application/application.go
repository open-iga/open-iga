package application

import (
	"log/slog"

	"github.com/open-iga/core/internal/application/oauth"
	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/contract/application"
	"github.com/open-iga/core/internal/contract/remote"
)

func NewApplication(appConfig *common.AppConfig, remotes *remote.RuntimeRemote, logger *slog.Logger) *application.RuntimeApplication {
	return &application.RuntimeApplication{
		LoginService: oauth.NewLoginService(appConfig, remotes, logger),
	}
}
