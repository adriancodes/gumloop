package agent

import (
	"testing"
)

// setupGeminiTests ensures the gemini agent is registered before tests run.
func setupGeminiTests() {
	// Re-register gemini agent for testing
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

func TestGeminiAgent_Registration(t *testing.T) {
	setupGeminiTests()
	agent, err := GetAgent("gemini")
	if err != nil {
		t.Fatalf("Gemini agent not registered: %v", err)
	}

	if agent.ID != "gemini" {
		t.Errorf("Expected ID 'gemini', got '%s'", agent.ID)
	}

	if agent.Name != "Google Gemini" {
		t.Errorf("Expected Name 'Google Gemini', got '%s'", agent.Name)
	}

	if agent.Command != "gemini" {
		t.Errorf("Expected Command 'gemini', got '%s'", agent.Command)
	}

	if agent.CheckCommand != "gemini" {
		t.Errorf("Expected CheckCommand 'gemini', got '%s'", agent.CheckCommand)
	}

	if agent.ModelFlag != "--model" {
		t.Errorf("Expected ModelFlag '--model', got '%s'", agent.ModelFlag)
	}

	if agent.PromptStyle != PromptStyleArg {
		t.Errorf("Expected PromptStyle 'arg', got '%s'", agent.PromptStyle)
	}
}

func TestGeminiAgent_AutonomousFlags(t *testing.T) {
	setupGeminiTests()
	agent, _ := GetAgent("gemini")

	expectedFlags := []string{
		"-p",
		"--yolo",
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

func TestGeminiAgent_InteractiveFlags(t *testing.T) {
	setupGeminiTests()
	agent, _ := GetAgent("gemini")

	expectedFlags := []string{
		"-p",
		"--output-format",
		"text",
	}

	if len(agent.InteractiveFlags) != len(expectedFlags) {
		t.Fatalf("Expected %d interactive flags, got %d", len(expectedFlags), len(agent.InteractiveFlags))
	}

	for i, expected := range expectedFlags {
		if agent.InteractiveFlags[i] != expected {
			t.Errorf("InteractiveFlags[%d]: expected '%s', got '%s'", i, expected, agent.InteractiveFlags[i])
		}
	}

	// Verify --yolo is NOT in interactive mode
	for _, flag := range agent.InteractiveFlags {
		if flag == "--yolo" {
			t.Error("InteractiveFlags should not contain '--yolo'")
		}
	}
}

func TestGeminiAgent_BuildCommand_Autonomous(t *testing.T) {
	setupGeminiTests()
	agent, _ := GetAgent("gemini")

	prompt := "Fix the tests"
	model := "gemini-pro"
	cmd := agent.BuildCommand(prompt, model, true)

	// Expected: gemini -p --yolo --output-format text --model gemini-pro "Fix the tests"
	expectedCmd := []string{
		"gemini",
		"-p",
		"--yolo",
		"--output-format",
		"text",
		"--model",
		"gemini-pro",
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

func TestGeminiAgent_BuildCommand_Interactive(t *testing.T) {
	setupGeminiTests()
	agent, _ := GetAgent("gemini")

	prompt := "Add a feature"
	model := "gemini-pro"
	cmd := agent.BuildCommand(prompt, model, false)

	// Expected: gemini -p --output-format text --model gemini-pro "Add a feature"
	// Note: --yolo should NOT be present
	expectedCmd := []string{
		"gemini",
		"-p",
		"--output-format",
		"text",
		"--model",
		"gemini-pro",
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

	// Verify --yolo is NOT present
	for _, arg := range cmd {
		if arg == "--yolo" {
			t.Error("Interactive command should not contain '--yolo'")
		}
	}
}

func TestGeminiAgent_BuildCommand_NoModel(t *testing.T) {
	setupGeminiTests()
	agent, _ := GetAgent("gemini")

	prompt := "Test without model"
	cmd := agent.BuildCommand(prompt, "", true)

	// Expected: gemini -p --yolo --output-format text "Test without model"
	// Model flag should NOT be present
	expectedCmd := []string{
		"gemini",
		"-p",
		"--yolo",
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

func TestGeminiAgent_CommandStructure(t *testing.T) {
	setupGeminiTests()
	agent, _ := GetAgent("gemini")

	// Test that the command properly starts with "gemini"
	cmd := agent.BuildCommand("test", "", true)

	if cmd[0] != "gemini" {
		t.Errorf("Expected first command element to be 'gemini', got '%s'", cmd[0])
	}

	// Gemini command should have flags immediately after the command
	if len(cmd) < 2 {
		t.Fatal("Expected command to have at least 2 elements")
	}

	if cmd[1] != "-p" {
		t.Errorf("Expected second command element to be '-p', got '%s'", cmd[1])
	}
}
