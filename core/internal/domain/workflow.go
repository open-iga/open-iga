package domain

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
)

var ErrInvalidWorkflow = errors.New("invalid workflow")

// WorkflowOperation is the connector call a step makes: an operation on a resource of a managed system.
// Everything else the call needs is derived: the identity from the run, the account from the identity and the tenant.
type WorkflowOperation struct {
	System   string        `json:"system"`
	Resource string        `json:"resource"`
	Kind     ConnectorKind `json:"kind"`
	Action   string        `json:"action"`
	// Tenant is the tenant's name; needed only when the system has several tenants and no entitlement tells which.
	Tenant string `json:"tenant,omitempty"`
	// Entitlement is the entitlement to grant or revoke; the only extra an operation can need.
	Entitlement uuid.UUID `json:"entitlement,omitzero"`
}

// workflowActions is the one action a workflow can run per kind: onboarding creates an account and grants entitlements.
var workflowActions = map[ConnectorKind]string{AccountActionKind: "create", EntitlementKind: "grant"}

func (o WorkflowOperation) Validate() error {
	if o.System == "" || o.Resource == "" {
		return fmt.Errorf("%w: operation needs a system and a resource", ErrInvalidWorkflow)
	}

	if workflowActions[o.Kind] != o.Action {
		return fmt.Errorf("%w: %s %q is not a workflow action, onboarding only creates accounts and grants entitlements", ErrInvalidWorkflow, o.Kind, o.Action)
	}

	needsEntitlement := needsEntitlement(o.Kind, o.Action)
	if needsEntitlement && o.Entitlement == uuid.Nil {
		return fmt.Errorf("%w: %s %q needs an entitlement", ErrInvalidWorkflow, o.Kind, o.Action)
	}

	if !needsEntitlement && o.Entitlement != uuid.Nil {
		return fmt.Errorf("%w: %s %q does not take an entitlement", ErrInvalidWorkflow, o.Kind, o.Action)
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
	Steps []WorkflowStep `json:"steps"`
}

// WorkflowRun is one run of a definition, about an identity. Workflows that touch no identity (e.g. listing accounts) leave it empty.
type WorkflowRun struct {
	Definition WorkflowDefinition `json:"definition"`
	IdentityId uuid.UUID          `json:"identityId"`
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

func (d WorkflowDefinition) Validate() error {
	if err := validateName("workflow", d.Name); err != nil {
		return err
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

		if err := step.Operation.Validate(); err != nil {
			return fmt.Errorf("step %q: %w", step.Name, err)
		}
	}

	return nil
}

// Validate also checks the run has an identity when a step works on one.
func (r WorkflowRun) Validate() error {
	if err := r.Definition.Validate(); err != nil {
		return err
	}

	if r.IdentityId != uuid.Nil {
		return nil
	}

	for _, step := range r.Definition.Steps {
		if needsIdentity(step.Operation.Action) {
			return fmt.Errorf("%w: step %q works on an identity but the run has none", ErrInvalidWorkflow, step.Name)
		}
	}

	return nil
}
