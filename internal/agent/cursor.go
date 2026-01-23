package agent

func init() {
	RegisterAgent(&Agent{
		ID:           "cursor",
		Name:         "Cursor Agent",
		Command:      "cursor-agent",
		CheckCommand: "cursor-agent",
		// Autonomous: -p for prompt, --force for auto-approval, --output-format for text
		AutonomousFlags: []string{"-p", "--force", "--output-format", "text"},
		// Interactive: no flags (empty)
		InteractiveFlags: []string{},
		ModelFlag:        "--model",
		// Prompt is passed as argument
		PromptStyle: PromptStyleArg,
	})
}
