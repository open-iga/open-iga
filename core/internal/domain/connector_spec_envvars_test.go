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
}
