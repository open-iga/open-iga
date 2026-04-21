package application

import (
	"log/slog"

	"github.com/open-iga/core/internal/application/oauth"
	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/contract"
)

func NewApplication(appConfig *common.AppConfig, remotes *contract.RuntimeRemote, logger *slog.Logger) *contract.RuntimeApplication {
	return &contract.RuntimeApplication{
		LoginService: oauth.NewLoginService(appConfig, remotes, logger),
	}
}
