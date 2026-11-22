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

const (
	WorkflowTypeOnboarding WorkflowType = "onboarding"
	WorkflowTypeDiscovery  WorkflowType = "discovery"
)

// WorkflowQueue limits how many runs of a type run at the same time; the other runs wait.
type WorkflowQueue struct {
	Concurrency int
}

// WorkflowTypeSpec says what the runs of a type do and how the engine starts them.
type WorkflowTypeSpec struct {
	// Actions is the action a run runs per kind of operation.
	Actions map[ConnectorKind]string
	// NeedsIdentity is true when a run is about one identity.
	NeedsIdentity bool
	AllowsGates   bool
	// Queue is nil when runs start at once. A queued run holds its slot until it ends and a gated step waits up to 30 days,
	// so gates in a queued type would block the queue. A type that allows gates must not have a queue.
	Queue *WorkflowQueue
}

var workflowTypes = map[WorkflowType]WorkflowTypeSpec{
	WorkflowTypeOnboarding: {
		Actions:       map[ConnectorKind]string{AccountActionKind: "create", EntitlementKind: "grant"},
		NeedsIdentity: true,
		AllowsGates:   true,
		Queue:         nil,
	},
	WorkflowTypeDiscovery: {
		Actions:       map[ConnectorKind]string{EntitlementKind: "discover"},
		NeedsIdentity: false,
		AllowsGates:   false,
		Queue:         &WorkflowQueue{Concurrency: 1},
	},
}

// QueuedWorkflowTypes is the queue of each type that has one; the engine creates them when it launches.
func QueuedWorkflowTypes() map[WorkflowType]WorkflowQueue {
	queues := map[WorkflowType]WorkflowQueue{}
	for workflowType, spec := range workflowTypes {
		if spec.Queue != nil {
			queues[workflowType] = *spec.Queue
		}
	}

	return queues
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

	takesEntitlement := needsEntitlement(o.Kind, actions[o.Kind])
	if takesEntitlement && o.Entitlement == uuid.Nil {
		return fmt.Errorf("%w: %s needs an entitlement", ErrInvalidWorkflow, o.Kind)
	}

	if !takesEntitlement && o.Entitlement != uuid.Nil {
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
	// Gate is a prerequisite: the run waits before this step until someone continues it manually
	Gate bool `json:"gate,omitempty"`
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
		d.Steps[index].Operation.Action = workflowTypes[d.Type].Actions[d.Steps[index].Operation.Kind]
	}
}

func (d *WorkflowDefinition) Validate() error {
	if err := validateName("workflow", d.Name); err != nil {
		return err
	}

	spec, ok := workflowTypes[d.Type]
	if !ok {
		return fmt.Errorf("%w: workflow %q has type %q, use one of %v", ErrInvalidWorkflow, d.Name, d.Type, slices.Sorted(maps.Keys(workflowTypes)))
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

		if err := validateStepForType(step, d.Type, spec); err != nil {
			return err
		}
	}

	return nil
}

// validateStepForType checks the step against what its workflow type allows.
func validateStepForType(step WorkflowStep, workflowType WorkflowType, spec WorkflowTypeSpec) error {
	if step.Gate && !spec.AllowsGates {
		return fmt.Errorf("%w: step %q has a gate, but a %s workflow has none", ErrInvalidWorkflow, step.Name, workflowType)
	}

	if err := step.Operation.Validate(spec.Actions); err != nil {
		return fmt.Errorf("step %q: %w", step.Name, err)
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

// Validate also checks the run has an identity when its type needs one.
func (r WorkflowRun) Validate() error {
	if err := r.Definition.Validate(); err != nil {
		return err
	}

	if workflowTypes[r.Definition.Type].NeedsIdentity && r.IdentityId == uuid.Nil {
		return fmt.Errorf("%w: the run has no identity", ErrInvalidWorkflow)
	}

	return nil
}
