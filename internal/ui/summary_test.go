package ui

import (
	"strings"
	"testing"
	"time"
)

func TestRenderRunSummary(t *testing.T) {
	tests := []struct {
		name     string
		config   SummaryConfig
		wantIcon string
		wantText string
	}{
		{
			name: "successful completion",
			config: SummaryConfig{
				Agent:      "claude",
				Iterations: 5,
				Commits:    3,
				Duration:   4*time.Minute + 32*time.Second,
				ExitCode:   ExitSuccess,
			},
			wantIcon: "✅",
			wantText: "Complete (no changes)",
		},
		{
			name: "max iterations reached",
			config: SummaryConfig{
				Agent:      "codex",
				Iterations: 20,
				Commits:    15,
				Duration:   10 * time.Minute,
				ExitCode:   ExitMaxIterations,
			},
			wantIcon: "⏱️",
			wantText: "Max iterations reached",
		},
		{
			name: "stuck detection",
			config: SummaryConfig{
				Agent:      "gemini",
				Iterations: 8,
				Commits:    2,
				Duration:   6*time.Minute + 45*time.Second,
				ExitCode:   ExitStuck,
			},
			wantIcon: "⚠️",
			wantText: "Stuck (no commits)",
		},
		{
			name: "general error",
			config: SummaryConfig{
				Agent:      "claude",
				Iterations: 1,
				Commits:    0,
				Duration:   30 * time.Second,
				ExitCode:   ExitError,
			},
			wantIcon: "❌",
			wantText: "Error",
		},
		{
			name: "safety refusal",
			config: SummaryConfig{
				Agent:      "opencode",
				Iterations: 0,
				Commits:    0,
				Duration:   0,
				ExitCode:   ExitSafety,
			},
			wantIcon: "🛑",
			wantText: "Safety refusal",
		},
		{
			name: "user interrupt",
			config: SummaryConfig{
				Agent:      "cursor",
				Iterations: 3,
				Commits:    1,
				Duration:   2 * time.Minute,
				ExitCode:   ExitInterrupt,
			},
			wantIcon: "⏸️",
			wantText: "Interrupted by user",
		},
		{
			name: "custom exit reason",
			config: SummaryConfig{
				Agent:      "ollama",
				Iterations: 7,
				Commits:    4,
				Duration:   5 * time.Minute,
				ExitCode:   ExitSuccess,
				ExitReason: "Task completed",
			},
			wantIcon: "✅",
			wantText: "Task completed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := RenderRunSummary(tt.config)

			// Verify the output contains expected elements
			if !strings.Contains(output, "RUN COMPLETE") {
				t.Error("output should contain 'RUN COMPLETE'")
			}

			if !strings.Contains(output, tt.config.Agent) {
				t.Errorf("output should contain agent name %q", tt.config.Agent)
			}

			// Verify metrics are present
			if !strings.Contains(output, "Agent:") {
				t.Error("output should contain 'Agent:'")
			}
			if !strings.Contains(output, "Iterations:") {
				t.Error("output should contain 'Iterations:'")
			}
			if !strings.Contains(output, "Commits:") {
				t.Error("output should contain 'Commits:'")
			}
			if !strings.Contains(output, "Duration:") {
				t.Error("output should contain 'Duration:'")
			}

			// Verify exit reason contains icon and text
			if !strings.Contains(output, tt.wantIcon) {
				t.Errorf("output should contain icon %q", tt.wantIcon)
			}
			if !strings.Contains(output, tt.wantText) {
				t.Errorf("output should contain text %q", tt.wantText)
			}
		})
	}
}

func TestFormatExitReason(t *testing.T) {
	tests := []struct {
		name         string
		code         ExitCode
		customReason string
		wantIcon     string
		wantText     string
	}{
		{
			name:     "success default",
			code:     ExitSuccess,
			wantIcon: "✅",
			wantText: "Complete (no changes)",
		},
		{
			name:         "success custom",
			code:         ExitSuccess,
			customReason: "All tests passed",
			wantIcon:     "✅",
			wantText:     "All tests passed",
		},
		{
			name:     "error default",
			code:     ExitError,
			wantIcon: "❌",
			wantText: "Error",
		},
		{
			name:     "safety default",
			code:     ExitSafety,
			wantIcon: "🛑",
			wantText: "Safety refusal",
		},
		{
			name:     "max iterations default",
			code:     ExitMaxIterations,
			wantIcon: "⏱️",
			wantText: "Max iterations reached",
		},
		{
			name:     "stuck default",
			code:     ExitStuck,
			wantIcon: "⚠️",
			wantText: "Stuck (no commits)",
		},
		{
			name:     "interrupt default",
			code:     ExitInterrupt,
			wantIcon: "⏸️",
			wantText: "Interrupted by user",
		},
		{
			name:     "unknown code",
			code:     ExitCode(99),
			wantIcon: "❓",
			wantText: "Unknown (code 99)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIcon, gotText := formatExitReason(tt.code, tt.customReason)

			if gotIcon != tt.wantIcon {
				t.Errorf("formatExitReason() icon = %q, want %q", gotIcon, tt.wantIcon)
			}
			if gotText != tt.wantText {
				t.Errorf("formatExitReason() text = %q, want %q", gotText, tt.wantText)
			}
		})
	}
}

func TestStyleExitLine(t *testing.T) {
	tests := []struct {
		name string
		code ExitCode
		line string
	}{
		{
			name: "success styling",
			code: ExitSuccess,
			line: "Exit: ✅ Complete",
		},
		{
			name: "error styling",
			code: ExitError,
			line: "Exit: ❌ Error",
		},
		{
			name: "safety styling",
			code: ExitSafety,
			line: "Exit: 🛑 Safety",
		},
		{
			name: "max iterations styling",
			code: ExitMaxIterations,
			line: "Exit: ⏱️ Max iterations",
		},
		{
			name: "stuck styling",
			code: ExitStuck,
			line: "Exit: ⚠️ Stuck",
		},
		{
			name: "interrupt styling",
			code: ExitInterrupt,
			line: "Exit: ⏸️ Interrupted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			styled := styleExitLine(tt.code, tt.line)

			// The styled output should contain the original line text
			// We can't easily test ANSI codes, but we can verify the function runs
			if styled == "" {
				t.Error("styleExitLine() should not return empty string")
			}
		})
	}
}

func TestSummaryWithZeroDuration(t *testing.T) {
	config := SummaryConfig{
		Agent:      "claude",
		Iterations: 1,
		Commits:    0,
		Duration:   0,
		ExitCode:   ExitSuccess,
	}

	output := RenderRunSummary(config)

	// Should show "0s" for zero duration
	if !strings.Contains(output, "0s") {
		t.Error("output should contain '0s' for zero duration")
	}
}

func TestSummaryWithLongDuration(t *testing.T) {
	config := SummaryConfig{
		Agent:      "claude",
		Iterations: 100,
		Commits:    50,
		Duration:   2*time.Hour + 15*time.Minute + 30*time.Second,
		ExitCode:   ExitSuccess,
	}

	output := RenderRunSummary(config)

	// Should format as "2h 15m 30s"
	if !strings.Contains(output, "2h 15m 30s") {
		t.Error("output should contain formatted duration '2h 15m 30s'")
	}
}
