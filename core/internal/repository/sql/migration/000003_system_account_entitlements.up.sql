-- A managed system is an external target, onboarded with a WASM connector
CREATE TABLE managed_system (
    id              uuid PRIMARY KEY    DEFAULT gen_random_uuid(),
    name            VARCHAR(255) UNIQUE NOT NULL,
    connector_url   TEXT                NOT NULL,
    connector_hash  VARCHAR(64)         NOT NULL,
    connector_spec  JSONB               NOT NULL,
    created_at      TIMESTAMPTZ(6)      DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ(6)      DEFAULT CURRENT_TIMESTAMP
);

-- A tenant/scope within a managed system: a Keycloak realm, etc.
-- Single-tenant systems (e.g. one AWS account) have exactly one row here.
CREATE TABLE managed_system_tenant (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    managed_system_id uuid REFERENCES managed_system(id) NOT NULL,
    -- tenant id in the target system (realm name, account id, ...)
    external_id       VARCHAR(255)   NOT NULL,
    name              VARCHAR(255)   NOT NULL,
    created_at        TIMESTAMPTZ(6) DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMPTZ(6) DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX idx_unique_managed_system_tenant ON managed_system_tenant(managed_system_id, external_id);
CREATE UNIQUE INDEX idx_unique_managed_system_tenant_name ON managed_system_tenant(managed_system_id, name);

-- The entitlement catalog: what each connector resource of a system (or of one of its tenants) can grant, as found by discovery
CREATE TABLE entitlement_catalog (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    managed_system_id uuid REFERENCES managed_system(id) NOT NULL,
    -- empty for a single-tenant system, which has no tenant rows
    tenant_id         uuid REFERENCES managed_system_tenant(id),
    -- connector resource that offers it
    resource     VARCHAR(255)     NOT NULL,
    -- entitlement id in the target system, sent by grant and revoke
    external_id  VARCHAR(255)     NOT NULL,
    display_name VARCHAR(255)     NOT NULL,
    description  TEXT,
    metadata     JSONB            NOT NULL DEFAULT '{}',
    -- set on every discovery run that finds it; a row a complete run did not find is stale
    last_seen_at TIMESTAMPTZ(6)   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at   TIMESTAMPTZ(6)   DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMPTZ(6)   DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX idx_unique_entitlement_catalog ON entitlement_catalog(managed_system_id, tenant_id, resource, external_id) NULLS NOT DISTINCT;

-- An identity's managed account on a system, or on one of its tenants. Created before any entitlement is assigned.
CREATE TABLE account (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    managed_system_id uuid REFERENCES managed_system(id) NOT NULL,
    -- empty for a single-tenant system, which has no tenant rows
    tenant_id         uuid REFERENCES managed_system_tenant(id),
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
