package agent

func init() {
	RegisterAgent(&Agent{
		ID:           "opencode",
		Name:         "OpenCode",
		Command:      "opencode",
		CheckCommand: "opencode",
		// Autonomous: -p for prompt, -q hides spinner, -f for format
		AutonomousFlags: []string{"-p", "-q", "-f", "text"},
		// Interactive: -p for prompt, -f for format (no -q to show spinner)
		InteractiveFlags: []string{"-p", "-f", "text"},
		// OpenCode uses config file (~/.opencode.json), not CLI flag
		ModelFlag: "",
		// Prompt is passed via -p flag as argument
		PromptStyle: PromptStyleArg,
	})
}
