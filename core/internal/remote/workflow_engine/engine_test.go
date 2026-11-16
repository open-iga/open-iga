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

func step(name string, gate bool, after ...string) domain.WorkflowStep {
	return domain.WorkflowStep{
		Name:      name,
		Operation: domain.WorkflowOperation{System: "s", Resource: "r", Kind: domain.AccountActionKind},
		After:     after,
		Gate:      gate,
	}
}

func run(steps ...domain.WorkflowStep) domain.WorkflowRun {
	return domain.WorkflowRun{Definition: domain.WorkflowDefinition{Name: "onboarding", Type: domain.WorkflowTypeOnboarding, Steps: steps}, IdentityId: uuid.New()}
}

// Two lanes (aws, github): aws has a gated grant, github does not.
func buildTwoLanesRun() domain.WorkflowRun {
	return run(
		step("create-aws", false),
		step("grant-aws", true, "create-aws"),
		step("after-grant-aws", false, "grant-aws"),
		step("create-github", false),
		step("grant-github", false, "create-github"),
	)
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

	t.Run("a gate holds its lane only", func(t *testing.T) {
		recorded := &calls{}
		engine := launch(t, recorded.run)
		require.NoError(t, engine.Start(ctx, "gate-run", buildTwoLanesRun()))

		waitFor(t, engine, "gate-run", map[string]domain.WorkflowStepStatus{
			"create-aws": domain.WorkflowStepSucceeded, "grant-aws": domain.WorkflowStepGated, "after-grant-aws": domain.WorkflowStepPending,
			"create-github": domain.WorkflowStepSucceeded, "grant-github": domain.WorkflowStepSucceeded,
		})
		assert.Zero(t, recorded.count("grant-aws"))

		require.NoError(t, engine.Continue(ctx, "gate-run", "grant-aws"))
		waitFor(t, engine, "gate-run", map[string]domain.WorkflowStepStatus{
			"create-aws": domain.WorkflowStepSucceeded, "grant-aws": domain.WorkflowStepSucceeded, "after-grant-aws": domain.WorkflowStepSucceeded,
			"create-github": domain.WorkflowStepSucceeded, "grant-github": domain.WorkflowStepSucceeded,
		})
	})

	t.Run("a failure blocks its lane only", func(t *testing.T) {
		recorded := &calls{failing: map[string]bool{"create-aws": true}}
		engine := launch(t, recorded.run)
		failing := run(
			step("create-aws", false),
			step("grant-aws", false, "create-aws"),
			step("create-github", false),
			step("grant-github", false, "create-github"),
		)
		require.NoError(t, engine.Start(ctx, "fail-run", failing))

		waitFor(t, engine, "fail-run", map[string]domain.WorkflowStepStatus{
			"create-aws": domain.WorkflowStepFailed, "grant-aws": domain.WorkflowStepBlocked,
			"create-github": domain.WorkflowStepSucceeded, "grant-github": domain.WorkflowStepSucceeded,
		})
		assert.Zero(t, recorded.count("grant-aws"))
		assert.Equal(t, 1, recorded.count("create-aws"), "a failed step is not retried")
	})

	t.Run("a run waiting at a gate survives a restart", func(t *testing.T) {
		recorded := &calls{}
		first := NewEngine(databaseUrl, testutil.NewTestLogger())
		require.NoError(t, first.Launch(ctx, recorded.run))
		require.NoError(t, first.Start(ctx, "restart-run", buildTwoLanesRun()))
		waitFor(t, first, "restart-run", map[string]domain.WorkflowStepStatus{
			"create-aws": domain.WorkflowStepSucceeded, "grant-aws": domain.WorkflowStepGated, "after-grant-aws": domain.WorkflowStepPending,
			"create-github": domain.WorkflowStepSucceeded, "grant-github": domain.WorkflowStepSucceeded,
		})
		first.Shutdown(5 * time.Second) // the restart

		second := launch(t, recorded.run)
		require.Eventually(t, func() bool { return second.Continue(ctx, "restart-run", "grant-aws") == nil }, 15*time.Second, 100*time.Millisecond)
		waitFor(t, second, "restart-run", map[string]domain.WorkflowStepStatus{
			"create-aws": domain.WorkflowStepSucceeded, "grant-aws": domain.WorkflowStepSucceeded, "after-grant-aws": domain.WorkflowStepSucceeded,
			"create-github": domain.WorkflowStepSucceeded, "grant-github": domain.WorkflowStepSucceeded,
		})
		assert.Equal(t, 1, recorded.count("create-aws"), "a finished step is not repeated after the restart")
	})

	t.Run("rejects an invalid run", func(t *testing.T) {
		engine := launch(t, (&calls{}).run)
		invalid := run(domain.WorkflowStep{Name: "a", Operation: domain.WorkflowOperation{}, After: nil, Gate: false})
		assert.ErrorIs(t, engine.Start(ctx, "bad-run", invalid), domain.ErrInvalidWorkflow)
	})

	t.Run("continuing a step that is not at a gate is refused", func(t *testing.T) {
		engine := launch(t, (&calls{}).run)
		require.NoError(t, engine.Start(ctx, "plain-run", run(step("create-aws", false))))
		waitFor(t, engine, "plain-run", map[string]domain.WorkflowStepStatus{"create-aws": domain.WorkflowStepSucceeded})
		assert.ErrorIs(t, engine.Continue(ctx, "plain-run", "create-aws"), domain.ErrWorkflowStepNotGated)
	})

	t.Run("an unknown run is not found", func(t *testing.T) {
		engine := launch(t, (&calls{}).run)
		_, err := engine.Get(ctx, "no-such-run")
		assert.ErrorIs(t, err, domain.ErrWorkflowRunNotFound)
	})
}
