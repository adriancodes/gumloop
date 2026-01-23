package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/adriancodes/gumloop/internal/agent"
	"github.com/adriancodes/gumloop/internal/config"
	"github.com/adriancodes/gumloop/internal/git"
	"github.com/adriancodes/gumloop/internal/memory"
	"github.com/adriancodes/gumloop/internal/runner"
	"github.com/adriancodes/gumloop/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Run command flags
	runPrompt      string
	runPromptFile  string
	runCLI         string
	runModel       string
	runChooChoo    int
	runChooChooSet bool // Track if --choo-choo was explicitly set
	runNoPush      bool
	runStuck       int
	runVerify      string
	runMemory      bool
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run [flags]",
	Short: "Execute agent with prompt",
	Long: `Execute an AI coding agent with a prompt.

The run command is the main way to use gumloop. It executes your chosen
AI agent (Claude, Codex, Gemini, etc.) with a prompt, either once or in
a loop until the work is complete.

Examples:
  # Single run with inline prompt
  gumloop run -p "Fix the failing tests"

  # Loop mode until complete
  gumloop run --choo-choo -p "Implement the auth module"

  # Loop with max iterations
  gumloop run --choo-choo 20 -p "Migrate JS to TS"

  # Use specific agent and model
  gumloop run --cli codex --model gpt-4 -p "Add tests"

  # With verification
  gumloop run --choo-choo --verify "npm test" -p "Fix bugs"`,
	RunE: runRun,
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Define flags per SPEC section 2.2
	runCmd.Flags().StringVarP(&runPrompt, "prompt", "p", "", "Inline prompt text (required if no --prompt-file)")
	runCmd.Flags().StringVar(&runPromptFile, "prompt-file", "", "Path to prompt file (default from config)")
	runCmd.Flags().StringVar(&runCLI, "cli", "", "Agent to use (claude, codex, gemini, opencode, cursor, ollama)")
	runCmd.Flags().StringVar(&runModel, "model", "", "Model override")
	runCmd.Flags().IntVar(&runChooChoo, "choo-choo", 0, "Loop mode. Optional max iterations (0 = unlimited)")
	runCmd.Flags().BoolVar(&runNoPush, "no-push", false, "Don't push to remote")
	runCmd.Flags().IntVar(&runStuck, "stuck-threshold", 0, "Exit after N iterations without commits")
	runCmd.Flags().StringVar(&runVerify, "verify", "", "Command to run after each iteration")
	runCmd.Flags().BoolVar(&runMemory, "memory", false, "Enable session memory (persists context between runs)")

	// Track if --choo-choo was explicitly set (for distinguishing between not set and set to 0)
	runCmd.Flags().Lookup("choo-choo").NoOptDefVal = "-1" // Special value to indicate flag without value
}

func runRun(cmd *cobra.Command, args []string) error {
	// Load configuration using the cascade system
	cfg, err := loadRunConfig()
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}

	// Validate configuration
	if err := validateRunConfig(cfg); err != nil {
		// Check if this is a safety error that needs a special exit code
		if safetyErr, ok := err.(*SafetyError); ok {
			fmt.Fprintf(os.Stderr, "Error: %s\n", safetyErr.Message)
			os.Exit(int(safetyErr.Code))
		}
		return err
	}

	// Debug output
	if Debug {
		fmt.Fprintf(os.Stderr, "Run configuration:\n")
		fmt.Fprintf(os.Stderr, "  CLI: %s\n", cfg.CLI)
		fmt.Fprintf(os.Stderr, "  Model: %s\n", cfg.Model)
		fmt.Fprintf(os.Stderr, "  Prompt: %s\n", cfg.Prompt)
		fmt.Fprintf(os.Stderr, "  PromptFile: %s\n", cfg.PromptFile)
		fmt.Fprintf(os.Stderr, "  ChooChoo: %v (max: %d)\n", cfg.ChooChoo, cfg.MaxIterations)
		fmt.Fprintf(os.Stderr, "  AutoPush: %v\n", cfg.AutoPush)
		fmt.Fprintf(os.Stderr, "  StuckThreshold: %d\n", cfg.StuckThreshold)
		fmt.Fprintf(os.Stderr, "  Verify: %s\n", cfg.Verify)
	}

	// Get the agent
	ag, err := agent.GetAgent(cfg.CLI)
	if err != nil {
		return fmt.Errorf("agent error: %w", err)
	}

	// Load session memory if enabled
	var mem *memory.SessionMemory
	if cfg.Memory {
		existing, err := memory.Load(memory.DefaultFileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Warning: failed to load session memory: %v\n", err)
		}

		// Inject previous session context into the prompt
		if existing != nil {
			context := existing.ToPromptContext()
			if context != "" {
				cfg.Prompt = context + "\n" + cfg.Prompt
			}
		}

		// Create a fresh memory for this session
		branch, _ := git.GetBranch()
		mem = &memory.SessionMemory{
			StartedAt: time.Now(),
			Branch:    branch,
			AgentName: ag.Name,
		}
	}

	// Create and run the runner
	r := runner.New(&cfg.Config, cfg.Prompt, ag, cfg.ChooChoo, cfg.MaxIterations, mem)
	exitCode := r.Run()

	// Display run summary
	metrics := r.GetMetrics()
	summary := ui.RenderRunSummary(ui.SummaryConfig{
		Agent:      ag.Name,
		Iterations: metrics.Iterations,
		Commits:    metrics.Commits,
		Duration:   metrics.Duration(),
		ExitCode:   ui.ExitCode(exitCode),
	})
	fmt.Println()
	fmt.Println(summary)

	// Exit with the appropriate code
	os.Exit(int(exitCode))
	return nil
}

// RunConfig extends the base Config with run-specific fields
type RunConfig struct {
	config.Config
	Prompt        string // The actual prompt text (from -p or file)
	ChooChoo      bool   // Whether loop mode is enabled
	MaxIterations int    // Max iterations (0 = unlimited)
}

// loadRunConfig loads config from cascade (defaults → global → project → flags)
func loadRunConfig() (*RunConfig, error) {
	// Start with defaults
	defaults := config.Defaults()

	// Create base config from viper (which has already loaded files via initConfig)
	cfg := &RunConfig{
		Config: config.Config{
			CLI:            viper.GetString("cli"),
			Model:          viper.GetString("model"),
			PromptFile:     viper.GetString("prompt_file"),
			AutoPush:       viper.GetBool("auto_push"),
			StuckThreshold: viper.GetInt("stuck_threshold"),
			Verify:         viper.GetString("verify"),
		},
	}

	// Apply flag overrides (flags have highest priority)
	if runCLI != "" {
		cfg.CLI = runCLI
	}
	if runModel != "" {
		cfg.Model = runModel
	}
	if runPromptFile != "" {
		cfg.PromptFile = runPromptFile
	}
	if runNoPush {
		cfg.AutoPush = false // --no-push overrides config
	}
	if runStuck > 0 {
		cfg.StuckThreshold = runStuck
	}
	if runVerify != "" {
		cfg.Verify = runVerify
	}
	if runMemory {
		cfg.Memory = true
	}

	// Handle --choo-choo flag
	// The flag can be: not set, set without value (use -1 as signal), or set with value
	if runChooChoo == -1 {
		// Flag set without value (e.g., --choo-choo)
		cfg.ChooChoo = true
		cfg.MaxIterations = 0 // unlimited
	} else if runChooChoo > 0 {
		// Flag set with value (e.g., --choo-choo 20)
		cfg.ChooChoo = true
		cfg.MaxIterations = runChooChoo
	}
	// If runChooChoo == 0, flag was not set, so ChooChoo stays false

	// Handle prompt: inline (-p) takes precedence over file
	if runPrompt != "" {
		cfg.Prompt = runPrompt
	} else {
		// Load from prompt file
		promptFile := cfg.PromptFile
		if promptFile == "" {
			promptFile = defaults.PromptFile // Use default if not set
		}

		// Read prompt file if it exists
		if _, err := os.Stat(promptFile); err == nil {
			content, err := os.ReadFile(promptFile)
			if err != nil {
				return nil, fmt.Errorf("failed to read prompt file %s: %w", promptFile, err)
			}
			cfg.Prompt = string(content)
		}
	}

	return cfg, nil
}

// validateRunConfig validates the run configuration
func validateRunConfig(cfg *RunConfig) error {
	// Must have a prompt
	if cfg.Prompt == "" {
		return fmt.Errorf("prompt required: use -p flag or create %s", cfg.PromptFile)
	}

	// Validate stuck threshold
	if cfg.StuckThreshold < 0 {
		return fmt.Errorf("stuck_threshold must be a positive integer, got %d", cfg.StuckThreshold)
	}

	// Validate max iterations
	if cfg.MaxIterations < 0 {
		return fmt.Errorf("max iterations must be non-negative, got %d", cfg.MaxIterations)
	}

	// Validate agent exists
	if _, err := agent.GetAgent(cfg.CLI); err != nil {
		return fmt.Errorf("invalid agent: %w", err)
	}

	// Safety check: Must be in a git repository
	if !git.IsInsideWorkTree() {
		return &SafetyError{
			Code:    runner.ExitSafety,
			Message: "not in a git repository. Initialize with: git init",
		}
	}

	// Safety check: Refuse dangerous paths (no override)
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	if git.IsDangerousPath(cwd) {
		return &SafetyError{
			Code:    runner.ExitSafety,
			Message: fmt.Sprintf("refusing to run in dangerous path: %s\n\nFor safety, gumloop refuses to run in system directories.\nPlease run from a project directory.", cwd),
		}
	}

	// Safety check: Warn if in home subdirectory with choo-choo mode
	if cfg.ChooChoo && git.IsHomeSubdirectory(cwd) {
		if !git.ConfirmHomeSubdirectory() {
			return &SafetyError{
				Code:    runner.ExitInterrupt,
				Message: "cancelled by user",
			}
		}
	}

	return nil
}

// SafetyError represents a safety check failure with an associated exit code
type SafetyError struct {
	Code    runner.ExitCode
	Message string
}

func (e *SafetyError) Error() string {
	return e.Message
}
