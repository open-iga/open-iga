-- name: CreateManagedSystem :one
INSERT INTO managed_system(name, connector_url, connector_hash)
VALUES ($1, $2, $3)
RETURNING *;
