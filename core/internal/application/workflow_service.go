package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/open-iga/core/internal/contract"
	"github.com/open-iga/core/internal/domain"
)

const engineLaunchRetryInterval = 10 * time.Second

type WorkflowService struct {
	logger                       *slog.Logger
	workflowEngine               contract.WorkflowEngine
	connectorService             contract.ConnectorService
	discoveryService             contract.DiscoveryService
	workflowRepository           contract.WorkflowRepository
	managedSystemRepository      contract.ManagedSystemRepository
	entitlementCatalogRepository contract.EntitlementCatalogRepository
	identityRepository           contract.IdentityRepository
	accountRepository            contract.AccountRepository
	// engineReady is closed once the engine has launched; the calls that need the engine fail until then.
	engineReady chan struct{}
	// engineLaunchRetryInterval is the wait before launching the engine again after a failed launch.
	engineLaunchRetryInterval time.Duration
}

var _ contract.WorkflowService = (*WorkflowService)(nil)

func NewWorkflowService(logger *slog.Logger, workflowEngine contract.WorkflowEngine, connectorService contract.ConnectorService, discoveryService contract.DiscoveryService, workflowRepository contract.WorkflowRepository, managedSystemRepository contract.ManagedSystemRepository, entitlementCatalogRepository contract.EntitlementCatalogRepository, identityRepository contract.IdentityRepository, accountRepository contract.AccountRepository) *WorkflowService {
	return &WorkflowService{
		logger,
		workflowEngine,
		connectorService,
		discoveryService,
		workflowRepository,
		managedSystemRepository,
		entitlementCatalogRepository,
		identityRepository,
		accountRepository,
		make(chan struct{}),
		engineLaunchRetryInterval,
	}
}

// LaunchEngine launches the workflow engine, trying again until it works, then opens the engine calls.
// Run it in the background so the server need not wait.
func (w *WorkflowService) LaunchEngine(ctx context.Context) {
	for {
		err := w.workflowEngine.Launch(ctx, w.ExecuteConnectorOperation, domain.QueuedWorkflowTypes())
		if err == nil {
			break
		}

		w.logger.Error("failed to launch the workflow engine, workflow runs stay unavailable", "err", err, "retryIn", w.engineLaunchRetryInterval)
		select {
		case <-ctx.Done():
			return
		case <-time.After(w.engineLaunchRetryInterval):
		}
	}

	close(w.engineReady)
	w.logger.Info("workflow engine launched")
}

func (w *WorkflowService) checkEngineReady() error {
	select {
	case <-w.engineReady:
		return nil
	default:
		return domain.ErrWorkflowEngineNotReady
	}
}

func (w *WorkflowService) CreateWorkflow(ctx context.Context, definition domain.WorkflowDefinition) error {
	if err := definition.Validate(); err != nil {
		return fmt.Errorf("create workflow: %w", err)
	}

	definition.SetActions()

	// A wrong step fails when the admin saves it, not halfway through someone's run.
	// An identity has one account per system and tenant, so two create steps on the same one would fight over it.
	createdOn := map[string]string{}
	for _, step := range definition.Steps {
		managedSystem, ref, err := w.resolveOperation(ctx, step.Operation)
		if err != nil {
			return fmt.Errorf("create workflow: step %q: %w", step.Name, err)
		}

		if step.Operation.Kind == domain.AccountActionKind {
			accountKey := managedSystem.Id + "/" + ref.TenantId.String()

			if firstStep, taken := createdOn[accountKey]; taken {
				return fmt.Errorf("%w: steps %q and %q both create an account on system %q", domain.ErrInvalidWorkflow, firstStep, step.Name, step.Operation.System)
			}

			createdOn[accountKey] = step.Name
		}
	}

	if err := w.workflowRepository.CreateWorkflow(ctx, definition); err != nil {
		return fmt.Errorf("create workflow: %w", err)
	}

	return nil
}

func (w *WorkflowService) GetWorkflow(ctx context.Context, name string) (domain.WorkflowDefinition, error) {
	definition, err := w.workflowRepository.GetWorkflow(ctx, name)
	if err != nil {
		return domain.WorkflowDefinition{}, fmt.Errorf("get workflow: %w", err)
	}

	return definition, nil
}

func (w *WorkflowService) StartWorkflow(ctx context.Context, name string, identityId uuid.UUID) (string, error) {
	if err := w.checkEngineReady(); err != nil {
		return "", fmt.Errorf("start workflow: %w", err)
	}

	definition, err := w.workflowRepository.GetWorkflow(ctx, name)
	if err != nil {
		return "", fmt.Errorf("start workflow: %w", err)
	}

	if _, err := w.identityRepository.GetIdentityById(ctx, identityId); err != nil {
		return "", fmt.Errorf("start workflow: %w", err)
	}

	runId := uuid.NewString()
	if err := w.workflowEngine.Start(ctx, runId, domain.WorkflowRun{Definition: definition, IdentityId: identityId}); err != nil {
		return "", fmt.Errorf("start workflow: %w", err)
	}

	return runId, nil
}

func (w *WorkflowService) StartDiscovery(ctx context.Context) (string, error) {
	if err := w.checkEngineReady(); err != nil {
		return "", fmt.Errorf("start discovery: %w", err)
	}

	targets, err := w.discoveryService.ListDiscoveryTargets(ctx)
	if err != nil {
		return "", fmt.Errorf("start discovery: %w", err)
	}

	if len(targets) == 0 {
		return "", nil
	}

	definition := buildDiscoveryDefinition(targets)
	runId := uuid.NewString()
	if err := w.workflowEngine.Start(ctx, runId, domain.WorkflowRun{Definition: definition, IdentityId: uuid.Nil}); err != nil {
		return "", fmt.Errorf("start discovery: %w", err)
	}

	return runId, nil
}

// buildDiscoveryDefinition makes one step per target; the steps depend on none, so they start together.
func buildDiscoveryDefinition(targets []domain.DiscoveryTarget) domain.WorkflowDefinition {
	definition := domain.WorkflowDefinition{Name: "discovery", Type: domain.WorkflowTypeDiscovery, Steps: make([]domain.WorkflowStep, 0, len(targets))}
	for _, target := range targets {
		operation := domain.WorkflowOperation{System: target.SystemName, Resource: target.Resource, Kind: domain.EntitlementKind, Action: "", Tenant: target.TenantName, Entitlement: uuid.Nil}
		definition.Steps = append(definition.Steps, domain.WorkflowStep{Name: buildDiscoveryStepName(target), Operation: operation, After: nil, Gate: false})
	}
	definition.SetActions()

	return definition
}

// buildDiscoveryStepName is unique per target, as it is made of ids and the resource.
func buildDiscoveryStepName(target domain.DiscoveryTarget) string {
	return fmt.Sprintf("%s-%s-%s", target.SystemId, target.TenantId, target.Resource)
}

func (w *WorkflowService) Continue(ctx context.Context, workflowName string, runId string, stepName string) error {
	if _, err := w.GetRun(ctx, workflowName, runId); err != nil {
		return err
	}

	if err := w.workflowEngine.Continue(ctx, runId, stepName); err != nil {
		return fmt.Errorf("continue workflow: %w", err)
	}

	return nil
}

// GetRun treats a run of another workflow as not found.
func (w *WorkflowService) GetRun(ctx context.Context, workflowName string, runId string) (domain.WorkflowRunState, error) {
	if err := w.checkEngineReady(); err != nil {
		return domain.WorkflowRunState{}, fmt.Errorf("get workflow run: %w", err)
	}

	state, err := w.workflowEngine.Get(ctx, runId)
	if err != nil {
		return domain.WorkflowRunState{}, fmt.Errorf("get workflow run: %w", err)
	}

	if state.Name != workflowName {
		return domain.WorkflowRunState{}, fmt.Errorf("%w: %s", domain.ErrWorkflowRunNotFound, runId)
	}

	return state, nil
}
