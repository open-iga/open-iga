package application

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/open-iga/core/internal/contract"
	"github.com/open-iga/core/internal/domain"
)

type ConnectorService struct {
	connectorRuntime        contract.ConnectorRuntime
	managedSystemRepository contract.ManagedSystemRepository
	logger                  *slog.Logger
	mu                      sync.RWMutex
	onboardingResults       map[string]domain.ConnectorValidationResult
}

const validationTimeout = 2 * time.Minute

var _ contract.ConnectorService = (*ConnectorService)(nil)

func NewConnectorService(connectorRuntime contract.ConnectorRuntime, managedSystemRepository contract.ManagedSystemRepository, logger *slog.Logger) *ConnectorService {
	return &ConnectorService{
		connectorRuntime:        connectorRuntime,
		managedSystemRepository: managedSystemRepository,
		logger:                  logger,
		onboardingResults:       make(map[string]domain.ConnectorValidationResult),
		mu:                      sync.RWMutex{}, // for exhaustruct lint
	}
}

// LoadConnectorsOnStartup lists onboarded connectors and validates each one against the stored
// hash, which also compiles and caches the wasm so later invocations skip the download+compile.
// A missing required env var aborts startup (returns an error) so the operator fixes the config
// before the server runs. A connector that has no spec, fails to fetch, or whose hash no longer
// matches is skipped with an error log.
func (c *ConnectorService) LoadConnectorsOnStartup(ctx context.Context) error {
	managedSystems, err := c.managedSystemRepository.ListManagedSystems(ctx)
	if err != nil {
		return fmt.Errorf("list managed systems: %w", err)
	}

	for _, managedSystem := range managedSystems {
		if managedSystem.ConnectorSpec == nil {
			c.logger.Error("skipping connector with no spec", "managedSystemId", managedSystem.Id)
			continue
		}

		if missing := managedSystem.ConnectorSpec.MissingRequiredEnvVars(); len(missing) > 0 {
			return fmt.Errorf("connector %q (%s) missing required env vars %v", managedSystem.Name, managedSystem.Id, missing)
		}

		if _, err := c.connectorRuntime.ValidateConnectorByURL(ctx, managedSystem.ConnectorUrl, managedSystem.ConnectorHash); err != nil {
			c.logger.Error("skipping connector that failed validation",
				"managedSystemId", managedSystem.Id, "name", managedSystem.Name, "err", err)
			continue
		}

		c.logger.Info("loaded connector", "managedSystemId", managedSystem.Id, "name", managedSystem.Name)
	}
	return nil
}

func (c *ConnectorService) ValidateByUrl(ctx context.Context, connectorUrl string, connectorHash string) string {
	c.mu.Lock()
	id := uuid.New().String()
	c.onboardingResults[id] = domain.ConnectorValidationResult{
		Error:          nil,
		ConnectorHash:  connectorHash,
		ConnectorUrl:   connectorUrl,
		Status:         domain.ConnectorValidationPending,
		ConnectorSpec:  nil,
		MissingEnvVars: nil,
	}
	c.mu.Unlock()

	go func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), validationTimeout)
		defer cancel()

		connectorSpec, err := c.connectorRuntime.ValidateConnectorByURL(ctx, connectorUrl, connectorHash)
		status := domain.ConnectorValidationSuccess
		if err != nil {
			c.logger.Error("Validation failed", "err", err)
			status = domain.ConnectorValidationFailed
		}

		c.logger.Debug("Validation completed", "connectorUrl", connectorUrl, "id", id, "status", status)
		c.mu.Lock()
		c.onboardingResults[id] = domain.ConnectorValidationResult{
			Error:          err,
			ConnectorHash:  connectorHash,
			ConnectorUrl:   connectorUrl,
			Status:         status,
			ConnectorSpec:  connectorSpec,
			MissingEnvVars: nil,
		}
		c.mu.Unlock()
	}()

	return id
}

// Execute routes an invocation to the connector runtime, which runs it against the cached connector.
func (c *ConnectorService) Execute(ctx context.Context, connectorUrl string, connectorHash string, envelope domain.ConnectorDispatchEnvelope) ([]byte, error) {
	return c.connectorRuntime.Execute(ctx, connectorUrl, connectorHash, envelope)
}

func (c *ConnectorService) GetOnboardedConnectorDetails(_ context.Context, onboardingId string) (*domain.ConnectorValidationResult, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, ok := c.onboardingResults[onboardingId]
	if !ok {
		return nil, false
	}
	// Computed on read: the core's environment can change between validation and onboarding.
	if value.ConnectorSpec != nil {
		value.MissingEnvVars = value.ConnectorSpec.MissingRequiredEnvVars()
	}
	return &value, true
}
