package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/adriancodes/gumloop/internal/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withTempDir changes to a temp directory for the test, then restores the original.
func withTempDir(t *testing.T) string {
	t.Helper()
	orig, err := os.Getwd()
	require.NoError(t, err)
	dir := t.TempDir()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { os.Chdir(orig) })
	return dir
}

// captureStdout runs fn while capturing os.Stdout and returns the output.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

// --- memory show ---

func TestMemoryShow_NoFile(t *testing.T) {
	withTempDir(t)

	output := captureStdout(t, func() {
		err := runMemoryShow(nil, nil)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "No session memory found.")
	// Should NOT contain the formatted display header
	assert.NotContains(t, output, "Session Memory")
}

func TestMemoryShow_WithFullMemory(t *testing.T) {
	dir := withTempDir(t)

	mem := &memory.SessionMemory{
		StartedAt:  time.Date(2026, 2, 4, 14, 30, 5, 0, time.UTC),
		Branch:     "feat/add-auth",
		AgentName:  "Claude Code",
		Iterations: 7,
		Commits:    5,
		ExitReason: "Max iterations reached",
		CommitLog: []memory.CommitRecord{
			{Hash: "a1b2c3d", Message: "Add JWT middleware and token validation"},
			{Hash: "d4e5f6g", Message: "Add auth routes and login handler"},
		},
		Remaining: "Refresh token rotation has not been implemented yet.",
	}
	require.NoError(t, mem.Save(filepath.Join(dir, memory.DefaultFileName)))

	output := captureStdout(t, func() {
		err := runMemoryShow(nil, nil)
		assert.NoError(t, err)
	})

	// Header
	assert.Contains(t, output, "Session Memory")

	// Metadata — use specific label prefixes to avoid false matches
	assert.Contains(t, output, "Started:    2026-02-04 14:30:05 UTC")
	assert.Contains(t, output, "Branch:     feat/add-auth")
	assert.Contains(t, output, "Agent:      Claude Code")
	assert.Contains(t, output, "Iterations: 7")
	assert.Contains(t, output, "Commits:    5")
	assert.Contains(t, output, "Exit:       Max iterations reached")

	// Commit log section — use "\nCommits:\n" to match the section header specifically
	assert.Contains(t, output, "\nCommits:\n")
	assert.Contains(t, output, "a1b2c3d  Add JWT middleware and token validation")
	assert.Contains(t, output, "d4e5f6g  Add auth routes and login handler")

	// Remaining section
	assert.Contains(t, output, "Remaining:")
	assert.Contains(t, output, "Refresh token rotation has not been implemented yet.")
}

func TestMemoryShow_MinimalMemory(t *testing.T) {
	// Memory with only required fields — no ExitReason, no CommitLog, no Remaining.
	// Exercises the 3 negative conditional branches in runMemoryShow.
	dir := withTempDir(t)

	mem := &memory.SessionMemory{
		StartedAt:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Branch:     "main",
		AgentName:  "Codex",
		Iterations: 2,
		Commits:    0,
	}
	require.NoError(t, mem.Save(filepath.Join(dir, memory.DefaultFileName)))

	output := captureStdout(t, func() {
		err := runMemoryShow(nil, nil)
		assert.NoError(t, err)
	})

	// Core fields present
	assert.Contains(t, output, "Session Memory")
	assert.Contains(t, output, "Branch:     main")
	assert.Contains(t, output, "Agent:      Codex")
	assert.Contains(t, output, "Iterations: 2")
	assert.Contains(t, output, "Commits:    0")

	// Optional sections absent
	// Use "\nCommits:\n" to match the standalone section header, not the metadata line "  Commits:    0"
	assert.NotContains(t, output, "Exit:")
	assert.NotContains(t, output, "\nCommits:\n")
	assert.NotContains(t, output, "Remaining:")
}

func TestMemoryShow_MalformedFile(t *testing.T) {
	// memory.Load returns &SessionMemory{} for malformed YAML.
	// show should handle this gracefully — display the header with zero-values.
	dir := withTempDir(t)

	path := filepath.Join(dir, memory.DefaultFileName)
	require.NoError(t, os.WriteFile(path, []byte("{{not yaml!"), 0644))

	output := captureStdout(t, func() {
		err := runMemoryShow(nil, nil)
		assert.NoError(t, err)
	})

	// Should display header (mem is non-nil, just empty)
	assert.Contains(t, output, "Session Memory")
	assert.Contains(t, output, "Iterations: 0")
	assert.Contains(t, output, "Commits:    0")

	// No optional sections
	assert.NotContains(t, output, "Exit:")
	assert.NotContains(t, output, "\nCommits:\n")
	assert.NotContains(t, output, "Remaining:")
}

// --- memory clear ---

func TestMemoryClear_NoFile(t *testing.T) {
	withTempDir(t)

	output := captureStdout(t, func() {
		err := runMemoryClear(nil, nil)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "No session memory to clear.")
}

func TestMemoryClear_WithFile(t *testing.T) {
	dir := withTempDir(t)

	// Create a memory file
	mem := &memory.SessionMemory{Branch: "main", AgentName: "test"}
	path := filepath.Join(dir, memory.DefaultFileName)
	require.NoError(t, mem.Save(path))

	// Verify file exists before clear
	_, err := os.Stat(path)
	require.NoError(t, err)

	output := captureStdout(t, func() {
		err := runMemoryClear(nil, nil)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "Session memory cleared.")

	// Verify file is gone
	_, err = os.Stat(path)
	assert.True(t, os.IsNotExist(err))
}

func TestMemoryClear_Idempotent(t *testing.T) {
	dir := withTempDir(t)

	// Create and clear
	mem := &memory.SessionMemory{Branch: "main"}
	require.NoError(t, mem.Save(filepath.Join(dir, memory.DefaultFileName)))

	output1 := captureStdout(t, func() {
		err := runMemoryClear(nil, nil)
		assert.NoError(t, err)
	})
	assert.Contains(t, output1, "Session memory cleared.")

	// Clear again — should report nothing to clear
	output2 := captureStdout(t, func() {
		err := runMemoryClear(nil, nil)
		assert.NoError(t, err)
	})
	assert.Contains(t, output2, "No session memory to clear.")
}

// --- command registration ---

func TestMemoryCommandRegistration(t *testing.T) {
	// Verify memory command is registered on rootCmd
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "memory" {
			found = true

			// Verify subcommands
			subNames := make(map[string]bool)
			for _, sub := range cmd.Commands() {
				subNames[sub.Name()] = true
			}
			assert.True(t, subNames["show"], "memory should have 'show' subcommand")
			assert.True(t, subNames["clear"], "memory should have 'clear' subcommand")
			break
		}
	}
	assert.True(t, found, "'memory' command should be registered on rootCmd")
}
