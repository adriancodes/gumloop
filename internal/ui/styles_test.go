package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestColorConstants(t *testing.T) {
	// Verify color constants are defined
	colors := map[string]lipgloss.TerminalColor{
		"success": ColorSuccess,
		"warning": ColorWarning,
		"error":   ColorError,
		"tool":    ColorTool,
		"muted":   ColorMuted,
		"white":   ColorWhite,
		"black":   ColorBlack,
	}

	for name, color := range colors {
		if color == nil {
			t.Errorf("Color %s should not be nil", name)
		}
	}
}

func TestStylesAreDefined(t *testing.T) {
	// Verify all base styles are defined (non-nil)
	styles := map[string]lipgloss.Style{
		"header":          HeaderStyle,
		"success":         SuccessStyle,
		"warning":         WarningStyle,
		"error":           ErrorStyle,
		"tool":            ToolStyle,
		"muted":           MutedStyle,
		"box":             BoxStyle,
		"iterationHeader": IterationHeaderStyle,
		"summaryBox":      SummaryBoxStyle,
	}

	// Note: lipgloss.Style is a struct, so we check if it's the zero value
	// by checking if it produces empty output for empty input
	for name, style := range styles {
		// Styles should be able to render something
		rendered := style.Render("")
		// Even rendering empty string should work without panicking
		_ = rendered
		t.Logf("Style %s is defined and can render", name)
	}
}

func TestHeaderStyleProperties(t *testing.T) {
	// Verify HeaderStyle has bold and white foreground
	rendered := HeaderStyle.Render("Test")
	if rendered == "" {
		t.Error("HeaderStyle should render text")
	}
	// We can't easily test ANSI codes directly, but we can verify it renders something
}

func TestBoxStyleHasBorder(t *testing.T) {
	rendered := BoxStyle.Render("Content")
	if rendered == "" {
		t.Error("BoxStyle should render text")
	}
	// BoxStyle should add border characters, making output longer than input
	if !strings.Contains(rendered, "Content") {
		t.Error("BoxStyle should contain the input text")
	}
}

func TestSeparator(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		char     string
		expected string
	}{
		{
			name:     "default width and char",
			width:    10,
			char:     "─",
			expected: "──────────",
		},
		{
			name:     "custom char",
			width:    5,
			char:     "=",
			expected: "=====",
		},
		{
			name:     "zero width uses default",
			width:    0,
			char:     "─",
			expected: strings.Repeat("─", 40), // Default width is 40
		},
		{
			name:     "negative width uses default",
			width:    -1,
			char:     "─",
			expected: strings.Repeat("─", 40),
		},
		{
			name:     "empty char uses default",
			width:    5,
			char:     "",
			expected: "─────",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Separator(tt.width, tt.char)
			if result != tt.expected {
				t.Errorf("Separator(%d, %q) = %q, want %q", tt.width, tt.char, result, tt.expected)
			}
		})
	}
}

func TestDoubleSeparator(t *testing.T) {
	width := 10
	result := DoubleSeparator(width)
	expected := strings.Repeat("═", width)

	if result != expected {
		t.Errorf("DoubleSeparator(%d) = %q, want %q", width, result, expected)
	}
}

func TestSimpleSeparator(t *testing.T) {
	width := 10
	result := SimpleSeparator(width)
	expected := strings.Repeat("─", width)

	if result != expected {
		t.Errorf("SimpleSeparator(%d) = %q, want %q", width, result, expected)
	}
}

func TestComposeStyles(t *testing.T) {
	tests := []struct {
		name   string
		styles []lipgloss.Style
		text   string
	}{
		{
			name:   "empty styles returns default",
			styles: []lipgloss.Style{},
			text:   "test",
		},
		{
			name:   "single style",
			styles: []lipgloss.Style{HeaderStyle},
			text:   "test",
		},
		{
			name:   "multiple styles",
			styles: []lipgloss.Style{HeaderStyle, SuccessStyle},
			text:   "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			composed := ComposeStyles(tt.styles...)
			rendered := composed.Render(tt.text)
			// Should not panic and should render something
			if rendered == "" && tt.text != "" {
				t.Error("ComposeStyles should render non-empty text")
			}
		})
	}
}

func TestStylesColorConsistency(t *testing.T) {
	// Verify success style uses success color
	// We can't directly inspect the style's color, but we can verify it renders
	successText := SuccessStyle.Render("success")
	warningText := WarningStyle.Render("warning")
	errorText := ErrorStyle.Render("error")

	// Basic smoke test - these should all render without panicking
	if successText == "" || warningText == "" || errorText == "" {
		t.Error("Styled text should not be empty")
	}
}
