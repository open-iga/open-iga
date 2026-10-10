package application

import (
	"context"
	"errors"
	"testing"

	"github.com/open-iga/core/internal/domain"
	"github.com/open-iga/core/internal/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func setupManagedSystemServiceWithMocks(t *testing.T) (*ManagedSystemService, *ConnectorService, *testutil.MockManagedSystemRepository) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := testutil.NewMockManagedSystemRepository(ctrl)
	// runtime is unused by the Onboard path (it only reads already-recorded results)
	connectorService := NewConnectorService(nil, testutil.NewTestLogger())
	svc := NewManagedSystemService(testutil.NewTestLogger(), connectorService, repo)

	return svc, connectorService, repo
}

// seedConnector records a validation result directly so Onboard can read it.
func seedConnector(cs *ConnectorService, id string, result domain.ConnectorValidationResult) {
	cs.mu.Lock()
	cs.onboardingResults[id] = result
	cs.mu.Unlock()
}

func TestManagedSystemService_Onboard(t *testing.T) {
	t.Run("returns ErrConnectorNotFound for an unknown connector", func(t *testing.T) {
		svc, _, _ := setupManagedSystemServiceWithMocks(t)

		ms, err := svc.Onboard(context.TODO(), "unknown")

		assert.Nil(t, ms)
		assert.ErrorIs(t, err, domain.ErrConnectorNotFound)
	})

	t.Run("returns ErrConnectorNotOnboardable when the connector has not passed validation", func(t *testing.T) {
		svc, cs, _ := setupManagedSystemServiceWithMocks(t)
		seedConnector(cs, "pending-id", domain.ConnectorValidationResult{Status: domain.ConnectorValidationPending})

		ms, err := svc.Onboard(context.TODO(), "pending-id")

		assert.Nil(t, ms)
		assert.ErrorIs(t, err, domain.ErrConnectorNotOnboardable)
	})

	t.Run("creates a managed system when the connector passed validation", func(t *testing.T) {
		svc, cs, repo := setupManagedSystemServiceWithMocks(t)
		seedConnector(cs, "ok-id", domain.ConnectorValidationResult{
			Status:        domain.ConnectorValidationSuccess,
			ConnectorUrl:  "https://conn",
			ConnectorHash: "hash",
			ConnectorSpec: &domain.ConnectorSpec{Name: "aws"},
		})
		want := &domain.ManagedSystem{Id: "ms-1", Name: "aws"}
		repo.EXPECT().CreateManagedSystem(gomock.Any(), "aws", "https://conn", "hash", &domain.ConnectorSpec{Name: "aws"}).Return(want, nil)

		ms, err := svc.Onboard(context.TODO(), "ok-id")

		assert.NoError(t, err)
		assert.Equal(t, want, ms)
	})

	t.Run("propagates repository errors", func(t *testing.T) {
		svc, cs, repo := setupManagedSystemServiceWithMocks(t)
		seedConnector(cs, "ok-id", domain.ConnectorValidationResult{
			Status:        domain.ConnectorValidationSuccess,
			ConnectorUrl:  "https://conn",
			ConnectorHash: "hash",
			ConnectorSpec: &domain.ConnectorSpec{Name: "aws"},
		})
		repo.EXPECT().CreateManagedSystem(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("db down"))

		ms, err := svc.Onboard(context.TODO(), "ok-id")

		assert.Nil(t, ms)
		assert.ErrorContains(t, err, "db down")
	})
}

func TestManagedSystemService_ListManagedSystems(t *testing.T) {
	t.Run("returns the managed systems from the repository", func(t *testing.T) {
		svc, _, repo := setupManagedSystemServiceWithMocks(t)
		want := []*domain.ManagedSystem{{Id: "a", Name: "aws"}, {Id: "b", Name: "gcp"}}
		repo.EXPECT().ListManagedSystems(gomock.Any()).Return(want, nil)

		got, err := svc.ListManagedSystems(context.TODO())

		assert.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("propagates repository errors", func(t *testing.T) {
		svc, _, repo := setupManagedSystemServiceWithMocks(t)
		repo.EXPECT().ListManagedSystems(gomock.Any()).Return(nil, errors.New("db down"))

		got, err := svc.ListManagedSystems(context.TODO())

		assert.Nil(t, got)
		assert.ErrorContains(t, err, "db down")
	})
}
