package application

import (
	"context"
	"encoding/json"
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
	identityRepository      contract.IdentityRepository
	accountRepository       contract.AccountRepository
	entitlementRepository   contract.EntitlementRepository
	logger                  *slog.Logger
	mu                      sync.RWMutex
	onboardingResults       map[string]domain.ConnectorValidationResult
}

const validationTimeout = 2 * time.Minute

var _ contract.ConnectorService = (*ConnectorService)(nil)

func NewConnectorService(connectorRuntime contract.ConnectorRuntime, managedSystemRepository contract.ManagedSystemRepository, identityRepository contract.IdentityRepository, accountRepository contract.AccountRepository, entitlementRepository contract.EntitlementRepository, logger *slog.Logger) *ConnectorService {
	return &ConnectorService{
		connectorRuntime:        connectorRuntime,
		managedSystemRepository: managedSystemRepository,
		identityRepository:      identityRepository,
		accountRepository:       accountRepository,
		entitlementRepository:   entitlementRepository,
		logger:                  logger,
		onboardingResults:       make(map[string]domain.ConnectorValidationResult),
		mu:                      sync.RWMutex{}, // for exhaustruct lint
	}
}

// LoadConnectorsOnStartup lists onboarded connectors and compiles each into the runtime cache
// (resolving its endpoints from the environment). A missing required env var aborts startup
// (returns an error) so the operator fixes the config before the server runs. A connector with
// no spec is skipped with an error log.
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

		c.CompileConnector(ctx, managedSystem.ConnectorUrl, managedSystem.ConnectorHash, managedSystem.ConnectorSpec)
		c.logger.Info("loading connector", "managedSystemId", managedSystem.Id, "name", managedSystem.Name)
	}
	return nil
}

func (c *ConnectorService) ValidateByUrl(ctx context.Context, connectorUrl string, connectorHash string) string {
	c.mu.Lock()
	id := uuid.New().String()
	c.onboardingResults[id] = domain.ConnectorValidationResult{
		Error:         nil,
		ConnectorHash: connectorHash,
		ConnectorUrl:  connectorUrl,
		Status:        domain.ConnectorValidationPending,
		ConnectorSpec: nil,
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
			Error:         err,
			ConnectorHash: connectorHash,
			ConnectorUrl:  connectorUrl,
			Status:        status,
			ConnectorSpec: connectorSpec,
		}
		c.mu.Unlock()
	}()

	return id
}

// CompileConnector compiles the connector into the runtime cache (which resolves its endpoints
// from the environment). Called post-approval (onboarding) and at startup.
func (c *ConnectorService) CompileConnector(ctx context.Context, connectorUrl string, connectorHash string, spec *domain.ConnectorSpec) {
	c.connectorRuntime.CompileConnector(ctx, connectorUrl, connectorHash, spec)
}

// Execute looks up the connector by name, builds the typed input of the kind and action only from rows loaded from the DB, runs it on the connector and returns the typed output.
// TODO: save the call on an approval request and run it in the background once approved.
func (c *ConnectorService) Execute(ctx context.Context, connectorName string, managedResource string, kind domain.ConnectorKind, action string, ref domain.OperationRef) (domain.ConnectorOutput, error) {
	managedSystem, err := c.managedSystemRepository.GetManagedSystemByName(ctx, connectorName)
	if err != nil {
		return nil, fmt.Errorf("execute connector %s: %w", connectorName, err)
	}
	data, err := c.loadOperationData(ctx, managedSystem, kind, action, ref)
	if err != nil {
		return nil, fmt.Errorf("execute connector %s: %w", connectorName, err)
	}

	operation, err := domain.NewConnectorOperation(kind, action, *data)
	if err != nil {
		return nil, fmt.Errorf("execute connector %s: %w", connectorName, err)
	}

	input, err := json.Marshal(operation)
	if err != nil {
		return nil, fmt.Errorf("execute connector %s: marshal input: %w", connectorName, err)
	}

	raw, err := c.connectorRuntime.Execute(ctx, managedSystem.ConnectorUrl, managedSystem.ConnectorHash, domain.ConnectorInvocation{
		Kind:            operation.Kind(),
		ManagedResource: managedResource,
		Type:            operation.Type(),
		Input:           input,
	})
	if err != nil {
		return nil, fmt.Errorf("execute connector %s: %w", connectorName, err)
	}

	output, err := operation.DecodeOutput(raw)
	if err != nil {
		return nil, fmt.Errorf("execute connector %s: %w", connectorName, err)
	}

	// TODO: save the result on the account and entitlement rows.
	c.logger.Info("connector executed", "connector", connectorName, "resource", managedResource, "action", action, "output", output)

	return output, nil
}

// loadOperationData loads the rows the ref points to. The tenant and entitlement must belong to the
// connector's managed system. The account is loaded only for operations that need an existing one.
func (c *ConnectorService) loadOperationData(ctx context.Context, managedSystem *domain.ManagedSystem, kind domain.ConnectorKind, action string, ref domain.OperationRef) (*domain.OperationData, error) {
	tenant, err := c.loadTenant(ctx, managedSystem, ref.TenantId)
	if err != nil {
		return nil, err
	}
	data := &domain.OperationData{Tenant: tenant, Identity: nil, Account: nil, Entitlement: nil}

	if ref.IdentityId != uuid.Nil {
		if data.Identity, err = c.identityRepository.GetIdentityById(ctx, ref.IdentityId); err != nil {
			return nil, fmt.Errorf("load identity: %w", err)
		}
	}

	if ref.NeedsAccount(kind, action) {
		if data.Account, err = c.accountRepository.GetAccountByTenantAndIdentity(ctx, ref.TenantId, ref.IdentityId); err != nil {
			return nil, fmt.Errorf("load account: %w", err)
		}
	}

	if ref.NeedsEntitlement(kind, action) {
		if data.Entitlement, err = c.entitlementRepository.GetEntitlement(ctx, ref.EntitlementId); err != nil {
			return nil, fmt.Errorf("load entitlement: %w", err)
		}

		if data.Entitlement.TenantId != tenant.Id {
			return nil, fmt.Errorf("%w: %s is not an entitlement of tenant %s", domain.ErrEntitlementNotFound, ref.EntitlementId, tenant.Id)
		}
	}

	return data, nil
}

// loadTenant loads the tenant and checks it belongs to the managed system.
// TODO: temporary, so the wasm can be tested before a tenant exists: no tenant id gives an empty tenant. Remove once tenants are always set.
func (c *ConnectorService) loadTenant(ctx context.Context, managedSystem *domain.ManagedSystem, tenantId uuid.UUID) (*domain.ManagedSystemTenant, error) {
	if tenantId == uuid.Nil {
		return new(domain.ManagedSystemTenant), nil
	}

	tenant, err := c.managedSystemRepository.GetManagedSystemTenant(ctx, tenantId)
	if err != nil {
		return nil, fmt.Errorf("load tenant: %w", err)
	}
	if tenant.ManagedSystemId != managedSystem.Id {
		return nil, fmt.Errorf("%w: %s is not a tenant of %s", domain.ErrTenantNotFound, tenantId, managedSystem.Name)
	}

	return tenant, nil
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
