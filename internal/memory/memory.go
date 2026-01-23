package memory

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	// DefaultFileName is the default memory file name
	DefaultFileName = ".gumloop-memory.yaml"

	// MaxCommitLog is the maximum number of commits to keep in memory
	MaxCommitLog = 20
)

// SessionMemory represents the persisted state between loop sessions.
type SessionMemory struct {
	StartedAt  time.Time      `yaml:"started"`
	Branch     string         `yaml:"branch"`
	AgentName  string         `yaml:"agent"`
	Iterations int            `yaml:"iterations"`
	Commits    int            `yaml:"commits"`
	ExitReason string         `yaml:"exit_reason"`
	CommitLog  []CommitRecord `yaml:"commit_log"`
	Remaining  string         `yaml:"remaining,omitempty"`
}

// CommitRecord is a single commit entry.
type CommitRecord struct {
	Hash    string `yaml:"hash"`
	Message string `yaml:"message"`
}

// Load reads the memory file from disk and parses it.
// Returns nil (not error) if the file does not exist.
func Load(path string) (*SessionMemory, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read memory file: %w", err)
	}

	var mem SessionMemory
	if err := yaml.Unmarshal(data, &mem); err != nil {
		// Malformed file - return empty memory rather than failing
		return &SessionMemory{}, nil
	}

	return &mem, nil
}

// Save writes the memory to disk as YAML with a header comment.
func (m *SessionMemory) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create memory file: %w", err)
	}
	defer f.Close()

	// Write header comment
	if _, err := f.WriteString("# gumloop session memory (auto-generated, safe to edit \"remaining\" field)\n\n"); err != nil {
		return err
	}

	encoder := yaml.NewEncoder(f)
	encoder.SetIndent(2)
	if err := encoder.Encode(m); err != nil {
		return fmt.Errorf("failed to write memory file: %w", err)
	}

	return encoder.Close()
}

// ToPromptContext renders the memory as compact plain text for prompt injection.
// Returns empty string if there's nothing useful to inject.
func (m *SessionMemory) ToPromptContext() string {
	if m.Iterations == 0 {
		return ""
	}

	var b strings.Builder

	b.WriteString("--- PREVIOUS SESSION ---\n")
	b.WriteString(fmt.Sprintf("Last session: %d iterations, %d commits on branch %s\n",
		m.Iterations, m.Commits, m.Branch))
	b.WriteString(fmt.Sprintf("Agent: %s | Exited: %s\n", m.AgentName, m.ExitReason))

	if len(m.CommitLog) > 0 {
		b.WriteString("\nCommits made:\n")
		for _, c := range m.CommitLog {
			b.WriteString(fmt.Sprintf("- %s %s\n", c.Hash, c.Message))
		}
	}

	if m.Remaining != "" {
		b.WriteString(fmt.Sprintf("\nNote: %s\n", strings.TrimSpace(m.Remaining)))
	}

	b.WriteString("--- END PREVIOUS SESSION ---\n")

	return b.String()
}

// RecordIteration updates the memory with results from the latest iteration.
func (m *SessionMemory) RecordIteration(commitsMade int, newCommits []CommitRecord) {
	m.Iterations++
	m.Commits += commitsMade

	// Prepend new commits (most recent first)
	m.CommitLog = append(newCommits, m.CommitLog...)

	// Cap the commit log
	if len(m.CommitLog) > MaxCommitLog {
		m.CommitLog = m.CommitLog[:MaxCommitLog]
	}
}

// SetExit records why the loop stopped.
func (m *SessionMemory) SetExit(reason string) {
	m.ExitReason = reason
}
