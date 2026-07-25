/*
With ORM entity is the single source truth. But with migrations and SQLC, it's difficult to see the state of the DB in single place
The aim of this file is to check the state of DB in a single place and check if the DB is actually in an expected state post migration.
*/
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


CREATE TABLE role (
    id  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name    VARCHAR(50) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ(3) DEFAULT CURRENT_TIMESTAMP
);
-- seed default roles in DB during migration
INSERT INTO role(name) VALUES ('admin'), ('member');

CREATE TABLE identity_role (
       id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
       role_id uuid REFERENCES role(id) NOT NULL ,
       identity_id uuid REFERENCES identity(id) NOT NULL
);
CREATE UNIQUE INDEX idx_unique_identity_role ON identity_role(identity_id, role_id);
