package agent

func init() {
	RegisterAgent(&Agent{
		ID:           "claude",
		Name:         "Claude Code",
		Command:      "claude",
		CheckCommand: "claude",
		AutonomousFlags: []string{
			"-p",
			"--dangerously-skip-permissions",
			"--verbose",
			"--output-format",
			"stream-json",
		},
		InteractiveFlags: []string{
			"-p",
			"--verbose",
			"--output-format",
			"stream-json",
		},
		ModelFlag:   "--model",
		PromptStyle: PromptStyleStream,
	})
}
