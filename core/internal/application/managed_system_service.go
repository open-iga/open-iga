package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/open-iga/core/internal/contract"
	"github.com/open-iga/core/internal/domain"
)

type ManagedSystemService struct {
	logger                  *slog.Logger
	connectorService        *ConnectorService
	managedSystemRepository contract.ManagedSystemRepository
}

var _ contract.ManagedSystemService = (*ManagedSystemService)(nil)

func NewManagedSystemService(logger *slog.Logger, connectorService *ConnectorService, managedSystemRepository contract.ManagedSystemRepository) *ManagedSystemService {
	return &ManagedSystemService{logger: logger, connectorService: connectorService, managedSystemRepository: managedSystemRepository}
}

func (m *ManagedSystemService) Onboard(ctx context.Context, connectorOnboardId string) (*domain.ManagedSystem, error) {
	connectorDetails, ok := m.connectorService.GetOnboardedConnectorDetails(ctx, connectorOnboardId)
	if !ok {
		return nil, fmt.Errorf("connector '%s': %w", connectorOnboardId, domain.ErrConnectorNotFound)
	}

	if !connectorDetails.CanBeOnboarded() {
		return nil, fmt.Errorf("connector '%s' in state %s: %w", connectorOnboardId, connectorDetails.Status, domain.ErrConnectorNotOnboardable)
	}

	managedSystem, err := m.managedSystemRepository.CreateManagedSystem(ctx, connectorDetails.ConnectorSpec.Name, connectorDetails.ConnectorUrl, connectorDetails.ConnectorHash)
	if err != nil {
		m.logger.Warn("failed to create managed system", "connector", connectorDetails.ConnectorSpec.Name)
		return nil, fmt.Errorf("failed to create managed system: %w", err)
	}

	return managedSystem, nil
}
