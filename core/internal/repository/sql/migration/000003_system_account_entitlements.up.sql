-- A managed system is an external target, onboarded with a WASM connector
CREATE TABLE managed_system (
    id              uuid PRIMARY KEY    DEFAULT gen_random_uuid(),
    name            VARCHAR(255) UNIQUE NOT NULL,
    connector_url   TEXT                NOT NULL,
    connector_hash  VARCHAR(64)         NOT NULL,
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
    metadata          JSONB          NOT NULL DEFAULT '{}',
    created_at        TIMESTAMPTZ(6) DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMPTZ(6) DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX idx_unique_managed_system_tenant ON managed_system_tenant(managed_system_id, external_id);

-- Entitlements are exposed by connectors. These are onboarded by business owners/admins
CREATE TABLE entitlement (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id  uuid REFERENCES managed_system_tenant(id) NOT NULL,
    -- entitlement name from the connector
    name       VARCHAR(255)     NOT NULL,
    created_at TIMESTAMPTZ(6)   DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ(6)   DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX idx_unique_entitlement ON entitlement(tenant_id, name);

-- An identity's managed account on a tenant. Created before any entitlement is assigned.
CREATE TABLE account (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   uuid REFERENCES managed_system_tenant(id) NOT NULL,
    identity_id uuid REFERENCES identity(id) NOT NULL,
    -- account id in the target system, set once provisioned
    external_id VARCHAR(255),
    enabled     BOOLEAN          NOT NULL DEFAULT TRUE,
    metadata    JSONB            NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ(6)   DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ(6)   DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX idx_unique_account ON account(tenant_id, identity_id);

-- An entitlement assigned to an account
CREATE TABLE account_entitlement (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id     uuid REFERENCES account(id) NOT NULL,
    entitlement_id uuid REFERENCES entitlement(id) NOT NULL,
    external_id    VARCHAR(255),
    metadata       JSONB            NOT NULL DEFAULT '{}',
    created_at     TIMESTAMPTZ(6)   DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMPTZ(6)   DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX idx_unique_account_entitlement ON account_entitlement(account_id, entitlement_id);
