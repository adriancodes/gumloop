package agent

import (
	"testing"
)

// setupClaudeTests ensures the claude agent is registered before tests run.
// This is needed because other tests may clear the registry.
func setupClaudeTests() {
	// Re-register claude agent for testing
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

func TestClaudeAgent_Registration(t *testing.T) {
	setupClaudeTests()
	agent, err := GetAgent("claude")
	if err != nil {
		t.Fatalf("Claude agent not registered: %v", err)
	}

	if agent.ID != "claude" {
		t.Errorf("Expected ID 'claude', got '%s'", agent.ID)
	}

	if agent.Name != "Claude Code" {
		t.Errorf("Expected Name 'Claude Code', got '%s'", agent.Name)
	}

	if agent.Command != "claude" {
		t.Errorf("Expected Command 'claude', got '%s'", agent.Command)
	}

	if agent.CheckCommand != "claude" {
		t.Errorf("Expected CheckCommand 'claude', got '%s'", agent.CheckCommand)
	}

	if agent.ModelFlag != "--model" {
		t.Errorf("Expected ModelFlag '--model', got '%s'", agent.ModelFlag)
	}

	if agent.PromptStyle != PromptStyleStream {
		t.Errorf("Expected PromptStyle 'stream', got '%s'", agent.PromptStyle)
	}
}

func TestClaudeAgent_AutonomousFlags(t *testing.T) {
	setupClaudeTests()
	agent, _ := GetAgent("claude")

	expectedFlags := []string{
		"-p",
		"--dangerously-skip-permissions",
		"--verbose",
		"--output-format",
		"stream-json",
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

func TestClaudeAgent_InteractiveFlags(t *testing.T) {
	setupClaudeTests()
	agent, _ := GetAgent("claude")

	expectedFlags := []string{
		"-p",
		"--verbose",
		"--output-format",
		"stream-json",
	}

	if len(agent.InteractiveFlags) != len(expectedFlags) {
		t.Fatalf("Expected %d interactive flags, got %d", len(expectedFlags), len(agent.InteractiveFlags))
	}

	for i, expected := range expectedFlags {
		if agent.InteractiveFlags[i] != expected {
			t.Errorf("InteractiveFlags[%d]: expected '%s', got '%s'", i, expected, agent.InteractiveFlags[i])
		}
	}

	// Verify --dangerously-skip-permissions is NOT in interactive mode
	for _, flag := range agent.InteractiveFlags {
		if flag == "--dangerously-skip-permissions" {
			t.Error("InteractiveFlags should not contain '--dangerously-skip-permissions'")
		}
	}
}

func TestClaudeAgent_BuildCommand_Autonomous(t *testing.T) {
	setupClaudeTests()
	agent, _ := GetAgent("claude")

	prompt := "Fix the tests"
	model := "sonnet"
	cmd := agent.BuildCommand(prompt, model, true)

	// Expected: claude -p --dangerously-skip-permissions --verbose --output-format stream-json --model sonnet "Fix the tests"
	expectedCmd := []string{
		"claude",
		"-p",
		"--dangerously-skip-permissions",
		"--verbose",
		"--output-format",
		"stream-json",
		"--model",
		"sonnet",
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

func TestClaudeAgent_BuildCommand_Interactive(t *testing.T) {
	setupClaudeTests()
	agent, _ := GetAgent("claude")

	prompt := "Add a feature"
	model := "opus"
	cmd := agent.BuildCommand(prompt, model, false)

	// Expected: claude -p --verbose --output-format stream-json --model opus "Add a feature"
	// Note: --dangerously-skip-permissions should NOT be present
	expectedCmd := []string{
		"claude",
		"-p",
		"--verbose",
		"--output-format",
		"stream-json",
		"--model",
		"opus",
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

	// Verify --dangerously-skip-permissions is NOT present
	for _, arg := range cmd {
		if arg == "--dangerously-skip-permissions" {
			t.Error("Interactive command should not contain '--dangerously-skip-permissions'")
		}
	}
}

func TestClaudeAgent_BuildCommand_NoModel(t *testing.T) {
	setupClaudeTests()
	agent, _ := GetAgent("claude")

	prompt := "Test without model"
	cmd := agent.BuildCommand(prompt, "", true)

	// Expected: claude -p --dangerously-skip-permissions --verbose --output-format stream-json "Test without model"
	// Model flag should NOT be present
	expectedCmd := []string{
		"claude",
		"-p",
		"--dangerously-skip-permissions",
		"--verbose",
		"--output-format",
		"stream-json",
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
