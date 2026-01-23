package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Build information (set via ldflags during build)
var (
	GitCommit = "unknown" // Git commit hash
	BuildDate = "unknown" // Build date
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display the current version of gumloop along with build information.`,
	Run:   runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func runVersion(cmd *cobra.Command, args []string) {
	// Format: gumloop vX.Y.Z (goX.Y, OS/arch)
	// Version is already defined in root.go
	fmt.Printf("gumloop %s (go%s, %s/%s)\n",
		Version,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	)

	// Show additional build info in debug mode
	if Debug {
		fmt.Printf("\nBuild info:\n")
		fmt.Printf("  Commit:    %s\n", GitCommit)
		fmt.Printf("  BuildDate: %s\n", BuildDate)
	}
}
