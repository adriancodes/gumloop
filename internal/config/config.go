package config

// Config represents the gumloop configuration structure.
// It can be loaded from multiple sources (defaults, global file, project file, CLI flags)
// and merged using a cascade priority system.
type Config struct {
	// CLI is the agent to use (claude, codex, gemini, opencode, cursor, ollama)
	CLI string `yaml:"cli" mapstructure:"cli"`

	// Model is the model override (agent-specific, empty string uses agent default)
	Model string `yaml:"model" mapstructure:"model"`

	// PromptFile is the default prompt file path
	PromptFile string `yaml:"prompt_file" mapstructure:"prompt_file"`

	// AutoPush determines whether to push to remote after commits
	AutoPush bool `yaml:"auto_push" mapstructure:"auto_push"`

	// StuckThreshold is the number of iterations with changes but no commits before exiting
	StuckThreshold int `yaml:"stuck_threshold" mapstructure:"stuck_threshold"`

	// Verify is the verification command to run after each iteration
	Verify string `yaml:"verify" mapstructure:"verify"`

	// Memory enables session memory persistence between runs
	Memory bool `yaml:"memory" mapstructure:"memory"`
}
