package domain

import (
	"slices"
	"testing"
)

func TestMissingRequiredEnvVars(t *testing.T) {
	required, optional := true, false
	spec := ConnectorSpec{
		Config: []Config{
			{Name: "OPEN_IGA_TEST_ENDPOINT", Required: &required}, // required, not set -> missing
			{Name: "OPEN_IGA_TEST_REGION", Required: &optional},   // optional, not set -> allowed
			{Name: "OPEN_IGA_TEST_TOKEN", Required: &required},    // required and set -> allowed
			{Name: "OPEN_IGA_TEST_EMPTY", Required: &required},    // required, set but empty -> missing
		},
	}

	t.Setenv("OPEN_IGA_TEST_TOKEN", "set") // required and present -> allowed
	t.Setenv("OPEN_IGA_TEST_EMPTY", "")    // empty counts as unset

	got := spec.MissingRequiredEnvVars()

	if !slices.Equal(got, []string{"OPEN_IGA_TEST_ENDPOINT", "OPEN_IGA_TEST_EMPTY"}) {
		t.Fatalf("expected ENDPOINT and EMPTY missing, got %v", got)
	}

	t.Run("operation-level config", func(t *testing.T) {
		spec := ConnectorSpec{Actions: map[string]AccountActions{
			"user": {Create: &Operation{Config: []Config{
				{Name: "OPEN_IGA_TEST_ROLE_ARN", Required: &required},
				{Name: "OPEN_IGA_TEST_ROLE_OPT", Required: &optional},
			}}},
		}}
		if got := spec.MissingRequiredEnvVars(); !slices.Equal(got, []string{"OPEN_IGA_TEST_ROLE_ARN"}) {
			t.Fatalf("expected ROLE_ARN missing, got %v", got)
		}
	})
}

func TestConfigRejectsAppEnvKey(t *testing.T) {
	required := true
	if err := (Config{Name: "DATABASE_URL", Description: "x", Required: &required}).Validate(); err == nil {
		t.Fatal("expected app env key to be rejected")
	}
	if err := (Config{Name: "AWS_REGION", Description: "x", Required: &required}).Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
