package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/open-iga/core/internal/contract"
	"github.com/open-iga/core/internal/domain"
)

type WorkflowService struct {
	logger                       *slog.Logger
	workflowEngine               contract.WorkflowEngine
	connectorService             contract.ConnectorService
	workflowRepository           contract.WorkflowRepository
	managedSystemRepository      contract.ManagedSystemRepository
	entitlementCatalogRepository contract.EntitlementCatalogRepository
	identityRepository           contract.IdentityRepository
	accountRepository            contract.AccountRepository
}

var _ contract.WorkflowService = (*WorkflowService)(nil)

func NewWorkflowService(logger *slog.Logger, workflowEngine contract.WorkflowEngine, connectorService contract.ConnectorService, workflowRepository contract.WorkflowRepository, managedSystemRepository contract.ManagedSystemRepository, entitlementCatalogRepository contract.EntitlementCatalogRepository, identityRepository contract.IdentityRepository, accountRepository contract.AccountRepository) *WorkflowService {
	return &WorkflowService{
		logger,
		workflowEngine,
		connectorService,
		workflowRepository,
		managedSystemRepository,
		entitlementCatalogRepository,
		identityRepository,
		accountRepository,
	}
}

func (w *WorkflowService) CreateWorkflow(ctx context.Context, definition domain.WorkflowDefinition) error {
	if err := definition.Validate(); err != nil {
		return fmt.Errorf("create workflow: %w", err)
	}

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

// GetRun treats a run of another workflow as not found.
func (w *WorkflowService) GetRun(ctx context.Context, workflowName string, runId string) (domain.WorkflowRunState, error) {
	state, err := w.workflowEngine.Get(ctx, runId)
	if err != nil {
		return domain.WorkflowRunState{}, fmt.Errorf("get workflow run: %w", err)
	}

	if state.Name != workflowName {
		return domain.WorkflowRunState{}, fmt.Errorf("%w: %s", domain.ErrWorkflowRunNotFound, runId)
	}

	return state, nil
}
