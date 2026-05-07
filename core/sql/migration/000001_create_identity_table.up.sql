CREATE TYPE identity_type AS ENUM ('user');

CREATE TABLE identity
(
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name VARCHAR(255)     DEFAULT NULL,
    last_name  VARCHAR(255)     DEFAULT NULL,
    type       identity_type       NOT NULL,
    email      VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ(6)   DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ(6)   DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_identity_email on identity (email);


CREATE TABLE session
(
    id          uuid PRIMARY KEY        DEFAULT gen_random_uuid(),
-- session_id is the high entropy value used in cookie
    session_id  VARCHAR(64)    NOT NULL UNIQUE,
    identity_id uuid           NOT NULL REFERENCES identity (id),
    active      boolean        NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ(6)          DEFAULT CURRENT_TIMESTAMP,
    expires_at  TIMESTAMPTZ(6) NOT NULL
);

-- Ensure that only one active session exists per identity
CREATE UNIQUE INDEX idx_unique_session ON session (identity_id) WHERE active = TRUE;
CREATE INDEX idx_session_identity_id ON session (identity_id);
