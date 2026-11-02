package domain

import (
	"encoding/json"
	"fmt"
)

// ConnectorOperation is a typed handler input that knows its operation and how to decode that operation's output.
type ConnectorOperation interface {
	Kind() ConnectorKind
	Type() string
	DecodeOutput(raw []byte) (ConnectorOutput, error)
}

// ConnectorOutput is the output of one operation; callers type-switch on it.
// connectorOutput is an empty marker method: it does nothing, it only limits the interface to the output types in this package.
type ConnectorOutput interface{ connectorOutput() }

func decodeOutput[Out ConnectorOutput](raw []byte) (ConnectorOutput, error) {
	var out Out
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode %T: %w", out, err)
	}
	return out, nil
}

// Metadata is free-form JSON that connectors attach to accounts and entitlements.
type Metadata map[string]any

// ConnectorTenant is the tenant an operation runs against; ID is the tenant's external id.
type ConnectorTenant struct {
	ID string `json:"id"`
}
