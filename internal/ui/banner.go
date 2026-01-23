package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// BannerConfig contains all the information needed to render a startup banner.
type BannerConfig struct {
	Version        string // e.g., "v2.0.0"
	CLI            string // e.g., "claude"
	Model          string // e.g., "sonnet" or "" for default
	Autonomous     bool   // true for --choo-choo mode
	PromptFile     string // e.g., "PROMPT.md"
	Branch         string // e.g., "main"
	MaxIterations  int    // 0 for unlimited, >0 for specific max
	ShowRalphQuote bool   // true to include Ralph ASCII art and quote
}

// RenderBanner creates the startup banner display.
// If ShowRalphQuote is true, includes Ralph ASCII art and a random quote (for help output).
// Otherwise, shows a clean config summary (for run command).
func RenderBanner(cfg BannerConfig) string {
	var sb strings.Builder

	// If requested, show Ralph ASCII art and quote first
	if cfg.ShowRalphQuote {
		sb.WriteString(RalphStyle.Render(RalphASCII))
		sb.WriteString("\n\n")
		sb.WriteString(MutedStyle.Render(fmt.Sprintf(`"%s"`, RandomQuote())))
		sb.WriteString("\n\n")
	}

	// Create the title box
	title := fmt.Sprintf("gumloop %s", cfg.Version)
	titleBox := BoxStyle.
		Width(35).
		Align(lipgloss.Center).
		Render(title)

	sb.WriteString(titleBox)
	sb.WriteString("\n")

	// Build the config summary lines
	lines := []string{
		fmt.Sprintf(" CLI:    %s", cfg.CLI),
	}

	// Model line (only show if specified)
	if cfg.Model != "" {
		lines = append(lines, fmt.Sprintf(" Model:  %s", cfg.Model))
	}

	// Mode line
	mode := "Single run"
	if cfg.Autonomous {
		mode = "🚂 Choo-choo (autonomous)"
	}
	lines = append(lines, fmt.Sprintf(" Mode:   %s", mode))

	// Prompt file
	lines = append(lines, fmt.Sprintf(" Prompt: %s", cfg.PromptFile))

	// Branch
	if cfg.Branch != "" {
		lines = append(lines, fmt.Sprintf(" Branch: %s", cfg.Branch))
	}

	// Max iterations (only show if in autonomous mode)
	if cfg.Autonomous {
		maxDisplay := "unlimited"
		if cfg.MaxIterations > 0 {
			maxDisplay = fmt.Sprintf("%d iterations", cfg.MaxIterations)
		}
		lines = append(lines, fmt.Sprintf(" Max:    %s", maxDisplay))
	}

	// Render all config lines
	for _, line := range lines {
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	return sb.String()
}

// RenderHelpBanner is a convenience function for rendering the banner in help output.
// It includes Ralph ASCII art and a random quote.
func RenderHelpBanner(version string) string {
	return RenderBanner(BannerConfig{
		Version:        version,
		ShowRalphQuote: true,
	})
}

// RenderStartupBanner is a convenience function for rendering the banner at run start.
// It shows the full configuration without Ralph art.
func RenderStartupBanner(version, cli, model, promptFile, branch string, autonomous bool, maxIterations int) string {
	return RenderBanner(BannerConfig{
		Version:        version,
		CLI:            cli,
		Model:          model,
		Autonomous:     autonomous,
		PromptFile:     promptFile,
		Branch:         branch,
		MaxIterations:  maxIterations,
		ShowRalphQuote: false,
	})
}
