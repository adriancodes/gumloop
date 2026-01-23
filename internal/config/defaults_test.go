package config

import (
	"testing"
)

func TestDefaults(t *testing.T) {
	cfg := Defaults()

	// Verify all default values match SPEC section 3.3
	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"CLI", cfg.CLI, "claude"},
		{"Model", cfg.Model, ""},
		{"PromptFile", cfg.PromptFile, "PROMPT.md"},
		{"AutoPush", cfg.AutoPush, true},
		{"StuckThreshold", cfg.StuckThreshold, 3},
		{"Verify", cfg.Verify, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s: got %v, expected %v", tt.name, tt.got, tt.expected)
			}
		})
	}
}
