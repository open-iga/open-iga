-- name: CreateManagedSystem :one
INSERT INTO managed_system(name, connector_url, connector_hash, connector_spec)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListManagedSystems :many
SELECT * FROM managed_system
ORDER BY created_at DESC;

-- name: GetManagedSystem :one
SELECT * FROM managed_system
WHERE id = $1;

-- name: GetManagedSystemByName :one
SELECT * FROM managed_system
WHERE name = $1;
