package domain

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/google/uuid"
)

var ErrInvalidWorkflow = errors.New("invalid workflow")

// WorkflowType says what a workflow does, and so which action each kind of step runs. Names are free, the type is not.
type WorkflowType string

const WorkflowTypeOnboarding WorkflowType = "onboarding"

// workflowActions is the action a workflow of a type runs per kind of operation. Offboarding would delete accounts and revoke entitlements.
var workflowActions = map[WorkflowType]map[ConnectorKind]string{
	WorkflowTypeOnboarding: {AccountActionKind: "create", EntitlementKind: "grant"},
}

// WorkflowOperation is the connector call a step makes: an operation on a resource of a managed system.
// Everything else the call needs is derived: the identity from the run, the account from the identity and the tenant.
type WorkflowOperation struct {
	System   string        `json:"system"`
	Resource string        `json:"resource"`
	Kind     ConnectorKind `json:"kind"`
	// Action is not sent by the client: saving the workflow sets it from the workflow type and the kind.
	Action string `json:"action,omitempty"`
	// Tenant is the tenant's name; needed only when the system has several tenants and no entitlement tells which.
	Tenant string `json:"tenant,omitempty"`
	// Entitlement is the entitlement to grant or revoke; the only extra an operation can need.
	Entitlement uuid.UUID `json:"entitlement,omitzero"`
}

func (o WorkflowOperation) Validate(actions map[ConnectorKind]string) error {
	if o.System == "" || o.Resource == "" {
		return fmt.Errorf("%w: operation needs a system and a resource", ErrInvalidWorkflow)
	}

	if _, ok := actions[o.Kind]; !ok {
		return fmt.Errorf("%w: kind %q is not one this workflow runs", ErrInvalidWorkflow, o.Kind)
	}

	needsEntitlement := o.Kind == EntitlementKind
	if needsEntitlement && o.Entitlement == uuid.Nil {
		return fmt.Errorf("%w: %s needs an entitlement", ErrInvalidWorkflow, o.Kind)
	}

	if !needsEntitlement && o.Entitlement != uuid.Nil {
		return fmt.Errorf("%w: %s does not take an entitlement", ErrInvalidWorkflow, o.Kind)
	}

	return nil
}

// WorkflowStep is one step of a workflow: a connector operation.
type WorkflowStep struct {
	// Name is unique within the workflow; it is also what continues a gate.
	Name      string            `json:"name"`
	Operation WorkflowOperation `json:"operation"`
	// After lists the steps that must finish first. A step with none starts at once, so independent steps run in parallel.
	After []string `json:"after,omitempty"`
}

// WorkflowDefinition is a workflow: an ordered list of steps.
type WorkflowDefinition struct {
	Name  string         `json:"name"`
	Type  WorkflowType   `json:"type"`
	Steps []WorkflowStep `json:"steps"`
}

// SetActions sets the action of every step from the workflow type and the kind, in place.
func (d *WorkflowDefinition) SetActions() {
	for index := range d.Steps {
		d.Steps[index].Operation.Action = workflowActions[d.Type][d.Steps[index].Operation.Kind]
	}
}

func (d *WorkflowDefinition) Validate() error {
	if err := validateName("workflow", d.Name); err != nil {
		return err
	}

	actions, ok := workflowActions[d.Type]
	if !ok {
		return fmt.Errorf("%w: workflow %q has type %q, use one of %v", ErrInvalidWorkflow, d.Name, d.Type, slices.Sorted(maps.Keys(workflowActions)))
	}

	if len(d.Steps) == 0 {
		return fmt.Errorf("%w: workflow %q has no steps", ErrInvalidWorkflow, d.Name)
	}

	existingStepNames := make([]string, 0, len(d.Steps))
	for _, step := range d.Steps {
		if err := validateName("step", step.Name); err != nil {
			return err
		}

		if slices.Contains(existingStepNames, step.Name) {
			return fmt.Errorf("%w: step name %q is used twice", ErrInvalidWorkflow, step.Name)
		}
		existingStepNames = append(existingStepNames, step.Name)

		for _, previous := range step.After {
			if !slices.Contains(existingStepNames[:len(existingStepNames)-1], previous) {
				return fmt.Errorf("%w: step %q runs after %q, which is not an earlier step", ErrInvalidWorkflow, step.Name, previous)
			}
		}

		if err := step.Operation.Validate(actions); err != nil {
			return fmt.Errorf("step %q: %w", step.Name, err)
		}
	}

	return nil
}

// validateName rejects an empty name and one with a "/": names are path parts of the API, and the engine joins them with "/".
func validateName(kind string, name string) error {
	if name == "" {
		return fmt.Errorf("%w: %s has no name", ErrInvalidWorkflow, kind)
	}

	if strings.Contains(name, "/") {
		return fmt.Errorf("%w: %s name %q has a /", ErrInvalidWorkflow, kind, name)
	}

	return nil
}

// WorkflowRun is one run of a definition, about an identity.
type WorkflowRun struct {
	Definition WorkflowDefinition `json:"definition"`
	IdentityId uuid.UUID          `json:"identityId"`
}

// Validate also checks the run has an identity: every step works on one.
func (r WorkflowRun) Validate() error {
	if err := r.Definition.Validate(); err != nil {
		return err
	}

	if r.IdentityId == uuid.Nil {
		return fmt.Errorf("%w: the run has no identity", ErrInvalidWorkflow)
	}

	return nil
}
