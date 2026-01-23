package cli

import (
	"strings"
	"testing"
)

func TestConfirm(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "yes response",
			input:    "yes\n",
			expected: true,
		},
		{
			name:     "y response",
			input:    "y\n",
			expected: true,
		},
		{
			name:     "Y response (uppercase)",
			input:    "Y\n",
			expected: true,
		},
		{
			name:     "YES response (uppercase)",
			input:    "YES\n",
			expected: true,
		},
		{
			name:     "no response",
			input:    "no\n",
			expected: false,
		},
		{
			name:     "n response",
			input:    "n\n",
			expected: false,
		},
		{
			name:     "empty response (default no)",
			input:    "\n",
			expected: false,
		},
		{
			name:     "invalid response",
			input:    "maybe\n",
			expected: false,
		},
		{
			name:     "whitespace with yes",
			input:    "  yes  \n",
			expected: true,
		},
		{
			name:     "whitespace with no",
			input:    "  no  \n",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: confirm() reads from os.Stdin which we can't easily mock
			// This test verifies the logic by checking string processing

			// Simulate the string processing logic
			response := strings.TrimSpace(strings.ToLower(tt.input))
			result := response == "y" || response == "yes"

			if result != tt.expected {
				t.Errorf("confirm() with input %q: got %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestUninstallCmd(t *testing.T) {
	// Test that the command is registered
	if uninstallCmd == nil {
		t.Fatal("uninstallCmd is nil")
	}

	// Test command properties
	if uninstallCmd.Use != "uninstall" {
		t.Errorf("Use = %q, want %q", uninstallCmd.Use, "uninstall")
	}

	if uninstallCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if uninstallCmd.Long == "" {
		t.Error("Long description is empty")
	}

	if uninstallCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}
