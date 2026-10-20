package domain

import (
	"encoding/json"
	"slices"
	"testing"
)

func TestCollectEnvVars(t *testing.T) {
	spec := ConnectorSpec{
		Actions: map[string]AccountActions{
			"user": {
				Create: &Operation{Endpoints: []Endpoint{{URL: "https://{{AWS_ENDPOINT}}/users/{{ AWS-ENDPOINT }}"}}},
			},
		},
		Entitlements: map[string]UserEntitlements{
			"role": {
				Read: &Operation{Endpoints: []Endpoint{{URL: "https://{{AWS_ENDPOINT}}/roles"}}},
			},
		},
	}

	got := spec.collectEnvVarsFromTemplateUrl()
	for _, want := range []string{"AWS_ENDPOINT", "AWS-ENDPOINT"} {
		if !slices.Contains(got, want) {
			t.Fatalf("expected env var %q in %v", want, got)
		}
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 unique env vars, got %v", got)
	}
}

func TestMissingRequiredEnvVars(t *testing.T) {
	required, optional := true, false
	spec := ConnectorSpec{
		EnvVars: []string{"AWS_ENDPOINT", "AWS_REGION", "AWS_TOKEN"},
		Config: []Config{
			{Name: "AWS_ENDPOINT", Required: &required}, // required, not set -> missing
			{Name: "AWS_REGION", Required: &optional},   // optional, not set -> allowed
			{Name: "AWS_TOKEN", Required: &required},    // required and set -> allowed
		},
	}

	t.Setenv("AWS_TOKEN", "set") // required and present -> allowed

	got := spec.MissingRequiredEnvVars()

	if !slices.Equal(got, []string{"AWS_ENDPOINT"}) {
		t.Fatalf("expected only AWS_ENDPOINT missing, got %v", got)
	}
}

// EnvVars is not persisted (json:"-"), so a spec reloaded from storage must recompute it
// from the endpoint URLs on unmarshal — otherwise the startup env check is a silent no-op.
func TestConnectorSpec_UnmarshalJSON_RecomputesEnvVars(t *testing.T) {
	stored := `{
		"actions": {"user": {"create": {"endpoints": [{"url": "https://{{AWS_ENDPOINT}}/users"}]}}},
		"entitlements": {"role": {"read": {"endpoints": [{"url": "https://{{AWS_REGION}}/roles"}]}}}
	}`

	var spec ConnectorSpec
	if err := json.Unmarshal([]byte(stored), &spec); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if !slices.Contains(spec.EnvVars, "AWS_ENDPOINT") || !slices.Contains(spec.EnvVars, "AWS_REGION") {
		t.Fatalf("expected EnvVars recomputed from URLs, got %v", spec.EnvVars)
	}
}
