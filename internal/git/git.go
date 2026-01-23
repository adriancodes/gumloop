package git

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// IsInsideWorkTree checks if the current directory is inside a git repository
func IsInsideWorkTree() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	err := cmd.Run()
	return err == nil
}

// GetBranch returns the current branch name
func GetBranch() (string, error) {
	// Try symbolic-ref first (works when on a branch)
	cmd := exec.Command("git", "symbolic-ref", "--short", "HEAD")
	output, err := cmd.Output()
	if err == nil {
		branch := strings.TrimSpace(string(output))
		return branch, nil
	}

	// Fallback to rev-parse (works in detached HEAD, but may return "HEAD")
	cmd = exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err = cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	branch := strings.TrimSpace(string(output))
	return branch, nil
}

// CountCommits returns the number of commits on the current branch
func CountCommits() (int, error) {
	cmd := exec.Command("git", "rev-list", "--count", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		// If there are no commits yet (new repo), git exits non-zero
		// Check stderr for the typical error message
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := string(exitErr.Stderr)
			if strings.Contains(stderr, "does not have any commits yet") ||
			   strings.Contains(stderr, "bad revision") ||
			   strings.Contains(stderr, "unknown revision") {
				return 0, nil
			}
		}
		return 0, fmt.Errorf("failed to count commits: %w", err)
	}

	countStr := strings.TrimSpace(string(output))
	count, err := strconv.Atoi(countStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse commit count '%s': %w", countStr, err)
	}

	return count, nil
}

// HasChanges checks if there are any uncommitted changes (modified, staged, or untracked files)
func HasChanges() (bool, error) {
	// Check for changes using git status --porcelain
	// This returns empty string if working tree is clean
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check for changes: %w", err)
	}

	// If output is empty, there are no changes
	return len(strings.TrimSpace(string(output))) > 0, nil
}

// GetChangedFiles returns counts of changed files by category
func GetChangedFiles() (modified int, staged int, untracked int, err error) {
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get changed files: %w", err)
	}

	// Don't trim the output as it will remove the leading space from status codes like " M"
	outputStr := string(output)
	if outputStr == "" {
		return 0, 0, 0, nil
	}

	lines := strings.Split(strings.TrimSuffix(outputStr, "\n"), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		// git status --porcelain format:
		// XY PATH
		// X = staged status, Y = unstaged status
		// ?? = untracked
		// M = modified
		// A = added
		// D = deleted
		// R = renamed
		// C = copied
		// U = updated but unmerged

		if len(line) < 3 {
			continue
		}

		statusCode := line[0:2]

		// Untracked files
		if statusCode == "??" {
			untracked++
			continue
		}

		// Staged changes (first character is not space)
		if statusCode[0] != ' ' && statusCode[0] != '?' {
			staged++
		}

		// Unstaged changes (second character is not space)
		if statusCode[1] != ' ' && statusCode[1] != '?' {
			modified++
		}
	}

	return modified, staged, untracked, nil
}

// Push pushes the current branch to the remote
func Push(branch string) error {
	cmd := exec.Command("git", "push", "origin", branch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git push failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// ResetHard resets the working tree to the specified ref
func ResetHard(ref string) error {
	cmd := exec.Command("git", "reset", "--hard", ref)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git reset failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// CommitInfo holds a short hash and message for a single commit.
type CommitInfo struct {
	Hash    string
	Message string
}

// GetRecentCommits returns the N most recent commits (short hash + first line of message).
func GetRecentCommits(n int) ([]CommitInfo, error) {
	if n <= 0 {
		return nil, nil
	}

	cmd := exec.Command("git", "log", "--oneline", "-n", strconv.Itoa(n))
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get recent commits: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var commits []CommitInfo
	for _, line := range lines {
		if line == "" {
			continue
		}
		// Format: "hash message..."
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			commits = append(commits, CommitInfo{Hash: parts[0], Message: parts[1]})
		} else if len(parts) == 1 {
			commits = append(commits, CommitInfo{Hash: parts[0], Message: ""})
		}
	}

	return commits, nil
}

// Clean removes all untracked files and directories
func Clean() error {
	cmd := exec.Command("git", "clean", "-fd")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clean failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}
