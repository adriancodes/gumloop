package runner

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewMetrics(t *testing.T) {
	m := NewMetrics()

	assert.NotNil(t, m)
	assert.Equal(t, 0, m.Iterations)
	assert.Equal(t, 0, m.Commits)
	assert.Equal(t, "", m.ExitReason)
	assert.True(t, m.StartTime.Before(time.Now().Add(time.Second)))
	assert.True(t, m.StartTime.After(time.Now().Add(-time.Second)))
}

func TestMetricsDuration(t *testing.T) {
	m := NewMetrics()

	// Should be very recent
	assert.True(t, m.Duration() < 100*time.Millisecond)

	// Wait and check again
	time.Sleep(10 * time.Millisecond)
	assert.True(t, m.Duration() >= 10*time.Millisecond)
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "seconds only",
			duration: 45 * time.Second,
			expected: "45s",
		},
		{
			name:     "minutes and seconds",
			duration: 2*time.Minute + 30*time.Second,
			expected: "2m 30s",
		},
		{
			name:     "minutes only",
			duration: 3 * time.Minute,
			expected: "3m 0s",
		},
		{
			name:     "hours, minutes, and seconds",
			duration: 2*time.Hour + 15*time.Minute + 45*time.Second,
			expected: "2h 15m 45s",
		},
		{
			name:     "hours only",
			duration: 5 * time.Hour,
			expected: "5h 0m 0s",
		},
		{
			name:     "less than a second",
			duration: 500 * time.Millisecond,
			expected: "0s", // Truncates to 0
		},
		{
			name:     "zero duration",
			duration: 0,
			expected: "0s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExitReasonString(t *testing.T) {
	tests := []struct {
		code     ExitCode
		expected string
	}{
		{ExitSuccess, "✅ Complete (no changes)"},
		{ExitError, "❌ Error"},
		{ExitSafety, "⛔ Safety refusal"},
		{ExitMaxIterations, "🔄 Max iterations reached"},
		{ExitStuck, "🔁 Stuck (no commits)"},
		{ExitInterrupt, "⚠️  Interrupted"},
		{ExitCode(99), "Unknown exit code: 99"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := ExitReasonString(tt.code)
			assert.Equal(t, tt.expected, result)
		})
	}
}
