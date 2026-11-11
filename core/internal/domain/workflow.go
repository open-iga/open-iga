package domain

import (
	"errors"
	"fmt"
	"slices"
)

var ErrInvalidWorkflow = errors.New("invalid workflow")

// WorkflowStep is one step of a workflow: a call to a function by name with static arguments.
type WorkflowStep struct {
	// Name is unique within the workflow; it is also what continues a gate.
	Name     string            `json:"name"`
	Function string            `json:"function"`
	Args     map[string]string `json:"args,omitempty"`
	// Gate is a prerequisite: the run waits before this step until someone continues it manually (from the UI).
	Gate bool `json:"gate,omitempty"`
}

// WorkflowDefinition is a workflow: an ordered list of steps.
type WorkflowDefinition struct {
	Name  string         `json:"name"`
	Steps []WorkflowStep `json:"steps"`
}

// WorkflowRun is one run of a definition. Input holds what the run is about (e.g. the identity id) and is passed to every function
type WorkflowRun struct {
	Definition WorkflowDefinition `json:"definition"`
	Input      map[string]string  `json:"input"`
}

func (d WorkflowDefinition) Validate() error {
	if d.Name == "" {
		return fmt.Errorf("%w: workflow has no name", ErrInvalidWorkflow)
	}

	if len(d.Steps) == 0 {
		return fmt.Errorf("%w: workflow %q has no steps", ErrInvalidWorkflow, d.Name)
	}

	existingStepNames := make([]string, len(d.Steps))
	for _, step := range d.Steps {
		if step.Name == "" {
			return fmt.Errorf("%w: step has no name", ErrInvalidWorkflow)
		}

		if slices.Contains(existingStepNames, step.Name) {
			return fmt.Errorf("%w: step name %q is used twice", ErrInvalidWorkflow, step.Name)
		}
		existingStepNames = append(existingStepNames, step.Name)

		if step.Function == "" {
			return fmt.Errorf("%w: step %q has no function", ErrInvalidWorkflow, step.Name)
		}
	}
	return nil
}
