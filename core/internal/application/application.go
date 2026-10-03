package application

import (
	"log/slog"

	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/contract"
)

func NewApplication(_ *common.AppConfig, logger *slog.Logger, remotes *contract.RuntimeRemote, repository *contract.Repository) *contract.RuntimeApplication {
	connectorService := NewConnectorService(remotes.ConnectorRuntime, logger)

	return &contract.RuntimeApplication{
		AuthService:          NewAuthService(remotes.Oauth2Clients, logger, repository.SessionRepository, repository.IdentityRepository),
		ConnectorService:     connectorService,
		ManagedSystemService: NewManagedSystemService(logger, connectorService, repository.ManagedSystemRepository),
	}
}
