package adapter

import (
	"strings"
	"testing"
)

func TestCodexAdapter_Process_ContentField(t *testing.T) {
	adapter := &CodexAdapter{}
	input := `{"type":"message","content":"I'll help you with that."}`

	events := make(chan Event, 10)
	done := make(chan error)

	go func() {
		done <- adapter.Process(strings.NewReader(input), events)
	}()

	// Should receive one AssistantMessage event
	event := <-events
	msg, ok := event.(AssistantMessage)
	if !ok {
		t.Fatalf("expected AssistantMessage, got %T", event)
	}
	if msg.Text != "I'll help you with that." {
		t.Errorf("expected text 'I'll help you with that.', got %q", msg.Text)
	}

	// Wait for processing to complete
	if err := <-done; err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCodexAdapter_Process_TextField(t *testing.T) {
	adapter := &CodexAdapter{}
	input := `{"type":"output","text":"Output from Codex"}`

	events := make(chan Event, 10)
	done := make(chan error)

	go func() {
		done <- adapter.Process(strings.NewReader(input), events)
	}()

	// Should receive one AssistantMessage event using text field
	event := <-events
	msg, ok := event.(AssistantMessage)
	if !ok {
		t.Fatalf("expected AssistantMessage, got %T", event)
	}
	if msg.Text != "Output from Codex" {
		t.Errorf("expected text 'Output from Codex', got %q", msg.Text)
	}

	if err := <-done; err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCodexAdapter_Process_MessageField(t *testing.T) {
	adapter := &CodexAdapter{}
	input := `{"type":"status","message":"Processing request"}`

	events := make(chan Event, 10)
	done := make(chan error)

	go func() {
		done <- adapter.Process(strings.NewReader(input), events)
	}()

	// Should receive one AssistantMessage event using message field
	event := <-events
	msg, ok := event.(AssistantMessage)
	if !ok {
		t.Fatalf("expected AssistantMessage, got %T", event)
	}
	if msg.Text != "Processing request" {
		t.Errorf("expected text 'Processing request', got %q", msg.Text)
	}

	if err := <-done; err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCodexAdapter_Process_ToolUse(t *testing.T) {
	adapter := &CodexAdapter{}
	input := `{"type":"tool","tool":"Read","content":"Reading file"}`

	events := make(chan Event, 10)
	done := make(chan error)

	go func() {
		done <- adapter.Process(strings.NewReader(input), events)
	}()

	// Should receive ToolUse event followed by AssistantMessage
	event1 := <-events
	tool, ok := event1.(ToolUse)
	if !ok {
		t.Fatalf("expected ToolUse, got %T", event1)
	}
	if tool.Name != "Read" {
		t.Errorf("expected tool name 'Read', got %q", tool.Name)
	}

	event2 := <-events
	msg, ok := event2.(AssistantMessage)
	if !ok {
		t.Fatalf("expected AssistantMessage, got %T", event2)
	}
	if msg.Text != "Reading file" {
		t.Errorf("expected text 'Reading file', got %q", msg.Text)
	}

	if err := <-done; err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCodexAdapter_Process_Error(t *testing.T) {
	adapter := &CodexAdapter{}
	input := `{"type":"error","error":"Failed to execute command"}`

	events := make(chan Event, 10)
	done := make(chan error)

	go func() {
		done <- adapter.Process(strings.NewReader(input), events)
	}()

	// Should receive one Error event
	event := <-events
	errEvent, ok := event.(Error)
	if !ok {
		t.Fatalf("expected Error, got %T", event)
	}
	if errEvent.Message != "Failed to execute command" {
		t.Errorf("expected error 'Failed to execute command', got %q", errEvent.Message)
	}

	if err := <-done; err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCodexAdapter_Process_ErrorWithContent(t *testing.T) {
	adapter := &CodexAdapter{}
	// Error should take priority over content
	input := `{"type":"error","error":"Failed","content":"This should be ignored"}`

	events := make(chan Event, 10)
	done := make(chan error)

	go func() {
		done <- adapter.Process(strings.NewReader(input), events)
	}()

	// Should receive only Error event, not AssistantMessage
	event := <-events
	errEvent, ok := event.(Error)
	if !ok {
		t.Fatalf("expected Error, got %T", event)
	}
	if errEvent.Message != "Failed" {
		t.Errorf("expected error 'Failed', got %q", errEvent.Message)
	}

	// No more events
	if err := <-done; err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	select {
	case event := <-events:
		t.Errorf("expected no more events, got %T", event)
	default:
		// OK
	}
}

func TestCodexAdapter_Process_MultipleEvents(t *testing.T) {
	adapter := &CodexAdapter{}
	input := `{"type":"tool","tool":"Read","content":"Reading file"}
{"type":"message","content":"File contents here"}
{"type":"tool","tool":"Edit"}
{"type":"message","content":"Done!"}`

	events := make(chan Event, 10)
	done := make(chan error)

	go func() {
		done <- adapter.Process(strings.NewReader(input), events)
	}()

	// Should receive: ToolUse, AssistantMessage, AssistantMessage, ToolUse, AssistantMessage
	// First event has both tool and content, so emits both ToolUse and AssistantMessage
	expectedEvents := []struct {
		typ  string
		data string
	}{
		{"ToolUse", "Read"},
		{"AssistantMessage", "Reading file"},
		{"AssistantMessage", "File contents here"},
		{"ToolUse", "Edit"},
		{"AssistantMessage", "Done!"},
	}

	for i, expected := range expectedEvents {
		event := <-events
		switch e := event.(type) {
		case ToolUse:
			if expected.typ != "ToolUse" {
				t.Errorf("event %d: expected %s, got ToolUse", i, expected.typ)
			}
			if e.Name != expected.data {
				t.Errorf("event %d: expected name %q, got %q", i, expected.data, e.Name)
			}
		case AssistantMessage:
			if expected.typ != "AssistantMessage" {
				t.Errorf("event %d: expected %s, got AssistantMessage", i, expected.typ)
			}
			if e.Text != expected.data {
				t.Errorf("event %d: expected text %q, got %q", i, expected.data, e.Text)
			}
		default:
			t.Errorf("event %d: unexpected type %T", i, event)
		}
	}

	if err := <-done; err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCodexAdapter_Process_MalformedJSON_FallbackToPlainText(t *testing.T) {
	adapter := &CodexAdapter{}
	// Mix of valid and invalid JSON - adapter should treat invalid as plain text
	input := `{"type":"tool","tool":"Read"}
This is plain text, not JSON
{"type":"message","content":"Success"}`

	events := make(chan Event, 10)
	done := make(chan error)

	go func() {
		done <- adapter.Process(strings.NewReader(input), events)
	}()

	// Should receive ToolUse, AssistantMessage (plain text), AssistantMessage
	event1 := <-events
	if tool, ok := event1.(ToolUse); !ok || tool.Name != "Read" {
		t.Errorf("expected ToolUse(Read), got %v", event1)
	}

	event2 := <-events
	if msg, ok := event2.(AssistantMessage); !ok || msg.Text != "This is plain text, not JSON" {
		t.Errorf("expected AssistantMessage(plain text), got %v", event2)
	}

	event3 := <-events
	if msg, ok := event3.(AssistantMessage); !ok || msg.Text != "Success" {
		t.Errorf("expected AssistantMessage(Success), got %v", event3)
	}

	if err := <-done; err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCodexAdapter_Process_EmptyLines(t *testing.T) {
	adapter := &CodexAdapter{}
	input := `{"type":"tool","tool":"Read"}

{"type":"message","content":"Done"}`

	events := make(chan Event, 10)
	done := make(chan error)

	go func() {
		done <- adapter.Process(strings.NewReader(input), events)
	}()

	// Should skip empty line and process valid events
	event1 := <-events
	if tool, ok := event1.(ToolUse); !ok || tool.Name != "Read" {
		t.Errorf("expected ToolUse(Read), got %v", event1)
	}

	event2 := <-events
	if msg, ok := event2.(AssistantMessage); !ok || msg.Text != "Done" {
		t.Errorf("expected AssistantMessage(Done), got %v", event2)
	}

	if err := <-done; err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCodexAdapter_Process_ContentPriorityOrder(t *testing.T) {
	adapter := &CodexAdapter{}
	// When multiple text fields present, content takes priority over text over message
	input := `{"type":"test","content":"from content","text":"from text","message":"from message"}`

	events := make(chan Event, 10)
	done := make(chan error)

	go func() {
		done <- adapter.Process(strings.NewReader(input), events)
	}()

	// Should use content field (highest priority)
	event := <-events
	msg, ok := event.(AssistantMessage)
	if !ok {
		t.Fatalf("expected AssistantMessage, got %T", event)
	}
	if msg.Text != "from content" {
		t.Errorf("expected text from 'content' field, got %q", msg.Text)
	}

	if err := <-done; err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCodexAdapter_Process_TextPriorityOverMessage(t *testing.T) {
	adapter := &CodexAdapter{}
	// When text and message present (no content), text takes priority
	input := `{"type":"test","text":"from text","message":"from message"}`

	events := make(chan Event, 10)
	done := make(chan error)

	go func() {
		done <- adapter.Process(strings.NewReader(input), events)
	}()

	// Should use text field (second priority)
	event := <-events
	msg, ok := event.(AssistantMessage)
	if !ok {
		t.Fatalf("expected AssistantMessage, got %T", event)
	}
	if msg.Text != "from text" {
		t.Errorf("expected text from 'text' field, got %q", msg.Text)
	}

	if err := <-done; err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCodexAdapter_Process_EmptyToolName(t *testing.T) {
	adapter := &CodexAdapter{}
	// Tool with empty name should not emit ToolUse event
	input := `{"type":"tool","tool":"","content":"Some content"}`

	events := make(chan Event, 10)
	done := make(chan error)

	go func() {
		done <- adapter.Process(strings.NewReader(input), events)
	}()

	// Should only receive AssistantMessage, not ToolUse
	event := <-events
	if _, ok := event.(ToolUse); ok {
		t.Errorf("expected no ToolUse for empty tool name")
	}
	if msg, ok := event.(AssistantMessage); !ok || msg.Text != "Some content" {
		t.Errorf("expected AssistantMessage(Some content), got %v", event)
	}

	if err := <-done; err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCodexAdapter_Process_NoTextFields(t *testing.T) {
	adapter := &CodexAdapter{}
	// Event with no content, text, or message fields should not emit AssistantMessage
	input := `{"type":"status","code":200}`

	events := make(chan Event, 10)
	done := make(chan error)

	go func() {
		done <- adapter.Process(strings.NewReader(input), events)
	}()

	if err := <-done; err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should not emit any events
	select {
	case event := <-events:
		t.Errorf("expected no events for entry without text fields, got %T", event)
	default:
		// OK
	}
}

func TestCodexAdapter_Process_ToolWithoutContent(t *testing.T) {
	adapter := &CodexAdapter{}
	// Tool event with no content should still emit ToolUse
	input := `{"type":"tool","tool":"Bash"}`

	events := make(chan Event, 10)
	done := make(chan error)

	go func() {
		done <- adapter.Process(strings.NewReader(input), events)
	}()

	// Should only receive ToolUse event
	event := <-events
	tool, ok := event.(ToolUse)
	if !ok {
		t.Fatalf("expected ToolUse, got %T", event)
	}
	if tool.Name != "Bash" {
		t.Errorf("expected tool name 'Bash', got %q", tool.Name)
	}

	if err := <-done; err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// No more events
	select {
	case event := <-events:
		t.Errorf("expected no more events, got %T", event)
	default:
		// OK
	}
}

func TestCodexAdapter_Process_EmptyInput(t *testing.T) {
	adapter := &CodexAdapter{}
	input := ``

	events := make(chan Event, 10)
	done := make(chan error)

	go func() {
		done <- adapter.Process(strings.NewReader(input), events)
	}()

	// Should not error and should not emit events
	if err := <-done; err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	select {
	case event := <-events:
		t.Errorf("expected no events for empty input, got %T", event)
	default:
		// OK
	}
}
