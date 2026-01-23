package adapter

// Event represents a normalized event emitted by an adapter during agent output processing.
type Event interface {
	isEvent()
}

// ToolUse indicates the agent is using a tool.
type ToolUse struct {
	Name  string // Tool name (e.g., "Read", "Edit", "Bash")
	Input string // Optional: tool input/parameters
}

func (ToolUse) isEvent() {}

// AssistantMessage contains text output from the agent.
type AssistantMessage struct {
	Text string // Message text
}

func (AssistantMessage) isEvent() {}

// Error indicates the agent reported an error.
type Error struct {
	Message string // Error message
}

func (Error) isEvent() {}
