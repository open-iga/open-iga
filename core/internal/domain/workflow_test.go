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
	grant := operation(EntitlementKind)
	grant.Entitlement = uuid.New()
	assert.NoError(t, operation(AccountActionKind).Validate())
	assert.NoError(t, grant.Validate())

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
			assert.ErrorIs(t, invalidOperation.Validate(), ErrInvalidWorkflow)
		})
	}
}

func TestWorkflowDefinitionValidate(t *testing.T) {
	create := operation(AccountActionKind)
	grant := operation(EntitlementKind)
	grant.Entitlement = uuid.New()
	valid := WorkflowDefinition{Name: "onboarding", Steps: []WorkflowStep{
		{Name: "create-aws", Operation: create},
		{Name: "create-github", Operation: create},
		{Name: "grant-aws", Operation: grant, After: []string{"create-aws"}},
	}}
	assert.NoError(t, valid.Validate())

	invalid := map[string]WorkflowDefinition{
		"no name":            {Steps: valid.Steps},
		"no steps":           {Name: "w"},
		"slash in name":      {Name: "a/b", Steps: valid.Steps},
		"slash in step name": {Name: "w", Steps: []WorkflowStep{{Name: "a/b", Operation: create}}},
		"unnamed step":       {Name: "w", Steps: []WorkflowStep{{Operation: create}}},
		"duplicate step":     {Name: "w", Steps: []WorkflowStep{{Name: "a", Operation: create}, {Name: "a", Operation: create}}},
		"after unknown step": {Name: "w", Steps: []WorkflowStep{{Name: "a", Operation: create, After: []string{"b"}}}},
		"after later step":   {Name: "w", Steps: []WorkflowStep{{Name: "a", Operation: create, After: []string{"b"}}, {Name: "b", Operation: create}}},
		"after itself":       {Name: "w", Steps: []WorkflowStep{{Name: "a", Operation: create, After: []string{"a"}}}},
		"invalid operation":  {Name: "w", Steps: []WorkflowStep{{Name: "a", Operation: WorkflowOperation{}}}},
	}
	for name, definition := range invalid {
		t.Run(name, func(t *testing.T) {
			assert.ErrorIs(t, definition.Validate(), ErrInvalidWorkflow)
		})
	}
}

func TestWorkflowRunValidate(t *testing.T) {
	createDefinition := WorkflowDefinition{Name: "w", Steps: []WorkflowStep{{Name: "a", Operation: operation(AccountActionKind)}}}

	assert.NoError(t, WorkflowRun{Definition: createDefinition, IdentityId: uuid.New()}.Validate())
	assert.ErrorIs(t, WorkflowRun{Definition: createDefinition}.Validate(), ErrInvalidWorkflow)
}
