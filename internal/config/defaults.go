package config

// Defaults returns the default configuration values as defined in SPEC section 3.3.
func Defaults() Config {
	return Config{
		CLI:            "claude",
		Model:          "",
		PromptFile:     "PROMPT.md",
		AutoPush:       true,
		StuckThreshold: 3,
		Verify:         "",
		Memory:         false,
	}
}
