package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// ExitCode represents the exit status of a gumloop run
type ExitCode int

const (
	ExitSuccess        ExitCode = 0   // Work complete (no changes detected)
	ExitError          ExitCode = 1   // General error (config, validation, runtime)
	ExitSafety         ExitCode = 2   // Safety refusal (dangerous path, no git)
	ExitMaxIterations  ExitCode = 3   // Max iterations reached
	ExitStuck          ExitCode = 4   // Stuck (changes but no commits for N iterations)
	ExitInterrupt      ExitCode = 130 // User interrupted (Ctrl+C)
)

// SummaryConfig contains all the information needed to render a run summary.
type SummaryConfig struct {
	Agent      string        // Agent name (e.g., "claude")
	Iterations int           // Total iterations run
	Commits    int           // Total commits made
	Duration   time.Duration // Total run duration
	ExitCode   ExitCode      // Exit code
	ExitReason string        // Optional custom exit reason message
}

// RenderRunSummary renders the summary shown at the end of a gumloop run.
// Uses the Simpsons color theme for a distinctive, branded appearance.
//
// Example output:
//   ╭─────────────────────────────────────╮
//   │         🍩 RUN COMPLETE 🍩          │
//   ├─────────────────────────────────────┤
//   │  Agent:       claude                │
//   │  Iterations:  5                     │
//   │  Commits:     3                     │
//   │  Duration:    4m 32s                │
//   ├─────────────────────────────────────┤
//   │  Exit: ✅ Complete (no changes)     │
//   ╰─────────────────────────────────────╯
func RenderRunSummary(cfg SummaryConfig) string {
	// Determine exit message and styling
	exitIcon, exitText := formatExitReason(cfg.ExitCode, cfg.ExitReason)

	// Style definitions using Simpsons theme
	labelStyle := lipgloss.NewStyle().Foreground(ColorMargeBlue)
	valueStyle := lipgloss.NewStyle().Foreground(ColorWhite)
	borderStyle := lipgloss.NewStyle().Foreground(ColorSimpsonYellow)

	// Title style - bold Simpson Yellow with donut emoji
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorSimpsonYellow)

	// Box dimensions
	innerWidth := 35

	// Helper to pad content to width
	pad := func(s string, width int) string {
		// Account for ANSI escape codes when calculating visible length
		visible := lipgloss.Width(s)
		if visible >= width {
			return s
		}
		return s + strings.Repeat(" ", width-visible)
	}

	// Build separator line
	separator := borderStyle.Render("├" + strings.Repeat("─", innerWidth) + "┤")

	// Build content lines
	var lines []string

	// Top border
	lines = append(lines, borderStyle.Render("╭"+strings.Repeat("─", innerWidth)+"╮"))

	// Title line
	title := titleStyle.Render("🍩 RUN COMPLETE 🍩")
	titlePadded := lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(title)
	lines = append(lines, borderStyle.Render("│")+titlePadded+borderStyle.Render("│"))

	// Separator
	lines = append(lines, separator)

	// Metrics
	metrics := []struct{ label, value string }{
		{"Agent:", cfg.Agent},
		{"Iterations:", fmt.Sprintf("%d", cfg.Iterations)},
		{"Commits:", fmt.Sprintf("%d", cfg.Commits)},
		{"Duration:", FormatDuration(cfg.Duration)},
	}
	for _, m := range metrics {
		content := fmt.Sprintf("  %s %s", labelStyle.Render(fmt.Sprintf("%-12s", m.label)), valueStyle.Render(m.value))
		lines = append(lines, borderStyle.Render("│")+pad(content, innerWidth)+borderStyle.Render("│"))
	}

	// Separator
	lines = append(lines, separator)

	// Exit status line
	exitContent := fmt.Sprintf("  Exit: %s %s", exitIcon, exitText)
	styledExit := styleExitLine(cfg.ExitCode, exitContent)
	lines = append(lines, borderStyle.Render("│")+pad(styledExit, innerWidth)+borderStyle.Render("│"))

	// Bottom border
	lines = append(lines, borderStyle.Render("╰"+strings.Repeat("─", innerWidth)+"╯"))

	return strings.Join(lines, "\n")
}

// formatExitReason returns the icon and text for an exit code
func formatExitReason(code ExitCode, customReason string) (icon string, text string) {
	if customReason != "" {
		text = customReason
	}

	switch code {
	case ExitSuccess:
		icon = "✅"
		if text == "" {
			text = "Complete (no changes)"
		}
	case ExitError:
		icon = "❌"
		if text == "" {
			text = "Error"
		}
	case ExitSafety:
		icon = "🛑"
		if text == "" {
			text = "Safety refusal"
		}
	case ExitMaxIterations:
		icon = "⏱️"
		if text == "" {
			text = "Max iterations reached"
		}
	case ExitStuck:
		icon = "⚠️"
		if text == "" {
			text = "Stuck (no commits)"
		}
	case ExitInterrupt:
		icon = "⏸️"
		if text == "" {
			text = "Interrupted by user"
		}
	default:
		icon = "❓"
		if text == "" {
			text = fmt.Sprintf("Unknown (code %d)", code)
		}
	}

	return icon, text
}

// styleExitLine applies appropriate styling to the exit line based on exit code
func styleExitLine(code ExitCode, line string) string {
	switch code {
	case ExitSuccess:
		return SuccessStyle.Render(line)
	case ExitError, ExitSafety:
		return ErrorStyle.Render(line)
	case ExitMaxIterations, ExitStuck:
		return WarningStyle.Render(line)
	case ExitInterrupt:
		return MutedStyle.Render(line)
	default:
		return line
	}
}
