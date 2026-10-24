package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/open-iga/core/internal/domain"
	"github.com/open-iga/core/internal/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func setupConnectorServiceWithMocks(t *testing.T) (*ConnectorService, *testutil.MockConnectorRuntime, *testutil.MockManagedSystemRepository) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	runtime := testutil.NewMockConnectorRuntime(ctrl)
	repo := testutil.NewMockManagedSystemRepository(ctrl)
	return NewConnectorService(runtime, repo, testutil.NewTestLogger()), runtime, repo
}

func TestConnectorService_ValidateByUrl(t *testing.T) {
	t.Run("returns an id and records success once validation completes", func(t *testing.T) {
		cs, runtime, _ := setupConnectorServiceWithMocks(t)
		spec := &domain.ConnectorSpec{Name: "aws"}
		runtime.EXPECT().ValidateConnectorByURL(gomock.Any(), "https://conn", "hash").Return(spec, nil)

		id := cs.ValidateByUrl(context.TODO(), "https://conn", "hash")
		assert.NotEmpty(t, id)

		// validation runs in a goroutine; wait for the terminal state
		assert.Eventually(t, func() bool {
			res, ok := cs.GetOnboardedConnectorDetails(context.TODO(), id)
			return ok && res.Status == domain.ConnectorValidationSuccess
		}, time.Second, 5*time.Millisecond)

		res, _ := cs.GetOnboardedConnectorDetails(context.TODO(), id)
		assert.Equal(t, spec, res.ConnectorSpec)
		assert.NoError(t, res.Error)
	})

	t.Run("records failure when validation errors", func(t *testing.T) {
		cs, runtime, _ := setupConnectorServiceWithMocks(t)
		runtime.EXPECT().ValidateConnectorByURL(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("boom"))

		id := cs.ValidateByUrl(context.TODO(), "https://conn", "hash")

		assert.Eventually(t, func() bool {
			res, ok := cs.GetOnboardedConnectorDetails(context.TODO(), id)
			return ok && res.Status == domain.ConnectorValidationFailed
		}, time.Second, 5*time.Millisecond)

		res, _ := cs.GetOnboardedConnectorDetails(context.TODO(), id)
		assert.ErrorContains(t, res.Error, "boom")
		assert.Nil(t, res.ConnectorSpec)
	})
}

func TestConnectorService_LoadConnectorsOnStartup(t *testing.T) {
	t.Run("verifies connectors and skips no-spec / bad-hash without erroring", func(t *testing.T) {
		cs, runtime, repo := setupConnectorServiceWithMocks(t)
		required := true
		t.Setenv("OPEN_IGA_LOAD_ENV", "set")

		okSpec := &domain.ConnectorSpec{
			Name:   "ok",
			Config: []domain.Config{{Name: "OPEN_IGA_LOAD_ENV", Required: &required}},
		}

		systems := []*domain.ManagedSystem{
			{Id: "ok-id", Name: "ok", ConnectorUrl: "https://ok", ConnectorHash: "h1", ConnectorSpec: okSpec},
			{Id: "bad-hash-id", Name: "bad-hash", ConnectorUrl: "https://bad", ConnectorHash: "h2", ConnectorSpec: &domain.ConnectorSpec{Name: "bad"}},
			{Id: "no-spec-id", Name: "no-spec", ConnectorUrl: "https://ns", ConnectorHash: "h4", ConnectorSpec: nil},
		}

		repo.EXPECT().ListManagedSystems(gomock.Any()).Return(systems, nil)
		// no-spec is skipped before validation; ok and bad-hash are validated (bad-hash skipped, not fatal).
		runtime.EXPECT().ValidateConnectorByURL(gomock.Any(), "https://ok", "h1").Return(okSpec, nil)
		runtime.EXPECT().ValidateConnectorByURL(gomock.Any(), "https://bad", "h2").Return(nil, errors.New("connector hash mismatch"))

		assert.NoError(t, cs.LoadConnectorsOnStartup(context.TODO()))
	})

	t.Run("aborts startup when a required env var is missing", func(t *testing.T) {
		cs, _, repo := setupConnectorServiceWithMocks(t)
		required := true
		spec := &domain.ConnectorSpec{
			Name:   "missing-env",
			Config: []domain.Config{{Name: "OPEN_IGA_LOAD_UNSET_ENV", Required: &required}},
		}
		repo.EXPECT().ListManagedSystems(gomock.Any()).Return([]*domain.ManagedSystem{
			{Id: "me-id", Name: "missing-env", ConnectorUrl: "https://me", ConnectorHash: "h3", ConnectorSpec: spec},
		}, nil)

		err := cs.LoadConnectorsOnStartup(context.TODO())

		assert.ErrorContains(t, err, "OPEN_IGA_LOAD_UNSET_ENV")
	})
}

func TestConnectorService_GetOnboardedConnectorDetails(t *testing.T) {
	t.Run("returns false for an unknown id", func(t *testing.T) {
		cs, _, _ := setupConnectorServiceWithMocks(t)

		res, ok := cs.GetOnboardedConnectorDetails(context.TODO(), "does-not-exist")

		assert.False(t, ok)
		assert.Nil(t, res)
	})

	t.Run("computes missing required env vars against the core environment on read", func(t *testing.T) {
		cs, _, _ := setupConnectorServiceWithMocks(t)
		required := true
		cs.onboardingResults["id"] = domain.ConnectorValidationResult{
			Status: domain.ConnectorValidationSuccess,
			ConnectorSpec: &domain.ConnectorSpec{
				Config: []domain.Config{{Name: "OPEN_IGA_TEST_UNSET_ENV", Required: &required}},
			},
		}

		res, ok := cs.GetOnboardedConnectorDetails(context.TODO(), "id")

		assert.True(t, ok)
		assert.Equal(t, []string{"OPEN_IGA_TEST_UNSET_ENV"}, res.MissingEnvVars)
	})
}
