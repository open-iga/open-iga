-- name: GetAccountByTenantAndIdentity :one
SELECT * FROM account
WHERE tenant_id = $1
  AND identity_id = $2;
