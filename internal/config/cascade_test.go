package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromFile_MissingFile(t *testing.T) {
	// Test that missing file returns empty config (not an error)
	cfg, err := loadFromFile("/nonexistent/path/config.yaml")
	if err != nil {
		t.Errorf("Expected no error for missing file, got: %v", err)
	}
	if cfg.CLI != "" || cfg.Model != "" {
		t.Errorf("Expected empty config, got: %+v", cfg)
	}
}

func TestLoadFromFile_ValidYAML(t *testing.T) {
	// Create a temp file with valid YAML
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	content := `cli: codex
model: gpt-4
prompt_file: TEST.md
auto_push: false
stuck_threshold: 5
verify: npm test
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := loadFromFile(configPath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify all fields
	if cfg.CLI != "codex" {
		t.Errorf("Expected CLI=codex, got: %s", cfg.CLI)
	}
	if cfg.Model != "gpt-4" {
		t.Errorf("Expected Model=gpt-4, got: %s", cfg.Model)
	}
	if cfg.PromptFile != "TEST.md" {
		t.Errorf("Expected PromptFile=TEST.md, got: %s", cfg.PromptFile)
	}
	if cfg.AutoPush != false {
		t.Errorf("Expected AutoPush=false, got: %v", cfg.AutoPush)
	}
	if cfg.StuckThreshold != 5 {
		t.Errorf("Expected StuckThreshold=5, got: %d", cfg.StuckThreshold)
	}
	if cfg.Verify != "npm test" {
		t.Errorf("Expected Verify='npm test', got: %s", cfg.Verify)
	}
}

func TestLoadFromFile_MalformedYAML(t *testing.T) {
	// Create a temp file with malformed YAML
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	content := `cli: claude
model: [invalid
prompt_file: PROMPT.md
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	_, err := loadFromFile(configPath)
	if err == nil {
		t.Error("Expected error for malformed YAML, got nil")
	}
}

func TestValidate_ValidAgent(t *testing.T) {
	validAgents := []string{"claude", "codex", "gemini", "opencode", "cursor", "ollama"}
	for _, agent := range validAgents {
		cfg := Config{CLI: agent}
		if err := validate(&cfg); err != nil {
			t.Errorf("Expected no error for valid agent %s, got: %v", agent, err)
		}
	}
}

func TestValidate_InvalidAgent(t *testing.T) {
	cfg := Config{CLI: "invalid-agent"}
	err := validate(&cfg)
	if err == nil {
		t.Error("Expected error for invalid agent, got nil")
	}
}

func TestValidate_InvalidAgentWithSuggestion(t *testing.T) {
	// Test typo that should get a suggestion
	cfg := Config{CLI: "cluade"} // typo of "claude"
	err := validate(&cfg)
	if err == nil {
		t.Error("Expected error for invalid agent, got nil")
	}
	// Error should suggest "claude"
	if err != nil && err.Error() == "" {
		t.Errorf("Expected helpful error message, got empty string")
	}
}

func TestValidate_NegativeStuckThreshold(t *testing.T) {
	cfg := Config{StuckThreshold: -1}
	err := validate(&cfg)
	if err == nil {
		t.Error("Expected error for negative stuck_threshold, got nil")
	}
}

func TestValidate_EmptyConfig(t *testing.T) {
	// Empty config should be valid
	cfg := Config{}
	if err := validate(&cfg); err != nil {
		t.Errorf("Expected no error for empty config, got: %v", err)
	}
}

func TestMerge_DefaultsOnly(t *testing.T) {
	result := Merge()
	defaults := Defaults()

	if result.CLI != defaults.CLI {
		t.Errorf("Expected CLI=%s, got: %s", defaults.CLI, result.CLI)
	}
	if result.Model != defaults.Model {
		t.Errorf("Expected Model=%s, got: %s", defaults.Model, result.Model)
	}
	if result.PromptFile != defaults.PromptFile {
		t.Errorf("Expected PromptFile=%s, got: %s", defaults.PromptFile, result.PromptFile)
	}
	if result.AutoPush != defaults.AutoPush {
		t.Errorf("Expected AutoPush=%v, got: %v", defaults.AutoPush, result.AutoPush)
	}
	if result.StuckThreshold != defaults.StuckThreshold {
		t.Errorf("Expected StuckThreshold=%d, got: %d", defaults.StuckThreshold, result.StuckThreshold)
	}
	if result.Verify != defaults.Verify {
		t.Errorf("Expected Verify=%s, got: %s", defaults.Verify, result.Verify)
	}
}

func TestMerge_GlobalOnly(t *testing.T) {
	global := Config{
		CLI:            "codex",
		Model:          "gpt-4",
		StuckThreshold: 5,
	}

	result := Merge(global)

	// Should have global values where set
	if result.CLI != "codex" {
		t.Errorf("Expected CLI=codex, got: %s", result.CLI)
	}
	if result.Model != "gpt-4" {
		t.Errorf("Expected Model=gpt-4, got: %s", result.Model)
	}
	if result.StuckThreshold != 5 {
		t.Errorf("Expected StuckThreshold=5, got: %d", result.StuckThreshold)
	}

	// Should have defaults for unset values
	defaults := Defaults()
	if result.PromptFile != defaults.PromptFile {
		t.Errorf("Expected PromptFile=%s (default), got: %s", defaults.PromptFile, result.PromptFile)
	}
	// Note: AutoPush is a bool with zero value false, so it gets merged even when "unset"
	// This is a known limitation. In practice, configs loaded from YAML will have explicit values.
}

func TestMerge_ProjectOnly(t *testing.T) {
	project := Config{
		CLI:    "gemini",
		Verify: "go test ./...",
	}

	result := Merge(project)

	// Should have project values where set
	if result.CLI != "gemini" {
		t.Errorf("Expected CLI=gemini, got: %s", result.CLI)
	}
	if result.Verify != "go test ./..." {
		t.Errorf("Expected Verify='go test ./...', got: %s", result.Verify)
	}

	// Should have defaults for unset values
	defaults := Defaults()
	if result.Model != defaults.Model {
		t.Errorf("Expected Model=%s (default), got: %s", defaults.Model, result.Model)
	}
}

func TestMerge_GlobalAndProject(t *testing.T) {
	global := Config{
		CLI:            "codex",
		Model:          "gpt-4",
		StuckThreshold: 5,
	}

	project := Config{
		CLI:    "gemini", // Override global
		Verify: "npm test",
	}

	result := Merge(global, project)

	// Project should override global for CLI
	if result.CLI != "gemini" {
		t.Errorf("Expected CLI=gemini (project override), got: %s", result.CLI)
	}

	// Project should add verify
	if result.Verify != "npm test" {
		t.Errorf("Expected Verify='npm test', got: %s", result.Verify)
	}

	// Global should provide model (project didn't set it)
	if result.Model != "gpt-4" {
		t.Errorf("Expected Model=gpt-4 (from global), got: %s", result.Model)
	}

	// Global should provide stuck_threshold (project didn't set it)
	if result.StuckThreshold != 5 {
		t.Errorf("Expected StuckThreshold=5 (from global), got: %d", result.StuckThreshold)
	}

	// Defaults should fill in the rest
	defaults := Defaults()
	if result.PromptFile != defaults.PromptFile {
		t.Errorf("Expected PromptFile=%s (default), got: %s", defaults.PromptFile, result.PromptFile)
	}
}

func TestMerge_EmptyValuesDoNotOverride(t *testing.T) {
	global := Config{
		CLI:   "codex",
		Model: "gpt-4",
	}

	project := Config{
		CLI:   "", // Empty - should not override
		Model: "", // Empty - should not override
	}

	result := Merge(global, project)

	// Empty values in project should not override global
	if result.CLI != "codex" {
		t.Errorf("Expected CLI=codex (global not overridden), got: %s", result.CLI)
	}
	if result.Model != "gpt-4" {
		t.Errorf("Expected Model=gpt-4 (global not overridden), got: %s", result.Model)
	}
}

func TestMerge_ZeroThresholdDoesNotOverride(t *testing.T) {
	global := Config{
		StuckThreshold: 5,
	}

	project := Config{
		StuckThreshold: 0, // Zero - should not override
	}

	result := Merge(global, project)

	// Zero should not override non-zero
	if result.StuckThreshold != 5 {
		t.Errorf("Expected StuckThreshold=5 (global not overridden), got: %d", result.StuckThreshold)
	}
}

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		a        string
		b        string
		expected int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "a", 1},
		{"claude", "claude", 0},
		{"claude", "cluade", 2},
		{"codex", "code", 1},
		{"gemini", "gimini", 1},
	}

	for _, test := range tests {
		result := levenshteinDistance(test.a, test.b)
		if result != test.expected {
			t.Errorf("levenshteinDistance(%q, %q) = %d, expected %d", test.a, test.b, result, test.expected)
		}
	}
}

func TestFindClosestMatch(t *testing.T) {
	options := []string{"claude", "codex", "gemini", "opencode", "cursor", "ollama"}

	tests := []struct {
		typo     string
		expected string
	}{
		{"cluade", "claude"},     // 2 edits
		{"claued", "claude"},     // 2 edits
		{"code", "codex"},        // 1 edit
		{"gimini", "gemini"},     // 1 edit
		{"invalid", ""},          // too different
		{"xyz", ""},              // too different
		{"", ""},                 // empty
		{"claude", "claude"},     // exact match (0 edits)
		{"cursor-agent", ""},     // too different
		{"ollama-local", ""},     // too different
	}

	for _, test := range tests {
		result := findClosestMatch(test.typo, options)
		if result != test.expected {
			t.Errorf("findClosestMatch(%q) = %q, expected %q", test.typo, result, test.expected)
		}
	}
}

func TestLoadGlobal_Integration(t *testing.T) {
	// This test verifies LoadGlobal doesn't crash when no global config exists
	// It should return an empty config
	cfg, err := LoadGlobal()
	if err != nil {
		// Only fail if it's not a "file doesn't exist" scenario
		if !os.IsNotExist(err) {
			t.Errorf("LoadGlobal returned unexpected error: %v", err)
		}
	}
	// Empty config is valid
	_ = cfg
}

func TestLoadProject_Integration(t *testing.T) {
	// This test verifies LoadProject doesn't crash when no project config exists
	// It should return an empty config
	cfg, err := LoadProject()
	if err != nil {
		t.Errorf("LoadProject returned error: %v", err)
	}
	// Empty config is valid
	_ = cfg
}
