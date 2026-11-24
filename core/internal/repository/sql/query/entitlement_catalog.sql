-- name: GetEntitlementCatalog :one
SELECT * FROM entitlement_catalog
WHERE id = $1
  AND managed_system_id = $2
  AND resource = $3;

-- name: UpsertEntitlementCatalog :one
INSERT INTO entitlement_catalog (managed_system_id, tenant_id, resource, external_id, display_name, description, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (managed_system_id, tenant_id, resource, external_id) DO UPDATE
SET display_name = EXCLUDED.display_name, description = EXCLUDED.description, metadata = EXCLUDED.metadata, last_seen_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: SearchEntitlementCatalog :many
-- query_pattern is the query with its LIKE wildcards escaped; query is the query as typed, for ranking.
SELECT sqlc.embed(entitlement_catalog), COUNT(*) OVER () AS total
FROM entitlement_catalog
WHERE (sqlc.narg('managed_system_id')::uuid IS NULL OR managed_system_id = sqlc.narg('managed_system_id'))
  AND (sqlc.narg('tenant_id')::uuid IS NULL OR tenant_id = sqlc.narg('tenant_id'))
  AND (sqlc.narg('resource')::text IS NULL OR resource = sqlc.narg('resource'))
  AND (sqlc.narg('query_pattern')::text IS NULL OR display_name ILIKE '%' || sqlc.narg('query_pattern')::text || '%' OR external_id ILIKE '%' || sqlc.narg('query_pattern')::text || '%')
ORDER BY
  (display_name ILIKE sqlc.narg('query_pattern')::text || '%' OR external_id ILIKE sqlc.narg('query_pattern')::text || '%') DESC,
  GREATEST(word_similarity(sqlc.narg('query')::text, display_name), word_similarity(sqlc.narg('query')::text, external_id)) DESC,
  display_name, id
LIMIT sqlc.arg('page_size') OFFSET sqlc.arg('page_offset');
