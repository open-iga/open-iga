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

func setupConnectorServiceWithMocks(t *testing.T) (*ConnectorService, *testutil.MockConnectorRuntime) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	runtime := testutil.NewMockConnectorRuntime(ctrl)
	return NewConnectorService(runtime, testutil.NewTestLogger()), runtime
}

func TestConnectorService_ValidateByUrl(t *testing.T) {
	t.Run("returns an id and records success once validation completes", func(t *testing.T) {
		cs, runtime := setupConnectorServiceWithMocks(t)
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
		cs, runtime := setupConnectorServiceWithMocks(t)
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

func TestConnectorService_GetOnboardedConnectorDetails(t *testing.T) {
	t.Run("returns false for an unknown id", func(t *testing.T) {
		cs, _ := setupConnectorServiceWithMocks(t)

		res, ok := cs.GetOnboardedConnectorDetails(context.TODO(), "does-not-exist")

		assert.False(t, ok)
		assert.Nil(t, res)
	})
}
