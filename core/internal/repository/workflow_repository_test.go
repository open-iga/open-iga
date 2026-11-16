package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/open-iga/core/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkflowRepository(t *testing.T) {
	t.Cleanup(func() {
		_, _ = conn.Exec(t.Context(), "DELETE FROM workflow_definition WHERE name = 'repository-test'")
	})
	definition := domain.WorkflowDefinition{Name: "repository-test", Steps: []domain.WorkflowStep{
		{Name: "create", Operation: domain.WorkflowOperation{System: "aws", Resource: "iam-user", Kind: domain.AccountActionKind, Action: "create"}},
		{
			Name:      "grant",
			Operation: domain.WorkflowOperation{System: "aws", Resource: "iam-user", Kind: domain.EntitlementKind, Action: "grant", Entitlement: uuid.New()},
			After:     []string{"create"},
			Gate:      true,
		},
	}}

	t.Run("creates a workflow and reads it back unchanged", func(t *testing.T) {
		require.NoError(t, repository.WorkflowRepository.CreateWorkflow(t.Context(), definition))

		loaded, err := repository.WorkflowRepository.GetWorkflow(t.Context(), "repository-test")
		require.NoError(t, err)
		assert.Equal(t, definition, loaded)
	})

	t.Run("a second workflow with the same name is refused and the first stays", func(t *testing.T) {
		other := domain.WorkflowDefinition{Name: "repository-test", Steps: definition.Steps[:1]}
		assert.ErrorIs(t, repository.WorkflowRepository.CreateWorkflow(t.Context(), other), domain.ErrWorkflowAlreadyExists)

		loaded, err := repository.WorkflowRepository.GetWorkflow(t.Context(), "repository-test")
		require.NoError(t, err)
		assert.Equal(t, definition, loaded)
	})

	t.Run("returns not found for an unknown workflow", func(t *testing.T) {
		_, err := repository.WorkflowRepository.GetWorkflow(t.Context(), "no-such-workflow")
		assert.ErrorIs(t, err, domain.ErrWorkflowNotFound)
	})
}
