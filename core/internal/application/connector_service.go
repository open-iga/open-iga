package application

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/open-iga/core/internal/contract"
	"github.com/open-iga/core/internal/domain"
)

type ConnectorService struct {
	connectorRuntime  contract.ConnectorRuntime
	logger            *slog.Logger
	mu                sync.RWMutex
	onboardingResults map[string]domain.ConnectorValidationResult
}

const validationTimeout = 2 * time.Minute

var _ contract.ConnectorService = (*ConnectorService)(nil)

func NewConnectorService(connectorRuntime contract.ConnectorRuntime, logger *slog.Logger) *ConnectorService {
	return &ConnectorService{
		connectorRuntime:  connectorRuntime,
		logger:            logger,
		onboardingResults: make(map[string]domain.ConnectorValidationResult),
	}
}

func (c *ConnectorService) ValidateByUrl(_ context.Context, connectorUrl string, connectorHash string) string {
	c.mu.Lock()
	id := uuid.New().String()
	c.onboardingResults[id] = domain.ConnectorValidationResult{Error: nil, Status: domain.ConnectorValidationPending, ConnectorSpec: nil}
	c.mu.Unlock()

	go func() {
		time.Sleep(10 * time.Second)

		ctx, cancel := context.WithTimeout(context.Background(), validationTimeout)
		defer cancel()

		connectorSpec, err := c.connectorRuntime.ValidateConnectorByURL(ctx, connectorUrl, connectorHash)
		status := domain.ConnectorValidationSuccess
		if err != nil {
			c.logger.Error("Validation failed", "err", err)
			status = domain.ConnectorValidationFailed
		}

		c.logger.Debug("Validation completed", "connectorUrl", connectorUrl, "id", id, "status", status)
		c.mu.Lock()
		c.onboardingResults[id] = domain.ConnectorValidationResult{Error: err, Status: status, ConnectorSpec: connectorSpec}
		c.mu.Unlock()
	}()

	return id
}

func (c *ConnectorService) GetOnboardedConnectorDetails(_ context.Context, onboardingId string) (*domain.ConnectorValidationResult, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, ok := c.onboardingResults[onboardingId]
	if !ok {
		return nil, false
	}
	return &value, true
}
