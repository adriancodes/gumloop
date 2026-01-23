package ui

import (
	"strings"
	"testing"
)

// TestRalphASCIINotEmpty verifies that the ASCII art is present.
func TestRalphASCIINotEmpty(t *testing.T) {
	if RalphASCII == "" {
		t.Error("RalphASCII should not be empty")
	}
}

// TestRalphASCIIContainsExpectedPattern verifies the ASCII art contains
// recognizable braille characters (the main pattern used in the art).
func TestRalphASCIIContainsExpectedPattern(t *testing.T) {
	// The ASCII art uses braille Unicode characters (U+2800 to U+28FF)
	// Check that it contains these patterns
	if !strings.Contains(RalphASCII, "⠀") {
		t.Error("RalphASCII should contain braille patterns")
	}
}

// TestRalphASCIILineCount verifies the ASCII art has the expected number of lines.
func TestRalphASCIILineCount(t *testing.T) {
	lines := strings.Split(RalphASCII, "\n")
	// The ASCII art has 28 lines (including leading/trailing newlines from raw string)
	expectedLines := 28
	if len(lines) != expectedLines {
		t.Errorf("RalphASCII should have %d lines, got %d", expectedLines, len(lines))
	}
}

// TestRalphQuotesNotEmpty verifies that quotes array is not empty.
func TestRalphQuotesNotEmpty(t *testing.T) {
	if len(ralphQuotes) == 0 {
		t.Error("ralphQuotes should not be empty")
	}
}

// TestRalphQuotesCount verifies we have exactly 12 quotes as in the bash script.
func TestRalphQuotesCount(t *testing.T) {
	expectedCount := 12
	if len(ralphQuotes) != expectedCount {
		t.Errorf("ralphQuotes should have %d quotes, got %d", expectedCount, len(ralphQuotes))
	}
}

// TestRalphQuotesContainExpectedQuotes verifies specific iconic quotes are present.
func TestRalphQuotesContainExpectedQuotes(t *testing.T) {
	expectedQuotes := []string{
		"I'm learnding!",
		"Me fail English? That's unpossible!",
		"I'm Idaho!",
	}

	for _, expected := range expectedQuotes {
		found := false
		for _, quote := range ralphQuotes {
			if quote == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected quote not found: %s", expected)
		}
	}
}

// TestRandomQuoteReturnsValidQuote verifies RandomQuote returns one of the known quotes.
func TestRandomQuoteReturnsValidQuote(t *testing.T) {
	quote := RandomQuote()

	// Verify the returned quote is one of the known quotes
	found := false
	for _, knownQuote := range ralphQuotes {
		if quote == knownQuote {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("RandomQuote returned unknown quote: %s", quote)
	}
}

// TestRandomQuoteNotEmpty verifies RandomQuote never returns empty string.
func TestRandomQuoteNotEmpty(t *testing.T) {
	// Run multiple times to increase confidence
	for i := 0; i < 100; i++ {
		quote := RandomQuote()
		if quote == "" {
			t.Error("RandomQuote should never return empty string")
			break
		}
	}
}

// TestRandomQuoteVariety verifies RandomQuote has some variety.
// This is a probabilistic test - with 12 quotes and 100 calls,
// we should see at least 5 different quotes (very conservative threshold).
func TestRandomQuoteVariety(t *testing.T) {
	seen := make(map[string]bool)

	// Call RandomQuote many times
	for i := 0; i < 100; i++ {
		quote := RandomQuote()
		seen[quote] = true
	}

	// With 100 calls and 12 quotes, we should see at least 5 different ones
	minVariety := 5
	if len(seen) < minVariety {
		t.Errorf("RandomQuote should show variety, expected at least %d different quotes, got %d", minVariety, len(seen))
	}
}

// TestAllQuotesNonEmpty verifies none of the quotes are empty strings.
func TestAllQuotesNonEmpty(t *testing.T) {
	for i, quote := range ralphQuotes {
		if quote == "" {
			t.Errorf("Quote at index %d should not be empty", i)
		}
	}
}

// TestQuotesMatchSpec verifies the quotes exactly match the SPEC.md list.
func TestQuotesMatchSpec(t *testing.T) {
	// These are the exact quotes from SPEC.md section 5.3
	expectedQuotes := []string{
		"I'm learnding!",
		"Me fail English? That's unpossible!",
		"My cat's breath smells like cat food.",
		"I bent my wookiee.",
		"It tastes like burning!",
		"I'm a unitard!",
		"That's where I saw the leprechaun. He told me to burn things.",
		"I found a moon rock in my nose!",
		"My knob tastes funny.",
		"I eated the purple berries.",
		"When I grow up, I want to be a principal or a caterpillar.",
		"I'm Idaho!",
	}

	// Verify count matches
	if len(ralphQuotes) != len(expectedQuotes) {
		t.Errorf("Quote count mismatch: expected %d, got %d", len(expectedQuotes), len(ralphQuotes))
	}

	// Verify each quote matches exactly (order matters for consistency)
	for i, expected := range expectedQuotes {
		if i >= len(ralphQuotes) {
			t.Errorf("Missing quote at index %d: %s", i, expected)
			continue
		}
		if ralphQuotes[i] != expected {
			t.Errorf("Quote mismatch at index %d:\nExpected: %s\nGot: %s", i, expected, ralphQuotes[i])
		}
	}
}
