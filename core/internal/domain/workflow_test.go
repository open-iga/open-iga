package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func operation(kind ConnectorKind, action string) WorkflowOperation {
	return WorkflowOperation{System: "aws", Resource: "iam-user", Kind: kind, Action: action}
}

func TestWorkflowOperationValidate(t *testing.T) {
	grant := operation(EntitlementKind, "grant")
	grant.Entitlement = uuid.New()
	assert.NoError(t, operation(AccountActionKind, "create").Validate())
	assert.NoError(t, grant.Validate())

	withEntitlement := operation(AccountActionKind, "create")
	withEntitlement.Entitlement = uuid.New()

	invalid := map[string]WorkflowOperation{
		"no system":                 {Resource: "r", Kind: AccountActionKind, Action: "create"},
		"no resource":               {System: "s", Kind: AccountActionKind, Action: "create"},
		"unknown action":            operation(AccountActionKind, "nope"),
		"action of another kind":    operation(AccountActionKind, "grant"),
		"grant without entitlement": operation(EntitlementKind, "grant"),
		"entitlement on create":     withEntitlement,
	}
	for name, invalidOperation := range invalid {
		t.Run(name, func(t *testing.T) {
			assert.ErrorIs(t, invalidOperation.Validate(), ErrInvalidWorkflow)
		})
	}
}

func TestWorkflowDefinitionValidate(t *testing.T) {
	create := operation(AccountActionKind, "create")
	valid := WorkflowDefinition{Name: "onboarding", Steps: []WorkflowStep{
		{Name: "create-aws", Operation: create},
		{Name: "create-github", Operation: create},
		{Name: "disable-aws", Operation: operation(AccountActionKind, "disable"), After: []string{"create-aws"}},
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
	createDefinition := WorkflowDefinition{Name: "w", Steps: []WorkflowStep{{Name: "a", Operation: operation(AccountActionKind, "create")}}}
	listDefinition := WorkflowDefinition{Name: "w", Steps: []WorkflowStep{{Name: "a", Operation: operation(AccountActionKind, "list")}}}

	assert.NoError(t, WorkflowRun{Definition: createDefinition, IdentityId: uuid.New()}.Validate())
	assert.NoError(t, WorkflowRun{Definition: listDefinition}.Validate(), "listing accounts needs no identity")
	assert.ErrorIs(t, WorkflowRun{Definition: createDefinition}.Validate(), ErrInvalidWorkflow)
}
