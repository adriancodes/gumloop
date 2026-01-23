package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adriancodes/gumloop/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestWriteConfigFile(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "gumloop-init-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)
	require.NoError(t, os.Chdir(tmpDir))

	// Create test config
	cfg := config.Config{
		CLI:            "claude",
		Model:          "sonnet",
		PromptFile:     "PROMPT.md",
		AutoPush:       true,
		StuckThreshold: 3,
		Verify:         "go test ./...",
	}

	// Write config file
	initGlobal = false // Ensure we're writing project config
	err = writeConfigFileToPath(cfg, ".gumloop.yaml")
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(".gumloop.yaml")
	require.NoError(t, err)

	// Read and parse file
	data, err := os.ReadFile(".gumloop.yaml")
	require.NoError(t, err)

	var parsed config.Config
	err = yaml.Unmarshal(data, &parsed)
	require.NoError(t, err)

	// Verify values
	assert.Equal(t, "claude", parsed.CLI)
	assert.Equal(t, "sonnet", parsed.Model)
	assert.Equal(t, "PROMPT.md", parsed.PromptFile)
	assert.True(t, parsed.AutoPush)
	assert.Equal(t, 3, parsed.StuckThreshold)
	assert.Equal(t, "go test ./...", parsed.Verify)
}

func TestWriteConfigFileWithEmptyValues(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "gumloop-init-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)
	require.NoError(t, os.Chdir(tmpDir))

	// Create config with empty optional values
	cfg := config.Config{
		CLI:            "codex",
		Model:          "",
		PromptFile:     "PROMPT.md",
		AutoPush:       false,
		StuckThreshold: 5,
		Verify:         "",
	}

	// Write config file
	initGlobal = false // Ensure we're writing project config
	err = writeConfigFileToPath(cfg, ".gumloop.yaml")
	require.NoError(t, err)

	// Read and parse file
	data, err := os.ReadFile(".gumloop.yaml")
	require.NoError(t, err)

	var parsed config.Config
	err = yaml.Unmarshal(data, &parsed)
	require.NoError(t, err)

	// Verify values (empty strings should be preserved)
	assert.Equal(t, "codex", parsed.CLI)
	assert.Equal(t, "", parsed.Model)
	assert.Equal(t, "PROMPT.md", parsed.PromptFile)
	assert.False(t, parsed.AutoPush)
	assert.Equal(t, 5, parsed.StuckThreshold)
	assert.Equal(t, "", parsed.Verify)
}

func TestWritePromptTemplate(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "gumloop-init-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)
	require.NoError(t, os.Chdir(tmpDir))

	// Write template
	err = writePromptTemplate()
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat("PROMPT.md")
	require.NoError(t, err)

	// Read file
	content, err := os.ReadFile("PROMPT.md")
	require.NoError(t, err)

	// Verify content contains expected sections
	contentStr := string(content)
	assert.Contains(t, contentStr, "# Task")
	assert.Contains(t, contentStr, "# Rules")
	assert.Contains(t, contentStr, "# Guardrails")
	assert.Contains(t, contentStr, "ONE task per session")
	assert.Contains(t, contentStr, "No placeholders or TODOs")
	assert.Contains(t, contentStr, "Run tests before committing")
}

func TestInitCmdNonInteractive(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "gumloop-init-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)
	require.NoError(t, os.Chdir(tmpDir))

	// Reset flag
	nonInteractive = true

	// Run init command directly (not via Execute which shows help)
	err = runInit(nil, []string{})
	require.NoError(t, err)

	// Verify .gumloop.yaml was created
	_, err = os.Stat(".gumloop.yaml")
	require.NoError(t, err)

	// Verify PROMPT.md was created
	_, err = os.Stat("PROMPT.md")
	require.NoError(t, err)

	// Read and verify config
	data, err := os.ReadFile(".gumloop.yaml")
	require.NoError(t, err)

	var cfg config.Config
	err = yaml.Unmarshal(data, &cfg)
	require.NoError(t, err)

	// Should use defaults
	defaults := config.Defaults()
	assert.Equal(t, defaults.CLI, cfg.CLI)
	assert.Equal(t, defaults.Model, cfg.Model)
	assert.Equal(t, defaults.PromptFile, cfg.PromptFile)
	assert.Equal(t, defaults.AutoPush, cfg.AutoPush)
	assert.Equal(t, defaults.StuckThreshold, cfg.StuckThreshold)
	assert.Equal(t, defaults.Verify, cfg.Verify)
}

func TestInitCmdRejectsExistingConfig(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "gumloop-init-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)
	require.NoError(t, os.Chdir(tmpDir))

	// Create existing .gumloop.yaml
	err = os.WriteFile(".gumloop.yaml", []byte("cli: claude\n"), 0644)
	require.NoError(t, err)

	// Reset flag
	nonInteractive = true

	// Try to run init command
	err = runInit(nil, []string{})

	// Should fail with error about existing file
	require.Error(t, err)
	assert.Contains(t, err.Error(), "config file already exists")
	assert.Contains(t, err.Error(), ".gumloop.yaml")
}

func TestWriteConfigFileContainsHeader(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "gumloop-init-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)
	require.NoError(t, os.Chdir(tmpDir))

	// Create test config
	cfg := config.Defaults()

	// Write config file
	initGlobal = false // Ensure we're writing project config
	err = writeConfigFileToPath(cfg, ".gumloop.yaml")
	require.NoError(t, err)

	// Read raw file content
	data, err := os.ReadFile(".gumloop.yaml")
	require.NoError(t, err)

	content := string(data)

	// Verify header comments are present
	assert.Contains(t, content, "# gumloop project configuration")
	assert.Contains(t, content, "# Docs: https://github.com/adriancodes/gumloop")
}

func TestWriteConfigFileCreatesValidYAML(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "gumloop-init-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)
	require.NoError(t, os.Chdir(tmpDir))

	// Test various config combinations
	testCases := []struct {
		name   string
		config config.Config
	}{
		{
			name: "all fields populated",
			config: config.Config{
				CLI:            "gemini",
				Model:          "gemini-2.0-flash-exp",
				PromptFile:     "PROMPT.md",
				AutoPush:       true,
				StuckThreshold: 5,
				Verify:         "npm test",
			},
		},
		{
			name: "minimal config",
			config: config.Config{
				CLI:            "ollama",
				Model:          "qwen2.5-coder",
				PromptFile:     "PROMPT.md",
				AutoPush:       false,
				StuckThreshold: 3,
				Verify:         "",
			},
		},
		{
			name: "special characters in verify",
			config: config.Config{
				CLI:            "cursor",
				Model:          "",
				PromptFile:     "PROMPT.md",
				AutoPush:       true,
				StuckThreshold: 3,
				Verify:         "npm run test && npm run lint",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Write config
			filename := filepath.Join(tmpDir, tc.name+".yaml")
			f, err := os.Create(filename)
			require.NoError(t, err)

			encoder := yaml.NewEncoder(f)
			encoder.SetIndent(2)
			err = encoder.Encode(&tc.config)
			require.NoError(t, err)
			encoder.Close()
			f.Close()

			// Read back and verify it's valid YAML
			data, err := os.ReadFile(filename)
			require.NoError(t, err)

			var parsed config.Config
			err = yaml.Unmarshal(data, &parsed)
			require.NoError(t, err)

			// Verify all fields match
			assert.Equal(t, tc.config.CLI, parsed.CLI)
			assert.Equal(t, tc.config.Model, parsed.Model)
			assert.Equal(t, tc.config.PromptFile, parsed.PromptFile)
			assert.Equal(t, tc.config.AutoPush, parsed.AutoPush)
			assert.Equal(t, tc.config.StuckThreshold, parsed.StuckThreshold)
			assert.Equal(t, tc.config.Verify, parsed.Verify)
		})
	}
}
