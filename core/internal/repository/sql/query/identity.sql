-- name: UpsertIdentity :one
INSERT INTO identity (first_name, last_name, type, email)
VALUES ($1, $2, $3, $4)
ON CONFLICT (email) DO UPDATE
    SET email      = excluded.email,
        first_name = excluded.first_name,
        last_name  = excluded.last_name,
        updated_at = CURRENT_TIMESTAMP
RETURNING *;
