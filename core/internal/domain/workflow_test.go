package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func operation(kind ConnectorKind) WorkflowOperation {
	return WorkflowOperation{System: "aws", Resource: "iam-user", Kind: kind}
}

func TestWorkflowOperationValidate(t *testing.T) {
	actions := workflowTypes[WorkflowTypeOnboarding].Actions
	grant := operation(EntitlementKind)
	grant.Entitlement = uuid.New()
	assert.NoError(t, operation(AccountActionKind).Validate(actions))
	assert.NoError(t, grant.Validate(actions))

	withEntitlement := operation(AccountActionKind)
	withEntitlement.Entitlement = uuid.New()

	invalid := map[string]WorkflowOperation{
		"no system":                    {Resource: "r", Kind: AccountActionKind},
		"no resource":                  {System: "s", Kind: AccountActionKind},
		"kind a workflow does not run": operation("other"),
		"grant without entitlement":    operation(EntitlementKind),
		"entitlement on create":        withEntitlement,
	}
	for name, invalidOperation := range invalid {
		t.Run(name, func(t *testing.T) {
			assert.ErrorIs(t, invalidOperation.Validate(actions), ErrInvalidWorkflow)
		})
	}
}

func TestWorkflowDefinitionValidate(t *testing.T) {
	create := operation(AccountActionKind)
	grant := operation(EntitlementKind)
	grant.Entitlement = uuid.New()
	valid := WorkflowDefinition{Name: "onboarding", Type: WorkflowTypeOnboarding, Steps: []WorkflowStep{
		{Name: "create-aws", Operation: create},
		{Name: "create-github", Operation: create, Gate: true},
		{Name: "grant-aws", Operation: grant, After: []string{"create-aws"}},
	}}
	assert.NoError(t, valid.Validate())
	valid.Name = "contractors"
	assert.NoError(t, valid.Validate(), "the name is free, the type decides")
	valid.Name = "onboarding"

	invalid := map[string]WorkflowDefinition{
		"no name":            {Steps: valid.Steps},
		"unknown type":       {Name: "w", Type: "offboarding", Steps: valid.Steps},
		"no steps":           {Name: "onboarding", Type: WorkflowTypeOnboarding},
		"slash in name":      {Name: "a/b", Steps: valid.Steps},
		"slash in step name": {Name: "onboarding", Steps: []WorkflowStep{{Name: "a/b", Operation: create}}},
		"unnamed step":       {Name: "onboarding", Steps: []WorkflowStep{{Operation: create}}},
		"duplicate step":     {Name: "onboarding", Steps: []WorkflowStep{{Name: "a", Operation: create}, {Name: "a", Operation: create}}},
		"after unknown step": {Name: "onboarding", Steps: []WorkflowStep{{Name: "a", Operation: create, After: []string{"b"}}}},
		"after later step":   {Name: "onboarding", Steps: []WorkflowStep{{Name: "a", Operation: create, After: []string{"b"}}, {Name: "b", Operation: create}}},
		"after itself":       {Name: "onboarding", Steps: []WorkflowStep{{Name: "a", Operation: create, After: []string{"a"}}}},
		"invalid operation":  {Name: "onboarding", Steps: []WorkflowStep{{Name: "a", Operation: WorkflowOperation{}}}},
	}
	for name, definition := range invalid {
		t.Run(name, func(t *testing.T) {
			assert.ErrorIs(t, definition.Validate(), ErrInvalidWorkflow)
		})
	}
}

// A queued run holds its queue slot until it ends, and a gated step waits up to 30 days.
// With gates in a queued type, a few runs waiting at a gate would block every other run of the type.
func TestWorkflowTypes_QueuedTypeAllowsNoGates(t *testing.T) {
	for workflowType, spec := range workflowTypes {
		if spec.Queue == nil {
			continue
		}

		assert.False(t, spec.AllowsGates, "type %q has a queue, so it must not allow gates: a gated step would hold a queue slot for days", workflowType)
		assert.Positive(t, spec.Queue.Concurrency, "type %q has a queue without a limit", workflowType)
	}
}

func TestWorkflowTypes_OnboardingHasNoQueueAndDiscoveryHasOne(t *testing.T) {
	queues := QueuedWorkflowTypes()

	assert.NotContains(t, queues, WorkflowTypeOnboarding)
	assert.Equal(t, WorkflowQueue{Concurrency: 1}, queues[WorkflowTypeDiscovery])
}

func TestDiscoveryWorkflowValidate(t *testing.T) {
	discover := operation(EntitlementKind)
	definition := WorkflowDefinition{Name: "discovery", Type: WorkflowTypeDiscovery, Steps: []WorkflowStep{{Name: "aws", Operation: discover}}}

	assert.NoError(t, definition.Validate(), "discover takes no entitlement")
	assert.NoError(t, WorkflowRun{Definition: definition}.Validate(), "a discovery run is about no identity")

	definition.Steps[0].Gate = true
	assert.ErrorIs(t, definition.Validate(), ErrInvalidWorkflow, "a queued type has no gates")
}

func TestWorkflowRunValidate(t *testing.T) {
	createDefinition := WorkflowDefinition{Name: "onboarding", Type: WorkflowTypeOnboarding, Steps: []WorkflowStep{{Name: "a", Operation: operation(AccountActionKind)}}}

	assert.NoError(t, WorkflowRun{Definition: createDefinition, IdentityId: uuid.New()}.Validate())
	assert.ErrorIs(t, WorkflowRun{Definition: createDefinition}.Validate(), ErrInvalidWorkflow)
}
