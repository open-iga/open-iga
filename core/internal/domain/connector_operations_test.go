package domain

import (
	"encoding/json"
	"testing"

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
		{"create reports created=false", AccountCreateInput{}, `{"id":"arn:1","created":false}`, AccountCreateOutput{ID: "arn:1", Created: false}},
		{"list returns a page with a cursor", AccountListInput{}, `{"accounts":[{"id":"a","email":"a@b.co","enabled":true}],"nextCursor":"p2"}`,
			AccountListOutput{Accounts: []ListedAccount{{ID: "a", Email: "a@b.co", Enabled: true}}, NextCursor: "p2"}},
		{"discover returns the last page without a cursor", EntitlementDiscoverInput{}, `{"entitlements":[{"entitlementId":"g1","name":"G1"}]}`,
			EntitlementDiscoverOutput{Entitlements: []DiscoveredEntitlement{{EntitlementID: "g1", Name: "G1"}}}},
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

	nextPage, err := json.Marshal(EntitlementDiscoverInput{Tenant: ConnectorTenant{ID: "t1"}, Cursor: "p2"})
	require.NoError(t, err)
	assert.JSONEq(t, `{"tenant":{"id":"t1"},"cursor":"p2"}`, string(nextPage))
}
