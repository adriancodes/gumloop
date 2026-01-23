package ui

import (
	"fmt"
	"strings"
	"time"
)

// IterationConfig contains all the information needed to render an iteration display.
type IterationConfig struct {
	Number       int           // Current iteration number (e.g., 3)
	MaxIteration int           // Max iterations (0 for unlimited)
	Timestamp    time.Time     // When the iteration started
	CLI          string        // Agent name (e.g., "claude")
	ToolCalls    []ToolCall    // Tools used during iteration
	Duration     time.Duration // How long the iteration took
	Commits      int           // Number of commits made
	Modified     int           // Number of modified files
	Staged       int           // Number of staged files
	Untracked    int           // Number of untracked files
	Verified     bool          // Whether verification passed
	VerifyFailed bool          // Whether verification failed (if verify command was run)
	Pushed       bool          // Whether changes were pushed
	PushFailed   bool          // Whether push failed
}

// ToolCall represents a single tool use during iteration
type ToolCall struct {
	Name  string // Tool name (e.g., "Read", "Edit", "Bash")
	Extra string // Optional extra info (e.g., file path)
}

// RenderIterationHeader renders the header shown at the start of an iteration.
//
// Example output:
//   ══════════════════════════════════════
//     🚂 ITERATION 3 of 20
//     14:32:15 | claude
//   ══════════════════════════════════════
func RenderIterationHeader(cfg IterationConfig) string {
	var sb strings.Builder

	// Top separator
	sb.WriteString(DoubleSeparator(38))
	sb.WriteString("\n")

	// Iteration line
	maxDisplay := ""
	if cfg.MaxIteration > 0 {
		maxDisplay = fmt.Sprintf(" of %d", cfg.MaxIteration)
	}
	iterLine := fmt.Sprintf("  🚂 ITERATION %d%s", cfg.Number, maxDisplay)
	sb.WriteString(IterationHeaderStyle.Render(iterLine))
	sb.WriteString("\n")

	// Timestamp and CLI line
	timestamp := cfg.Timestamp.Format("15:04:05")
	infoLine := fmt.Sprintf("  %s | %s", timestamp, cfg.CLI)
	sb.WriteString(MutedStyle.Render(infoLine))
	sb.WriteString("\n")

	// Bottom separator
	sb.WriteString(DoubleSeparator(38))
	sb.WriteString("\n")

	return sb.String()
}

// RenderToolCall renders a single tool call line.
//
// Example output:
//   🔧 Read (src/main.go)
//   🔧 Edit (src/main.go)
//   🔧 Bash (go test ./...)
func RenderToolCall(tc ToolCall) string {
	var sb strings.Builder

	sb.WriteString("🔧 ")
	sb.WriteString(ToolStyle.Render(tc.Name))

	if tc.Extra != "" {
		sb.WriteString(" (")
		sb.WriteString(MutedStyle.Render(tc.Extra))
		sb.WriteString(")")
	}

	return sb.String()
}

// RenderIterationSummary renders the summary shown at the end of an iteration.
//
// Example output:
//   ──────────────────────────────────────
//     Iteration 3 complete (45s)
//     ✓ Commits: 1
//     📝 Changes: 2 modified, 0 staged, 0 new
//     ✓ Verification passed
//     ☁️  Pushed to origin/main
//   ──────────────────────────────────────
func RenderIterationSummary(cfg IterationConfig) string {
	var sb strings.Builder

	// Top separator
	sb.WriteString(SimpleSeparator(38))
	sb.WriteString("\n")

	// Completion line with duration
	durationStr := FormatDuration(cfg.Duration)
	completeLine := fmt.Sprintf("  Iteration %d complete (%s)", cfg.Number, durationStr)
	sb.WriteString(completeLine)
	sb.WriteString("\n")

	// Commits line
	commitIcon := "✓"
	commitStyle := SuccessStyle
	if cfg.Commits == 0 {
		commitIcon = "○"
		commitStyle = MutedStyle
	}
	commitsLine := fmt.Sprintf("  %s Commits: %d", commitIcon, cfg.Commits)
	sb.WriteString(commitStyle.Render(commitsLine))
	sb.WriteString("\n")

	// Changes line
	totalChanges := cfg.Modified + cfg.Staged + cfg.Untracked
	changesIcon := "📝"
	changesLine := fmt.Sprintf("  %s Changes: %d modified, %d staged, %d new",
		changesIcon, cfg.Modified, cfg.Staged, cfg.Untracked)
	if totalChanges == 0 {
		sb.WriteString(MutedStyle.Render(changesLine))
	} else {
		sb.WriteString(changesLine)
	}
	sb.WriteString("\n")

	// Verification line (only show if verify command was configured)
	if cfg.Verified || cfg.VerifyFailed {
		verifyIcon := "✓"
		verifyText := "Verification passed"
		verifyStyle := SuccessStyle
		if cfg.VerifyFailed {
			verifyIcon = "✗"
			verifyText = "Verification failed"
			verifyStyle = ErrorStyle
		}
		verifyLine := fmt.Sprintf("  %s %s", verifyIcon, verifyText)
		sb.WriteString(verifyStyle.Render(verifyLine))
		sb.WriteString("\n")
	}

	// Push line (only show if push was attempted)
	if cfg.Pushed || cfg.PushFailed {
		if cfg.Pushed {
			pushLine := "  ☁️  Pushed to origin"
			sb.WriteString(SuccessStyle.Render(pushLine))
		} else if cfg.PushFailed {
			pushLine := "  ⚠️  Push failed"
			sb.WriteString(WarningStyle.Render(pushLine))
		}
		sb.WriteString("\n")
	}

	// Bottom separator
	sb.WriteString(SimpleSeparator(38))
	sb.WriteString("\n")

	return sb.String()
}

// RenderToolCalls renders all tool calls from an iteration.
// Each tool call is rendered on its own line.
func RenderToolCalls(toolCalls []ToolCall) string {
	if len(toolCalls) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n")

	for _, tc := range toolCalls {
		sb.WriteString(RenderToolCall(tc))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	return sb.String()
}

// FormatDuration formats a duration in a human-readable format.
// Examples: "45s", "2m 15s", "1h 5m 30s"
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return "0s"
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

// RenderCompleteIteration is a convenience function that renders a complete iteration display:
// header + tool calls + summary.
func RenderCompleteIteration(cfg IterationConfig) string {
	var sb strings.Builder

	// Header
	sb.WriteString(RenderIterationHeader(cfg))

	// Tool calls
	if len(cfg.ToolCalls) > 0 {
		sb.WriteString(RenderToolCalls(cfg.ToolCalls))
	}

	// Summary
	sb.WriteString(RenderIterationSummary(cfg))

	return sb.String()
}
