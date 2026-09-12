-- A managed system is an external target, onboarded with a single WASM connector
CREATE TABLE managed_system (
    id            uuid PRIMARY KEY    DEFAULT gen_random_uuid(),
    name          VARCHAR(255) UNIQUE NOT NULL,
    connector_url TEXT                NOT NULL,
    connector_sha VARCHAR(64)         NOT NULL,
    created_at    TIMESTAMPTZ(6)      DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMPTZ(6)      DEFAULT CURRENT_TIMESTAMP
);

-- Entitlements are exposed by business owners/admins. These are exposed by the connectors
CREATE TABLE entitlement (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    managed_system_id  uuid REFERENCES managed_system(id) NOT NULL,
    name       VARCHAR(255)     NOT NULL,
    metadata   JSONB            NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ(6)   DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ(6)   DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX idx_unique_entitlement ON entitlement(managed_system_id, name);

-- An identity's managed account on a system. Created before any entitlement is assigned.
CREATE TABLE account (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    managed_system_id   uuid REFERENCES managed_system(id) NOT NULL,
    identity_id uuid REFERENCES identity(id) NOT NULL,
    -- account id in the target system, set once provisioned
    external_id VARCHAR(255),
    enabled     BOOLEAN          NOT NULL DEFAULT TRUE,
    metadata    JSONB            NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ(6)   DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ(6)   DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX idx_unique_account ON account(managed_system_id, identity_id);

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
