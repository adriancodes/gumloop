package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadGlobal loads the global configuration from ~/.config/gumloop/config.yaml.
// If the file doesn't exist, it returns an empty config (which will be filled with defaults).
// If the file exists but is malformed, it returns an error with context.
func LoadGlobal() (Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Config{}, fmt.Errorf("failed to get home directory: %w", err)
	}

	globalPath := filepath.Join(homeDir, ".config", "gumloop", "config.yaml")
	return loadFromFile(globalPath)
}

// LoadProject loads the project configuration from ./.gumloop.yaml in the current directory.
// If the file doesn't exist, it returns an empty config (which will be filled with defaults).
// If the file exists but is malformed, it returns an error with context.
func LoadProject() (Config, error) {
	projectPath := ".gumloop.yaml"
	return loadFromFile(projectPath)
}

// loadFromFile loads a config from the specified path.
// Returns an empty config if the file doesn't exist (not an error).
// Returns an error if the file exists but cannot be parsed.
func loadFromFile(path string) (Config, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// File doesn't exist - return empty config (will use defaults)
		return Config{}, nil
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	// Parse YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		// Try to extract line number from yaml error
		var yamlErr *yaml.TypeError
		if errors.As(err, &yamlErr) {
			return Config{}, fmt.Errorf("invalid config at %s: %s", path, yamlErr.Error())
		}
		return Config{}, fmt.Errorf("invalid config at %s: %w", path, err)
	}

	// Validate the config
	if err := validate(&cfg); err != nil {
		return Config{}, fmt.Errorf("invalid config at %s: %w", path, err)
	}

	return cfg, nil
}

// validate checks if the config values are valid.
// Returns an error if any values are invalid with helpful suggestions.
func validate(cfg *Config) error {
	// Validate CLI agent
	if cfg.CLI != "" {
		validAgents := []string{"claude", "codex", "gemini", "opencode", "cursor", "ollama"}
		valid := false
		for _, agent := range validAgents {
			if cfg.CLI == agent {
				valid = true
				break
			}
		}
		if !valid {
			// Try to suggest a close match
			suggestion := findClosestMatch(cfg.CLI, validAgents)
			if suggestion != "" {
				return fmt.Errorf("unknown agent '%s' (available: %v). Did you mean '%s'?", cfg.CLI, validAgents, suggestion)
			}
			return fmt.Errorf("unknown agent '%s' (available: %v)", cfg.CLI, validAgents)
		}
	}

	// Validate stuck_threshold
	if cfg.StuckThreshold < 0 {
		return fmt.Errorf("stuck_threshold must be a positive integer, got '%d'", cfg.StuckThreshold)
	}

	return nil
}

// findClosestMatch finds the closest match for a typo using simple edit distance.
// Returns empty string if no close match found.
func findClosestMatch(typo string, options []string) string {
	if len(typo) == 0 {
		return ""
	}

	// Simple heuristic: if typo is one character different, suggest it
	for _, option := range options {
		if levenshteinDistance(typo, option) <= 2 {
			return option
		}
	}
	return ""
}

// levenshteinDistance calculates the edit distance between two strings.
func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	// Create a matrix
	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	// Fill the matrix
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(a)][len(b)]
}

// min returns the minimum of three integers.
func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// Merge merges multiple configs with priority: later configs override earlier ones.
// Empty string and zero values in higher-priority configs are ignored (don't override).
func Merge(configs ...Config) Config {
	result := Defaults()

	for _, cfg := range configs {
		// CLI: override if non-empty
		if cfg.CLI != "" {
			result.CLI = cfg.CLI
		}

		// Model: override if non-empty
		if cfg.Model != "" {
			result.Model = cfg.Model
		}

		// PromptFile: override if non-empty
		if cfg.PromptFile != "" {
			result.PromptFile = cfg.PromptFile
		}

		// AutoPush: always override (bool has no "empty" value, so we need to track if it was set)
		// For now, we override if it's different from the default
		// This is a limitation of using plain bool - consider using *bool in the future
		result.AutoPush = cfg.AutoPush

		// StuckThreshold: override if non-zero
		if cfg.StuckThreshold != 0 {
			result.StuckThreshold = cfg.StuckThreshold
		}

		// Verify: override if non-empty
		if cfg.Verify != "" {
			result.Verify = cfg.Verify
		}

		// Memory: always override (same limitation as AutoPush)
		result.Memory = cfg.Memory
	}

	return result
}
