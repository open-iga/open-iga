-- name: CreateManagedSystem :one
INSERT INTO managed_system(name, connector_url, connector_hash)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListManagedSystems :many
SELECT *
FROM managed_system
ORDER BY created_at DESC;
