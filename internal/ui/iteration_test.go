package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderIterationHeader(t *testing.T) {
	tests := []struct {
		name     string
		cfg      IterationConfig
		contains []string
	}{
		{
			name: "with max iterations",
			cfg: IterationConfig{
				Number:       3,
				MaxIteration: 20,
				Timestamp:    time.Date(2024, 1, 1, 14, 32, 15, 0, time.UTC),
				CLI:          "claude",
			},
			contains: []string{
				"ITERATION 3 of 20",
				"14:32:15",
				"claude",
				"🚂",
				"═", // Double separator
			},
		},
		{
			name: "without max iterations (unlimited)",
			cfg: IterationConfig{
				Number:       5,
				MaxIteration: 0,
				Timestamp:    time.Date(2024, 1, 1, 9, 15, 30, 0, time.UTC),
				CLI:          "codex",
			},
			contains: []string{
				"ITERATION 5", // No "of N"
				"09:15:30",
				"codex",
				"🚂",
			},
		},
		{
			name: "first iteration",
			cfg: IterationConfig{
				Number:       1,
				MaxIteration: 10,
				Timestamp:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				CLI:          "gemini",
			},
			contains: []string{
				"ITERATION 1 of 10",
				"00:00:00",
				"gemini",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := RenderIterationHeader(tt.cfg)

			// Check all expected strings are present
			for _, expected := range tt.contains {
				assert.Contains(t, output, expected,
					"Expected output to contain '%s'", expected)
			}

			// Check output has reasonable length (not empty)
			assert.Greater(t, len(output), 50,
				"Output should have substantial content")

			// Check output has newlines (multiline)
			assert.Contains(t, output, "\n",
				"Output should be multiline")
		})
	}
}

func TestRenderToolCall(t *testing.T) {
	tests := []struct {
		name     string
		tc       ToolCall
		contains []string
	}{
		{
			name: "tool with extra info",
			tc: ToolCall{
				Name:  "Read",
				Extra: "src/main.go",
			},
			contains: []string{
				"🔧",
				"Read",
				"(src/main.go)",
			},
		},
		{
			name: "tool without extra info",
			tc: ToolCall{
				Name:  "Edit",
				Extra: "",
			},
			contains: []string{
				"🔧",
				"Edit",
			},
		},
		{
			name: "bash tool with command",
			tc: ToolCall{
				Name:  "Bash",
				Extra: "go test ./...",
			},
			contains: []string{
				"🔧",
				"Bash",
				"(go test ./...)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := RenderToolCall(tt.tc)

			for _, expected := range tt.contains {
				assert.Contains(t, output, expected,
					"Expected output to contain '%s'", expected)
			}

			// Single line output (no newlines)
			assert.NotContains(t, output, "\n",
				"Tool call should be single line")
		})
	}
}

func TestRenderIterationSummary(t *testing.T) {
	tests := []struct {
		name     string
		cfg      IterationConfig
		contains []string
		notContains []string
	}{
		{
			name: "successful iteration with commits and changes",
			cfg: IterationConfig{
				Number:   3,
				Duration: 45 * time.Second,
				Commits:  1,
				Modified: 2,
				Staged:   0,
				Untracked: 0,
				Verified: true,
				Pushed:   true,
			},
			contains: []string{
				"Iteration 3 complete",
				"45s",
				"Commits: 1",
				"Changes: 2 modified, 0 staged, 0 new",
				"Verification passed",
				"Pushed to origin",
				"✓",
				"☁️",
				"─", // Single separator
			},
		},
		{
			name: "iteration with no commits",
			cfg: IterationConfig{
				Number:   1,
				Duration: 10 * time.Second,
				Commits:  0,
				Modified: 1,
				Staged:   1,
				Untracked: 1,
			},
			contains: []string{
				"Iteration 1 complete",
				"10s",
				"Commits: 0",
				"Changes: 1 modified, 1 staged, 1 new",
				"○", // Empty circle for no commits
			},
		},
		{
			name: "iteration with verification failure",
			cfg: IterationConfig{
				Number:       2,
				Duration:     120 * time.Second,
				Commits:      1,
				VerifyFailed: true,
			},
			contains: []string{
				"Iteration 2 complete",
				"2m 0s",
				"Verification failed",
				"✗",
			},
		},
		{
			name: "iteration with push failure",
			cfg: IterationConfig{
				Number:     4,
				Duration:   30 * time.Second,
				Commits:    2,
				PushFailed: true,
			},
			contains: []string{
				"Iteration 4 complete",
				"30s",
				"Push failed",
				"⚠️",
			},
		},
		{
			name: "long duration formatting",
			cfg: IterationConfig{
				Number:   5,
				Duration: 1*time.Hour + 5*time.Minute + 30*time.Second,
				Commits:  3,
			},
			contains: []string{
				"1h 5m 30s",
			},
		},
		{
			name: "no verification line when not configured",
			cfg: IterationConfig{
				Number:   1,
				Duration: 10 * time.Second,
				Commits:  1,
				Verified: false,
				VerifyFailed: false,
			},
			notContains: []string{
				"Verification",
			},
		},
		{
			name: "no push line when not attempted",
			cfg: IterationConfig{
				Number:   1,
				Duration: 10 * time.Second,
				Commits:  1,
				Pushed:   false,
				PushFailed: false,
			},
			notContains: []string{
				"Push",
				"☁️",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := RenderIterationSummary(tt.cfg)

			for _, expected := range tt.contains {
				assert.Contains(t, output, expected,
					"Expected output to contain '%s'", expected)
			}

			for _, unexpected := range tt.notContains {
				assert.NotContains(t, output, unexpected,
					"Expected output to NOT contain '%s'", unexpected)
			}

			// Check output has reasonable length
			assert.Greater(t, len(output), 50,
				"Output should have substantial content")

			// Check output is multiline
			assert.Contains(t, output, "\n",
				"Output should be multiline")
		})
	}
}

func TestRenderToolCalls(t *testing.T) {
	tests := []struct {
		name     string
		calls    []ToolCall
		expected int // Expected number of tool lines
	}{
		{
			name:     "empty tool calls",
			calls:    []ToolCall{},
			expected: 0,
		},
		{
			name: "single tool call",
			calls: []ToolCall{
				{Name: "Read", Extra: "file.go"},
			},
			expected: 1,
		},
		{
			name: "multiple tool calls",
			calls: []ToolCall{
				{Name: "Read", Extra: "file1.go"},
				{Name: "Edit", Extra: "file2.go"},
				{Name: "Bash", Extra: "go test"},
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := RenderToolCalls(tt.calls)

			if tt.expected == 0 {
				assert.Empty(t, output, "Empty tool calls should produce empty output")
				return
			}

			// Count lines with tool emoji
			toolLines := strings.Count(output, "🔧")
			assert.Equal(t, tt.expected, toolLines,
				"Expected %d tool lines, got %d", tt.expected, toolLines)

			// Check all tool names are present
			for _, tc := range tt.calls {
				assert.Contains(t, output, tc.Name,
					"Expected output to contain tool name '%s'", tc.Name)
				if tc.Extra != "" {
					assert.Contains(t, output, tc.Extra,
						"Expected output to contain extra info '%s'", tc.Extra)
				}
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "less than a second",
			duration: 500 * time.Millisecond,
			expected: "0s",
		},
		{
			name:     "exactly one second",
			duration: 1 * time.Second,
			expected: "1s",
		},
		{
			name:     "seconds only",
			duration: 45 * time.Second,
			expected: "45s",
		},
		{
			name:     "minutes and seconds",
			duration: 2*time.Minute + 15*time.Second,
			expected: "2m 15s",
		},
		{
			name:     "hours, minutes, and seconds",
			duration: 1*time.Hour + 5*time.Minute + 30*time.Second,
			expected: "1h 5m 30s",
		},
		{
			name:     "hours only",
			duration: 3 * time.Hour,
			expected: "3h 0m 0s",
		},
		{
			name:     "minutes only",
			duration: 5 * time.Minute,
			expected: "5m 0s",
		},
		{
			name:     "large duration",
			duration: 24*time.Hour + 30*time.Minute + 45*time.Second,
			expected: "24h 30m 45s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.duration)
			assert.Equal(t, tt.expected, result,
				"FormatDuration(%v) = %s, expected %s",
				tt.duration, result, tt.expected)
		})
	}
}

func TestRenderCompleteIteration(t *testing.T) {
	cfg := IterationConfig{
		Number:       3,
		MaxIteration: 20,
		Timestamp:    time.Date(2024, 1, 1, 14, 32, 15, 0, time.UTC),
		CLI:          "claude",
		ToolCalls: []ToolCall{
			{Name: "Read", Extra: "src/main.go"},
			{Name: "Edit", Extra: "src/main.go"},
			{Name: "Bash", Extra: "go test ./..."},
		},
		Duration:  45 * time.Second,
		Commits:   1,
		Modified:  2,
		Staged:    0,
		Untracked: 0,
		Verified:  true,
		Pushed:    true,
	}

	output := RenderCompleteIteration(cfg)

	// Check all major components are present
	require.Contains(t, output, "ITERATION 3 of 20", "Should contain iteration header")
	require.Contains(t, output, "Read", "Should contain tool calls")
	require.Contains(t, output, "Edit", "Should contain tool calls")
	require.Contains(t, output, "Bash", "Should contain tool calls")
	require.Contains(t, output, "Iteration 3 complete", "Should contain summary")
	require.Contains(t, output, "Commits: 1", "Should contain commits")
	require.Contains(t, output, "Verification passed", "Should contain verification")
	require.Contains(t, output, "Pushed to origin", "Should contain push status")

	// Check separators are present
	require.Contains(t, output, "═", "Should contain double separator")
	require.Contains(t, output, "─", "Should contain single separator")

	// Check output is substantial and multiline
	assert.Greater(t, len(output), 200, "Complete iteration should have substantial content")
	assert.Greater(t, strings.Count(output, "\n"), 10, "Complete iteration should have many lines")
}

func TestIterationConfig_EdgeCases(t *testing.T) {
	t.Run("zero values", func(t *testing.T) {
		cfg := IterationConfig{
			Number: 1,
		}

		// Should not panic with zero values
		header := RenderIterationHeader(cfg)
		assert.NotEmpty(t, header, "Should handle zero values gracefully")

		summary := RenderIterationSummary(cfg)
		assert.NotEmpty(t, summary, "Should handle zero values gracefully")
	})

	t.Run("negative values", func(t *testing.T) {
		cfg := IterationConfig{
			Number:   1,
			Commits:  -1, // Shouldn't happen in practice
			Modified: -1,
		}

		// Should not panic with negative values
		summary := RenderIterationSummary(cfg)
		assert.NotEmpty(t, summary, "Should handle negative values gracefully")
	})

	t.Run("very large iteration number", func(t *testing.T) {
		cfg := IterationConfig{
			Number:       999999,
			MaxIteration: 1000000,
		}

		header := RenderIterationHeader(cfg)
		assert.Contains(t, header, "999999", "Should handle large iteration numbers")
	})
}

// TestIterationDisplay_VisualInspection outputs a complete iteration for manual visual inspection.
// This is helpful during development to see how the output actually looks.
func TestIterationDisplay_VisualInspection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping visual inspection test in short mode")
	}

	cfg := IterationConfig{
		Number:       3,
		MaxIteration: 20,
		Timestamp:    time.Now(),
		CLI:          "claude",
		ToolCalls: []ToolCall{
			{Name: "Read", Extra: "src/main.go"},
			{Name: "Edit", Extra: "src/main.go"},
			{Name: "Bash", Extra: "go test ./..."},
			{Name: "Write", Extra: "internal/config/config.go"},
		},
		Duration:  2*time.Minute + 15*time.Second,
		Commits:   1,
		Modified:  2,
		Staged:    0,
		Untracked: 0,
		Verified:  true,
		Pushed:    true,
	}

	t.Log("\n" + RenderCompleteIteration(cfg))
}
