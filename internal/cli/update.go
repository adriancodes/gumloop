package cli

import (
	"github.com/adriancodes/gumloop/internal/update"
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update gumloop to the latest version",
	Long: `Update gumloop to the latest version from GitHub releases.

This command will:
  1. Check for the latest release
  2. Download the appropriate binary for your OS/architecture
  3. Replace the current binary with the new version

The update is performed safely with a backup in case of failure.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return update.Update(Version)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
