package workflow_engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/dbos-inc/dbos-transact-golang/dbos"
	"github.com/open-iga/core/internal/contract"
	"github.com/open-iga/core/internal/domain"
)

const (
	appName = "open-iga"
	// applicationVersion changes only when a code change would break the replay of a run in flight.
	// The default is a hash of the binary, which would orphan every run waiting at a gate on each deploy.
	applicationVersion = "1"
	// gateTimeout is how long a step waits at a gate before it fails; DBOS needs a timeout, so it is very long.
	gateTimeout = 30 * 24 * time.Hour
	// eventReadTimeout is how long reading a step's gate state waits for the event to exist.
	eventReadTimeout = 10 * time.Millisecond

	runWorkflowName  = "run"
	stepWorkflowName = "step"
	gateTopic        = "continue"
	gateEvent        = "gate"
	gateWaiting      = "waiting"
	gateContinued    = "continued"
	blockedPrefix    = "blocked by "

	// getRunTimeout stops a read that DBOS keeps retrying while the database is down.
	getRunTimeout = 10 * time.Second
)

// Engine runs workflow runs on DBOS, which keeps their state in the core's database (schema dbos).
// A run is one DBOS workflow that starts one child workflow per step, so a gate or a failure holds only the steps after it.
type Engine struct {
	databaseUrl string
	logger      *slog.Logger
	runner      domain.WorkflowStepRunner
	dbosCtx     dbos.Context
}

var _ contract.WorkflowEngine = (*Engine)(nil)

func NewEngine(databaseUrl string, logger *slog.Logger) *Engine {
	return &Engine{databaseUrl: databaseUrl, logger: logger, runner: nil, dbosCtx: nil}
}

func (e *Engine) Launch(ctx context.Context, runner domain.WorkflowStepRunner) error {
	dbosCtx, err := dbos.NewContext(ctx, dbos.Config{
		AppName:            appName,
		DatabaseURL:        e.databaseUrl,
		ApplicationVersion: applicationVersion,
		Logger:             e.logger,
	})
	if err != nil {
		return fmt.Errorf("create dbos context: %w", err)
	}

	e.runner = runner
	dbos.RegisterWorkflow(dbosCtx, e.executeWorkflowRun, dbos.WithWorkflowName(runWorkflowName))
	dbos.RegisterWorkflow(dbosCtx, e.executeWorkflowStep, dbos.WithWorkflowName(stepWorkflowName))
	if err := dbos.Launch(dbosCtx); err != nil {
		return fmt.Errorf("launch dbos: %w", err)
	}

	e.dbosCtx = dbosCtx

	return nil
}

func (e *Engine) Start(_ context.Context, runId string, run domain.WorkflowRun) error {
	if err := run.Validate(); err != nil {
		return fmt.Errorf("start run %s: %w", runId, err)
	}

	if _, err := dbos.RunWorkflow(e.dbosCtx, e.executeWorkflowRun, run, dbos.WithWorkflowID(runId)); err != nil {
		return fmt.Errorf("start run %s: %w", runId, err)
	}

	return nil
}

func (e *Engine) Continue(_ context.Context, runId, stepName string) error {
	stepId := buildStepWorkflowId(runId, stepName)

	if e.getGateState(stepId) != gateWaiting {
		return fmt.Errorf("%w: %s", domain.ErrWorkflowStepNotGated, stepId)
	}

	// dbos.Recv in waitForGate receives this event and set the next stages for the gated step
	if err := dbos.Send(e.dbosCtx, stepId, gateContinued, gateTopic); err != nil {
		return fmt.Errorf("continue step %s: %w", stepId, err)
	}

	return nil
}

func (e *Engine) Get(_ context.Context, runId string) (domain.WorkflowRunState, error) {
	readCtx, cancel := dbos.WithTimeout(e.dbosCtx, getRunTimeout)
	defer cancel()

	runs, err := dbos.ListWorkflows(readCtx, dbos.WithFilterWorkflowIDs(runId), dbos.WithFilterLoadInput(true))
	if err != nil {
		return domain.WorkflowRunState{}, fmt.Errorf("get run %s: %w", runId, err)
	}

	if len(runs) == 0 {
		return domain.WorkflowRunState{}, fmt.Errorf("%w: %s", domain.ErrWorkflowRunNotFound, runId)
	}

	run, err := decodeInput[domain.WorkflowRun](runs[0].Input)
	if err != nil {
		return domain.WorkflowRunState{}, fmt.Errorf("decode run %s: %w", runId, err)
	}

	workflowSteps, err := dbos.ListWorkflows(readCtx, dbos.WithFilterWorkflowIDPrefix(runId+"/"))
	if err != nil {
		return domain.WorkflowRunState{}, fmt.Errorf("get steps of run %s: %w", runId, err)
	}

	// This additional step is to preserve ordering
	workflowStepById := make(map[string]dbos.WorkflowStatus, len(workflowSteps))
	for _, workflowStep := range workflowSteps {
		workflowStepById[workflowStep.ID] = workflowStep
	}

	workflowRunState := domain.WorkflowRunState{
		Id:    runId,
		Name:  run.Definition.Name,
		Steps: make([]domain.WorkflowStepState, 0, len(run.Definition.Steps)),
	}
	for _, step := range run.Definition.Steps {
		workflowStep := workflowStepById[buildStepWorkflowId(runId, step.Name)]
		workflowRunState.Steps = append(workflowRunState.Steps, e.toWorkflowStepState(workflowStep, step.Name))
	}

	return workflowRunState, nil
}

func (e *Engine) Shutdown(timeout time.Duration) {
	if e.dbosCtx == nil {
		return
	}

	if err := dbos.Shutdown(e.dbosCtx, timeout); err != nil {
		e.logger.Error("failed to shut down workflow engine", "err", err)
	}
}

func (e *Engine) toWorkflowStepState(workflowStatus dbos.WorkflowStatus, name string) domain.WorkflowStepState {
	workflowStepState := domain.WorkflowStepState{
		Name:      name,
		Status:    domain.WorkflowStepPending,
		Error:     "",
		CreatedAt: workflowStatus.CreatedAt,
		UpdatedAt: workflowStatus.UpdatedAt,
		StartedAt: workflowStatus.StartedAt,
	}

	switch workflowStatus.Status {
	case dbos.WorkflowStatusSuccess:
		workflowStepState.Status = domain.WorkflowStepSucceeded
	case dbos.WorkflowStatusError, dbos.WorkflowStatusCancelled, dbos.WorkflowStatusMaxRecoveryAttemptsExceeded:
		workflowStepState.Status = domain.WorkflowStepFailed
		if workflowStatus.Error != nil {
			workflowStepState.Error = workflowStatus.Error.Error()
		}

		if strings.HasPrefix(workflowStepState.Error, blockedPrefix) {
			workflowStepState.Status = domain.WorkflowStepBlocked
		}
	case dbos.WorkflowStatusPending, dbos.WorkflowStatusEnqueued, dbos.WorkflowStatusDelayed:
		if e.getGateState(workflowStatus.ID) == gateWaiting {
			workflowStepState.Status = domain.WorkflowStepGated
		}
	}

	return workflowStepState
}

// getGateState is empty when the step has no gate or has not reached it: DBOS reports a missing event as a timeout.
func (e *Engine) getGateState(stepId string) string {
	gate, err := dbos.GetEvent[string](e.dbosCtx, stepId, gateEvent, eventReadTimeout)
	if err != nil {
		return ""
	}

	return gate
}

func buildStepWorkflowId(runId, stepName string) string {
	return runId + "/" + stepName
}

// decodeInput reads a stored workflow input; DBOS returns it either as generic JSON or as a JSON string.
func decodeInput[T any](stored any) (T, error) {
	var input T
	raw, ok := stored.(string)
	if !ok {
		marshaled, err := json.Marshal(stored)
		if err != nil {
			return input, fmt.Errorf("marshal: %w", err)
		}
		raw = string(marshaled)
	}

	if err := json.Unmarshal([]byte(raw), &input); err != nil {
		return input, fmt.Errorf("unmarshal: %w", err)
	}

	return input, nil
}
