-- name: GetRolesByIdentityId :many
SELECT r.name FROM role r
    JOIN identity_role ir ON ir.role_id = r.id
              WHERE ir.identity_id = $1;

-- name: UpsertRoleByIdentityId :one
WITH inserted AS (
    INSERT INTO identity_role(role_id, identity_id)
        VALUES ((SELECT id from role where role.name = $1), $2) ON CONFLICT (role_id, identity_id) DO NOTHING
        RETURNING role_id
) SELECT r.name from role r JOIN inserted i ON r.id = i.role_id;

-- atleast one admin should present for the application to start
-- name: CountAdmin :one
SELECT COUNT(*) FROM identity_role ir
    JOIN role r ON r.id = ir.role_id
    WHERE r.name = 'admin';
