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
