package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRepo creates a temporary git repository for testing
func setupTestRepo(t *testing.T) (repoPath string, cleanup func()) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "gumloop-git-test-*")
	require.NoError(t, err)

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	err = cmd.Run()
	require.NoError(t, err)

	// Configure git user for commits
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tmpDir
	err = cmd.Run()
	require.NoError(t, err)

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	err = cmd.Run()
	require.NoError(t, err)

	// Save original dir and change to test repo
	origDir, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	cleanup = func() {
		os.Chdir(origDir)
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

// createCommit creates a commit in the current repo
func createCommit(t *testing.T, filename string, content string) {
	// Create file
	err := os.WriteFile(filename, []byte(content), 0644)
	require.NoError(t, err)

	// Add to git
	cmd := exec.Command("git", "add", filename)
	err = cmd.Run()
	require.NoError(t, err)

	// Commit
	cmd = exec.Command("git", "commit", "-m", "test commit")
	err = cmd.Run()
	require.NoError(t, err)
}

func TestIsInsideWorkTree(t *testing.T) {
	t.Run("inside git repo", func(t *testing.T) {
		_, cleanup := setupTestRepo(t)
		defer cleanup()

		assert.True(t, IsInsideWorkTree())
	})

	t.Run("outside git repo", func(t *testing.T) {
		// Create temp dir without git init
		tmpDir, err := os.MkdirTemp("", "gumloop-no-git-*")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		origDir, err := os.Getwd()
		require.NoError(t, err)
		defer os.Chdir(origDir)

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		assert.False(t, IsInsideWorkTree())
	})
}

func TestGetBranch(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// New repo should be on main or master
	branch, err := GetBranch()
	require.NoError(t, err)
	assert.Contains(t, []string{"main", "master"}, branch)

	// Create initial commit so we can create branches
	createCommit(t, "test.txt", "content")

	// Create and switch to new branch
	cmd := exec.Command("git", "checkout", "-b", "feature-branch")
	err = cmd.Run()
	require.NoError(t, err)

	branch, err = GetBranch()
	require.NoError(t, err)
	assert.Equal(t, "feature-branch", branch)
}

func TestGetBranchError(t *testing.T) {
	// Test outside git repo
	tmpDir, err := os.MkdirTemp("", "gumloop-no-git-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origDir)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	_, err = GetBranch()
	assert.Error(t, err)
}

func TestCountCommits(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	t.Run("new repo with no commits", func(t *testing.T) {
		count, err := CountCommits()
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("after creating commits", func(t *testing.T) {
		createCommit(t, "file1.txt", "content1")
		count, err := CountCommits()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		createCommit(t, "file2.txt", "content2")
		count, err = CountCommits()
		require.NoError(t, err)
		assert.Equal(t, 2, count)

		createCommit(t, "file3.txt", "content3")
		count, err = CountCommits()
		require.NoError(t, err)
		assert.Equal(t, 3, count)
	})
}

func TestCountCommitsError(t *testing.T) {
	// Test outside git repo
	tmpDir, err := os.MkdirTemp("", "gumloop-no-git-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origDir)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	_, err = CountCommits()
	assert.Error(t, err)
}

func TestHasChanges(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	t.Run("clean working tree", func(t *testing.T) {
		has, err := HasChanges()
		require.NoError(t, err)
		assert.False(t, has)
	})

	t.Run("with uncommitted changes", func(t *testing.T) {
		// Create initial commit
		createCommit(t, "file1.txt", "content1")

		// Modify file
		err := os.WriteFile("file1.txt", []byte("modified"), 0644)
		require.NoError(t, err)

		has, err := HasChanges()
		require.NoError(t, err)
		assert.True(t, has)
	})

	t.Run("with untracked file", func(t *testing.T) {
		// Reset to clean state
		cmd := exec.Command("git", "checkout", ".")
		cmd.Run()

		// Create untracked file
		err := os.WriteFile("untracked.txt", []byte("new file"), 0644)
		require.NoError(t, err)

		has, err := HasChanges()
		require.NoError(t, err)
		assert.True(t, has)
	})

	t.Run("with staged changes", func(t *testing.T) {
		// Clean up
		cmd := exec.Command("git", "clean", "-fd")
		cmd.Run()
		cmd = exec.Command("git", "checkout", ".")
		cmd.Run()

		// Modify and stage
		err := os.WriteFile("file1.txt", []byte("staged content"), 0644)
		require.NoError(t, err)

		cmd = exec.Command("git", "add", "file1.txt")
		err = cmd.Run()
		require.NoError(t, err)

		has, err := HasChanges()
		require.NoError(t, err)
		assert.True(t, has)
	})
}

func TestGetChangedFiles(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create initial commit
	createCommit(t, "file1.txt", "content1")

	t.Run("clean working tree", func(t *testing.T) {
		modified, staged, untracked, err := GetChangedFiles()
		require.NoError(t, err)
		assert.Equal(t, 0, modified)
		assert.Equal(t, 0, staged)
		assert.Equal(t, 0, untracked)
	})

	t.Run("with modified file", func(t *testing.T) {
		err := os.WriteFile("file1.txt", []byte("modified"), 0644)
		require.NoError(t, err)

		modified, staged, untracked, err := GetChangedFiles()
		require.NoError(t, err)
		assert.Equal(t, 1, modified)
		assert.Equal(t, 0, staged)
		assert.Equal(t, 0, untracked)
	})

	t.Run("with staged file", func(t *testing.T) {
		cmd := exec.Command("git", "add", "file1.txt")
		err := cmd.Run()
		require.NoError(t, err)

		modified, staged, untracked, err := GetChangedFiles()
		require.NoError(t, err)
		assert.Equal(t, 0, modified)
		assert.Equal(t, 1, staged)
		assert.Equal(t, 0, untracked)
	})

	t.Run("with untracked file", func(t *testing.T) {
		// Reset to clean
		cmd := exec.Command("git", "reset", "HEAD", "file1.txt")
		cmd.Run()
		cmd = exec.Command("git", "checkout", ".")
		cmd.Run()

		// Create untracked file
		err := os.WriteFile("untracked.txt", []byte("new"), 0644)
		require.NoError(t, err)

		modified, staged, untracked, err := GetChangedFiles()
		require.NoError(t, err)
		assert.Equal(t, 0, modified)
		assert.Equal(t, 0, staged)
		assert.Equal(t, 1, untracked)
	})

	t.Run("with mixed changes", func(t *testing.T) {
		// Clean first
		cmd := exec.Command("git", "clean", "-fd")
		cmd.Run()
		cmd = exec.Command("git", "checkout", ".")
		cmd.Run()

		// Create multiple types of changes
		// 1. Modified unstaged
		err := os.WriteFile("file1.txt", []byte("modified1"), 0644)
		require.NoError(t, err)

		// 2. Modified and staged
		createCommit(t, "file2.txt", "content2")
		err = os.WriteFile("file2.txt", []byte("modified2"), 0644)
		require.NoError(t, err)
		cmd = exec.Command("git", "add", "file2.txt")
		err = cmd.Run()
		require.NoError(t, err)

		// 3. Untracked
		err = os.WriteFile("untracked1.txt", []byte("new1"), 0644)
		require.NoError(t, err)
		err = os.WriteFile("untracked2.txt", []byte("new2"), 0644)
		require.NoError(t, err)

		modified, staged, untracked, err := GetChangedFiles()
		require.NoError(t, err)
		assert.Equal(t, 1, modified)
		assert.Equal(t, 1, staged)
		assert.Equal(t, 2, untracked)
	})
}

func TestPush(t *testing.T) {
	t.Run("push fails without remote", func(t *testing.T) {
		_, cleanup := setupTestRepo(t)
		defer cleanup()

		createCommit(t, "file.txt", "content")

		err := Push("main")
		// Should fail because there's no remote configured
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "git push failed")
	})

	// Note: We can't easily test successful push without a real remote
	// Integration tests should cover this scenario
}

func TestResetHard(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create initial commit
	createCommit(t, "file1.txt", "content1")

	// Create second commit
	createCommit(t, "file2.txt", "content2")

	// Verify we have 2 commits
	count, err := CountCommits()
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	// Reset to first commit
	err = ResetHard("HEAD~1")
	require.NoError(t, err)

	// Verify we're back to 1 commit
	count, err = CountCommits()
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Verify file2.txt is gone
	_, err = os.Stat("file2.txt")
	assert.True(t, os.IsNotExist(err))
}

func TestResetHardError(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// Try to reset to invalid ref
	err := ResetHard("invalid-ref-123456")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git reset failed")
}

func TestClean(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create initial commit
	createCommit(t, "tracked.txt", "tracked content")

	// Create untracked files
	err := os.WriteFile("untracked1.txt", []byte("untracked1"), 0644)
	require.NoError(t, err)

	err = os.WriteFile("untracked2.txt", []byte("untracked2"), 0644)
	require.NoError(t, err)

	// Create untracked directory with file
	err = os.Mkdir("untracked_dir", 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join("untracked_dir", "file.txt"), []byte("content"), 0644)
	require.NoError(t, err)

	// Verify untracked files exist
	_, err = os.Stat("untracked1.txt")
	require.NoError(t, err)

	// Run clean
	err = Clean()
	require.NoError(t, err)

	// Verify untracked files are gone
	_, err = os.Stat("untracked1.txt")
	assert.True(t, os.IsNotExist(err))

	_, err = os.Stat("untracked2.txt")
	assert.True(t, os.IsNotExist(err))

	_, err = os.Stat("untracked_dir")
	assert.True(t, os.IsNotExist(err))

	// Verify tracked file still exists
	_, err = os.Stat("tracked.txt")
	require.NoError(t, err)
}
