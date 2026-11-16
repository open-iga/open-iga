package workflow_engine

import (
	"context"
	"errors"
	"fmt"

	"github.com/dbos-inc/dbos-transact-golang/dbos"
	"github.com/google/uuid"
	"github.com/open-iga/core/internal/domain"
)

// stepInput is the input of a step workflow: one step of a run and who the run is about.
type stepInput struct {
	RunId      string              `json:"runId"`
	Step       domain.WorkflowStep `json:"step"`
	IdentityId uuid.UUID           `json:"identityId"`
}

// executeWorkflowRun starts every step at once and waits for all of them; each step waits for its own dependencies.
// A step only follows earlier steps (see WorkflowDefinition.Validate), so the ones it waits for already exist.
// TODO: for now workflow fans-out to all the steps; this is linear and with workflows x users x steps, number of queries to check the durable steps result in a lane increases; this is a problem to check when real users notices
func (e *Engine) executeWorkflowRun(ctx dbos.Context, run domain.WorkflowRun) (string, error) {
	workflowId, err := dbos.GetWorkflowID(ctx)
	if err != nil {
		return "", fmt.Errorf("get run id: %w", err)
	}

	handles := make([]dbos.WorkflowHandle[string], 0, len(run.Definition.Steps))
	for _, step := range run.Definition.Steps {
		handle, err := dbos.RunWorkflow(
			ctx,
			e.executeWorkflowStep,
			stepInput{RunId: workflowId, Step: step, IdentityId: run.IdentityId},
			dbos.WithWorkflowID(buildStepWorkflowId(workflowId, step.Name)),
		)
		if err != nil {
			return "", fmt.Errorf("start step %s: %w", step.Name, err)
		}

		handles = append(handles, handle)
	}

	var failures []error
	for _, handle := range handles {
		if _, err := handle.GetResult(); err != nil {
			failures = append(failures, err)
		}
	}

	return "", errors.Join(failures...)
}

// executeWorkflowStep waits for the steps it follows, then runs its step.
// It is a workflow of its own so a failure holds only the steps after it.
func (e *Engine) executeWorkflowStep(ctx dbos.Context, input stepInput) (string, error) {
	for _, previousStepName := range input.Step.After {
		previousWorkflowStepId := buildStepWorkflowId(input.RunId, previousStepName)
		handle, err := dbos.RetrieveWorkflow[string](ctx, previousWorkflowStepId)
		if err != nil {
			return "", fmt.Errorf("find step %s: %w", previousStepName, err)
		}

		if _, err := handle.GetResult(); err != nil {
			return "", fmt.Errorf("%s%s", blockedPrefix, previousStepName) //nolint:err113 - the text is what marks the step as blocked
		}
	}

	// A returned error is not retried. A crash can still replay the step, so connector operations must be idempotent.
	_, err := dbos.RunAsStep(ctx, func(stepCtx context.Context) (string, error) {
		return "", e.runner(stepCtx, input.Step, input.IdentityId)
	}, dbos.WithStepName(input.Step.Name))
	if err != nil {
		return "", fmt.Errorf("run step %s: %w", input.Step.Name, err)
	}

	return "", nil
}
