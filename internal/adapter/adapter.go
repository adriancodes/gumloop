package adapter

import "io"

// Adapter processes agent output and emits normalized events.
//
// Different agents produce output in different formats:
//   - Claude: NDJSON with stream-json format
//   - Codex: NDJSON with --json format
//   - Gemini, OpenCode, Cursor, Ollama: Plain text
//
// Adapters normalize these formats into a unified event stream
// (ToolUse, AssistantMessage, Error) that the UI can display consistently.
type Adapter interface {
	// Process reads from the agent output stream and sends normalized events
	// to the events channel. It returns when the stream ends or an error occurs.
	//
	// Implementations should:
	//   - Read the stream line-by-line or in chunks
	//   - Parse agent-specific format
	//   - Send events to the channel (non-blocking if possible)
	//   - Handle malformed output gracefully (log warnings, continue)
	//   - Return nil on clean stream end, error on read failures
	//
	// The events channel is managed by the caller - adapters should not close it.
	Process(reader io.Reader, events chan<- Event) error
}
