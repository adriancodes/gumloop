package ui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderBanner(t *testing.T) {
	tests := []struct {
		name     string
		cfg      BannerConfig
		contains []string // strings that should be in the output
		notContains []string // strings that should NOT be in the output
	}{
		{
			name: "basic config without Ralph",
			cfg: BannerConfig{
				Version:        "v2.0.0",
				CLI:            "claude",
				Model:          "sonnet",
				Autonomous:     false,
				PromptFile:     "PROMPT.md",
				Branch:         "main",
				MaxIterations:  0,
				ShowRalphQuote: false,
			},
			contains: []string{
				"gumloop v2.0.0",
				"CLI:    claude",
				"Model:  sonnet",
				"Mode:   Single run",
				"Prompt: PROMPT.md",
				"Branch: main",
			},
			notContains: []string{
				"Max:",           // Should not show max iterations in single run mode
				"I'm learnding", // Should not show Ralph quote
				RalphASCII,      // Should not show Ralph art
			},
		},
		{
			name: "autonomous mode with max iterations",
			cfg: BannerConfig{
				Version:        "v2.0.0",
				CLI:            "codex",
				Model:          "gpt-4",
				Autonomous:     true,
				PromptFile:     "task.md",
				Branch:         "feature/test",
				MaxIterations:  20,
				ShowRalphQuote: false,
			},
			contains: []string{
				"gumloop v2.0.0",
				"CLI:    codex",
				"Model:  gpt-4",
				"Mode:   🚂 Choo-choo (autonomous)",
				"Prompt: task.md",
				"Branch: feature/test",
				"Max:    20 iterations",
			},
		},
		{
			name: "autonomous mode unlimited iterations",
			cfg: BannerConfig{
				Version:        "v2.0.0",
				CLI:            "gemini",
				Model:          "",
				Autonomous:     true,
				PromptFile:     "PROMPT.md",
				Branch:         "main",
				MaxIterations:  0, // 0 means unlimited
				ShowRalphQuote: false,
			},
			contains: []string{
				"gumloop v2.0.0",
				"CLI:    gemini",
				"Mode:   🚂 Choo-choo (autonomous)",
				"Max:    unlimited",
			},
			notContains: []string{
				"Model:", // Should not show model line if empty
			},
		},
		{
			name: "with Ralph quote for help",
			cfg: BannerConfig{
				Version:        "v2.0.0",
				CLI:            "claude",
				Model:          "",
				Autonomous:     false,
				PromptFile:     "PROMPT.md",
				Branch:         "main",
				MaxIterations:  0,
				ShowRalphQuote: true,
			},
			contains: []string{
				RalphASCII, // Should include ASCII art
				// Note: We can't check for specific quote since it's random
				// But we can verify the structure
				"gumloop v2.0.0",
			},
		},
		{
			name: "minimal config",
			cfg: BannerConfig{
				Version:        "v1.0.0",
				CLI:            "cursor",
				Model:          "",
				Autonomous:     false,
				PromptFile:     "PROMPT.md",
				Branch:         "",
				MaxIterations:  0,
				ShowRalphQuote: false,
			},
			contains: []string{
				"gumloop v1.0.0",
				"CLI:    cursor",
				"Prompt: PROMPT.md",
			},
			notContains: []string{
				"Branch:", // Should not show branch if empty
				"Model:",  // Should not show model if empty
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderBanner(tt.cfg)

			// Check that all expected strings are present
			for _, expected := range tt.contains {
				assert.Contains(t, result, expected,
					"Banner should contain %q", expected)
			}

			// Check that unexpected strings are NOT present
			for _, unexpected := range tt.notContains {
				assert.NotContains(t, result, unexpected,
					"Banner should NOT contain %q", unexpected)
			}
		})
	}
}

func TestRenderHelpBanner(t *testing.T) {
	result := RenderHelpBanner("v2.0.0")

	// Should include version
	assert.Contains(t, result, "v2.0.0")

	// Should include Ralph ASCII art
	assert.Contains(t, result, RalphASCII)

	// Should include a quote (we can't check which one since it's random,
	// but we can verify the quote format with quotation marks)
	assert.Contains(t, result, `"`)
}

func TestRenderStartupBanner(t *testing.T) {
	tests := []struct {
		name          string
		version       string
		cli           string
		model         string
		promptFile    string
		branch        string
		autonomous    bool
		maxIterations int
		contains      []string
		notContains   []string
	}{
		{
			name:          "full config autonomous",
			version:       "v2.0.0",
			cli:           "claude",
			model:         "sonnet",
			promptFile:    "PROMPT.md",
			branch:        "main",
			autonomous:    true,
			maxIterations: 10,
			contains: []string{
				"gumloop v2.0.0",
				"CLI:    claude",
				"Model:  sonnet",
				"🚂 Choo-choo (autonomous)",
				"Prompt: PROMPT.md",
				"Branch: main",
				"Max:    10 iterations",
			},
			notContains: []string{
				RalphASCII, // Should NOT include Ralph art in startup banner
			},
		},
		{
			name:          "single run mode",
			version:       "v2.0.0",
			cli:           "codex",
			model:         "",
			promptFile:    "PROMPT.md",
			branch:        "develop",
			autonomous:    false,
			maxIterations: 0,
			contains: []string{
				"gumloop v2.0.0",
				"CLI:    codex",
				"Single run",
				"Prompt: PROMPT.md",
				"Branch: develop",
			},
			notContains: []string{
				"Max:",      // Should not show max in single run
				"Model:",    // Should not show model if empty
				RalphASCII,  // Should NOT include Ralph art
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderStartupBanner(
				tt.version,
				tt.cli,
				tt.model,
				tt.promptFile,
				tt.branch,
				tt.autonomous,
				tt.maxIterations,
			)

			for _, expected := range tt.contains {
				assert.Contains(t, result, expected)
			}

			for _, unexpected := range tt.notContains {
				assert.NotContains(t, result, unexpected)
			}
		})
	}
}

func TestBannerConfig_Validation(t *testing.T) {
	// Test that banner renders without panics even with unusual configs
	tests := []struct {
		name string
		cfg  BannerConfig
	}{
		{
			name: "empty strings",
			cfg: BannerConfig{
				Version:        "",
				CLI:            "",
				Model:          "",
				Autonomous:     false,
				PromptFile:     "",
				Branch:         "",
				MaxIterations:  0,
				ShowRalphQuote: false,
			},
		},
		{
			name: "negative max iterations",
			cfg: BannerConfig{
				Version:        "v2.0.0",
				CLI:            "claude",
				Model:          "sonnet",
				Autonomous:     true,
				PromptFile:     "PROMPT.md",
				Branch:         "main",
				MaxIterations:  -1, // Invalid, but should not panic
				ShowRalphQuote: false,
			},
		},
		{
			name: "very long strings",
			cfg: BannerConfig{
				Version:        "v2.0.0-very-long-version-string-that-might-cause-issues",
				CLI:            "claude-with-a-very-long-name-for-some-reason",
				Model:          strings.Repeat("x", 100),
				Autonomous:     true,
				PromptFile:     strings.Repeat("path/", 20) + "PROMPT.md",
				Branch:         strings.Repeat("feature/", 10) + "branch",
				MaxIterations:  999999,
				ShowRalphQuote: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			assert.NotPanics(t, func() {
				result := RenderBanner(tt.cfg)
				// Result should be a non-empty string
				assert.NotEmpty(t, result)
			})
		})
	}
}

func TestBannerFormatting(t *testing.T) {
	cfg := BannerConfig{
		Version:        "v2.0.0",
		CLI:            "claude",
		Model:          "sonnet",
		Autonomous:     true,
		PromptFile:     "PROMPT.md",
		Branch:         "main",
		MaxIterations:  20,
		ShowRalphQuote: false,
	}

	result := RenderBanner(cfg)

	// Should be non-empty
	assert.NotEmpty(t, result)

	// Should have multiple lines
	lines := strings.Split(result, "\n")
	assert.Greater(t, len(lines), 5, "Banner should have multiple lines")

	// Should include the box border characters (lipgloss rounded border)
	assert.Contains(t, result, "│", "Should contain box border")

	// Each config line should start with a space for alignment
	assert.Contains(t, result, " CLI:")
	assert.Contains(t, result, " Model:")
	assert.Contains(t, result, " Mode:")
	assert.Contains(t, result, " Prompt:")
	assert.Contains(t, result, " Branch:")
	assert.Contains(t, result, " Max:")
}
