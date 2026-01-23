package adapter

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
)

// ClaudeAdapter parses Claude's stream-json output format (NDJSON).
//
// Claude with --output-format stream-json emits newline-delimited JSON events.
// Each line is a complete JSON object representing a different event type.
//
// Example stream-json event types:
//   - type: "assistant" → Contains message.content with assistant text
//   - type: "tool_use" → Contains tool name and parameters
//   - type: "result" → Duplicate of assistant message, skip
//   - type: "stream_event" → Real-time text deltas for display
type ClaudeAdapter struct{}

// ClaudeStreamEvent represents a single line of stream-json output from Claude.
type ClaudeStreamEvent struct {
	Type    string          `json:"type"`    // Event type: "assistant", "tool_use", "result", "stream_event"
	Name    string          `json:"name"`    // Tool name (for tool_use events)
	Message ClaudeMessage   `json:"message"` // Message content (for assistant events)
	Event   ClaudeEventData `json:"event"`   // Stream event data (for stream_event type)
}

// ClaudeMessage contains the content of an assistant message.
type ClaudeMessage struct {
	Content []ClaudeContent `json:"content"`
}

// ClaudeContent represents a content block in a Claude message.
type ClaudeContent struct {
	Type string `json:"type"` // "text" or "tool_use"
	Text string `json:"text"` // Text content (for text type)
}

// ClaudeEventData contains stream event data for real-time updates.
type ClaudeEventData struct {
	Delta ClaudeDelta `json:"delta"`
}

// ClaudeDelta contains incremental text updates.
type ClaudeDelta struct {
	Type string `json:"type"` // "text_delta"
	Text string `json:"text"` // Incremental text
}

// Process reads Claude's stream-json output and emits normalized events.
//
// It reads the output line-by-line, parses each JSON object, and converts
// Claude-specific events into normalized Event types (ToolUse, AssistantMessage, Error).
//
// Malformed JSON lines are logged as warnings and skipped.
func (a *ClaudeAdapter) Process(reader io.Reader, events chan<- Event) error {
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue // Skip empty lines
		}

		var event ClaudeStreamEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			// Malformed JSON - log warning and continue
			log.Printf("Warning: failed to parse Claude JSON: %v (line: %s)", err, line)
			continue
		}

		// Process based on event type
		switch event.Type {
		case "assistant":
			// Extract text from message content
			for _, content := range event.Message.Content {
				if content.Type == "text" && content.Text != "" {
					events <- AssistantMessage{Text: content.Text}
				}
			}

		case "tool_use":
			// Emit tool use event
			if event.Name != "" {
				events <- ToolUse{Name: event.Name}
			}

		case "result":
			// Skip - this duplicates the assistant message

		case "stream_event":
			// Real-time text delta for display
			if event.Event.Delta.Type == "text_delta" && event.Event.Delta.Text != "" {
				events <- AssistantMessage{Text: event.Event.Delta.Text}
			}

		default:
			// Unknown event type - log but don't error
			log.Printf("Warning: unknown Claude event type: %s", event.Type)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading Claude output: %w", err)
	}

	return nil
}
