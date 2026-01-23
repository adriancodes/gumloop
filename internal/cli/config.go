package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adriancodes/gumloop/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	// globalFlag is set by the --global flag for config commands
	globalFlag bool
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration values",
	Long: `Manage gumloop configuration values.

Configuration is stored in two locations:
  - Global: ~/.config/gumloop/config.yaml
  - Project: ./.gumloop.yaml

Project config overrides global config. CLI flags override both.`,
}

// configSetCmd sets a configuration value
var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a config value",
	Long: `Set a configuration value.

By default, sets the value in the project config (./.gumloop.yaml).
Use --global to set in the global config (~/.config/gumloop/config.yaml).

Valid keys: cli, model, prompt_file, auto_push, stuck_threshold, verify`,
	Args: cobra.ExactArgs(2),
	RunE: runConfigSet,
}

// configGetCmd gets a configuration value
var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a config value",
	Long: `Get a configuration value.

By default, gets the effective value (merged from all sources).
Use --global to get only from the global config.

Valid keys: cli, model, prompt_file, auto_push, stuck_threshold, verify`,
	Args: cobra.ExactArgs(1),
	RunE: runConfigGet,
}

// configListCmd lists all configuration values
var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all config values",
	Long: `List all configuration values.

By default, lists the project config (./.gumloop.yaml).
Use --global to list the global config (~/.config/gumloop/config.yaml).`,
	Args: cobra.NoArgs,
	RunE: runConfigList,
}

// configShowCmd shows the effective merged configuration
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show effective merged config",
	Long: `Show the effective merged configuration.

This displays all config values with their sources:
  - default: Built-in default value
  - global: From ~/.config/gumloop/config.yaml
  - project: From ./.gumloop.yaml
  - flag: From CLI flag (if applicable)`,
	Args: cobra.NoArgs,
	RunE: runConfigShow,
}

func init() {
	rootCmd.AddCommand(configCmd)

	// Add subcommands
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configShowCmd)

	// Add --global flag to set, get, and list
	configSetCmd.Flags().BoolVar(&globalFlag, "global", false, "Use global config instead of project config")
	configGetCmd.Flags().BoolVar(&globalFlag, "global", false, "Use global config instead of project config")
	configListCmd.Flags().BoolVar(&globalFlag, "global", false, "Use global config instead of project config")
}

// runConfigSet sets a configuration value
func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	// Validate key
	validKeys := []string{"cli", "model", "prompt_file", "auto_push", "stuck_threshold", "verify", "memory"}
	if !contains(validKeys, key) {
		return fmt.Errorf("unknown config key '%s' (valid keys: %s)", key, strings.Join(validKeys, ", "))
	}

	// Determine which file to write to
	var configPath string
	if globalFlag {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		configDir := filepath.Join(homeDir, ".config", "gumloop")
		configPath = filepath.Join(configDir, "config.yaml")

		// Create directory if it doesn't exist
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
	} else {
		configPath = ".gumloop.yaml"
	}

	// Load existing config from file
	var cfg config.Config
	data, err := os.ReadFile(configPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	if err == nil {
		// File exists, parse it
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("failed to parse existing config: %w", err)
		}
	}

	// Update the value
	if err := setConfigValue(&cfg, key, value); err != nil {
		return err
	}

	// Write back to file
	data, err = yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	scope := "project"
	if globalFlag {
		scope = "global"
	}
	fmt.Printf("Set %s config: %s = %s\n", scope, key, value)

	return nil
}

// runConfigGet gets a configuration value
func runConfigGet(cmd *cobra.Command, args []string) error {
	key := args[0]

	// Validate key
	validKeys := []string{"cli", "model", "prompt_file", "auto_push", "stuck_threshold", "verify", "memory"}
	if !contains(validKeys, key) {
		return fmt.Errorf("unknown config key '%s' (valid keys: %s)", key, strings.Join(validKeys, ", "))
	}

	var cfg config.Config
	var err error

	if globalFlag {
		// Get from global only
		cfg, err = config.LoadGlobal()
		if err != nil {
			return fmt.Errorf("failed to load global config: %w", err)
		}
	} else {
		// Get effective value (merged)
		defaults := config.Defaults()
		global, err := config.LoadGlobal()
		if err != nil {
			return fmt.Errorf("failed to load global config: %w", err)
		}
		project, err := config.LoadProject()
		if err != nil {
			return fmt.Errorf("failed to load project config: %w", err)
		}
		cfg = config.Merge(defaults, global, project)
	}

	// Get the value
	value, err := getConfigValue(&cfg, key)
	if err != nil {
		return err
	}

	fmt.Println(value)
	return nil
}

// runConfigList lists all configuration values
func runConfigList(cmd *cobra.Command, args []string) error {
	var cfg config.Config
	var err error
	var source string

	if globalFlag {
		cfg, err = config.LoadGlobal()
		source = "global config (~/.config/gumloop/config.yaml)"
	} else {
		cfg, err = config.LoadProject()
		source = "project config (./.gumloop.yaml)"
	}

	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Printf("%s:\n\n", source)
	printConfig(cfg)
	return nil
}

// runConfigShow shows the effective merged configuration
func runConfigShow(cmd *cobra.Command, args []string) error {
	// Load all config layers
	defaults := config.Defaults()
	global, err := config.LoadGlobal()
	if err != nil {
		return fmt.Errorf("failed to load global config: %w", err)
	}
	project, err := config.LoadProject()
	if err != nil {
		return fmt.Errorf("failed to load project config: %w", err)
	}

	// Merge to get effective config
	effective := config.Merge(defaults, global, project)

	fmt.Println("Effective configuration:")
	fmt.Println()

	// Print each value with its source
	printValueWithSource("cli", effective.CLI, defaults, global, project)
	printValueWithSource("model", effective.Model, defaults, global, project)
	printValueWithSource("prompt_file", effective.PromptFile, defaults, global, project)
	printValueWithSource("auto_push", fmt.Sprintf("%t", effective.AutoPush), defaults, global, project)
	printValueWithSource("stuck_threshold", fmt.Sprintf("%d", effective.StuckThreshold), defaults, global, project)
	printValueWithSource("verify", effective.Verify, defaults, global, project)
	printValueWithSource("memory", fmt.Sprintf("%t", effective.Memory), defaults, global, project)

	return nil
}

// setConfigValue sets a value in the config struct based on the key
func setConfigValue(cfg *config.Config, key, value string) error {
	switch key {
	case "cli":
		// Validate agent name
		validAgents := []string{"claude", "codex", "gemini", "opencode", "cursor", "ollama"}
		if !contains(validAgents, value) {
			return fmt.Errorf("invalid agent '%s' (valid: %s)", value, strings.Join(validAgents, ", "))
		}
		cfg.CLI = value
	case "model":
		cfg.Model = value
	case "prompt_file":
		cfg.PromptFile = value
	case "auto_push":
		// Parse boolean
		if value == "true" {
			cfg.AutoPush = true
		} else if value == "false" {
			cfg.AutoPush = false
		} else {
			return fmt.Errorf("auto_push must be 'true' or 'false', got '%s'", value)
		}
	case "stuck_threshold":
		// Parse integer
		var threshold int
		if _, err := fmt.Sscanf(value, "%d", &threshold); err != nil {
			return fmt.Errorf("stuck_threshold must be an integer, got '%s'", value)
		}
		if threshold < 0 {
			return fmt.Errorf("stuck_threshold must be positive, got %d", threshold)
		}
		cfg.StuckThreshold = threshold
	case "verify":
		cfg.Verify = value
	case "memory":
		if value == "true" {
			cfg.Memory = true
		} else if value == "false" {
			cfg.Memory = false
		} else {
			return fmt.Errorf("memory must be 'true' or 'false', got '%s'", value)
		}
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}
	return nil
}

// getConfigValue gets a value from the config struct based on the key
func getConfigValue(cfg *config.Config, key string) (string, error) {
	switch key {
	case "cli":
		return cfg.CLI, nil
	case "model":
		return cfg.Model, nil
	case "prompt_file":
		return cfg.PromptFile, nil
	case "auto_push":
		return fmt.Sprintf("%t", cfg.AutoPush), nil
	case "stuck_threshold":
		return fmt.Sprintf("%d", cfg.StuckThreshold), nil
	case "verify":
		return cfg.Verify, nil
	case "memory":
		return fmt.Sprintf("%t", cfg.Memory), nil
	default:
		return "", fmt.Errorf("unknown config key: %s", key)
	}
}

// printConfig prints a config struct in a readable format
func printConfig(cfg config.Config) {
	fmt.Printf("  cli:             %s\n", formatValue(cfg.CLI))
	fmt.Printf("  model:           %s\n", formatValue(cfg.Model))
	fmt.Printf("  prompt_file:     %s\n", formatValue(cfg.PromptFile))
	fmt.Printf("  auto_push:       %t\n", cfg.AutoPush)
	fmt.Printf("  stuck_threshold: %d\n", cfg.StuckThreshold)
	fmt.Printf("  verify:          %s\n", formatValue(cfg.Verify))
	fmt.Printf("  memory:          %t\n", cfg.Memory)
}

// printValueWithSource prints a value with its source
func printValueWithSource(key, effectiveValue string, defaults, global, project config.Config) {
	source := "default"

	// Determine source by comparing with each layer
	switch key {
	case "cli":
		if project.CLI != "" && project.CLI == effectiveValue {
			source = "project"
		} else if global.CLI != "" && global.CLI == effectiveValue {
			source = "global"
		}
	case "model":
		if project.Model != "" && project.Model == effectiveValue {
			source = "project"
		} else if global.Model != "" && global.Model == effectiveValue {
			source = "global"
		}
	case "prompt_file":
		if project.PromptFile != "" && project.PromptFile == effectiveValue {
			source = "project"
		} else if global.PromptFile != "" && global.PromptFile == effectiveValue {
			source = "global"
		}
	case "auto_push":
		// For bool, we can't easily determine the source, so check if it differs from default
		defaultValue := defaults.AutoPush
		if project.AutoPush != defaultValue {
			source = "project"
		} else if global.AutoPush != defaultValue {
			source = "global"
		}
	case "stuck_threshold":
		if project.StuckThreshold != 0 && fmt.Sprintf("%d", project.StuckThreshold) == effectiveValue {
			source = "project"
		} else if global.StuckThreshold != 0 && fmt.Sprintf("%d", global.StuckThreshold) == effectiveValue {
			source = "global"
		}
	case "verify":
		if project.Verify != "" && project.Verify == effectiveValue {
			source = "project"
		} else if global.Verify != "" && global.Verify == effectiveValue {
			source = "global"
		}
	case "memory":
		defaultValue := defaults.Memory
		if project.Memory != defaultValue {
			source = "project"
		} else if global.Memory != defaultValue {
			source = "global"
		}
	}

	fmt.Printf("  %-17s %-15s (from: %s)\n", key+":", formatValue(effectiveValue), source)
}

// formatValue formats a value for display (empty string becomes "(not set)")
func formatValue(s string) string {
	if s == "" {
		return "(not set)"
	}
	return s
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
