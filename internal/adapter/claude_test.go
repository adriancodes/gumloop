package adapter

import (
	"strings"
	"testing"
)

func TestClaudeAdapter_Process_AssistantMessage(t *testing.T) {
	adapter := &ClaudeAdapter{}
	input := `{"type":"assistant","message":{"content":[{"type":"text","text":"I'll help you with that."}]}}`

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

func TestClaudeAdapter_Process_ToolUse(t *testing.T) {
	adapter := &ClaudeAdapter{}
	input := `{"type":"tool_use","name":"Read"}`

	events := make(chan Event, 10)
	done := make(chan error)

	go func() {
		done <- adapter.Process(strings.NewReader(input), events)
	}()

	// Should receive one ToolUse event
	event := <-events
	tool, ok := event.(ToolUse)
	if !ok {
		t.Fatalf("expected ToolUse, got %T", event)
	}
	if tool.Name != "Read" {
		t.Errorf("expected tool name 'Read', got %q", tool.Name)
	}

	if err := <-done; err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestClaudeAdapter_Process_StreamEvent(t *testing.T) {
	adapter := &ClaudeAdapter{}
	input := `{"type":"stream_event","event":{"delta":{"type":"text_delta","text":"Hello"}}}`

	events := make(chan Event, 10)
	done := make(chan error)

	go func() {
		done <- adapter.Process(strings.NewReader(input), events)
	}()

	// Should receive one AssistantMessage with delta text
	event := <-events
	msg, ok := event.(AssistantMessage)
	if !ok {
		t.Fatalf("expected AssistantMessage, got %T", event)
	}
	if msg.Text != "Hello" {
		t.Errorf("expected text 'Hello', got %q", msg.Text)
	}

	if err := <-done; err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestClaudeAdapter_Process_Result(t *testing.T) {
	adapter := &ClaudeAdapter{}
	// Result events should be skipped (they duplicate assistant messages)
	input := `{"type":"result"}`

	events := make(chan Event, 10)
	done := make(chan error)

	go func() {
		done <- adapter.Process(strings.NewReader(input), events)
	}()

	// Should not receive any events
	if err := <-done; err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	select {
	case event := <-events:
		t.Errorf("expected no events, got %T", event)
	default:
		// OK - no events received
	}
}

func TestClaudeAdapter_Process_MultipleEvents(t *testing.T) {
	adapter := &ClaudeAdapter{}
	input := `{"type":"tool_use","name":"Read"}
{"type":"assistant","message":{"content":[{"type":"text","text":"File contents here"}]}}
{"type":"tool_use","name":"Edit"}
{"type":"assistant","message":{"content":[{"type":"text","text":"Done!"}]}}`

	events := make(chan Event, 10)
	done := make(chan error)

	go func() {
		done <- adapter.Process(strings.NewReader(input), events)
	}()

	// Should receive: ToolUse, AssistantMessage, ToolUse, AssistantMessage
	expectedEvents := []struct {
		typ  string
		data string
	}{
		{"ToolUse", "Read"},
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

func TestClaudeAdapter_Process_MalformedJSON(t *testing.T) {
	adapter := &ClaudeAdapter{}
	// Mix of valid and invalid JSON - adapter should skip invalid and continue
	input := `{"type":"tool_use","name":"Read"}
{invalid json here}
{"type":"assistant","message":{"content":[{"type":"text","text":"Success"}]}}`

	events := make(chan Event, 10)
	done := make(chan error)

	go func() {
		done <- adapter.Process(strings.NewReader(input), events)
	}()

	// Should receive ToolUse and AssistantMessage (malformed line skipped)
	event1 := <-events
	if tool, ok := event1.(ToolUse); !ok || tool.Name != "Read" {
		t.Errorf("expected ToolUse(Read), got %v", event1)
	}

	event2 := <-events
	if msg, ok := event2.(AssistantMessage); !ok || msg.Text != "Success" {
		t.Errorf("expected AssistantMessage(Success), got %v", event2)
	}

	if err := <-done; err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestClaudeAdapter_Process_EmptyLines(t *testing.T) {
	adapter := &ClaudeAdapter{}
	input := `{"type":"tool_use","name":"Read"}

{"type":"assistant","message":{"content":[{"type":"text","text":"Done"}]}}`

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

func TestClaudeAdapter_Process_UnknownEventType(t *testing.T) {
	adapter := &ClaudeAdapter{}
	// Unknown event type should be logged but not error
	input := `{"type":"unknown_type","data":"something"}`

	events := make(chan Event, 10)
	done := make(chan error)

	go func() {
		done <- adapter.Process(strings.NewReader(input), events)
	}()

	// Should not error
	if err := <-done; err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should not emit any events
	select {
	case event := <-events:
		t.Errorf("expected no events, got %T", event)
	default:
		// OK
	}
}

func TestClaudeAdapter_Process_EmptyToolName(t *testing.T) {
	adapter := &ClaudeAdapter{}
	// tool_use with empty name should not emit event
	input := `{"type":"tool_use","name":""}`

	events := make(chan Event, 10)
	done := make(chan error)

	go func() {
		done <- adapter.Process(strings.NewReader(input), events)
	}()

	if err := <-done; err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	select {
	case event := <-events:
		t.Errorf("expected no events for empty tool name, got %T", event)
	default:
		// OK
	}
}

func TestClaudeAdapter_Process_EmptyText(t *testing.T) {
	adapter := &ClaudeAdapter{}
	// Assistant message with empty text should not emit event
	input := `{"type":"assistant","message":{"content":[{"type":"text","text":""}]}}`

	events := make(chan Event, 10)
	done := make(chan error)

	go func() {
		done <- adapter.Process(strings.NewReader(input), events)
	}()

	if err := <-done; err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	select {
	case event := <-events:
		t.Errorf("expected no events for empty text, got %T", event)
	default:
		// OK
	}
}

func TestClaudeAdapter_Process_MultipleContentBlocks(t *testing.T) {
	adapter := &ClaudeAdapter{}
	// Assistant message with multiple content blocks
	input := `{"type":"assistant","message":{"content":[{"type":"text","text":"First part"},{"type":"text","text":"Second part"}]}}`

	events := make(chan Event, 10)
	done := make(chan error)

	go func() {
		done <- adapter.Process(strings.NewReader(input), events)
	}()

	// Should receive two AssistantMessage events
	event1 := <-events
	if msg, ok := event1.(AssistantMessage); !ok || msg.Text != "First part" {
		t.Errorf("expected AssistantMessage(First part), got %v", event1)
	}

	event2 := <-events
	if msg, ok := event2.(AssistantMessage); !ok || msg.Text != "Second part" {
		t.Errorf("expected AssistantMessage(Second part), got %v", event2)
	}

	if err := <-done; err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
