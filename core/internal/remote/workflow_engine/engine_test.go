package workflow_engine

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/open-iga/core/internal/domain"
	"github.com/open-iga/core/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

// calls records which steps ran. A step fails when its name is in failing.
type calls struct {
	mutex   sync.Mutex
	names   []string
	failing map[string]bool
}

func (c *calls) run(_ context.Context, step domain.WorkflowStep, _ uuid.UUID) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.names = append(c.names, step.Name)
	if c.failing[step.Name] {
		return errors.New("boom")
	}
	return nil
}

func (c *calls) count(name string) int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	count := 0
	for _, recorded := range c.names {
		if recorded == name {
			count++
		}
	}
	return count
}

func step(name string, after ...string) domain.WorkflowStep {
	return domain.WorkflowStep{
		Name:      name,
		Operation: domain.WorkflowOperation{System: "s", Resource: "r", Kind: domain.AccountActionKind},
		After:     after,
	}
}

func run(steps ...domain.WorkflowStep) domain.WorkflowRun {
	return domain.WorkflowRun{Definition: domain.WorkflowDefinition{Name: "test", Steps: steps}, IdentityId: uuid.New()}
}

func getStepStatuses(state domain.WorkflowRunState) map[string]domain.WorkflowStepStatus {
	result := map[string]domain.WorkflowStepStatus{}
	for _, step := range state.Steps {
		result[step.Name] = step.Status
	}
	return result
}

func TestEngine(t *testing.T) {
	ctx := context.Background()
	container, err := postgres.Run(ctx, "postgres:18.3-alpine3.23",
		postgres.WithDatabase("open_iga"), postgres.WithUsername("test"), postgres.WithPassword("test"), postgres.BasicWaitStrategies())
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(ctx) })
	databaseUrl, err := container.ConnectionString(ctx)
	require.NoError(t, err)

	launch := func(t *testing.T, runner domain.WorkflowStepRunner) *Engine {
		engine := NewEngine(databaseUrl, testutil.NewTestLogger())
		require.NoError(t, engine.Launch(ctx, runner))
		t.Cleanup(func() { engine.Shutdown(5 * time.Second) })
		return engine
	}
	waitFor := func(t *testing.T, engine *Engine, runId string, want map[string]domain.WorkflowStepStatus) {
		require.Eventually(t, func() bool {
			state, err := engine.Get(ctx, runId)
			return err == nil && assert.ObjectsAreEqual(want, getStepStatuses(state))
		}, 15*time.Second, 100*time.Millisecond)
	}

	t.Run("every step of a run succeeds, each after the steps it follows", func(t *testing.T) {
		recorded := &calls{}
		engine := launch(t, recorded.run)
		require.NoError(t, engine.Start(ctx, "success-run", run(
			step("create-aws"),
			step("disable-aws", "create-aws"),
			step("create-github"),
		)))

		waitFor(t, engine, "success-run", map[string]domain.WorkflowStepStatus{
			"create-aws": domain.WorkflowStepSucceeded, "disable-aws": domain.WorkflowStepSucceeded, "create-github": domain.WorkflowStepSucceeded,
		})
	})

	t.Run("a failure blocks its lane only", func(t *testing.T) {
		recorded := &calls{failing: map[string]bool{"create-aws": true}}
		engine := launch(t, recorded.run)
		failing := run(
			step("create-aws"),
			step("grant-aws", "create-aws"),
			step("create-github"),
			step("grant-github", "create-github"),
		)
		require.NoError(t, engine.Start(ctx, "fail-run", failing))

		waitFor(t, engine, "fail-run", map[string]domain.WorkflowStepStatus{
			"create-aws": domain.WorkflowStepFailed, "grant-aws": domain.WorkflowStepBlocked,
			"create-github": domain.WorkflowStepSucceeded, "grant-github": domain.WorkflowStepSucceeded,
		})
		assert.Zero(t, recorded.count("grant-aws"))
		assert.Equal(t, 1, recorded.count("create-aws"), "a failed step is not retried")
	})

	t.Run("rejects an invalid run", func(t *testing.T) {
		engine := launch(t, (&calls{}).run)
		invalid := run(domain.WorkflowStep{Name: "a", Operation: domain.WorkflowOperation{}, After: nil})
		assert.ErrorIs(t, engine.Start(ctx, "bad-run", invalid), domain.ErrInvalidWorkflow)
	})

	t.Run("an unknown run is not found", func(t *testing.T) {
		engine := launch(t, (&calls{}).run)
		_, err := engine.Get(ctx, "no-such-run")
		assert.ErrorIs(t, err, domain.ErrWorkflowRunNotFound)
	})
}
