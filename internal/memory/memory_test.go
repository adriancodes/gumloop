package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "memory.yaml")

	original := &SessionMemory{
		StartedAt:  time.Date(2026, 2, 4, 14, 30, 5, 0, time.UTC),
		Branch:     "feat/add-auth",
		AgentName:  "Claude Code",
		Iterations: 7,
		Commits:    5,
		ExitReason: "Max iterations reached",
		CommitLog: []CommitRecord{
			{Hash: "a1b2c3d", Message: "Add JWT middleware"},
			{Hash: "d4e5f6g", Message: "Add auth routes"},
		},
		Remaining: "Refresh token rotation not implemented yet.",
	}

	// Save
	err := original.Save(path)
	require.NoError(t, err)

	// Load
	loaded, err := Load(path)
	require.NoError(t, err)
	require.NotNil(t, loaded)

	// Verify fields
	assert.Equal(t, original.Branch, loaded.Branch)
	assert.Equal(t, original.AgentName, loaded.AgentName)
	assert.Equal(t, original.Iterations, loaded.Iterations)
	assert.Equal(t, original.Commits, loaded.Commits)
	assert.Equal(t, original.ExitReason, loaded.ExitReason)
	assert.Equal(t, original.Remaining, loaded.Remaining)
	assert.Equal(t, len(original.CommitLog), len(loaded.CommitLog))
	assert.Equal(t, original.CommitLog[0].Hash, loaded.CommitLog[0].Hash)
	assert.Equal(t, original.CommitLog[0].Message, loaded.CommitLog[0].Message)
}

func TestLoad_NoFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.yaml")

	mem, err := Load(path)
	assert.NoError(t, err)
	assert.Nil(t, mem)
}

func TestLoad_MalformedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")

	err := os.WriteFile(path, []byte("{{{{not yaml at all!!!!"), 0644)
	require.NoError(t, err)

	mem, err := Load(path)
	assert.NoError(t, err)
	assert.NotNil(t, mem) // Returns empty struct, not nil
	assert.Equal(t, 0, mem.Iterations)
}

func TestToPromptContext_Empty(t *testing.T) {
	mem := &SessionMemory{}
	assert.Empty(t, mem.ToPromptContext())
}

func TestToPromptContext_WithHistory(t *testing.T) {
	mem := &SessionMemory{
		StartedAt:  time.Date(2026, 2, 4, 14, 30, 5, 0, time.UTC),
		Branch:     "main",
		AgentName:  "Claude Code",
		Iterations: 3,
		Commits:    2,
		ExitReason: "Complete (no changes)",
		CommitLog: []CommitRecord{
			{Hash: "abc1234", Message: "Fix login bug"},
			{Hash: "def5678", Message: "Add tests"},
		},
		Remaining: "All tests passing.",
	}

	ctx := mem.ToPromptContext()

	assert.Contains(t, ctx, "PREVIOUS SESSION")
	assert.Contains(t, ctx, "3 iterations")
	assert.Contains(t, ctx, "2 commits")
	assert.Contains(t, ctx, "main")
	assert.Contains(t, ctx, "Claude Code")
	assert.Contains(t, ctx, "abc1234 Fix login bug")
	assert.Contains(t, ctx, "def5678 Add tests")
	assert.Contains(t, ctx, "All tests passing")
	assert.Contains(t, ctx, "END PREVIOUS SESSION")
}

func TestCommitLogCapping(t *testing.T) {
	mem := &SessionMemory{}

	// Add 25 commits
	commits := make([]CommitRecord, 25)
	for i := range commits {
		commits[i] = CommitRecord{Hash: "hash" + string(rune('a'+i)), Message: "commit"}
	}

	mem.RecordIteration(25, commits)

	assert.Equal(t, MaxCommitLog, len(mem.CommitLog))
	assert.Equal(t, 25, mem.Commits)
}

func TestRecordIteration(t *testing.T) {
	mem := &SessionMemory{
		Iterations: 2,
		Commits:    3,
		CommitLog: []CommitRecord{
			{Hash: "old1", Message: "Old commit"},
		},
	}

	newCommits := []CommitRecord{
		{Hash: "new1", Message: "New commit 1"},
		{Hash: "new2", Message: "New commit 2"},
	}

	mem.RecordIteration(2, newCommits)

	assert.Equal(t, 3, mem.Iterations)
	assert.Equal(t, 5, mem.Commits)
	assert.Equal(t, 3, len(mem.CommitLog))
	// New commits should be first (most recent)
	assert.Equal(t, "new1", mem.CommitLog[0].Hash)
	assert.Equal(t, "new2", mem.CommitLog[1].Hash)
	assert.Equal(t, "old1", mem.CommitLog[2].Hash)
}

func TestToPromptContext_NoCommits(t *testing.T) {
	// Iterations happened but no commits were made (e.g., agent explored but didn't commit)
	mem := &SessionMemory{
		Branch:     "main",
		AgentName:  "Claude Code",
		Iterations: 2,
		Commits:    0,
		ExitReason: "Stuck (no commits)",
	}

	ctx := mem.ToPromptContext()

	assert.Contains(t, ctx, "PREVIOUS SESSION")
	assert.Contains(t, ctx, "2 iterations")
	assert.Contains(t, ctx, "0 commits")
	assert.NotContains(t, ctx, "Commits made:") // No commit section when log is empty
	assert.NotContains(t, ctx, "Note:")          // No remaining note
	assert.Contains(t, ctx, "END PREVIOUS SESSION")
}

func TestToPromptContext_NoRemaining(t *testing.T) {
	// Normal session with commits but no "remaining" note
	mem := &SessionMemory{
		Branch:     "feat/tests",
		AgentName:  "Codex",
		Iterations: 1,
		Commits:    1,
		ExitReason: "Complete (no changes)",
		CommitLog: []CommitRecord{
			{Hash: "abc1234", Message: "Add unit tests"},
		},
	}

	ctx := mem.ToPromptContext()

	assert.Contains(t, ctx, "abc1234 Add unit tests")
	assert.NotContains(t, ctx, "Note:") // No remaining note when field is empty
}

func TestRecordIteration_ZeroCommits(t *testing.T) {
	// Iteration where agent made changes but didn't commit
	mem := &SessionMemory{
		Iterations: 1,
		Commits:    2,
		CommitLog: []CommitRecord{
			{Hash: "old1", Message: "Previous commit"},
		},
	}

	mem.RecordIteration(0, nil)

	assert.Equal(t, 2, mem.Iterations) // Incremented
	assert.Equal(t, 2, mem.Commits)    // Unchanged
	assert.Equal(t, 1, len(mem.CommitLog)) // Unchanged
	assert.Equal(t, "old1", mem.CommitLog[0].Hash)
}

func TestRecordIteration_MultipleIterations(t *testing.T) {
	// Simulate 3 iterations accumulating state
	mem := &SessionMemory{}

	// Iteration 1: 1 commit
	mem.RecordIteration(1, []CommitRecord{
		{Hash: "aaa", Message: "First commit"},
	})
	assert.Equal(t, 1, mem.Iterations)
	assert.Equal(t, 1, mem.Commits)
	assert.Equal(t, 1, len(mem.CommitLog))

	// Iteration 2: 0 commits (agent explored)
	mem.RecordIteration(0, nil)
	assert.Equal(t, 2, mem.Iterations)
	assert.Equal(t, 1, mem.Commits)
	assert.Equal(t, 1, len(mem.CommitLog))

	// Iteration 3: 2 commits
	mem.RecordIteration(2, []CommitRecord{
		{Hash: "bbb", Message: "Second commit"},
		{Hash: "ccc", Message: "Third commit"},
	})
	assert.Equal(t, 3, mem.Iterations)
	assert.Equal(t, 3, mem.Commits)
	assert.Equal(t, 3, len(mem.CommitLog))
	// Most recent first
	assert.Equal(t, "bbb", mem.CommitLog[0].Hash)
	assert.Equal(t, "ccc", mem.CommitLog[1].Hash)
	assert.Equal(t, "aaa", mem.CommitLog[2].Hash)
}

func TestCommitLogCapping_AcrossIterations(t *testing.T) {
	mem := &SessionMemory{}

	// Add 15 commits in first iteration
	batch1 := make([]CommitRecord, 15)
	for i := range batch1 {
		batch1[i] = CommitRecord{Hash: fmt.Sprintf("batch1-%d", i), Message: "commit"}
	}
	mem.RecordIteration(15, batch1)
	assert.Equal(t, 15, len(mem.CommitLog))

	// Add 10 more in second iteration — should cap at 20
	batch2 := make([]CommitRecord, 10)
	for i := range batch2 {
		batch2[i] = CommitRecord{Hash: fmt.Sprintf("batch2-%d", i), Message: "commit"}
	}
	mem.RecordIteration(10, batch2)
	assert.Equal(t, MaxCommitLog, len(mem.CommitLog))
	assert.Equal(t, 25, mem.Commits) // Total commits tracked even though log is capped

	// Newest commits should be first
	assert.Equal(t, "batch2-0", mem.CommitLog[0].Hash)
}

func TestSetExit(t *testing.T) {
	mem := &SessionMemory{}
	mem.SetExit("Stuck (no commits)")
	assert.Equal(t, "Stuck (no commits)", mem.ExitReason)
}

func TestSetExit_Overwrite(t *testing.T) {
	// Exit reason can be overwritten (e.g., Ctrl+C replaces previous reason)
	mem := &SessionMemory{ExitReason: "Max iterations reached"}
	mem.SetExit("Interrupted")
	assert.Equal(t, "Interrupted", mem.ExitReason)
}

func TestSaveAndLoad_EmptyCommitLog(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "memory.yaml")

	original := &SessionMemory{
		StartedAt:  time.Date(2026, 2, 4, 14, 0, 0, 0, time.UTC),
		Branch:     "main",
		AgentName:  "Claude Code",
		Iterations: 2,
		Commits:    0,
		ExitReason: "Stuck (no commits)",
		CommitLog:  []CommitRecord{},
	}

	err := original.Save(path)
	require.NoError(t, err)

	loaded, err := Load(path)
	require.NoError(t, err)
	require.NotNil(t, loaded)

	assert.Equal(t, 0, loaded.Commits)
	assert.Equal(t, 0, len(loaded.CommitLog))
	assert.Equal(t, "Stuck (no commits)", loaded.ExitReason)
}

func TestLoad_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.yaml")

	err := os.WriteFile(path, []byte(""), 0644)
	require.NoError(t, err)

	mem, err := Load(path)
	assert.NoError(t, err)
	// Empty YAML unmarshals to zero-value struct
	assert.NotNil(t, mem)
	assert.Equal(t, 0, mem.Iterations)
}

func TestLoad_PartialFile(t *testing.T) {
	// File with only some fields (user hand-edited, or from older version)
	dir := t.TempDir()
	path := filepath.Join(dir, "partial.yaml")

	content := `branch: main
remaining: "Need to finish auth module"
`
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)

	mem, err := Load(path)
	require.NoError(t, err)
	require.NotNil(t, mem)

	assert.Equal(t, "main", mem.Branch)
	assert.Equal(t, "Need to finish auth module", mem.Remaining)
	assert.Equal(t, 0, mem.Iterations)  // Defaults to zero
	assert.Equal(t, "", mem.AgentName)   // Defaults to empty
}

func TestSave_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "memory.yaml")

	mem := &SessionMemory{
		Branch:    "main",
		AgentName: "test",
	}

	err := mem.Save(path)
	require.NoError(t, err)

	// Verify file exists and contains header comment
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "# gumloop session memory")
	assert.Contains(t, string(data), "branch: main")
}
