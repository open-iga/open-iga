package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/contract"
)

func NewApplication(ctx context.Context, appConfig *common.AppConfig, logger *slog.Logger, remotes *contract.RuntimeRemote, repository *contract.Repository) (*contract.RuntimeApplication, error) {
	connectorService := NewConnectorService(remotes.ConnectorRuntime, repository.ManagedSystemRepository, logger)

	application := &contract.RuntimeApplication{
		AuthService:          NewAuthService(remotes.Oauth2Clients, logger, repository.SessionRepository, repository.IdentityRepository),
		ConnectorService:     connectorService,
		ManagedSystemService: NewManagedSystemService(logger, connectorService, repository.ManagedSystemRepository),
	}

	if err := application.AuthService.EnsureAdmin(ctx, appConfig.AdminUser.Email); err != nil {
		return nil, fmt.Errorf("ensure admin: %w", err)
	}

	// TODO: for now the server will not start if the required env variable is missing; one better option is to start and prevent executing. For now stopping is preferred as this is a simpler solution
	if err := application.ConnectorService.LoadConnectorsOnStartup(ctx); err != nil {
		return nil, fmt.Errorf("load connectors: %w", err)
	}

	return application, nil
}
