-- A workflow an admin defined: connector operation steps that the workflow engine runs. The definition is the domain JSON as is.
CREATE TABLE workflow_definition (
    name       VARCHAR(255) PRIMARY KEY,
    definition JSONB          NOT NULL,
    created_at TIMESTAMPTZ(6) DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ(6) DEFAULT CURRENT_TIMESTAMP
);
