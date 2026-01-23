package agent

import (
	"fmt"
	"strings"
)

// PromptStyle defines how prompts are passed to agents.
type PromptStyle string

const (
	// PromptStyleArg passes prompt as command argument: <cmd> <flags> "<prompt>"
	PromptStyleArg PromptStyle = "arg"

	// PromptStylePipe pipes prompt to stdin: echo "<prompt>" | <cmd> <flags>
	PromptStylePipe PromptStyle = "pipe"

	// PromptStyleStream passes prompt as arg and parses JSON stream output
	PromptStyleStream PromptStyle = "stream"

	// PromptStyleOllama passes model and prompt as positional args: ollama run <model> "<prompt>"
	PromptStyleOllama PromptStyle = "ollama"
)

// Agent represents an AI coding agent with its configuration.
type Agent struct {
	// ID is the unique identifier (e.g., "claude", "codex")
	ID string

	// Name is the human-readable name (e.g., "Claude Code", "OpenAI Codex")
	Name string

	// Command is the CLI command to execute (e.g., "claude", "codex exec")
	Command string

	// CheckCommand is the command to verify agent is installed (e.g., "claude", "codex")
	CheckCommand string

	// AutonomousFlags are flags used in --choo-choo mode for autonomous operation
	AutonomousFlags []string

	// InteractiveFlags are flags used in single-run mode
	InteractiveFlags []string

	// ModelFlag is how to pass model (e.g., "--model", "-m", "" for none/positional)
	ModelFlag string

	// PromptStyle defines how to pass the prompt to the agent
	PromptStyle PromptStyle
}

// Registry stores all registered agents.
var Registry = make(map[string]*Agent)

// RegisterAgent adds an agent to the registry.
func RegisterAgent(agent *Agent) {
	Registry[agent.ID] = agent
}

// GetAgent retrieves an agent by ID.
// Returns an error with helpful suggestions if the agent is not found.
func GetAgent(id string) (*Agent, error) {
	agent, ok := Registry[id]
	if !ok {
		// Generate list of available agents
		available := ListAgents()

		// Try to find a close match for helpful error message
		suggestion := findClosestMatch(id, available)
		if suggestion != "" {
			return nil, fmt.Errorf("unknown agent '%s' (available: %v). Did you mean '%s'?",
				id, available, suggestion)
		}
		return nil, fmt.Errorf("unknown agent '%s' (available: %v)", id, available)
	}
	return agent, nil
}

// ListAgents returns a sorted list of all registered agent IDs.
func ListAgents() []string {
	agents := make([]string, 0, len(Registry))
	for id := range Registry {
		agents = append(agents, id)
	}
	// Sort for consistent output
	// Manual sort since we want to avoid importing sort package
	for i := 0; i < len(agents); i++ {
		for j := i + 1; j < len(agents); j++ {
			if agents[i] > agents[j] {
				agents[i], agents[j] = agents[j], agents[i]
			}
		}
	}
	return agents
}

// findClosestMatch finds the closest match for a typo using simple edit distance.
// Returns empty string if no close match found.
func findClosestMatch(typo string, options []string) string {
	if len(typo) == 0 {
		return ""
	}

	// Simple heuristic: if typo is within 2 characters different, suggest it
	for _, option := range options {
		if levenshteinDistance(typo, option) <= 2 {
			return option
		}
	}
	return ""
}

// levenshteinDistance calculates the edit distance between two strings.
func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	// Create a matrix
	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	// Fill the matrix
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(a)][len(b)]
}

// min returns the minimum of three integers.
func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// BuildCommand constructs the command array to execute the agent.
// Parameters:
//   - prompt: The prompt text to send to the agent
//   - model: The model to use (empty string uses agent default)
//   - autonomous: Whether to use autonomous mode (--choo-choo)
//
// Returns a command array suitable for exec.Command(args[0], args[1:]...)
func (a *Agent) BuildCommand(prompt string, model string, autonomous bool) []string {
	// Start with the base command
	cmdParts := strings.Fields(a.Command)
	args := make([]string, 0, len(cmdParts)+10)
	args = append(args, cmdParts...)

	// Add mode-specific flags
	var flags []string
	if autonomous {
		flags = a.AutonomousFlags
	} else {
		flags = a.InteractiveFlags
	}
	args = append(args, flags...)

	// Add model flag if specified and agent supports it
	if model != "" && a.ModelFlag != "" {
		args = append(args, a.ModelFlag, model)
	}

	// Handle prompt based on style
	switch a.PromptStyle {
	case PromptStyleOllama:
		// Ollama: model is positional, then prompt
		// Format: ollama run <model> "<prompt>"
		if model != "" {
			args = append(args, model)
		}
		args = append(args, prompt)

	case PromptStyleArg, PromptStyleStream:
		// Argument style: prompt is passed as final argument
		args = append(args, prompt)

	case PromptStylePipe:
		// Pipe style: prompt will be piped to stdin
		// Don't add prompt to args
	}

	return args
}
