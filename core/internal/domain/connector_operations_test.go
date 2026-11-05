package domain

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectorOperationsDecodeSdkOutputs(t *testing.T) {
	tests := []struct {
		name      string
		operation ConnectorOperation
		raw       string
		expected  ConnectorOutput
	}{
		{"create reports created=false", AccountCreateInput{}, `{"accountId":"arn:1","created":false}`, AccountCreateOutput{AccountID: "arn:1", Created: false}},
		{
			"list returns a page with a cursor",
			AccountListInput{},
			`{"accounts":[{"accountId":"a","email":"a@b.co","enabled":true}],"nextCursor":"p2"}`,
			AccountListOutput{Accounts: []ListedAccount{{AccountID: "a", Email: "a@b.co", Enabled: true}}, NextCursor: "p2"},
		},
		{
			"discover returns the last page without a cursor",
			EntitlementDiscoverInput{},
			`{"entitlements":[{"entitlementId":"g1","name":"G1"}]}`,
			EntitlementDiscoverOutput{Entitlements: []DiscoveredEntitlement{{EntitlementID: "g1", Name: "G1"}}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output, err := test.operation.DecodeOutput([]byte(test.raw))
			require.NoError(t, err)
			assert.Equal(t, test.expected, output)
		})
	}
}

func TestConnectorInputsEncodeTenantAndOptionalCursor(t *testing.T) {
	firstPage, err := json.Marshal(AccountListInput{Tenant: ConnectorTenant{ID: "t1"}})
	require.NoError(t, err)
	assert.JSONEq(t, `{"tenant":{"id":"t1"}}`, string(firstPage))

	nextCursor := "p2"
	nextPage, err := json.Marshal(EntitlementDiscoverInput{Tenant: ConnectorTenant{ID: "t1"}, Cursor: &nextCursor})
	require.NoError(t, err)
	assert.JSONEq(t, `{"tenant":{"id":"t1"},"cursor":"p2"}`, string(nextPage))
}

func TestNewConnectorOperation(t *testing.T) {
	data := OperationData{
		Tenant:      &ManagedSystemTenant{ExternalId: "acct-1"},
		Identity:    &Identity{Email: "a@b.c", FirstName: "Ada", LastName: "L"},
		Account:     &Account{ExternalId: "arn:u1"},
		Entitlement: &Entitlement{Name: "g1"},
	}

	create, err := NewConnectorOperation(AccountActionKind, "create", data)
	assert.NoError(t, err)
	assert.Equal(t, AccountCreateInput{Tenant: ConnectorTenant{ID: "acct-1"}, Email: "a@b.c", Firstname: "Ada", Lastname: "L"}, create)

	grant, err := NewConnectorOperation(EntitlementKind, "grant", data)
	assert.NoError(t, err)
	assert.Equal(t, EntitlementGrantInput{Tenant: ConnectorTenant{ID: "acct-1"}, AccountID: "arn:u1", EntitlementID: "g1"}, grant)

	_, err = NewConnectorOperation(AccountActionKind, "grant", data)
	assert.ErrorIs(t, err, ErrUnknownOperation)
}

func TestNewConnectorOperation_NeedsProvisionedAccount(t *testing.T) {
	tenant := &ManagedSystemTenant{ExternalId: "acct-1"}

	_, err := NewConnectorOperation(AccountActionKind, "enable", OperationData{Tenant: tenant})
	assert.ErrorIs(t, err, ErrMissingOperationData)

	_, err = NewConnectorOperation(AccountActionKind, "enable", OperationData{Tenant: tenant, Account: &Account{}})
	assert.ErrorIs(t, err, ErrMissingOperationData)
}

func TestOperationRefNeeds(t *testing.T) {
	ref := OperationRef{IdentityId: uuid.New(), EntitlementId: uuid.New()}

	assert.False(t, ref.NeedsAccount(AccountActionKind, "create"))
	assert.False(t, ref.NeedsAccount(EntitlementKind, "discover"))
	assert.True(t, ref.NeedsAccount(AccountActionKind, "enable"))
	assert.True(t, ref.NeedsAccount(EntitlementKind, "grant"))
	assert.False(t, OperationRef{}.NeedsAccount(AccountActionKind, "enable"))

	assert.True(t, ref.NeedsEntitlement(EntitlementKind, "grant"))
	assert.False(t, ref.NeedsEntitlement(EntitlementKind, "read"))
	assert.False(t, OperationRef{}.NeedsEntitlement(EntitlementKind, "grant"))
}

func TestDecodeOutputRejectsEmptyOutput(t *testing.T) {
	for _, raw := range []string{`null`, `{}`, ``} {
		_, err := AccountEnableInput{}.DecodeOutput([]byte(raw))
		assert.Error(t, err, "raw %q", raw)
	}

	// An empty account id would be saved as a provisioned account, so create refuses it.
	for _, raw := range []string{`{"created":true}`, `{"accountId":"","created":true}`} {
		_, err := AccountCreateInput{}.DecodeOutput([]byte(raw))
		assert.ErrorIs(t, err, ErrEmptyOutput, "raw %q", raw)
	}

	// false is a real answer, not an empty output.
	output, err := AccountEnableInput{}.DecodeOutput([]byte(`{"enabled":false}`))
	require.NoError(t, err)
	assert.Equal(t, AccountEnableOutput{Enabled: false}, output)
}
