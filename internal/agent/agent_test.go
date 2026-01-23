package agent

import (
	"testing"
)

func TestRegisterAgent(t *testing.T) {
	// Clear registry for clean test
	Registry = make(map[string]*Agent)

	agent := &Agent{
		ID:   "test",
		Name: "Test Agent",
	}

	RegisterAgent(agent)

	if len(Registry) != 1 {
		t.Errorf("expected 1 agent in registry, got %d", len(Registry))
	}

	retrieved, ok := Registry["test"]
	if !ok {
		t.Error("agent not found in registry")
	}

	if retrieved.ID != "test" {
		t.Errorf("expected ID 'test', got '%s'", retrieved.ID)
	}
}

func TestGetAgent(t *testing.T) {
	// Clear registry and add test agents
	Registry = make(map[string]*Agent)

	claude := &Agent{ID: "claude", Name: "Claude Code"}
	codex := &Agent{ID: "codex", Name: "OpenAI Codex"}

	RegisterAgent(claude)
	RegisterAgent(codex)

	tests := []struct {
		name        string
		agentID     string
		shouldExist bool
		expectedErr string
	}{
		{
			name:        "valid agent - claude",
			agentID:     "claude",
			shouldExist: true,
		},
		{
			name:        "valid agent - codex",
			agentID:     "codex",
			shouldExist: true,
		},
		{
			name:        "invalid agent - no suggestion",
			agentID:     "invalid",
			shouldExist: false,
			expectedErr: "unknown agent 'invalid'",
		},
		{
			name:        "invalid agent - with suggestion",
			agentID:     "clude",
			shouldExist: false,
			expectedErr: "Did you mean 'claude'?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := GetAgent(tt.agentID)

			if tt.shouldExist {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
				if agent == nil {
					t.Error("expected agent, got nil")
				}
				if agent != nil && agent.ID != tt.agentID {
					t.Errorf("expected ID '%s', got '%s'", tt.agentID, agent.ID)
				}
			} else {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if agent != nil {
					t.Error("expected nil agent, got non-nil")
				}
				if err != nil && tt.expectedErr != "" {
					errMsg := err.Error()
					if !contains(errMsg, tt.expectedErr) {
						t.Errorf("expected error containing '%s', got '%s'", tt.expectedErr, errMsg)
					}
				}
			}
		})
	}
}

func TestListAgents(t *testing.T) {
	// Clear registry and add test agents
	Registry = make(map[string]*Agent)

	agents := []*Agent{
		{ID: "claude", Name: "Claude Code"},
		{ID: "codex", Name: "OpenAI Codex"},
		{ID: "gemini", Name: "Google Gemini"},
	}

	for _, agent := range agents {
		RegisterAgent(agent)
	}

	list := ListAgents()

	if len(list) != 3 {
		t.Errorf("expected 3 agents, got %d", len(list))
	}

	// Should be sorted
	expected := []string{"claude", "codex", "gemini"}
	for i, id := range expected {
		if list[i] != id {
			t.Errorf("expected list[%d] = '%s', got '%s'", i, id, list[i])
		}
	}
}

func TestBuildCommand(t *testing.T) {
	tests := []struct {
		name       string
		agent      *Agent
		prompt     string
		model      string
		autonomous bool
		expected   []string
	}{
		{
			name: "claude autonomous with model",
			agent: &Agent{
				ID:               "claude",
				Command:          "claude",
				AutonomousFlags:  []string{"-p", "--dangerously-skip-permissions"},
				InteractiveFlags: []string{"-p"},
				ModelFlag:        "--model",
				PromptStyle:      PromptStyleStream,
			},
			prompt:     "Fix the tests",
			model:      "sonnet",
			autonomous: true,
			expected:   []string{"claude", "-p", "--dangerously-skip-permissions", "--model", "sonnet", "Fix the tests"},
		},
		{
			name: "claude interactive without model",
			agent: &Agent{
				ID:               "claude",
				Command:          "claude",
				AutonomousFlags:  []string{"-p", "--dangerously-skip-permissions"},
				InteractiveFlags: []string{"-p"},
				ModelFlag:        "--model",
				PromptStyle:      PromptStyleStream,
			},
			prompt:     "Fix the tests",
			model:      "",
			autonomous: false,
			expected:   []string{"claude", "-p", "Fix the tests"},
		},
		{
			name: "ollama with model (positional)",
			agent: &Agent{
				ID:               "ollama",
				Command:          "ollama run",
				AutonomousFlags:  []string{},
				InteractiveFlags: []string{},
				ModelFlag:        "",
				PromptStyle:      PromptStyleOllama,
			},
			prompt:     "Write tests",
			model:      "llama2",
			autonomous: true,
			expected:   []string{"ollama", "run", "llama2", "Write tests"},
		},
		{
			name: "codex with multi-word command",
			agent: &Agent{
				ID:               "codex",
				Command:          "codex exec",
				AutonomousFlags:  []string{"--full-auto", "--json"},
				InteractiveFlags: []string{"--json"},
				ModelFlag:        "--model",
				PromptStyle:      PromptStyleArg,
			},
			prompt:     "Fix bugs",
			model:      "gpt-4",
			autonomous: true,
			expected:   []string{"codex", "exec", "--full-auto", "--json", "--model", "gpt-4", "Fix bugs"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.agent.BuildCommand(tt.prompt, tt.model, tt.autonomous)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d args, got %d\nExpected: %v\nGot: %v",
					len(tt.expected), len(result), tt.expected, result)
				return
			}

			for i, arg := range tt.expected {
				if result[i] != arg {
					t.Errorf("arg[%d]: expected '%s', got '%s'", i, arg, result[i])
				}
			}
		})
	}
}

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		a        string
		b        string
		expected int
	}{
		{"", "", 0},
		{"", "abc", 3},
		{"abc", "", 3},
		{"abc", "abc", 0},
		{"abc", "abd", 1},
		{"claude", "clude", 1},
		{"claude", "claudee", 1},
		{"abc", "xyz", 3},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			result := levenshteinDistance(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("levenshteinDistance('%s', '%s') = %d, expected %d",
					tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && hasSubstring(s, substr))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
