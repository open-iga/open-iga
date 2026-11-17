/*
With ORM, entity is the single source truth. But with migrations and SQLC, it's difficult to see the state of the DB in single place
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

-- A managed system is an external target, onboarded with a single WASM connector
CREATE TABLE managed_system (
    id              uuid PRIMARY KEY    DEFAULT gen_random_uuid(),
    name            VARCHAR(255) UNIQUE NOT NULL,
    connector_url   TEXT                NOT NULL,
    connector_hash  VARCHAR(64)         NOT NULL,
    connector_spec  JSONB               NOT NULL,
    created_at      TIMESTAMPTZ(6)      DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ(6)      DEFAULT CURRENT_TIMESTAMP
);

-- A tenant/scope within a system: a Keycloak realm, an Azure AD tenant, etc.
-- Single-tenant systems (e.g. one AWS account) have exactly one row here.
CREATE TABLE managed_system_tenant (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    managed_system_id uuid REFERENCES managed_system(id) NOT NULL,
    external_id       VARCHAR(255)   NOT NULL,
    name              VARCHAR(255)   NOT NULL,
    created_at        TIMESTAMPTZ(6) DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMPTZ(6) DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX idx_unique_managed_system_tenant ON managed_system_tenant(managed_system_id, external_id);
CREATE UNIQUE INDEX idx_unique_managed_system_tenant_name ON managed_system_tenant(managed_system_id, name);

-- Entitlements are exposed by business owners/admins. These are owned by the connectors
CREATE TABLE entitlement_catalog (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    managed_system_id uuid REFERENCES managed_system(id) NOT NULL,
    tenant_id  uuid REFERENCES managed_system_tenant(id),
    resource     VARCHAR(255)   NOT NULL,
    external_id  VARCHAR(255)   NOT NULL,
    display_name VARCHAR(255)   NOT NULL,
    description  TEXT,
    metadata     JSONB          NOT NULL DEFAULT '{}',
    last_seen_at TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at   TIMESTAMPTZ(6) DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMPTZ(6) DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX idx_unique_entitlement_catalog ON entitlement_catalog(managed_system_id, tenant_id, resource, external_id) NULLS NOT DISTINCT;

-- An identity's managed account on a tenant. Created before any entitlement is assigned.
CREATE TABLE account (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    managed_system_id uuid REFERENCES managed_system(id) NOT NULL,
    tenant_id   uuid REFERENCES managed_system_tenant(id),
    identity_id uuid REFERENCES identity(id) NOT NULL,
    -- account id in the target system, set once provisioned
    external_id VARCHAR(255),
    enabled     BOOLEAN          NOT NULL DEFAULT TRUE,
    metadata    JSONB            NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ(6)   DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ(6)   DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX idx_unique_account ON account(managed_system_id, tenant_id, identity_id) NULLS NOT DISTINCT;

-- An entitlement assigned to an account
CREATE TABLE account_entitlement (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id     uuid REFERENCES account(id) NOT NULL,
    entitlement_catalog_id uuid REFERENCES entitlement_catalog(id) NOT NULL,
    external_id    VARCHAR(255),
    metadata       JSONB            NOT NULL DEFAULT '{}',
    created_at     TIMESTAMPTZ(6)   DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMPTZ(6)   DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX idx_unique_account_entitlement ON account_entitlement(account_id, entitlement_catalog_id);

-- A workflow an admin defined: connector operation steps that the workflow engine runs. The definition is the domain JSON as is.
CREATE TABLE workflow_definition (
    name       VARCHAR(255) PRIMARY KEY,
    definition JSONB          NOT NULL,
    created_at TIMESTAMPTZ(6) DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ(6) DEFAULT CURRENT_TIMESTAMP
);
