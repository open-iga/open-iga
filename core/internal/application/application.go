package application

import (
	"log/slog"

	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/contract"
)

func NewApplication(_ *common.AppConfig, logger *slog.Logger, remotes *contract.RuntimeRemote, repository *contract.Repository) *contract.RuntimeApplication {
	return &contract.RuntimeApplication{
		AuthService: NewAuthService(remotes.Oauth2Clients, logger, repository.SessionRepository, repository.IdentityRepository),
	}
}
