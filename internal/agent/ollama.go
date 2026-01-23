package agent

func init() {
	RegisterAgent(&Agent{
		ID:               "ollama",
		Name:             "Ollama",
		Command:          "ollama run",
		CheckCommand:     "ollama",
		AutonomousFlags:  []string{}, // No flags needed - Ollama is local, no permissions to skip
		InteractiveFlags: []string{}, // Same as autonomous - no difference for local execution
		ModelFlag:        "",          // Empty - model is positional, not a flag
		PromptStyle:      PromptStyleOllama,
	})
}
