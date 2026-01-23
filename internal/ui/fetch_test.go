package ui

import (
	"fmt"
	"testing"
)

func TestFetchModels(t *testing.T) {
	agents := []string{"claude", "codex", "gemini", "ollama", "cursor", "opencode"}
	for _, agent := range agents {
		models := modelsForAgent(agent)
		fmt.Printf("\n=== %s (%d models) ===\n", agent, len(models))
		for i, m := range models {
			fmt.Printf("  %d: %s (%s)\n", i, m.Name, m.ID)
		}
	}
}
