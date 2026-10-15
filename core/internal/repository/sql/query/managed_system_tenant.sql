-- name: CreateManagedSystemTenant :one
INSERT INTO managed_system_tenant(managed_system_id, external_id, name, metadata)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetManagedSystemTenantByExternalId :one
SELECT *
FROM managed_system_tenant
WHERE managed_system_id = $1
  AND external_id = $2;
