-- name: InsertWorkflowDefinition :exec
INSERT INTO workflow_definition(name, definition)
VALUES ($1, $2);

-- name: GetWorkflowDefinition :one
SELECT * FROM workflow_definition
WHERE name = $1;
