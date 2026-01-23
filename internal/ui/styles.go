package ui

import "github.com/charmbracelet/lipgloss"

// Simpsons-themed color palette
// Inspired by classic character colors and the iconic title card style
var (
	// Character colors
	ColorSimpsonYellow = lipgloss.Color("#FED90F") // Homer's "Woohoo!" moments
	ColorBartOrange    = lipgloss.Color("#FF7F00") // Bart's shirt
	ColorBartRed       = lipgloss.Color("#D6282B") // Danger/errors
	ColorBartmanPurple = lipgloss.Color("#5C4D9A") // Bartman costume
	ColorMargeBlue     = lipgloss.Color("#2F7BBA") // Marge's hair - calm, informative

	// Title card colors
	ColorSkyBlue   = lipgloss.Color("#70D1FE") // Title card clouds
	ColorTitlePink = lipgloss.Color("#F14E9B") // Title card accent

	// Neutral colors
	ColorWhite = lipgloss.Color("#FFFFFF")
	ColorBlack = lipgloss.Color("#000000")

	// Semantic aliases for backward compatibility
	ColorSuccess = ColorSimpsonYellow
	ColorWarning = ColorBartOrange
	ColorError   = ColorBartRed
	ColorTool    = ColorBartmanPurple
	ColorInfo    = ColorMargeBlue
	ColorMuted   = ColorSkyBlue
	ColorBorder  = ColorTitlePink
)

// Base styles that can be composed into more specific styles

// HeaderStyle is used for bold, prominent text like section headers
var HeaderStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(ColorMargeBlue)

// SuccessStyle indicates successful operations (commits, verification passed, etc.)
var SuccessStyle = lipgloss.NewStyle().
	Foreground(ColorSuccess)

// WarningStyle indicates warnings (non-fatal issues, suggestions)
var WarningStyle = lipgloss.NewStyle().
	Foreground(ColorWarning)

// ErrorStyle indicates errors or failures
var ErrorStyle = lipgloss.NewStyle().
	Foreground(ColorError)

// ToolStyle is used for tool names in iteration output
var ToolStyle = lipgloss.NewStyle().
	Foreground(ColorTool)

// MutedStyle is used for less important text (timestamps, metadata)
var MutedStyle = lipgloss.NewStyle().
	Foreground(ColorMuted)

// RalphStyle is used for the Ralph Wiggum ASCII art - Simpson Yellow!
var RalphStyle = lipgloss.NewStyle().
	Foreground(ColorSimpsonYellow)

// BoxStyle creates a border around content (used for banners, summaries)
var BoxStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(ColorTitlePink).
	Padding(0, 1)

// IterationHeaderStyle is used for the iteration number display
var IterationHeaderStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(ColorMargeBlue).
	Border(lipgloss.DoubleBorder(), true, false, true, false).
	BorderForeground(ColorTitlePink).
	Padding(0, 1)

// SummaryBoxStyle is specifically for the run summary
var SummaryBoxStyle = lipgloss.NewStyle().
	Border(lipgloss.ThickBorder()).
	BorderForeground(ColorSimpsonYellow).
	Padding(0, 1).
	Align(lipgloss.Left)

// ComposeStyles combines multiple styles (later styles override earlier ones)
func ComposeStyles(styles ...lipgloss.Style) lipgloss.Style {
	if len(styles) == 0 {
		return lipgloss.NewStyle()
	}

	result := styles[0]
	for _, style := range styles[1:] {
		result = result.Inherit(style)
	}
	return result
}

// Separator creates a horizontal line for visual separation
func Separator(width int, char string) string {
	if width <= 0 {
		width = 40
	}
	if char == "" {
		char = "─"
	}

	line := ""
	for i := 0; i < width; i++ {
		line += char
	}
	return line
}

// DoubleSeparator creates a double-line horizontal separator
func DoubleSeparator(width int) string {
	return Separator(width, "═")
}

// SimpleSeparator creates a single-line horizontal separator
func SimpleSeparator(width int) string {
	return Separator(width, "─")
}
