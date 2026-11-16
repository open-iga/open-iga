-- name: GetAccount :one
SELECT * FROM account
WHERE managed_system_id = $1
  AND tenant_id IS NOT DISTINCT FROM $2
  AND identity_id = $3;

-- name: UpsertAccount :one
INSERT INTO account (managed_system_id, tenant_id, identity_id, external_id)
VALUES ($1, $2, $3, $4)
ON CONFLICT (managed_system_id, tenant_id, identity_id) DO UPDATE
SET updated_at = CURRENT_TIMESTAMP
WHERE account.external_id = EXCLUDED.external_id
RETURNING *;
