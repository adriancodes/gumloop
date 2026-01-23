package agent

import (
	"testing"
)

// setupOpenCodeTests ensures the opencode agent is registered before tests run.
// This is needed because other tests may clear the registry.
func setupOpenCodeTests() {
	// Re-register opencode agent for testing
	RegisterAgent(&Agent{
		ID:              "opencode",
		Name:            "OpenCode",
		Command:         "opencode",
		CheckCommand:    "opencode",
		AutonomousFlags: []string{"-p", "-q", "-f", "text"},
		InteractiveFlags: []string{"-p", "-f", "text"},
		ModelFlag:       "",
		PromptStyle:     PromptStyleArg,
	})
}

func TestOpenCodeAgent(t *testing.T) {
	setupOpenCodeTests()
	agent, err := GetAgent("opencode")
	if err != nil {
		t.Fatalf("OpenCode agent not registered: %v", err)
	}

	// Verify basic properties
	if agent.ID != "opencode" {
		t.Errorf("Expected ID 'opencode', got '%s'", agent.ID)
	}
	if agent.Name != "OpenCode" {
		t.Errorf("Expected Name 'OpenCode', got '%s'", agent.Name)
	}
	if agent.Command != "opencode" {
		t.Errorf("Expected Command 'opencode', got '%s'", agent.Command)
	}
	if agent.CheckCommand != "opencode" {
		t.Errorf("Expected CheckCommand 'opencode', got '%s'", agent.CheckCommand)
	}
	if agent.PromptStyle != PromptStyleArg {
		t.Errorf("Expected PromptStyle 'arg', got '%s'", agent.PromptStyle)
	}

	// Verify model flag is empty (OpenCode uses config file)
	if agent.ModelFlag != "" {
		t.Errorf("Expected empty ModelFlag, got '%s'", agent.ModelFlag)
	}

	// Verify autonomous flags
	expectedAutonomous := []string{"-p", "-q", "-f", "text"}
	if len(agent.AutonomousFlags) != len(expectedAutonomous) {
		t.Fatalf("Expected %d autonomous flags, got %d", len(expectedAutonomous), len(agent.AutonomousFlags))
	}
	for i, flag := range expectedAutonomous {
		if agent.AutonomousFlags[i] != flag {
			t.Errorf("Expected autonomous flag[%d] '%s', got '%s'", i, flag, agent.AutonomousFlags[i])
		}
	}

	// Verify interactive flags (no -q)
	expectedInteractive := []string{"-p", "-f", "text"}
	if len(agent.InteractiveFlags) != len(expectedInteractive) {
		t.Fatalf("Expected %d interactive flags, got %d", len(expectedInteractive), len(agent.InteractiveFlags))
	}
	for i, flag := range expectedInteractive {
		if agent.InteractiveFlags[i] != flag {
			t.Errorf("Expected interactive flag[%d] '%s', got '%s'", i, flag, agent.InteractiveFlags[i])
		}
	}
}

func TestOpenCodeBuildCommand_Autonomous(t *testing.T) {
	setupOpenCodeTests()
	agent, err := GetAgent("opencode")
	if err != nil {
		t.Fatalf("OpenCode agent not registered: %v", err)
	}

	prompt := "Fix the tests"
	cmd := agent.BuildCommand(prompt, "", true)

	// Expected: opencode -p -q -f text "Fix the tests"
	expected := []string{"opencode", "-p", "-q", "-f", "text", prompt}

	if len(cmd) != len(expected) {
		t.Fatalf("Expected command length %d, got %d\nExpected: %v\nGot: %v",
			len(expected), len(cmd), expected, cmd)
	}

	for i, arg := range expected {
		if cmd[i] != arg {
			t.Errorf("Expected arg[%d] '%s', got '%s'", i, arg, cmd[i])
		}
	}
}

func TestOpenCodeBuildCommand_Interactive(t *testing.T) {
	setupOpenCodeTests()
	agent, err := GetAgent("opencode")
	if err != nil {
		t.Fatalf("OpenCode agent not registered: %v", err)
	}

	prompt := "Add tests"
	cmd := agent.BuildCommand(prompt, "", false)

	// Expected: opencode -p -f text "Add tests"
	// Note: no -q flag in interactive mode
	expected := []string{"opencode", "-p", "-f", "text", prompt}

	if len(cmd) != len(expected) {
		t.Fatalf("Expected command length %d, got %d\nExpected: %v\nGot: %v",
			len(expected), len(cmd), expected, cmd)
	}

	for i, arg := range expected {
		if cmd[i] != arg {
			t.Errorf("Expected arg[%d] '%s', got '%s'", i, arg, cmd[i])
		}
	}
}

func TestOpenCodeBuildCommand_WithModel(t *testing.T) {
	setupOpenCodeTests()
	agent, err := GetAgent("opencode")
	if err != nil {
		t.Fatalf("OpenCode agent not registered: %v", err)
	}

	prompt := "Refactor code"
	model := "gpt-4"
	cmd := agent.BuildCommand(prompt, model, true)

	// Expected: opencode -p -q -f text "Refactor code"
	// Model should NOT be added since OpenCode uses config file, not CLI
	expected := []string{"opencode", "-p", "-q", "-f", "text", prompt}

	if len(cmd) != len(expected) {
		t.Fatalf("Expected command length %d, got %d (model should not be added)\nExpected: %v\nGot: %v",
			len(expected), len(cmd), expected, cmd)
	}

	for i, arg := range expected {
		if cmd[i] != arg {
			t.Errorf("Expected arg[%d] '%s', got '%s'", i, arg, cmd[i])
		}
	}
}

func TestOpenCodeBuildCommand_PromptWithSpaces(t *testing.T) {
	setupOpenCodeTests()
	agent, err := GetAgent("opencode")
	if err != nil {
		t.Fatalf("OpenCode agent not registered: %v", err)
	}

	prompt := "Fix the failing tests in src/main.go"
	cmd := agent.BuildCommand(prompt, "", true)

	// Prompt should be preserved as a single argument
	expected := []string{"opencode", "-p", "-q", "-f", "text", prompt}

	if len(cmd) != len(expected) {
		t.Fatalf("Expected command length %d, got %d", len(expected), len(cmd))
	}

	// Check that prompt is intact as final argument
	if cmd[len(cmd)-1] != prompt {
		t.Errorf("Expected final arg to be full prompt '%s', got '%s'", prompt, cmd[len(cmd)-1])
	}
}
