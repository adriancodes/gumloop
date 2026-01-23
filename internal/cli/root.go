package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adriancodes/gumloop/internal/config"
	"github.com/adriancodes/gumloop/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Version is the current gumloop version (overridden by ldflags at build time)
	Version = "dev"
)

var (
	// Debug is set by the --debug flag
	Debug bool

	// cfgFile is set by the --config flag (optional)
	cfgFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gumloop",
	Short: "Run AI coding agents in autonomous loops",
	Long: `gumloop - Run AI coding agents in autonomous loops

Based on Geoffrey Huntley's "Ralph Wiggum" methodology: fresh context
per iteration, progress persists in git.

Loop termination IS the free(). Each iteration:
  fresh context → one task → commit → exit → repeat`,
	// Suppress default error handling - we'll handle it ourselves
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Persistent flags (available to all subcommands)
	rootCmd.PersistentFlags().BoolVar(&Debug, "debug", false, "Show debug output")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file (default is ./.gumloop.yaml or ~/.config/gumloop/config.yaml)")

	// Customize help template to include Ralph ASCII art and quote
	rootCmd.SetHelpTemplate(helpTemplate())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Set up config file search paths per SPEC section 3.4

		// Look for project config first (./.gumloop.yaml)
		viper.AddConfigPath(".")
		viper.SetConfigName(".gumloop")
		viper.SetConfigType("yaml")

		// Also check for global config (~/.config/gumloop/config.yaml)
		home, err := os.UserHomeDir()
		if err == nil {
			globalConfigDir := filepath.Join(home, ".config", "gumloop")
			viper.AddConfigPath(globalConfigDir)
			viper.SetConfigName("config")
		}
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil && Debug {
		fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
	}

	// Set defaults from the config package
	defaults := config.Defaults()
	viper.SetDefault("cli", defaults.CLI)
	viper.SetDefault("model", defaults.Model)
	viper.SetDefault("prompt_file", defaults.PromptFile)
	viper.SetDefault("auto_push", defaults.AutoPush)
	viper.SetDefault("stuck_threshold", defaults.StuckThreshold)
	viper.SetDefault("verify", defaults.Verify)
}

// helpTemplate returns a custom help template with Ralph ASCII art and a random quote
func helpTemplate() string {
	banner := ui.RenderHelpBanner(Version)

	// Add the standard cobra help template after the banner
	return banner + "\n" + `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
}
