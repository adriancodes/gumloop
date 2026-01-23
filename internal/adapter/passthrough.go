package adapter

import (
	"bufio"
	"io"
)

// PassThroughAdapter forwards lines as AssistantMessage events.
// Used for agents that output plain text: Gemini, OpenCode, Cursor, Ollama.
type PassThroughAdapter struct{}

// NewPassThroughAdapter creates a new pass-through adapter.
func NewPassThroughAdapter() *PassThroughAdapter {
	return &PassThroughAdapter{}
}

// Process reads lines from the reader and emits them as AssistantMessage events.
// This adapter does not parse structured output - it simply forwards all text.
func (a *PassThroughAdapter) Process(reader io.Reader, events chan<- Event) error {
	scanner := bufio.NewScanner(reader)

	// Read line by line
	for scanner.Scan() {
		line := scanner.Text()

		// Emit each line as an assistant message
		events <- AssistantMessage{Text: line}
	}

	return scanner.Err()
}
