package ui

import (
	"fmt"
	"testing"
	"time"
)

func TestPreviewSummary(t *testing.T) {
	fmt.Println("\n=== SUCCESS ===")
	fmt.Println(RenderRunSummary(SummaryConfig{
		Agent:      "claude",
		Iterations: 5,
		Commits:    3,
		Duration:   4*time.Minute + 32*time.Second,
		ExitCode:   ExitSuccess,
	}))

	fmt.Println("\n=== MAX ITERATIONS ===")
	fmt.Println(RenderRunSummary(SummaryConfig{
		Agent:      "codex",
		Iterations: 20,
		Commits:    15,
		Duration:   10 * time.Minute,
		ExitCode:   ExitMaxIterations,
	}))

	fmt.Println("\n=== ERROR ===")
	fmt.Println(RenderRunSummary(SummaryConfig{
		Agent:      "gemini",
		Iterations: 1,
		Commits:    0,
		Duration:   30 * time.Second,
		ExitCode:   ExitError,
	}))
}
