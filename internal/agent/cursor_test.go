package agent

import (
	"testing"
)

// setupCursorTests ensures the cursor agent is registered before tests run.
func setupCursorTests() {
	RegisterAgent(&Agent{
		ID:               "cursor",
		Name:             "Cursor Agent",
		Command:          "cursor-agent",
		CheckCommand:     "cursor-agent",
		AutonomousFlags:  []string{"-p", "--force", "--output-format", "text"},
		InteractiveFlags: []string{},
		ModelFlag:        "--model",
		PromptStyle:      PromptStyleArg,
	})
}

func TestCursorAgent_Registration(t *testing.T) {
	setupCursorTests()
	agent, err := GetAgent("cursor")
	if err != nil {
		t.Fatalf("Cursor agent not registered: %v", err)
	}

	if agent.ID != "cursor" {
		t.Errorf("Expected ID 'cursor', got '%s'", agent.ID)
	}

	if agent.Name != "Cursor Agent" {
		t.Errorf("Expected Name 'Cursor Agent', got '%s'", agent.Name)
	}

	if agent.Command != "cursor-agent" {
		t.Errorf("Expected Command 'cursor-agent', got '%s'", agent.Command)
	}

	if agent.CheckCommand != "cursor-agent" {
		t.Errorf("Expected CheckCommand 'cursor-agent', got '%s'", agent.CheckCommand)
	}

	if agent.ModelFlag != "--model" {
		t.Errorf("Expected ModelFlag '--model', got '%s'", agent.ModelFlag)
	}

	if agent.PromptStyle != PromptStyleArg {
		t.Errorf("Expected PromptStyle 'arg', got '%s'", agent.PromptStyle)
	}
}

func TestCursorAgent_AutonomousFlags(t *testing.T) {
	setupCursorTests()
	agent, _ := GetAgent("cursor")

	expectedFlags := []string{
		"-p",
		"--force",
		"--output-format",
		"text",
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

func TestCursorAgent_InteractiveFlags(t *testing.T) {
	setupCursorTests()
	agent, _ := GetAgent("cursor")

	// Interactive flags should be empty for Cursor
	if len(agent.InteractiveFlags) != 0 {
		t.Errorf("Expected 0 interactive flags, got %d: %v", len(agent.InteractiveFlags), agent.InteractiveFlags)
	}
}

func TestCursorAgent_BuildCommand_Autonomous(t *testing.T) {
	setupCursorTests()
	agent, _ := GetAgent("cursor")

	prompt := "Fix the tests"
	model := "gpt-4"
	cmd := agent.BuildCommand(prompt, model, true)

	// Expected: cursor-agent -p --force --output-format text --model gpt-4 "Fix the tests"
	expectedCmd := []string{
		"cursor-agent",
		"-p",
		"--force",
		"--output-format",
		"text",
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

func TestCursorAgent_BuildCommand_Interactive(t *testing.T) {
	setupCursorTests()
	agent, _ := GetAgent("cursor")

	prompt := "Add a feature"
	model := "gpt-4"
	cmd := agent.BuildCommand(prompt, model, false)

	// Expected: cursor-agent --model gpt-4 "Add a feature"
	// Note: No autonomous flags in interactive mode (InteractiveFlags is empty)
	expectedCmd := []string{
		"cursor-agent",
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

	// Verify autonomous flags are NOT present
	for _, arg := range cmd {
		if arg == "-p" || arg == "--force" || arg == "--output-format" {
			t.Errorf("Interactive command should not contain autonomous flags, but found '%s'", arg)
		}
	}
}

func TestCursorAgent_BuildCommand_NoModel(t *testing.T) {
	setupCursorTests()
	agent, _ := GetAgent("cursor")

	prompt := "Test without model"
	cmd := agent.BuildCommand(prompt, "", true)

	// Expected: cursor-agent -p --force --output-format text "Test without model"
	// Model flag should NOT be present
	expectedCmd := []string{
		"cursor-agent",
		"-p",
		"--force",
		"--output-format",
		"text",
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

func TestCursorAgent_BuildCommand_InteractiveNoModel(t *testing.T) {
	setupCursorTests()
	agent, _ := GetAgent("cursor")

	prompt := "Test interactive without model"
	cmd := agent.BuildCommand(prompt, "", false)

	// Expected: cursor-agent "Test interactive without model"
	// No flags at all in interactive mode with no model
	expectedCmd := []string{
		"cursor-agent",
		"Test interactive without model",
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
