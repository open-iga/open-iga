package domain

import (
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
