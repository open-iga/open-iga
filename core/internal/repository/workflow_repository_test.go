package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/open-iga/core/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildWorkflowDefinition(name string) domain.WorkflowDefinition {
	return domain.WorkflowDefinition{Name: name, Steps: []domain.WorkflowStep{
		{Name: "create", Operation: domain.WorkflowOperation{System: "aws", Resource: "iam-user", Kind: domain.AccountActionKind, Action: "create"}},
		{
			Name:      "grant",
			Operation: domain.WorkflowOperation{System: "aws", Resource: "iam-user", Kind: domain.EntitlementKind, Action: "grant", Entitlement: uuid.New()},
			After:     []string{"create"},
		},
	}}
}

func TestWorkflowRepository(t *testing.T) {
	t.Run("creates a workflow and reads it back unchanged", func(t *testing.T) {
		definition := buildWorkflowDefinition("repository-test-create")
		t.Cleanup(func() { deleteRows(t, "DELETE FROM workflow_definition WHERE name = $1", definition.Name) })

		require.NoError(t, repository.WorkflowRepository.CreateWorkflow(t.Context(), definition))

		loaded, err := repository.WorkflowRepository.GetWorkflow(t.Context(), definition.Name)
		require.NoError(t, err)
		assert.Equal(t, definition, loaded)
	})

	t.Run("a second workflow with the same name is refused and the first stays", func(t *testing.T) {
		definition := buildWorkflowDefinition("repository-test-duplicate")
		t.Cleanup(func() { deleteRows(t, "DELETE FROM workflow_definition WHERE name = $1", definition.Name) })
		require.NoError(t, repository.WorkflowRepository.CreateWorkflow(t.Context(), definition))

		other := domain.WorkflowDefinition{Name: definition.Name, Steps: definition.Steps[:1]}
		assert.ErrorIs(t, repository.WorkflowRepository.CreateWorkflow(t.Context(), other), domain.ErrWorkflowAlreadyExists)

		loaded, err := repository.WorkflowRepository.GetWorkflow(t.Context(), definition.Name)
		require.NoError(t, err)
		assert.Equal(t, definition, loaded)
	})

	t.Run("returns not found for an unknown workflow", func(t *testing.T) {
		_, err := repository.WorkflowRepository.GetWorkflow(t.Context(), "no-such-workflow")
		assert.ErrorIs(t, err, domain.ErrWorkflowNotFound)
	})
}
