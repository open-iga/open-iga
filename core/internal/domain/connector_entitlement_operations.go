package domain

// EntitlementAssignmentInput is the input shape of grant and revoke: AccountID is the account id,
// EntitlementID comes from discover. Each operation below gets its own type from it.
type EntitlementAssignmentInput struct {
	Tenant        ConnectorTenant `json:"tenant"`
	AccountID     string          `json:"accountId"`
	EntitlementID string          `json:"entitlementId"`
}

// EntitlementDiscoverInput asks for one page of the catalog; a nil Cursor means the first page.
type EntitlementDiscoverInput struct {
	Tenant ConnectorTenant `json:"tenant"`
	Cursor *string         `json:"cursor,omitempty"`
}

type DiscoveredEntitlement struct {
	EntitlementID string   `json:"entitlementId"`
	Name          string   `json:"name"`
	Description   *string  `json:"description,omitempty"`
	Metadata      Metadata `json:"metadata,omitempty"`
}

// EntitlementDiscoverOutput is one page; an empty NextCursor means it was the last page.
type EntitlementDiscoverOutput struct {
	Entitlements []DiscoveredEntitlement `json:"entitlements"`
	NextCursor   string                  `json:"nextCursor,omitempty"`
}

func (EntitlementDiscoverInput) Kind() ConnectorKind { return EntitlementKind }
func (EntitlementDiscoverInput) Type() string        { return "discover" }
func (EntitlementDiscoverInput) DecodeOutput(raw []byte) (ConnectorOutput, error) {
	return decodeOutput[EntitlementDiscoverOutput](raw)
}
func (EntitlementDiscoverOutput) connectorOutput() {} // marker only, no behavior

type EntitlementGrantInput EntitlementAssignmentInput

type EntitlementGrantOutput struct {
	Granted bool `json:"granted"`
}

func (EntitlementGrantInput) Kind() ConnectorKind { return EntitlementKind }
func (EntitlementGrantInput) Type() string        { return "grant" }
func (EntitlementGrantInput) DecodeOutput(raw []byte) (ConnectorOutput, error) {
	return decodeOutput[EntitlementGrantOutput](raw)
}
func (EntitlementGrantOutput) connectorOutput() {} // marker only, no behavior

type EntitlementRevokeInput EntitlementAssignmentInput

type EntitlementRevokeOutput struct {
	Revoked bool `json:"revoked"`
}

func (EntitlementRevokeInput) Kind() ConnectorKind { return EntitlementKind }
func (EntitlementRevokeInput) Type() string        { return "revoke" }
func (EntitlementRevokeInput) DecodeOutput(raw []byte) (ConnectorOutput, error) {
	return decodeOutput[EntitlementRevokeOutput](raw)
}
func (EntitlementRevokeOutput) connectorOutput() {} // marker only, no behavior

type EntitlementReadInput struct {
	Tenant    ConnectorTenant `json:"tenant"`
	AccountID string          `json:"accountId"`
}

type EntitlementReadOutput struct {
	EntitlementIDs []string `json:"entitlementIds"`
}

func (EntitlementReadInput) Kind() ConnectorKind { return EntitlementKind }
func (EntitlementReadInput) Type() string        { return "read" }
func (EntitlementReadInput) DecodeOutput(raw []byte) (ConnectorOutput, error) {
	return decodeOutput[EntitlementReadOutput](raw)
}
func (EntitlementReadOutput) connectorOutput() {} // marker only, no behavior
