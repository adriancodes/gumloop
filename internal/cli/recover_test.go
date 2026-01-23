package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRepo creates a temporary git repository for testing
func setupTestRepo(t *testing.T) (string, func()) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "gumloop-recover-test-*")
	require.NoError(t, err)

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run())

	// Configure git user for commits
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run())

	// Cleanup function
	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

// createCommit creates a commit in the test repository
func createCommit(t *testing.T, repoDir, filename, content string) {
	// Write file
	filePath := filepath.Join(repoDir, filename)
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))

	// Stage file
	cmd := exec.Command("git", "add", filename)
	cmd.Dir = repoDir
	require.NoError(t, cmd.Run())

	// Commit
	cmd = exec.Command("git", "commit", "-m", "test commit")
	cmd.Dir = repoDir
	require.NoError(t, cmd.Run())
}

// countCommits counts the number of commits in the repository
func countCommits(t *testing.T, repoDir string) int {
	cmd := exec.Command("git", "rev-list", "--count", "HEAD")
	cmd.Dir = repoDir
	output, err := cmd.Output()
	if err != nil {
		// No commits yet
		return 0
	}

	var count int
	_, err = fmt.Sscanf(string(output), "%d", &count)
	require.NoError(t, err)
	return count
}

// hasUncommittedChanges checks if there are uncommitted changes
func hasUncommittedChanges(t *testing.T, repoDir string) bool {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = repoDir
	output, err := cmd.Output()
	require.NoError(t, err)
	return len(output) > 0
}

func TestRecoverDiscardChanges(t *testing.T) {
	repoDir, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create initial commit
	createCommit(t, repoDir, "file1.txt", "initial content")

	// Make some changes
	require.NoError(t, os.WriteFile(filepath.Join(repoDir, "file1.txt"), []byte("modified"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(repoDir, "file2.txt"), []byte("new file"), 0644))

	// Verify changes exist
	assert.True(t, hasUncommittedChanges(t, repoDir))

	// Change to repo directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(repoDir))
	defer os.Chdir(originalDir)

	// Note: This test cannot fully test the interactive prompt
	// In a real scenario, we would need to mock stdin or refactor the code
	// to accept an io.Reader for testing

	// For now, we test that the validation logic works
	cmd := recoverCmd

	// We can't run it without mocking stdin, but we can verify the command structure
	assert.Equal(t, "recover [N]", cmd.Use)
	assert.Equal(t, "Discard changes or reset commits", cmd.Short)
}

func TestRecoverResetCommits(t *testing.T) {
	repoDir, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create multiple commits
	createCommit(t, repoDir, "file1.txt", "content 1")
	createCommit(t, repoDir, "file2.txt", "content 2")
	createCommit(t, repoDir, "file3.txt", "content 3")

	// Verify we have 3 commits
	assert.Equal(t, 3, countCommits(t, repoDir))

	// Change to repo directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(repoDir))
	defer os.Chdir(originalDir)

	// Test command structure
	cmd := recoverCmd
	cmd.SetArgs([]string{"2"})

	assert.Equal(t, "recover [N]", cmd.Use)
}

func TestRecoverValidation(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupRepo   bool
		expectError bool
		errorMsg    string
	}{
		{
			name:        "invalid N argument",
			args:        []string{"abc"},
			setupRepo:   true,
			expectError: true,
			errorMsg:    "invalid number of commits",
		},
		{
			name:        "negative N",
			args:        []string{"-1"},
			setupRepo:   true,
			expectError: true,
			errorMsg:    "number of commits must be positive",
		},
		{
			name:        "zero N",
			args:        []string{"0"},
			setupRepo:   true,
			expectError: true,
			errorMsg:    "number of commits must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var repoDir string
			var cleanup func()

			if tt.setupRepo {
				repoDir, cleanup = setupTestRepo(t)
				defer cleanup()

				// Create at least one commit for some tests
				createCommit(t, repoDir, "file.txt", "content")

				// Change to repo directory
				originalDir, err := os.Getwd()
				require.NoError(t, err)
				require.NoError(t, os.Chdir(repoDir))
				defer os.Chdir(originalDir)
			}

			// Run the command
			err := runRecover(recoverCmd, tt.args)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRecoverNotInGitRepo(t *testing.T) {
	// Create a temp directory that's NOT a git repo
	tmpDir, err := os.MkdirTemp("", "gumloop-not-repo-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Change to non-repo directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer os.Chdir(originalDir)

	// Try to run recover - should fail with "not in a git repository"
	err = runRecover(recoverCmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not in a git repository")
}

func TestRecoverNoChanges(t *testing.T) {
	repoDir, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create initial commit
	createCommit(t, repoDir, "file1.txt", "initial content")

	// Verify no uncommitted changes
	assert.False(t, hasUncommittedChanges(t, repoDir))

	// Change to repo directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(repoDir))
	defer os.Chdir(originalDir)

	// Running recover with no changes should succeed (no confirmation needed)
	err = recoverDiscardChanges()
	assert.NoError(t, err)
}

func TestRecoverResetTooManyCommits(t *testing.T) {
	repoDir, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create only 2 commits
	createCommit(t, repoDir, "file1.txt", "content 1")
	createCommit(t, repoDir, "file2.txt", "content 2")

	assert.Equal(t, 2, countCommits(t, repoDir))

	// Change to repo directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(repoDir))
	defer os.Chdir(originalDir)

	// Try to reset 5 commits (more than exist)
	err = recoverResetCommits(5)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot reset 5 commits")
}

func TestRecoverNoCommits(t *testing.T) {
	repoDir, cleanup := setupTestRepo(t)
	defer cleanup()

	// Don't create any commits

	assert.Equal(t, 0, countCommits(t, repoDir))

	// Change to repo directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(repoDir))
	defer os.Chdir(originalDir)

	// Try to reset commits when none exist
	err = recoverResetCommits(1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no commits to reset")
}
