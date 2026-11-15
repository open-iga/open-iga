package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/open-iga/core/internal/domain"
	"github.com/open-iga/core/internal/repository/db"
)

type WorkflowRepository struct {
	queries *db.Queries
}

func NewWorkflowRepository(queries *db.Queries) *WorkflowRepository {
	return &WorkflowRepository{queries: queries}
}

func (w *WorkflowRepository) SaveWorkflow(ctx context.Context, definition domain.WorkflowDefinition) error {
	raw, err := json.Marshal(definition)
	if err != nil {
		return fmt.Errorf("failed to marshal workflow %q: %w", definition.Name, err)
	}
	if err := w.queries.UpsertWorkflowDefinition(ctx, db.UpsertWorkflowDefinitionParams{Name: definition.Name, Definition: raw}); err != nil {
		return fmt.Errorf("failed to save workflow %q: %w", definition.Name, err)
	}
	return nil
}

func (w *WorkflowRepository) GetWorkflow(ctx context.Context, name string) (domain.WorkflowDefinition, error) {
	row, err := w.queries.GetWorkflowDefinition(ctx, name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.WorkflowDefinition{}, fmt.Errorf("%w: %s", domain.ErrWorkflowNotFound, name)
		}

		return domain.WorkflowDefinition{}, fmt.Errorf("failed to get workflow %q: %w", name, err)
	}

	var definition domain.WorkflowDefinition
	if err := json.Unmarshal(row.Definition, &definition); err != nil {
		return domain.WorkflowDefinition{}, fmt.Errorf("failed to decode workflow %q: %w", name, err)
	}

	return definition, nil
}
