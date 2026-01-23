package agent

import (
	"testing"
)

// setupOllamaTests ensures the ollama agent is registered before tests run.
// This is needed because other tests may clear the registry.
func setupOllamaTests() {
	// Re-register ollama agent for testing
	RegisterAgent(&Agent{
		ID:               "ollama",
		Name:             "Ollama",
		Command:          "ollama run",
		CheckCommand:     "ollama",
		AutonomousFlags:  []string{},
		InteractiveFlags: []string{},
		ModelFlag:        "",
		PromptStyle:      PromptStyleOllama,
	})
}

func TestOllamaAgent_Registration(t *testing.T) {
	setupOllamaTests()
	agent, err := GetAgent("ollama")
	if err != nil {
		t.Fatalf("Ollama agent not registered: %v", err)
	}

	if agent.ID != "ollama" {
		t.Errorf("Expected ID 'ollama', got '%s'", agent.ID)
	}

	if agent.Name != "Ollama" {
		t.Errorf("Expected Name 'Ollama', got '%s'", agent.Name)
	}

	if agent.Command != "ollama run" {
		t.Errorf("Expected Command 'ollama run', got '%s'", agent.Command)
	}

	if agent.CheckCommand != "ollama" {
		t.Errorf("Expected CheckCommand 'ollama', got '%s'", agent.CheckCommand)
	}

	if agent.ModelFlag != "" {
		t.Errorf("Expected ModelFlag to be empty (positional), got '%s'", agent.ModelFlag)
	}

	if agent.PromptStyle != PromptStyleOllama {
		t.Errorf("Expected PromptStyle 'ollama', got '%s'", agent.PromptStyle)
	}
}

func TestOllamaAgent_NoFlags(t *testing.T) {
	setupOllamaTests()
	agent, _ := GetAgent("ollama")

	// Ollama should have no autonomous flags (local execution, no permissions to skip)
	if len(agent.AutonomousFlags) != 0 {
		t.Errorf("Expected 0 autonomous flags (Ollama is local), got %d: %v",
			len(agent.AutonomousFlags), agent.AutonomousFlags)
	}

	// Interactive and autonomous should be identical for Ollama
	if len(agent.InteractiveFlags) != 0 {
		t.Errorf("Expected 0 interactive flags, got %d: %v",
			len(agent.InteractiveFlags), agent.InteractiveFlags)
	}
}

func TestOllamaAgent_BuildCommand_WithModel(t *testing.T) {
	setupOllamaTests()
	agent, _ := GetAgent("ollama")

	prompt := "Write a hello world program"
	model := "codellama"
	cmd := agent.BuildCommand(prompt, model, true)

	// Expected: ollama run codellama "Write a hello world program"
	// Model is positional, not a flag
	expectedCmd := []string{
		"ollama",
		"run",
		"codellama",
		"Write a hello world program",
	}

	if len(cmd) != len(expectedCmd) {
		t.Fatalf("Expected command length %d, got %d\nExpected: %v\nGot: %v",
			len(expectedCmd), len(cmd), expectedCmd, cmd)
	}

	for i, expected := range expectedCmd {
		if cmd[i] != expected {
			t.Errorf("cmd[%d]: expected '%s', got '%s'", i, expected, cmd[i])
		}
	}
}

func TestOllamaAgent_BuildCommand_InteractiveMode(t *testing.T) {
	setupOllamaTests()
	agent, _ := GetAgent("ollama")

	prompt := "Explain recursion"
	model := "llama2"
	cmd := agent.BuildCommand(prompt, model, false)

	// Expected: ollama run llama2 "Explain recursion"
	// Interactive and autonomous should produce identical commands for Ollama
	expectedCmd := []string{
		"ollama",
		"run",
		"llama2",
		"Explain recursion",
	}

	if len(cmd) != len(expectedCmd) {
		t.Fatalf("Expected command length %d, got %d\nExpected: %v\nGot: %v",
			len(expectedCmd), len(cmd), expectedCmd, cmd)
	}

	for i, expected := range expectedCmd {
		if cmd[i] != expected {
			t.Errorf("cmd[%d]: expected '%s', got '%s'", i, expected, cmd[i])
		}
	}
}

func TestOllamaAgent_BuildCommand_NoModel(t *testing.T) {
	setupOllamaTests()
	agent, _ := GetAgent("ollama")

	prompt := "Test without model"
	cmd := agent.BuildCommand(prompt, "", true)

	// Expected: ollama run "Test without model"
	// When model is empty, Ollama will fail at runtime, but BuildCommand should still work
	expectedCmd := []string{
		"ollama",
		"run",
		"Test without model",
	}

	if len(cmd) != len(expectedCmd) {
		t.Fatalf("Expected command length %d, got %d\nExpected: %v\nGot: %v",
			len(expectedCmd), len(cmd), expectedCmd, cmd)
	}

	for i, expected := range expectedCmd {
		if cmd[i] != expected {
			t.Errorf("cmd[%d]: expected '%s', got '%s'", i, expected, cmd[i])
		}
	}
}

func TestOllamaAgent_AutonomousVsInteractive(t *testing.T) {
	setupOllamaTests()
	agent, _ := GetAgent("ollama")

	prompt := "Same test for both modes"
	model := "mistral"

	autonomousCmd := agent.BuildCommand(prompt, model, true)
	interactiveCmd := agent.BuildCommand(prompt, model, false)

	// For Ollama, autonomous and interactive should be identical
	if len(autonomousCmd) != len(interactiveCmd) {
		t.Fatalf("Autonomous and interactive commands should be identical for Ollama\nAutonomous: %v\nInteractive: %v",
			autonomousCmd, interactiveCmd)
	}

	for i := range autonomousCmd {
		if autonomousCmd[i] != interactiveCmd[i] {
			t.Errorf("Mismatch at index %d: autonomous='%s', interactive='%s'",
				i, autonomousCmd[i], interactiveCmd[i])
		}
	}
}

func TestOllamaAgent_NoModelFlag(t *testing.T) {
	setupOllamaTests()
	agent, _ := GetAgent("ollama")

	// Build command with a model
	cmd := agent.BuildCommand("test", "llama2", true)

	// Verify that --model flag is NOT present anywhere
	for i, arg := range cmd {
		if arg == "--model" || arg == "-m" {
			t.Errorf("Command should not contain model flags (found '%s' at index %d). Model should be positional.", arg, i)
		}
	}

	// Verify the model appears as a positional argument after "run"
	// Expected structure: [ollama, run, llama2, test]
	if len(cmd) < 3 {
		t.Fatalf("Command too short to contain model as positional arg: %v", cmd)
	}

	if cmd[0] != "ollama" || cmd[1] != "run" {
		t.Fatalf("Expected command to start with 'ollama run', got: %v", cmd)
	}

	if cmd[2] != "llama2" {
		t.Errorf("Expected model 'llama2' at position 2, got '%s'", cmd[2])
	}
}
