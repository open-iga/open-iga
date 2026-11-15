-- name: UpsertWorkflowDefinition :exec
INSERT INTO workflow_definition(name, definition)
VALUES ($1, $2)
ON CONFLICT (name) DO UPDATE
SET definition = EXCLUDED.definition, updated_at = CURRENT_TIMESTAMP;

-- name: GetWorkflowDefinition :one
SELECT * FROM workflow_definition
WHERE name = $1;
