package agent

func init() {
	RegisterAgent(&Agent{
		ID:           "gemini",
		Name:         "Google Gemini",
		Command:      "gemini",
		CheckCommand: "gemini",
		AutonomousFlags: []string{
			"-p",
			"--yolo",
			"--output-format",
			"text",
		},
		InteractiveFlags: []string{
			"-p",
			"--output-format",
			"text",
		},
		ModelFlag:   "--model",
		PromptStyle: PromptStyleArg,
	})
}
