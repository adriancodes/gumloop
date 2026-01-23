package adapter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPassThroughAdapter_Process(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedEvents []Event
	}{
		{
			name:  "single line",
			input: "Hello, world!\n",
			expectedEvents: []Event{
				AssistantMessage{Text: "Hello, world!"},
			},
		},
		{
			name:  "multiple lines",
			input: "Line 1\nLine 2\nLine 3\n",
			expectedEvents: []Event{
				AssistantMessage{Text: "Line 1"},
				AssistantMessage{Text: "Line 2"},
				AssistantMessage{Text: "Line 3"},
			},
		},
		{
			name:           "empty input",
			input:          "",
			expectedEvents: []Event{},
		},
		{
			name:  "line with whitespace",
			input: "  spaced line  \n",
			expectedEvents: []Event{
				AssistantMessage{Text: "  spaced line  "},
			},
		},
		{
			name:  "empty lines",
			input: "\n\n\n",
			expectedEvents: []Event{
				AssistantMessage{Text: ""},
				AssistantMessage{Text: ""},
				AssistantMessage{Text: ""},
			},
		},
		{
			name:  "mixed content",
			input: "Starting task...\n\nProcessing file.go\nDone!\n",
			expectedEvents: []Event{
				AssistantMessage{Text: "Starting task..."},
				AssistantMessage{Text: ""},
				AssistantMessage{Text: "Processing file.go"},
				AssistantMessage{Text: "Done!"},
			},
		},
		{
			name:  "special characters",
			input: "Error: file not found ⚠️\n🔧 Using tool: Read\n",
			expectedEvents: []Event{
				AssistantMessage{Text: "Error: file not found ⚠️"},
				AssistantMessage{Text: "🔧 Using tool: Read"},
			},
		},
		{
			name:  "very long line",
			input: strings.Repeat("a", 10000) + "\n",
			expectedEvents: []Event{
				AssistantMessage{Text: strings.Repeat("a", 10000)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewPassThroughAdapter()
			reader := strings.NewReader(tt.input)
			events := make(chan Event, 100)

			// Run adapter in goroutine
			done := make(chan error, 1)
			go func() {
				done <- adapter.Process(reader, events)
			}()

			// Collect events
			var received []Event
		collectLoop:
			for {
				select {
				case event := <-events:
					received = append(received, event)
				case err := <-done:
					assert.NoError(t, err)
					close(events)
					// Drain any remaining events
					for event := range events {
						received = append(received, event)
					}
					break collectLoop
				}
			}

			// Verify events
			assert.Equal(t, len(tt.expectedEvents), len(received), "event count mismatch")
			for i, expected := range tt.expectedEvents {
				if i < len(received) {
					assert.Equal(t, expected, received[i], "event %d mismatch", i)
				}
			}
		})
	}
}

func TestPassThroughAdapter_NoTrailingNewline(t *testing.T) {
	adapter := NewPassThroughAdapter()
	reader := strings.NewReader("Line without newline")
	events := make(chan Event, 10)

	done := make(chan error, 1)
	go func() {
		done <- adapter.Process(reader, events)
	}()

	err := <-done
	close(events)

	assert.NoError(t, err)

	var received []Event
	for event := range events {
		received = append(received, event)
	}

	assert.Equal(t, 1, len(received))
	assert.Equal(t, AssistantMessage{Text: "Line without newline"}, received[0])
}

func TestPassThroughAdapter_ErrorFromScanner(t *testing.T) {
	// Create a reader that will fail
	reader := &failingReader{}
	adapter := NewPassThroughAdapter()
	events := make(chan Event, 10)

	err := adapter.Process(reader, events)
	close(events)

	// Should return the scanner error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "general error for testing")
}

// failingReader is a reader that always returns an error
type failingReader struct{}

func (r *failingReader) Read(p []byte) (n int, err error) {
	return 0, assert.AnError
}
