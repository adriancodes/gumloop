package agent

import (
	"testing"
)

// setupCodexTests ensures the codex agent is registered before tests run.
func setupCodexTests() {
	// Re-register codex agent for testing
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

func TestCodexAgent_Registration(t *testing.T) {
	setupCodexTests()
	agent, err := GetAgent("codex")
	if err != nil {
		t.Fatalf("Codex agent not registered: %v", err)
	}

	if agent.ID != "codex" {
		t.Errorf("Expected ID 'codex', got '%s'", agent.ID)
	}

	if agent.Name != "OpenAI Codex" {
		t.Errorf("Expected Name 'OpenAI Codex', got '%s'", agent.Name)
	}

	if agent.Command != "codex exec" {
		t.Errorf("Expected Command 'codex exec', got '%s'", agent.Command)
	}

	if agent.CheckCommand != "codex" {
		t.Errorf("Expected CheckCommand 'codex', got '%s'", agent.CheckCommand)
	}

	if agent.ModelFlag != "--model" {
		t.Errorf("Expected ModelFlag '--model', got '%s'", agent.ModelFlag)
	}

	if agent.PromptStyle != PromptStyleArg {
		t.Errorf("Expected PromptStyle 'arg', got '%s'", agent.PromptStyle)
	}
}

func TestCodexAgent_AutonomousFlags(t *testing.T) {
	setupCodexTests()
	agent, _ := GetAgent("codex")

	expectedFlags := []string{
		"--full-auto",
		"--json",
	}

	if len(agent.AutonomousFlags) != len(expectedFlags) {
		t.Fatalf("Expected %d autonomous flags, got %d", len(expectedFlags), len(agent.AutonomousFlags))
	}

	for i, expected := range expectedFlags {
		if agent.AutonomousFlags[i] != expected {
			t.Errorf("AutonomousFlags[%d]: expected '%s', got '%s'", i, expected, agent.AutonomousFlags[i])
		}
	}
}

func TestCodexAgent_InteractiveFlags(t *testing.T) {
	setupCodexTests()
	agent, _ := GetAgent("codex")

	expectedFlags := []string{
		"--json",
	}

	if len(agent.InteractiveFlags) != len(expectedFlags) {
		t.Fatalf("Expected %d interactive flags, got %d", len(expectedFlags), len(agent.InteractiveFlags))
	}

	for i, expected := range expectedFlags {
		if agent.InteractiveFlags[i] != expected {
			t.Errorf("InteractiveFlags[%d]: expected '%s', got '%s'", i, expected, agent.InteractiveFlags[i])
		}
	}

	// Verify --full-auto is NOT in interactive mode
	for _, flag := range agent.InteractiveFlags {
		if flag == "--full-auto" {
			t.Error("InteractiveFlags should not contain '--full-auto'")
		}
	}
}

func TestCodexAgent_BuildCommand_Autonomous(t *testing.T) {
	setupCodexTests()
	agent, _ := GetAgent("codex")

	prompt := "Fix the tests"
	model := "gpt-4"
	cmd := agent.BuildCommand(prompt, model, true)

	// Expected: codex exec --full-auto --json --model gpt-4 "Fix the tests"
	expectedCmd := []string{
		"codex",
		"exec",
		"--full-auto",
		"--json",
		"--model",
		"gpt-4",
		"Fix the tests",
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

func TestCodexAgent_BuildCommand_Interactive(t *testing.T) {
	setupCodexTests()
	agent, _ := GetAgent("codex")

	prompt := "Add a feature"
	model := "gpt-4"
	cmd := agent.BuildCommand(prompt, model, false)

	// Expected: codex exec --json --model gpt-4 "Add a feature"
	// Note: --full-auto should NOT be present
	expectedCmd := []string{
		"codex",
		"exec",
		"--json",
		"--model",
		"gpt-4",
		"Add a feature",
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

	// Verify --full-auto is NOT present
	for _, arg := range cmd {
		if arg == "--full-auto" {
			t.Error("Interactive command should not contain '--full-auto'")
		}
	}
}

func TestCodexAgent_BuildCommand_NoModel(t *testing.T) {
	setupCodexTests()
	agent, _ := GetAgent("codex")

	prompt := "Test without model"
	cmd := agent.BuildCommand(prompt, "", true)

	// Expected: codex exec --full-auto --json "Test without model"
	// Model flag should NOT be present
	expectedCmd := []string{
		"codex",
		"exec",
		"--full-auto",
		"--json",
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

	// Verify --model is NOT present
	for i, arg := range cmd {
		if arg == "--model" {
			t.Errorf("Command should not contain '--model' when model is empty (found at index %d)", i)
		}
	}
}

func TestCodexAgent_CommandStructure(t *testing.T) {
	setupCodexTests()
	agent, _ := GetAgent("codex")

	// Test that the command properly splits into ["codex", "exec"]
	cmd := agent.BuildCommand("test", "", true)

	if cmd[0] != "codex" {
		t.Errorf("Expected first command element to be 'codex', got '%s'", cmd[0])
	}

	if cmd[1] != "exec" {
		t.Errorf("Expected second command element to be 'exec', got '%s'", cmd[1])
	}
}
