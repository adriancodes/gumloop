package adapter

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
)

// CodexAdapter parses Codex's --json output format (NDJSON).
//
// Codex with --json emits newline-delimited JSON events.
// Each line is a complete JSON object representing different event types.
//
// The exact schema is determined through runtime observation, but we handle:
//   - type: event type indicator
//   - content: message content
//   - tool: tool name if applicable
//
// This adapter is designed to be resilient - if JSON parsing fails,
// it falls back to pass-through mode for that line.
type CodexAdapter struct{}

// CodexEvent represents a single line of --json output from Codex.
//
// Note: Codex documentation states it emits "newline-delimited JSON events
// (one per state change)" but doesn't specify the exact schema. This struct
// covers the expected fields based on common patterns in AI CLI tools.
type CodexEvent struct {
	Type    string `json:"type"`    // Event type
	Content string `json:"content"` // Message content
	Tool    string `json:"tool"`    // Tool name if applicable
	Text    string `json:"text"`    // Alternative text field (some events use this)
	Message string `json:"message"` // Alternative message field
	Error   string `json:"error"`   // Error message if applicable
}

// Process reads Codex's --json output and emits normalized events.
//
// It reads the output line-by-line, parses each JSON object, and converts
// Codex-specific events into normalized Event types (ToolUse, AssistantMessage, Error).
//
// If JSON parsing fails for a line, it treats the line as plain text and
// emits it as an AssistantMessage to ensure output is never lost.
func (a *CodexAdapter) Process(reader io.Reader, events chan<- Event) error {
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue // Skip empty lines
		}

		var event CodexEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			// Malformed JSON - treat as plain text and continue
			// This ensures we don't lose output if Codex format changes
			log.Printf("Warning: failed to parse Codex JSON, treating as plain text: %v", err)
			events <- AssistantMessage{Text: line}
			continue
		}

		// Process based on event content
		// Priority: Error > Tool > Content/Text/Message
		if event.Error != "" {
			events <- Error{Message: event.Error}
			continue
		}

		if event.Tool != "" {
			events <- ToolUse{Name: event.Tool}
		}

		// Extract text content from various possible fields
		// Different event types may use different field names
		text := ""
		if event.Content != "" {
			text = event.Content
		} else if event.Text != "" {
			text = event.Text
		} else if event.Message != "" {
			text = event.Message
		}

		if text != "" {
			events <- AssistantMessage{Text: text}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading Codex output: %w", err)
	}

	return nil
}
