package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adriancodes/gumloop/internal/config"
	"github.com/adriancodes/gumloop/internal/ui"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	// nonInteractive is set by the --non-interactive flag
	nonInteractive bool
	// initGlobal is set by the --global flag
	initGlobal bool
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Interactive setup wizard for new projects",
	Long: `Interactive setup wizard that guides you through configuring gumloop.

By default, creates a project config (.gumloop.yaml) in the current directory
and optionally creates a PROMPT.md template.

Use --global to create global config (~/.config/gumloop/config.yaml) instead,
which applies to all projects that don't have their own .gumloop.yaml.

Use --non-interactive to skip the wizard and use defaults.`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&nonInteractive, "non-interactive", false, "Skip wizard, use defaults")
	initCmd.Flags().BoolVar(&initGlobal, "global", false, "Create global config instead of project config")
}

// runInit executes the init command logic
func runInit(cmd *cobra.Command, args []string) error {
	// Determine config file path
	configPath, err := getInitConfigPath()
	if err != nil {
		return err
	}

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		// In non-interactive mode, don't overwrite
		if nonInteractive {
			if initGlobal {
				return fmt.Errorf("global config already exists: %s\n\nUse 'gumloop config set --global' to modify existing config", configPath)
			}
			return fmt.Errorf("config file already exists: .gumloop.yaml\n\nUse 'gumloop config set' to modify existing config")
		}
		// Ask user if they want to overwrite
		warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214")) // Orange
		if initGlobal {
			fmt.Println(warnStyle.Render(fmt.Sprintf("Global config already exists: %s", configPath)))
		} else {
			fmt.Println(warnStyle.Render("Config file already exists: .gumloop.yaml"))
		}
		if !confirmOverwrite() {
			fmt.Println()
			fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("No changes made."))
			fmt.Println()
			return nil
		}
		fmt.Println() // Add spacing before wizard
	}

	var wizardConfig *ui.WizardConfig

	if nonInteractive {
		// Use defaults in non-interactive mode
		defaults := config.Defaults()
		wizardConfig = &ui.WizardConfig{
			CLI:          defaults.CLI,
			Model:        defaults.Model,
			Verify:       defaults.Verify,
			CreatePrompt: !initGlobal, // Don't create PROMPT.md for global config
		}
	} else {
		// Launch interactive wizard
		wizardConfig, err = ui.RunWizard()
		if err != nil {
			// Check if user cancelled
			if errors.Is(err, ui.ErrWizardCancelled) {
				cancelStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("241")) // Gray
				fmt.Println()
				fmt.Println(cancelStyle.Render("Setup cancelled. No files were created."))
				fmt.Println()
				return nil // Return nil to avoid showing error message
			}
			return fmt.Errorf("wizard failed: %w", err)
		}
		// Don't create PROMPT.md for global config
		if initGlobal {
			wizardConfig.CreatePrompt = false
		}
	}

	// Create config struct from wizard values
	cfg := config.Config{
		CLI:            wizardConfig.CLI,
		Model:          wizardConfig.Model,
		PromptFile:     "PROMPT.md",           // Always use PROMPT.md
		AutoPush:       true,                  // Default to auto-push
		StuckThreshold: 3,                     // Default stuck threshold
		Verify:         wizardConfig.Verify,
	}

	// Write config file
	if err := writeConfigFileToPath(cfg, configPath); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Create PROMPT.md if requested (only for project config)
	if wizardConfig.CreatePrompt && !initGlobal {
		if err := writePromptTemplate(); err != nil {
			return fmt.Errorf("failed to write PROMPT.md: %w", err)
		}
	}

	// Print success message
	printInitSuccessMessage(configPath, wizardConfig.CreatePrompt)

	return nil
}

// getInitConfigPath returns the appropriate config file path based on --global flag
func getInitConfigPath() (string, error) {
	if initGlobal {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		return filepath.Join(homeDir, ".config", "gumloop", "config.yaml"), nil
	}
	return ".gumloop.yaml", nil
}

// writeConfigFileToPath writes the configuration to the specified path
func writeConfigFileToPath(cfg config.Config, configPath string) error {
	// Create parent directories if needed (for global config)
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	f, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write header comment
	var header string
	if initGlobal {
		header = `# gumloop global configuration
# Applies to all projects without a .gumloop.yaml
# Docs: https://github.com/adriancodes/gumloop

`
	} else {
		header = `# gumloop project configuration
# Docs: https://github.com/adriancodes/gumloop

`
	}
	if _, err := f.WriteString(header); err != nil {
		return err
	}

	// Marshal config to YAML
	encoder := yaml.NewEncoder(f)
	encoder.SetIndent(2)

	if err := encoder.Encode(&cfg); err != nil {
		return err
	}

	return encoder.Close()
}

// writePromptTemplate writes the default PROMPT.md template
func writePromptTemplate() error {
	template := `# Task

[Describe what you want the agent to do here]

Example:
  "Add input validation to the login form"
  "Fix all failing tests"
  "Migrate the utils module from JavaScript to TypeScript"

# Plan

Track progress here. The agent will check items off as it works.

- [ ] (your first task)
- [ ] (your second task)

# Rules

ONE task per session. Pick the next unchecked [ ] item from the Plan,
implement it, mark it [x], commit, and exit.

Search the codebase before implementing to avoid duplicating existing code.

No placeholders or TODOs. Implement completely or don't implement at all.

Run tests before committing. If tests fail, fix them first.

Keep commits atomic - one logical change per commit.

# Guardrails

Add new guardrails here when you observe failure patterns.
Use high numbers for critical rules (e.g., 99999).
`

	return os.WriteFile("PROMPT.md", []byte(template), 0644)
}

// printInitSuccessMessage displays the success message with next steps
func printInitSuccessMessage(configPath string, createdPrompt bool) {
	successStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")). // Green
		Bold(true)

	mutedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")) // Gray

	fmt.Println()
	if initGlobal {
		fmt.Println(successStyle.Render("✅ Created global config: " + configPath))
	} else {
		fmt.Println(successStyle.Render("✅ Created .gumloop.yaml"))
	}

	if createdPrompt {
		fmt.Println(successStyle.Render("✅ Created PROMPT.md"))
	}

	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Render("Next steps:"))

	if initGlobal {
		fmt.Println(mutedStyle.Render("  Global defaults are now set for all projects."))
		fmt.Println(mutedStyle.Render("  Run 'gumloop init' in a project to create project-specific config."))
	} else if createdPrompt {
		fmt.Println(mutedStyle.Render("  1. Edit PROMPT.md with your task"))
		fmt.Println(mutedStyle.Render("  2. Run: gumloop run --choo-choo"))
	} else {
		fmt.Println(mutedStyle.Render("  1. Create a PROMPT.md with your task"))
		fmt.Println(mutedStyle.Render("  2. Run: gumloop run --choo-choo"))
	}
	fmt.Println()
}

// confirmOverwrite prompts the user to confirm overwriting existing config
func confirmOverwrite() bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Overwrite? (y/N): ")

	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}
