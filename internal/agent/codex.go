package agent

func init() {
	RegisterAgent(&Agent{
		ID:           "codex",
		Name:         "OpenAI Codex",
		Command:      "codex exec",
		CheckCommand: "codex",
		AutonomousFlags: []string{
			"--full-auto",
			"--json",
		},
		InteractiveFlags: []string{
			"--json",
		},
		ModelFlag:   "--model",
		PromptStyle: PromptStyleArg,
	})
}
