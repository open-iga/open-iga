-- name: CreateSession :one
INSERT INTO session (session_id, identity_id, active, expires_at)
VALUES ($1, $2, TRUE, $3)
RETURNING *;


-- name: FindBySessionId :one
SELECT sqlc.embed(identity), sqlc.embed(session)
FROM identity
         JOIN session ON session.identity_id = identity.id
WHERE session.session_id = $1;

-- name: FindActiveSessionByIdentityId :one
SELECT *
from session
WHERE identity_id = $1
  AND active = TRUE;

-- name: DeactivateBySessionId :one
UPDATE session
SET active = FALSE
WHERE session_id = $1
RETURNING *;

-- name: DeactivateByIdentityId :one
UPDATE session
SET active = FALSE
WHERE identity_id = $1
RETURNING *;
