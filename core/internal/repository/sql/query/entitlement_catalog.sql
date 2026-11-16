-- name: GetEntitlementCatalog :one
SELECT * FROM entitlement_catalog
WHERE id = $1
  AND managed_system_id = $2
  AND resource = $3;

-- name: UpsertEntitlementCatalog :one
INSERT INTO entitlement_catalog (managed_system_id, tenant_id, resource, external_id, display_name, description, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (managed_system_id, tenant_id, resource, external_id) DO UPDATE
SET display_name = EXCLUDED.display_name, description = EXCLUDED.description, metadata = EXCLUDED.metadata, updated_at = CURRENT_TIMESTAMP
RETURNING *;
