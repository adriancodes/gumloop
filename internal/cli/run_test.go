package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adriancodes/gumloop/internal/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadRunConfig_Defaults(t *testing.T) {
	// Reset viper state
	viper.Reset()
	viper.SetDefault("cli", "claude")
	viper.SetDefault("model", "")
	viper.SetDefault("prompt_file", "PROMPT.md")
	viper.SetDefault("auto_push", true)
	viper.SetDefault("stuck_threshold", 3)
	viper.SetDefault("verify", "")

	// Reset flags
	runPrompt = ""
	runPromptFile = ""
	runCLI = ""
	runModel = ""
	runChooChoo = 0
	runNoPush = false
	runStuck = 0
	runVerify = ""

	cfg, err := loadRunConfig()
	require.NoError(t, err)

	assert.Equal(t, "claude", cfg.CLI)
	assert.Equal(t, "", cfg.Model)
	assert.Equal(t, "PROMPT.md", cfg.PromptFile)
	assert.Equal(t, true, cfg.AutoPush)
	assert.Equal(t, 3, cfg.StuckThreshold)
	assert.Equal(t, "", cfg.Verify)
	assert.Equal(t, false, cfg.ChooChoo)
	assert.Equal(t, 0, cfg.MaxIterations)
}

func TestLoadRunConfig_FlagOverrides(t *testing.T) {
	// Reset viper state
	viper.Reset()
	defaults := config.Defaults()
	viper.SetDefault("cli", defaults.CLI)
	viper.SetDefault("model", defaults.Model)
	viper.SetDefault("prompt_file", defaults.PromptFile)
	viper.SetDefault("auto_push", defaults.AutoPush)
	viper.SetDefault("stuck_threshold", defaults.StuckThreshold)
	viper.SetDefault("verify", defaults.Verify)

	// Set flags
	runPrompt = "test prompt"
	runCLI = "codex"
	runModel = "gpt-4"
	runNoPush = true
	runStuck = 5
	runVerify = "npm test"
	runChooChoo = 10

	cfg, err := loadRunConfig()
	require.NoError(t, err)

	assert.Equal(t, "codex", cfg.CLI)
	assert.Equal(t, "gpt-4", cfg.Model)
	assert.Equal(t, "test prompt", cfg.Prompt)
	assert.Equal(t, false, cfg.AutoPush) // --no-push overrides default
	assert.Equal(t, 5, cfg.StuckThreshold)
	assert.Equal(t, "npm test", cfg.Verify)
	assert.Equal(t, true, cfg.ChooChoo)
	assert.Equal(t, 10, cfg.MaxIterations)

	// Reset flags
	runPrompt = ""
	runCLI = ""
	runModel = ""
	runNoPush = false
	runStuck = 0
	runVerify = ""
	runChooChoo = 0
}

func TestLoadRunConfig_InlinePrompt(t *testing.T) {
	// Reset viper
	viper.Reset()
	defaults := config.Defaults()
	viper.SetDefault("cli", defaults.CLI)
	viper.SetDefault("prompt_file", defaults.PromptFile)

	// Set inline prompt
	runPrompt = "Fix the tests"

	cfg, err := loadRunConfig()
	require.NoError(t, err)

	assert.Equal(t, "Fix the tests", cfg.Prompt)

	// Reset
	runPrompt = ""
}

func TestLoadRunConfig_PromptFromFile(t *testing.T) {
	// Create temp prompt file
	tmpDir := t.TempDir()
	promptFile := filepath.Join(tmpDir, "PROMPT.md")
	err := os.WriteFile(promptFile, []byte("Prompt from file"), 0644)
	require.NoError(t, err)

	// Reset viper
	viper.Reset()
	defaults := config.Defaults()
	viper.SetDefault("cli", defaults.CLI)
	viper.SetDefault("prompt_file", defaults.PromptFile)

	// Set prompt file flag
	runPromptFile = promptFile

	cfg, err := loadRunConfig()
	require.NoError(t, err)

	assert.Equal(t, "Prompt from file", cfg.Prompt)

	// Reset
	runPromptFile = ""
}

func TestLoadRunConfig_ChooChooUnlimited(t *testing.T) {
	// Reset viper
	viper.Reset()
	defaults := config.Defaults()
	viper.SetDefault("cli", defaults.CLI)

	// Simulate --choo-choo without value (use -1 as signal)
	runChooChoo = -1

	cfg, err := loadRunConfig()
	require.NoError(t, err)

	assert.Equal(t, true, cfg.ChooChoo)
	assert.Equal(t, 0, cfg.MaxIterations) // 0 means unlimited

	// Reset
	runChooChoo = 0
}

func TestLoadRunConfig_ChooChooWithLimit(t *testing.T) {
	// Reset viper
	viper.Reset()
	defaults := config.Defaults()
	viper.SetDefault("cli", defaults.CLI)

	// Simulate --choo-choo 20
	runChooChoo = 20

	cfg, err := loadRunConfig()
	require.NoError(t, err)

	assert.Equal(t, true, cfg.ChooChoo)
	assert.Equal(t, 20, cfg.MaxIterations)

	// Reset
	runChooChoo = 0
}

func TestValidateRunConfig_NoPrompt(t *testing.T) {
	cfg := &RunConfig{
		Config: config.Config{
			PromptFile: "PROMPT.md",
		},
		Prompt: "", // Empty prompt
	}

	err := validateRunConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "prompt required")
}

func TestValidateRunConfig_ValidConfig(t *testing.T) {
	// This test validates the config structure validation, not safety checks
	// We use single-run mode (not choo-choo) to avoid the home directory prompt
	cfg := &RunConfig{
		Config: config.Config{
			CLI:            "claude",
			PromptFile:     "PROMPT.md",
			StuckThreshold: 3,
		},
		Prompt:        "Fix the tests",
		ChooChoo:      false, // Not in choo-choo mode, so no home directory check
		MaxIterations: 0,
	}

	err := validateRunConfig(cfg)
	// If we're in a git repo, this should pass
	// If not, it will fail with a safety error (which is correct)
	if err != nil {
		// Check if it's a safety error about git repo (expected if not in git repo)
		safetyErr, ok := err.(*SafetyError)
		assert.True(t, ok, "expected SafetyError or no error")
		if ok {
			assert.Contains(t, safetyErr.Message, "not in a git repository")
		}
	}
}

func TestValidateRunConfig_NegativeStuckThreshold(t *testing.T) {
	cfg := &RunConfig{
		Config: config.Config{
			StuckThreshold: -1,
		},
		Prompt: "test",
	}

	err := validateRunConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stuck_threshold must be a positive integer")
}

func TestValidateRunConfig_NegativeMaxIterations(t *testing.T) {
	cfg := &RunConfig{
		Config: config.Config{
			StuckThreshold: 3,
		},
		Prompt:        "test",
		MaxIterations: -5,
	}

	err := validateRunConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max iterations must be non-negative")
}

func TestValidateRunConfig_InvalidAgent(t *testing.T) {
	cfg := &RunConfig{
		Config: config.Config{
			CLI:            "nonexistent",
			StuckThreshold: 3,
		},
		Prompt:        "test",
		MaxIterations: 0,
	}

	err := validateRunConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid agent")
}

func TestValidateRunConfig_NotInGitRepo(t *testing.T) {
	// Create a temp directory that's not a git repo
	tmpDir := t.TempDir()

	// Change to the temp directory
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origDir)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	cfg := &RunConfig{
		Config: config.Config{
			CLI:            "claude",
			StuckThreshold: 3,
		},
		Prompt:        "test",
		MaxIterations: 0,
	}

	err = validateRunConfig(cfg)
	assert.Error(t, err)

	// Check that it's a SafetyError with ExitSafety code
	safetyErr, ok := err.(*SafetyError)
	assert.True(t, ok, "expected SafetyError")
	if ok {
		assert.Contains(t, safetyErr.Message, "not in a git repository")
	}
}

func TestSafetyError(t *testing.T) {
	err := &SafetyError{
		Message: "test safety error",
	}

	assert.Equal(t, "test safety error", err.Error())
}
