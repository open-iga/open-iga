package domain

// AccountIDInput is the input shape of every account operation on an existing account.
// Each operation below gets its own type from it, because the type names the operation.
type AccountIDInput struct {
	Tenant    ConnectorTenant `json:"tenant"`
	AccountID string          `json:"accountId"`
}

type AccountCreateInput struct {
	Tenant    ConnectorTenant `json:"tenant"`
	Email     string          `json:"email"`
	Firstname string          `json:"firstname"`
	Lastname  string          `json:"lastname"`
	Metadata  Metadata        `json:"metadata,omitempty"`
}

type AccountCreateOutput struct {
	AccountID string `json:"accountId"`
	// Created is false when the connector returned an existing account instead of creating one.
	Created  bool     `json:"created"`
	Metadata Metadata `json:"metadata,omitempty"`
}

func (AccountCreateInput) Kind() ConnectorKind { return AccountActionKind }
func (AccountCreateInput) Type() string        { return "create" }
func (AccountCreateInput) DecodeOutput(raw []byte) (ConnectorOutput, error) {
	return decodeOutput[AccountCreateOutput](raw)
}
func (AccountCreateOutput) connectorOutput() {} // marker only, no behavior

type AccountEnableInput AccountIDInput

type AccountEnableOutput struct {
	Enabled bool `json:"enabled"`
}

func (AccountEnableInput) Kind() ConnectorKind { return AccountActionKind }
func (AccountEnableInput) Type() string        { return "enable" }
func (AccountEnableInput) DecodeOutput(raw []byte) (ConnectorOutput, error) {
	return decodeOutput[AccountEnableOutput](raw)
}
func (AccountEnableOutput) connectorOutput() {} // marker only, no behavior

type AccountDisableInput AccountIDInput

type AccountDisableOutput struct {
	Disabled bool `json:"disabled"`
}

func (AccountDisableInput) Kind() ConnectorKind { return AccountActionKind }
func (AccountDisableInput) Type() string        { return "disable" }
func (AccountDisableInput) DecodeOutput(raw []byte) (ConnectorOutput, error) {
	return decodeOutput[AccountDisableOutput](raw)
}
func (AccountDisableOutput) connectorOutput() {} // marker only, no behavior

type AccountDeleteInput AccountIDInput

type AccountDeleteOutput struct {
	Deleted bool `json:"deleted"`
}

func (AccountDeleteInput) Kind() ConnectorKind { return AccountActionKind }
func (AccountDeleteInput) Type() string        { return "delete" }
func (AccountDeleteInput) DecodeOutput(raw []byte) (ConnectorOutput, error) {
	return decodeOutput[AccountDeleteOutput](raw)
}
func (AccountDeleteOutput) connectorOutput() {} // marker only, no behavior

type AccountReadInput AccountIDInput

// AccountReadOutput is an account-drift report, not profile attributes.
type AccountReadOutput struct {
	Exists  bool `json:"exists"`
	Enabled bool `json:"enabled"`
}

func (AccountReadInput) Kind() ConnectorKind { return AccountActionKind }
func (AccountReadInput) Type() string        { return "read" }
func (AccountReadInput) DecodeOutput(raw []byte) (ConnectorOutput, error) {
	return decodeOutput[AccountReadOutput](raw)
}
func (AccountReadOutput) connectorOutput() {} // marker only, no behavior

// AccountListInput asks for one page of the tenant's accounts; a nil Cursor means the first page.
type AccountListInput struct {
	Tenant ConnectorTenant `json:"tenant"`
	Cursor *string         `json:"cursor,omitempty"`
}

type ListedAccount struct {
	AccountID string `json:"accountId"`
	Email     string `json:"email"`
	Enabled   bool   `json:"enabled"`
	Firstname string `json:"firstname,omitempty"`
	Lastname  string `json:"lastname,omitempty"`
}

// AccountListOutput is one page; an empty NextCursor means it was the last page.
type AccountListOutput struct {
	Accounts   []ListedAccount `json:"accounts"`
	NextCursor string          `json:"nextCursor,omitempty"`
}

func (AccountListInput) Kind() ConnectorKind { return AccountActionKind }
func (AccountListInput) Type() string        { return "list" }
func (AccountListInput) DecodeOutput(raw []byte) (ConnectorOutput, error) {
	return decodeOutput[AccountListOutput](raw)
}
func (AccountListOutput) connectorOutput() {} // marker only, no behavior
