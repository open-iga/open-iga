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

// OperationData is what the core loads from its DB for one connector call. The connector input is
// built only from it, never from caller values. Tenant is always set; the rest only when needed.
type OperationData struct {
	Tenant      *ManagedSystemTenant
	Identity    *Identity
	Account     *Account
	Entitlement *Entitlement
}

// operationBuilders build the typed input of each operation, keyed by kind and action name.
var operationBuilders = map[ConnectorKind]map[string]func(OperationData) (ConnectorOperation, error){
	AccountActionKind: {
		"create": func(data OperationData) (ConnectorOperation, error) {
			if data.Identity == nil {
				return nil, fmt.Errorf("%w: identity", ErrMissingOperationData)
			}
			return AccountCreateInput{
				Tenant:    data.connectorTenant(),
				Email:     data.Identity.Email,
				Firstname: data.Identity.FirstName,
				Lastname:  data.Identity.LastName,
				Metadata:  nil,
			}, nil
		},
		"enable":  buildAccountIDInput(func(input AccountIDInput) ConnectorOperation { return AccountEnableInput(input) }),
		"disable": buildAccountIDInput(func(input AccountIDInput) ConnectorOperation { return AccountDisableInput(input) }),
		"delete":  buildAccountIDInput(func(input AccountIDInput) ConnectorOperation { return AccountDeleteInput(input) }),
		"read":    buildAccountIDInput(func(input AccountIDInput) ConnectorOperation { return AccountReadInput(input) }),
		"list": func(data OperationData) (ConnectorOperation, error) {
			return AccountListInput{Tenant: data.connectorTenant(), Cursor: nil}, nil
		},
	},
	EntitlementKind: {
		"discover": func(data OperationData) (ConnectorOperation, error) {
			// TODO: pass cursor from the application layer
			return EntitlementDiscoverInput{Tenant: data.connectorTenant(), Cursor: nil}, nil
		},
		"grant": buildEntitlementAssignment(func(input EntitlementAssignmentInput) ConnectorOperation {
			return EntitlementGrantInput(input)
		}),
		"revoke": buildEntitlementAssignment(func(input EntitlementAssignmentInput) ConnectorOperation {
			return EntitlementRevokeInput(input)
		}),
		"read": buildAccountIDInput(func(input AccountIDInput) ConnectorOperation { return EntitlementReadInput(input) }),
	},
}

func buildAccountIDInput(wrap func(AccountIDInput) ConnectorOperation) func(OperationData) (ConnectorOperation, error) {
	return func(data OperationData) (ConnectorOperation, error) {
		accountID, err := data.accountExternalID()
		if err != nil {
			return nil, err
		}
		return wrap(AccountIDInput{Tenant: data.connectorTenant(), AccountID: accountID}), nil
	}
}

func buildEntitlementAssignment(wrap func(EntitlementAssignmentInput) ConnectorOperation) func(OperationData) (ConnectorOperation, error) {
	return func(data OperationData) (ConnectorOperation, error) {
		accountID, err := data.accountExternalID()
		if err != nil {
			return nil, err
		}
		if data.Entitlement == nil {
			return nil, fmt.Errorf("%w: entitlement", ErrMissingOperationData)
		}
		return wrap(EntitlementAssignmentInput{
			Tenant:        data.connectorTenant(),
			AccountID:     accountID,
			EntitlementID: data.Entitlement.Name,
		}), nil
	}
}

func (d OperationData) connectorTenant() ConnectorTenant {
	return ConnectorTenant{ID: d.Tenant.ExternalId}
}

// accountExternalID is the account's id in the target system; it is empty until the account is provisioned.
func (d OperationData) accountExternalID() (string, error) {
	if d.Account == nil {
		return "", fmt.Errorf("%w: account", ErrMissingOperationData)
	}
	if d.Account.ExternalId == "" {
		return "", fmt.Errorf("%w: account is not provisioned yet", ErrMissingOperationData)
	}
	return d.Account.ExternalId, nil
}

// NewConnectorOperation builds the typed input of the action named by kind and name from the loaded data.
func NewConnectorOperation(kind ConnectorKind, action string, data OperationData) (ConnectorOperation, error) {
	build, ok := operationBuilders[kind][action]
	if !ok {
		return nil, fmt.Errorf("%w: %s %s", ErrUnknownOperation, kind, action)
	}
	return build(data)
}
