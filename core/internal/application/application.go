package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/contract"
)

func NewApplication(ctx context.Context, appConfig *common.AppConfig, logger *slog.Logger, remotes *contract.RuntimeRemote, repository *contract.Repository) (*contract.RuntimeApplication, error) {
	connectorService := NewConnectorService(remotes.ConnectorRuntime, logger)

	application := &contract.RuntimeApplication{
		AuthService:          NewAuthService(remotes.Oauth2Clients, logger, repository.SessionRepository, repository.IdentityRepository),
		ConnectorService:     connectorService,
		ManagedSystemService: NewManagedSystemService(logger, connectorService, repository.ManagedSystemRepository),
	}

	if err := application.AuthService.EnsureAdmin(ctx, appConfig.AdminUser.Email); err != nil {
		return nil, fmt.Errorf("ensure admin: %w", err)
	}

	return application, nil
}
