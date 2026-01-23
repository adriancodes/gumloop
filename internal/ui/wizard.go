package ui

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ErrWizardCancelled is returned when the user cancels the wizard with Escape or Ctrl+C
var ErrWizardCancelled = errors.New("wizard cancelled by user")

// WizardConfig holds the configuration values collected by the wizard
type WizardConfig struct {
	CLI           string
	Model         string
	Verify        string
	CreatePrompt  bool
}

// wizardStep represents the current step in the wizard
type wizardStep int

const (
	stepAgent wizardStep = iota
	stepModel
	stepVerify
	stepPrompt
	stepDone
)

// wizardModel is the BubbleTea model for the init wizard
type wizardModel struct {
	step            wizardStep
	agentIndex      int
	agents          []agentOption
	modelIndex      int
	models          []modelOption
	modelList       list.Model // Fuzzy searchable list for model selection
	customModelMode bool       // true when user selected "Custom..." and is typing
	modelInput      textinput.Model
	verifyInput     textinput.Model
	createPrompt    bool
	config          WizardConfig
	cancelled       bool // true if user cancelled with Escape/Ctrl+C
	width           int
	height          int
}

// agentOption represents an agent choice in the wizard
type agentOption struct {
	ID          string
	Name        string
	Description string
}

// modelOption represents a model choice in the wizard
type modelOption struct {
	ID       string
	Name     string
	Desc     string // Description of the model
	IsCustom bool   // If true, this option enables free text input
}

// Implement list.Item interface for modelOption
func (m modelOption) FilterValue() string { return m.Name }
func (m modelOption) Title() string       { return m.Name }
func (m modelOption) Description() string { return m.Desc }

// modelItemDelegate handles rendering of model items in the list
type modelItemDelegate struct{}

func (d modelItemDelegate) Height() int                             { return 1 }
func (d modelItemDelegate) Spacing() int                            { return 0 }
func (d modelItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d modelItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(modelOption)
	if !ok {
		return
	}

	cursor := "  "
	if index == m.Index() {
		cursor = "> "
	}

	nameStyle := lipgloss.NewStyle()
	if index == m.Index() {
		nameStyle = nameStyle.Foreground(lipgloss.Color("39")) // Blue
	}

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")) // Gray

	fmt.Fprintf(w, "%s%s %s",
		cursor,
		nameStyle.Render(item.Name),
		descStyle.Render("- "+item.Desc),
	)
}

var availableAgents = []agentOption{
	{ID: "claude", Name: "Claude Code", Description: "Anthropic Claude Code (AI-powered coding assistant)"},
	{ID: "codex", Name: "OpenAI Codex", Description: "OpenAI Codex (code generation model)"},
	{ID: "gemini", Name: "Google Gemini", Description: "Google Gemini (multimodal AI)"},
	{ID: "opencode", Name: "OpenCode", Description: "OpenCode (open source coding assistant)"},
	{ID: "cursor", Name: "Cursor Agent", Description: "Cursor Agent (AI pair programmer)"},
	{ID: "ollama", Name: "Ollama", Description: "Ollama (local LLM runtime)"},
}

// modelsForAgent returns the available models for a given agent.
// It fetches from the models.dev API, falling back to hardcoded models if fetch fails.
func modelsForAgent(agentID string) []modelOption {
	// Try to fetch from API first
	models := fetchModelsFromAPI(agentID)

	// Fall back to hardcoded models if fetch failed or returned empty
	if len(models) == 0 {
		models = fallbackModels(agentID)
	}

	// Always add (default) and Custom... options at the end
	models = append(models,
		modelOption{ID: "", Name: "(default)", Desc: "Use agent's default model"},
		modelOption{ID: "", Name: "Custom...", Desc: "Enter a custom model name", IsCustom: true},
	)

	return models
}

// RunWizard launches the interactive setup wizard
// Returns the collected configuration values or an error
func RunWizard() (*WizardConfig, error) {
	// Create model input
	modelInput := textinput.New()
	modelInput.Placeholder = "leave blank for agent default"
	modelInput.CharLimit = 50
	modelInput.Width = 50

	// Create verify input
	verifyInput := textinput.New()
	verifyInput.Placeholder = "e.g., npm test, go test ./..."
	verifyInput.CharLimit = 100
	verifyInput.Width = 50

	m := wizardModel{
		step:        stepAgent,
		agentIndex:  0,
		agents:      availableAgents,
		modelInput:  modelInput,
		verifyInput: verifyInput,
		createPrompt: true, // Default to yes
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("wizard failed: %w", err)
	}

	// Extract config from final model
	result := finalModel.(wizardModel)

	// Check if user cancelled
	if result.cancelled {
		return nil, ErrWizardCancelled
	}

	return &result.config, nil
}

// Init initializes the wizard model
func (m wizardModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the wizard state
func (m wizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			// In custom model mode, escape goes back to list
			if m.step == stepModel && m.customModelMode {
				m.customModelMode = false
				m.modelInput.Blur()
				m.modelInput.SetValue("")
				return m, nil
			}
			// Mark as cancelled and quit
			m.cancelled = true
			return m, tea.Quit

		case "enter":
			return m.handleEnter()

		case "up", "k":
			if m.step == stepAgent && m.agentIndex > 0 {
				m.agentIndex--
			} else if m.step == stepPrompt {
				m.createPrompt = !m.createPrompt
			}
			// Note: model step navigation is handled by the list component

		case "down", "j":
			if m.step == stepAgent && m.agentIndex < len(m.agents)-1 {
				m.agentIndex++
			} else if m.step == stepPrompt {
				m.createPrompt = !m.createPrompt
			}
			// Note: model step navigation is handled by the list component

		case "y", "Y":
			if m.step == stepPrompt {
				m.createPrompt = true
			}

		case "n", "N":
			if m.step == stepPrompt {
				m.createPrompt = false
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	// Update text inputs and lists if they're active
	var cmd tea.Cmd
	switch m.step {
	case stepModel:
		if m.customModelMode {
			m.modelInput, cmd = m.modelInput.Update(msg)
		} else {
			// Pass messages to the list for filtering/navigation
			m.modelList, cmd = m.modelList.Update(msg)
		}
	case stepVerify:
		m.verifyInput, cmd = m.verifyInput.Update(msg)
	}

	return m, cmd
}

// handleEnter processes the Enter key based on current step
func (m wizardModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.step {
	case stepAgent:
		// Store selected agent and load models for it
		m.config.CLI = m.agents[m.agentIndex].ID
		m.models = modelsForAgent(m.config.CLI)
		m.modelIndex = 0
		m.customModelMode = false

		// Initialize the fuzzy searchable list
		items := make([]list.Item, len(m.models))
		for i, model := range m.models {
			items[i] = model
		}

		// Create list with custom delegate
		delegate := modelItemDelegate{}
		m.modelList = list.New(items, delegate, 60, 10)
		m.modelList.Title = ""
		m.modelList.SetShowTitle(false)
		m.modelList.SetShowStatusBar(false)
		m.modelList.SetShowHelp(false)
		m.modelList.SetFilteringEnabled(true)
		m.modelList.Styles.Title = lipgloss.NewStyle()
		m.modelList.Styles.PaginationStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		m.modelList.Styles.HelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		m.modelList.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
		m.modelList.FilterInput.TextStyle = lipgloss.NewStyle()

		m.step = stepModel
		return m, nil

	case stepModel:
		if m.customModelMode {
			// Store custom model from text input
			m.config.Model = strings.TrimSpace(m.modelInput.Value())
			m.modelInput.Blur()
			m.step = stepVerify
			return m, m.verifyInput.Focus()
		}

		// Get selected item from the filtered list
		selectedItem := m.modelList.SelectedItem()
		if selectedItem == nil {
			return m, nil
		}
		selectedModel := selectedItem.(modelOption)

		// Check if selected option is "Custom..."
		if selectedModel.IsCustom {
			// Enter custom model input mode
			m.customModelMode = true
			m.modelInput.Placeholder = "enter model name"
			return m, m.modelInput.Focus()
		}

		// Store selected model ID (may be empty for default)
		m.config.Model = selectedModel.ID
		m.step = stepVerify
		return m, m.verifyInput.Focus()

	case stepVerify:
		// Store verify command (can be empty)
		m.config.Verify = strings.TrimSpace(m.verifyInput.Value())
		m.step = stepPrompt
		m.verifyInput.Blur()
		return m, nil

	case stepPrompt:
		// Store createPrompt choice
		m.config.CreatePrompt = m.createPrompt
		m.step = stepDone
		return m, tea.Quit
	}

	return m, nil
}

// View renders the wizard UI
func (m wizardModel) View() string {
	if m.step == stepDone {
		return ""
	}

	var s strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")). // Blue
		MarginBottom(1)
	s.WriteString(headerStyle.Render("🚂 gumloop setup"))
	s.WriteString("\n\n")

	// Render current step
	switch m.step {
	case stepAgent:
		s.WriteString(m.renderAgentStep())
	case stepModel:
		s.WriteString(m.renderModelStep())
	case stepVerify:
		s.WriteString(m.renderVerifyStep())
	case stepPrompt:
		s.WriteString(m.renderPromptStep())
	}

	// Footer
	s.WriteString("\n\n")
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")). // Gray
		Italic(true)
	s.WriteString(helpStyle.Render("↑/↓: navigate • enter: confirm • esc: cancel"))

	return s.String()
}

// renderAgentStep renders the agent selection step
func (m wizardModel) renderAgentStep() string {
	var s strings.Builder
	questionStyle := lipgloss.NewStyle().Bold(true)
	s.WriteString(questionStyle.Render("? Which AI agent?"))
	s.WriteString("\n\n")

	for i, agent := range m.agents {
		cursor := "  "
		if i == m.agentIndex {
			cursor = "> "
		}

		nameStyle := lipgloss.NewStyle()
		if i == m.agentIndex {
			nameStyle = nameStyle.Foreground(lipgloss.Color("39")) // Blue
		}

		s.WriteString(fmt.Sprintf("%s%s (%s)\n",
			cursor,
			nameStyle.Render(agent.Name),
			agent.Description,
		))
	}

	return s.String()
}

// renderModelStep renders the model selection step
func (m wizardModel) renderModelStep() string {
	var s strings.Builder
	questionStyle := lipgloss.NewStyle().Bold(true)
	s.WriteString(questionStyle.Render("? Which model?"))
	s.WriteString(" ")

	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")). // Gray
		Italic(true)

	// If in custom mode, show text input
	if m.customModelMode {
		s.WriteString("\n\n")
		s.WriteString(hintStyle.Render("Enter custom model name (esc to go back):"))
		s.WriteString("\n\n")
		s.WriteString(m.modelInput.View())
		return s.String()
	}

	// Show hint for filtering
	s.WriteString(hintStyle.Render("(type to filter)"))
	s.WriteString("\n\n")

	// Show the fuzzy-searchable list
	s.WriteString(m.modelList.View())

	return s.String()
}

// renderVerifyStep renders the verify command input step
func (m wizardModel) renderVerifyStep() string {
	var s strings.Builder
	questionStyle := lipgloss.NewStyle().Bold(true)
	s.WriteString(questionStyle.Render("? Run verification after each iteration?"))
	s.WriteString(" ")

	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")). // Gray
		Italic(true)
	s.WriteString(hintStyle.Render("(e.g., npm test)"))
	s.WriteString("\n\n")

	s.WriteString(m.verifyInput.View())

	return s.String()
}

// renderPromptStep renders the PROMPT.md creation confirmation step
func (m wizardModel) renderPromptStep() string {
	var s strings.Builder
	questionStyle := lipgloss.NewStyle().Bold(true)
	s.WriteString(questionStyle.Render("? Create PROMPT.md template?"))
	s.WriteString("\n\n")

	yesStyle := lipgloss.NewStyle()
	noStyle := lipgloss.NewStyle()

	if m.createPrompt {
		yesStyle = yesStyle.Foreground(lipgloss.Color("39")) // Blue
		s.WriteString(fmt.Sprintf("> %s\n", yesStyle.Render("Yes")))
		s.WriteString(fmt.Sprintf("  %s\n", noStyle.Render("No")))
	} else {
		noStyle = noStyle.Foreground(lipgloss.Color("39")) // Blue
		s.WriteString(fmt.Sprintf("  %s\n", yesStyle.Render("Yes")))
		s.WriteString(fmt.Sprintf("> %s\n", noStyle.Render("No")))
	}

	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")). // Gray
		Italic(true).
		MarginTop(1)
	s.WriteString("\n")
	s.WriteString(hintStyle.Render("(y/n or ↑/↓ to toggle)"))

	return s.String()
}
