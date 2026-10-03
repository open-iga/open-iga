package middleware

import (
	"reflect"
	"testing"

	"github.com/open-iga/core/internal/api/generated"
)

// TestOperationRolesMatchesGeneratedOperations keeps operationRoles in lockstep with
// the generated StrictServerInterface: every operation must have a policy, and every
// policy key must name a real operation. This catches a new endpoint added without a
// policy, and a stale key left behind after a rename/typo.
func TestOperationRolesMatchesGeneratedOperations(t *testing.T) {
	iface := reflect.TypeFor[generated.StrictServerInterface]()

	generatedOps := make(map[string]bool, iface.NumMethod())
	for i := 0; i < iface.NumMethod(); i++ {
		generatedOps[iface.Method(i).Name] = true
	}

	for op := range generatedOps {
		if _, ok := operationRoles[op]; !ok {
			t.Errorf("operation %q exists in generated.StrictServerInterface but has no entry in operationRoles", op)
		}
	}

	for op := range operationRoles {
		if !generatedOps[op] {
			t.Errorf("operationRoles has key %q with no matching operation in generated.StrictServerInterface", op)
		}
	}
}
