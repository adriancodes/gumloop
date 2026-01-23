package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adriancodes/gumloop/internal/ui"
	"github.com/spf13/cobra"
)

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove gumloop from your system",
	Long: `Remove gumloop from your system.

This command will:
  1. Remove the gumloop binary from its install location
  2. Optionally remove global configuration (~/.config/gumloop/)

Project-specific .gumloop.yaml files are not removed.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return uninstall()
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}

func uninstall() error {
	// Get current executable path
	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}

	// Resolve symlinks
	currentExe, err = filepath.EvalSymlinks(currentExe)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	// Get config directory path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	configDir := filepath.Join(homeDir, ".config", "gumloop")

	// Check if config directory exists
	configExists := false
	if _, err := os.Stat(configDir); err == nil {
		configExists = true
	}

	// Show what will be removed
	fmt.Println(ui.WarningStyle.Render("⚠️  Uninstall gumloop"))
	fmt.Println()
	fmt.Println("The following will be removed:")
	fmt.Printf("  • Binary: %s\n", ui.MutedStyle.Render(currentExe))
	if configExists {
		fmt.Printf("  • Config: %s (optional)\n", ui.MutedStyle.Render(configDir))
	}
	fmt.Println()
	fmt.Println(ui.MutedStyle.Render("Project .gumloop.yaml files will not be removed."))
	fmt.Println()

	// Confirm removal
	if !confirm("Remove gumloop binary?") {
		fmt.Println(ui.MutedStyle.Render("Cancelled."))
		return nil
	}

	// Remove binary
	if err := os.Remove(currentExe); err != nil {
		return fmt.Errorf("failed to remove binary: %w", err)
	}
	fmt.Println(ui.SuccessStyle.Render("✓ Removed binary"))

	// Ask about config directory
	if configExists {
		if confirm("Remove global config directory?") {
			if err := os.RemoveAll(configDir); err != nil {
				return fmt.Errorf("failed to remove config directory: %w", err)
			}
			fmt.Println(ui.SuccessStyle.Render("✓ Removed config directory"))
		} else {
			fmt.Println(ui.MutedStyle.Render("  Kept config directory"))
		}
	}

	fmt.Println()
	fmt.Println(ui.SuccessStyle.Render("✓ gumloop has been uninstalled"))
	return nil
}

// confirm prompts the user for yes/no confirmation
func confirm(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s (y/N): ", prompt)

	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}
