-- name: CreateManagedSystemTenant :one
INSERT INTO managed_system_tenant(managed_system_id, external_id, name)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListManagedSystemTenants :many
SELECT *
FROM managed_system_tenant
WHERE managed_system_id = $1
ORDER BY created_at DESC;

-- name: GetManagedSystemTenantByExternalId :one
SELECT *
FROM managed_system_tenant
WHERE managed_system_id = $1
  AND external_id = $2;
