package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrWorkflowEngineNotReady = errors.New("workflow engine is not ready")
	ErrWorkflowRunNotFound    = errors.New("workflow run not found")
	ErrWorkflowStepNotGated   = errors.New("workflow step is not waiting at a gate")
	ErrWorkflowNotFound       = errors.New("workflow not found")
)

// WorkflowStepRunner runs one step of a run about the given identity; the engine calls it, it knows nothing of what a step does.
type WorkflowStepRunner func(ctx context.Context, step WorkflowStep, identityId uuid.UUID) error

type WorkflowStepStatus string

const (
	// WorkflowStepPending covers waiting for the steps it follows and running.
	WorkflowStepPending WorkflowStepStatus = "pending"
	// WorkflowStepGated waits for someone to continue it manually.
	WorkflowStepGated     WorkflowStepStatus = "gated"
	WorkflowStepSucceeded WorkflowStepStatus = "succeeded"
	WorkflowStepFailed    WorkflowStepStatus = "failed"
	// WorkflowStepBlocked never ran because a step it follows did not succeed.
	WorkflowStepBlocked WorkflowStepStatus = "blocked"
)

type WorkflowStepState struct {
	Name   string             `json:"name"`
	Status WorkflowStepStatus `json:"status"`
	Error  string             `json:"error,omitempty"`
	// The times come from the step's workflow; StartedAt is empty until the step starts.
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	StartedAt time.Time `json:"startedAt,omitzero"`
}

// WorkflowRunState is where a run stands, step by step in definition order.
type WorkflowRunState struct {
	Id    string
	Name  string // name of the workflow the run belongs to
	Steps []WorkflowStepState
}
