package runner

import (
	"testing"
	"time"

	"github.com/adriancodes/gumloop/internal/agent"
	"github.com/adriancodes/gumloop/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	cfg := &config.Config{
		CLI:            "claude",
		StuckThreshold: 3,
	}

	mockAgent := &agent.Agent{
		ID:   "test-agent",
		Name: "Test Agent",
	}

	t.Run("creates runner in single-run mode", func(t *testing.T) {
		r := New(cfg, "test prompt", mockAgent, false, 0, nil)

		assert.NotNil(t, r)
		assert.Equal(t, cfg, r.config)
		assert.Equal(t, "test prompt", r.prompt)
		assert.Equal(t, mockAgent, r.agent)
		assert.True(t, r.singleRun)
		assert.Equal(t, 0, r.maxIters)
		assert.NotNil(t, r.metrics)
		assert.Equal(t, 0, r.metrics.Iterations)
		assert.Equal(t, 0, r.metrics.Commits)
	})

	t.Run("creates runner in choo-choo mode", func(t *testing.T) {
		r := New(cfg, "test prompt", mockAgent, true, 10, nil)

		assert.NotNil(t, r)
		assert.False(t, r.singleRun)
		assert.Equal(t, 10, r.maxIters)
	})

	t.Run("creates runner with unlimited iterations", func(t *testing.T) {
		r := New(cfg, "test prompt", mockAgent, true, 0, nil)

		assert.NotNil(t, r)
		assert.False(t, r.singleRun)
		assert.Equal(t, 0, r.maxIters) // 0 means unlimited
	})
}

func TestGetMetrics(t *testing.T) {
	cfg := &config.Config{
		CLI:            "claude",
		StuckThreshold: 3,
	}

	mockAgent := &agent.Agent{
		ID:   "test-agent",
		Name: "Test Agent",
	}
	r := New(cfg, "test prompt", mockAgent, true, 0, nil)

	// Initially should have zero metrics
	metrics := r.GetMetrics()
	assert.NotNil(t, metrics)
	assert.Equal(t, 0, metrics.Iterations)
	assert.Equal(t, 0, metrics.Commits)
	assert.True(t, metrics.Duration() >= 0)
	assert.True(t, metrics.Duration() < 100*time.Millisecond) // Should be very recent

	// Simulate running
	time.Sleep(10 * time.Millisecond)
	r.metrics.Iterations = 5
	r.metrics.Commits = 3

	metrics = r.GetMetrics()
	assert.Equal(t, 5, metrics.Iterations)
	assert.Equal(t, 3, metrics.Commits)
	assert.True(t, metrics.Duration() >= 10*time.Millisecond)
}

func TestExitCodes(t *testing.T) {
	// Verify exit code values match spec
	assert.Equal(t, ExitCode(0), ExitSuccess)
	assert.Equal(t, ExitCode(1), ExitError)
	assert.Equal(t, ExitCode(2), ExitSafety)
	assert.Equal(t, ExitCode(3), ExitMaxIterations)
	assert.Equal(t, ExitCode(4), ExitStuck)
	assert.Equal(t, ExitCode(130), ExitInterrupt)
}

// Note: Run() method integration tests will be added in CMD-005
// after iteration execution is implemented. For now, we verify
// that the runner structure is correct.
