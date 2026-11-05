-- name: GetEntitlement :one
SELECT * FROM entitlement
WHERE id = $1;
