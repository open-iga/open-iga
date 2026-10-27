package domain

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func endpoint(method, url string) Endpoint { return Endpoint{Method: method, URL: url} }

func specWith(actions map[string]AccountActions, ents map[string]UserEntitlements) *ConnectorSpec {
	return &ConnectorSpec{Actions: actions, Entitlements: ents}
}

func TestResolveEndpoints(t *testing.T) {
	spec := specWith(
		map[string]AccountActions{
			"iam-user": {Create: &Operation{Endpoints: []Endpoint{
				endpoint("post", "https://sts.{{AWS_REGION}}.amazonaws.com/"),
				endpoint("POST", "https://iam.amazonaws.com/"),
			}}},
		},
		map[string]UserEntitlements{
			"iam-user-group": {Grant: &Operation{Endpoints: []Endpoint{
				endpoint("POST", "https://sts.{{AWS_REGION}}.amazonaws.com/"),
			}}},
		},
	)

	t.Run("substitutes vars and keys endpoints by kind.resource.operation", func(t *testing.T) {
		t.Setenv("AWS_REGION", "us-east-1")

		resolved := spec.ResolveEndpoints(slog.Default())

		assert.Equal(t, map[string][]ResolvedEndpoint{
			"account-action.iam-user.create": {
				{Method: "POST", URL: "https://sts.us-east-1.amazonaws.com/"},
				{Method: "POST", URL: "https://iam.amazonaws.com/"},
			},
			"entitlement.iam-user-group.grant": {
				{Method: "POST", URL: "https://sts.us-east-1.amazonaws.com/"},
			},
		}, resolved)
	})

	t.Run("OperationKey separates the same resource and operation under both kinds", func(t *testing.T) {
		assert.NotEqual(t, OperationKey(AccountActionKind, "iam", "read"), OperationKey(EntitlementKind, "iam", "read"))
	})
}

func TestScopedConfigs(t *testing.T) {
	required := true
	spec := &ConnectorSpec{
		Config: []Config{{Name: "AWS_REGION", Required: &required}},
		Actions: map[string]AccountActions{
			"iam-user": {Create: &Operation{Config: []Config{{Name: "AWS_USER_MANAGEMENT_ROLE", Required: &required}}}},
		},
		Entitlements: map[string]UserEntitlements{
			"iam-user-group": {Grant: &Operation{Config: []Config{{Name: "AWS_ENTITLEMENT_ROLE", Required: &required}}}},
		},
	}

	t.Run("each operation gets module config plus only its own config", func(t *testing.T) {
		t.Setenv("AWS_REGION", "us-east-1")
		t.Setenv("AWS_USER_MANAGEMENT_ROLE", "arn:user")
		t.Setenv("AWS_ENTITLEMENT_ROLE", "arn:ent")

		configs := spec.ScopedConfigs()

		// An operation's role is scoped out of the other kind's config.
		assert.Equal(t, map[string]map[string]string{
			"account-action.iam-user.create": {
				"AWS_REGION":               "us-east-1",
				"AWS_USER_MANAGEMENT_ROLE": "arn:user",
			},
			"entitlement.iam-user-group.grant": {
				"AWS_REGION":           "us-east-1",
				"AWS_ENTITLEMENT_ROLE": "arn:ent",
			},
		}, configs)
	})

	t.Run("omits keys absent from the environment", func(t *testing.T) {
		t.Setenv("AWS_REGION", "us-east-1")
		os.Unsetenv("AWS_USER_MANAGEMENT_ROLE")

		configs := spec.ScopedConfigs()

		assert.Equal(t, map[string]string{"AWS_REGION": "us-east-1"}, configs["account-action.iam-user.create"])
	})
}
