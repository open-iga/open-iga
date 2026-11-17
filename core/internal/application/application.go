package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/contract"
)

func NewApplication(ctx context.Context, appConfig *common.AppConfig, logger *slog.Logger, remotes *contract.RuntimeRemote, repository *contract.Repository) (*contract.RuntimeApplication, error) {
	connectorService := NewConnectorService(remotes.ConnectorRuntime, repository.ManagedSystemRepository, repository.IdentityRepository, repository.AccountRepository, repository.EntitlementCatalogRepository, logger)
	workflowService := NewWorkflowService(logger, remotes.WorkflowEngine, connectorService, repository.WorkflowRepository, repository.ManagedSystemRepository, repository.EntitlementCatalogRepository, repository.IdentityRepository, repository.AccountRepository)

	application := &contract.RuntimeApplication{
		AuthService:          NewAuthService(remotes.Oauth2Clients, logger, repository.SessionRepository, repository.IdentityRepository),
		ConnectorService:     connectorService,
		DiscoveryService:     NewDiscoveryService(connectorService, repository.ManagedSystemRepository, repository.EntitlementCatalogRepository),
		ManagedSystemService: NewManagedSystemService(logger, connectorService, repository.ManagedSystemRepository),
		WorkflowService:      workflowService,
	}

	if err := application.AuthService.EnsureAdmin(ctx, appConfig.AdminUser.Email); err != nil {
		return nil, fmt.Errorf("ensure admin: %w", err)
	}

	// TODO: for now the server will not start if the required env variable is missing; one better option is to start and prevent executing. For now stopping is preferred as this is a simpler solution
	if err := application.ConnectorService.LoadConnectorsOnStartup(ctx); err != nil {
		return nil, fmt.Errorf("load connectors: %w", err)
	}

	// Launching resumes the runs left from before a restart, and they need the connectors loaded above.
	// It runs in the background so the server starts at once; workflow runs are refused until it is done.
	go workflowService.LaunchEngine(ctx)

	return application, nil
}
