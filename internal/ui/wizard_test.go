package ui

import (
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// TestWizardModelInit tests the initial state of the wizard
func TestWizardModelInit(t *testing.T) {
	m := wizardModel{
		step:         stepAgent,
		agentIndex:   0,
		agents:       availableAgents,
		createPrompt: true,
	}

	assert.Equal(t, stepAgent, m.step)
	assert.Equal(t, 0, m.agentIndex)
	assert.True(t, m.createPrompt)
	assert.Equal(t, 6, len(m.agents))
}

// TestWizardAgentNavigation tests navigation in agent selection step
func TestWizardAgentNavigation(t *testing.T) {
	m := wizardModel{
		step:       stepAgent,
		agentIndex: 0,
		agents:     availableAgents,
	}

	// Test down navigation
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = newModel.(wizardModel)
	assert.Equal(t, 1, m.agentIndex)

	// Test up navigation
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = newModel.(wizardModel)
	assert.Equal(t, 0, m.agentIndex)

	// Test down navigation at boundary (should not go below 0)
	m.agentIndex = 0
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = newModel.(wizardModel)
	assert.Equal(t, 0, m.agentIndex)

	// Test down navigation at boundary (should not exceed max)
	m.agentIndex = len(m.agents) - 1
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = newModel.(wizardModel)
	assert.Equal(t, len(m.agents)-1, m.agentIndex)
}

// TestWizardStepProgression tests progression through wizard steps
func TestWizardStepProgression(t *testing.T) {
	// Create properly initialized textinput models
	modelInput := textinput.New()
	modelInput.Placeholder = "leave blank for agent default"
	modelInput.CharLimit = 50
	modelInput.Width = 50

	verifyInput := textinput.New()
	verifyInput.Placeholder = "e.g., npm test, go test ./..."
	verifyInput.CharLimit = 100
	verifyInput.Width = 50

	m := wizardModel{
		step:         stepAgent,
		agentIndex:   0,
		agents:       availableAgents,
		modelInput:   modelInput,
		verifyInput:  verifyInput,
		createPrompt: true,
	}

	// Step 1: Select agent (claude at index 0)
	assert.Equal(t, stepAgent, m.step)
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(wizardModel)
	assert.Equal(t, stepModel, m.step)
	assert.Equal(t, "claude", m.config.CLI)

	// Step 2: Select model - navigate to "(default)" option and select it
	// "(default)" is always second-to-last in the list
	defaultIndex := len(m.models) - 2
	for i := 0; i < defaultIndex; i++ {
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = newModel.(wizardModel)
	}
	// With the list component, check the list's index instead of modelIndex
	assert.Equal(t, defaultIndex, m.modelList.Index())
	selectedItem := m.modelList.SelectedItem().(modelOption)
	assert.Equal(t, "(default)", selectedItem.Name)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(wizardModel)
	assert.Equal(t, stepVerify, m.step)
	assert.Equal(t, "", m.config.Model)

	// Step 3: Enter verify command (leave blank)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(wizardModel)
	assert.Equal(t, stepPrompt, m.step)
	assert.Equal(t, "", m.config.Verify)

	// Step 4: Create PROMPT.md (default yes)
	assert.True(t, m.createPrompt)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(wizardModel)
	assert.Equal(t, stepDone, m.step)
	assert.True(t, m.config.CreatePrompt)
}

// TestWizardPromptToggle tests toggling PROMPT.md creation
func TestWizardPromptToggle(t *testing.T) {
	m := wizardModel{
		step:         stepPrompt,
		createPrompt: true,
	}

	// Test 'n' key
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = newModel.(wizardModel)
	assert.False(t, m.createPrompt)

	// Test 'y' key
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	m = newModel.(wizardModel)
	assert.True(t, m.createPrompt)

	// Test arrow toggle
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = newModel.(wizardModel)
	assert.False(t, m.createPrompt)

	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = newModel.(wizardModel)
	assert.True(t, m.createPrompt)
}

// TestAvailableAgents tests that all expected agents are present
func TestAvailableAgents(t *testing.T) {
	assert.Equal(t, 6, len(availableAgents))

	expectedIDs := []string{"claude", "codex", "gemini", "opencode", "cursor", "ollama"}
	for i, expected := range expectedIDs {
		assert.Equal(t, expected, availableAgents[i].ID)
	}
}

// TestWizardView tests that View renders without panic
func TestWizardView(t *testing.T) {
	tests := []struct {
		name string
		step wizardStep
	}{
		{"Agent step", stepAgent},
		{"Model step", stepModel},
		{"Verify step", stepVerify},
		{"Prompt step", stepPrompt},
		{"Done step", stepDone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := wizardModel{
				step:       tt.step,
				agentIndex: 0,
				agents:     availableAgents,
			}

			// Should not panic
			view := m.View()

			// Done step should render empty
			if tt.step == stepDone {
				assert.Empty(t, view)
			} else {
				assert.NotEmpty(t, view)
				// Should contain the header
				assert.Contains(t, view, "gumloop setup")
			}
		})
	}
}

// TestWizardAgentStepRendering tests the agent step rendering
func TestWizardAgentStepRendering(t *testing.T) {
	m := wizardModel{
		step:       stepAgent,
		agentIndex: 1, // Codex
		agents:     availableAgents,
	}

	view := m.View()

	// Should contain all agent names
	for _, agent := range availableAgents {
		assert.Contains(t, view, agent.Name)
	}

	// Should contain help text
	assert.Contains(t, view, "navigate")
	assert.Contains(t, view, "confirm")
}

// TestWizardConfigPersistence tests that config values are properly stored
func TestWizardConfigPersistence(t *testing.T) {
	// Create properly initialized textinput models
	modelInput := textinput.New()
	verifyInput := textinput.New()

	m := wizardModel{
		step:        stepAgent,
		agents:      availableAgents,
		modelInput:  modelInput,
		verifyInput: verifyInput,
	}

	// Select gemini (index 2)
	m.agentIndex = 2
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(wizardModel)
	assert.Equal(t, "gemini", m.config.CLI)

	// Select model - navigate to "(default)" option (second-to-last)
	defaultIndex := len(m.models) - 2
	for i := 0; i < defaultIndex; i++ {
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = newModel.(wizardModel)
	}
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(wizardModel)

	// Skip verify
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(wizardModel)

	// Set createPrompt to false
	m.createPrompt = false
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(wizardModel)

	// Verify final config
	assert.Equal(t, "gemini", m.config.CLI)
	assert.Equal(t, "", m.config.Model)
	assert.Equal(t, "", m.config.Verify)
	assert.False(t, m.config.CreatePrompt)
}

// TestWizardEscape tests ESC key handling
func TestWizardEscape(t *testing.T) {
	m := wizardModel{
		step: stepAgent,
	}

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.NotNil(t, cmd)
	// Note: We can't easily test that it actually quits without running the full program
	// but we can verify the command is not nil
}

// TestWizardWindowSize tests window resize handling
func TestWizardWindowSize(t *testing.T) {
	m := wizardModel{
		step:   stepAgent,
		width:  0,
		height: 0,
	}

	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	m = newModel.(wizardModel)

	assert.Equal(t, 100, m.width)
	assert.Equal(t, 50, m.height)
}
