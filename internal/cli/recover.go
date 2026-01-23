package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/adriancodes/gumloop/internal/git"
	"github.com/adriancodes/gumloop/internal/ui"
	"github.com/spf13/cobra"
)

// recoverCmd represents the recover command
var recoverCmd = &cobra.Command{
	Use:   "recover [N]",
	Short: "Discard changes or reset commits",
	Long: `Discard uncommitted changes or reset commits.

With no arguments:
  Discards all uncommitted changes (git reset --hard && git clean -fd)

With N:
  Resets the last N commits (git reset --hard HEAD~N)

Examples:
  gumloop recover       # Discard uncommitted changes
  gumloop recover 3     # Reset last 3 commits`,
	Args: cobra.MaximumNArgs(1),
	RunE: runRecover,
}

func init() {
	rootCmd.AddCommand(recoverCmd)
}

func runRecover(cmd *cobra.Command, args []string) error {
	// Check if in git repository
	if !git.IsInsideWorkTree() {
		return fmt.Errorf("not in a git repository")
	}

	// Determine mode: discard changes or reset commits
	if len(args) == 0 {
		return recoverDiscardChanges()
	}

	// Parse N for reset commits
	n, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid number of commits: %s", args[0])
	}

	if n <= 0 {
		return fmt.Errorf("number of commits must be positive, got %d", n)
	}

	return recoverResetCommits(n)
}

// recoverDiscardChanges discards all uncommitted changes
func recoverDiscardChanges() error {
	// Check if there are any changes to discard
	hasChanges, err := git.HasChanges()
	if err != nil {
		return fmt.Errorf("failed to check for changes: %w", err)
	}

	if !hasChanges {
		fmt.Println(ui.SuccessStyle.Render("✓ Working tree is already clean"))
		return nil
	}

	// Get details of what will be affected
	modified, staged, untracked, err := git.GetChangedFiles()
	if err != nil {
		return fmt.Errorf("failed to get changed files: %w", err)
	}

	// Show what will be affected
	fmt.Println(ui.WarningStyle.Render("⚠ This will discard all uncommitted changes:"))
	fmt.Println()

	if modified > 0 {
		fmt.Printf("  • %d modified file(s)\n", modified)
	}
	if staged > 0 {
		fmt.Printf("  • %d staged file(s)\n", staged)
	}
	if untracked > 0 {
		fmt.Printf("  • %d untracked file(s)\n", untracked)
	}

	fmt.Println()

	// Confirm with user
	if !confirmAction("Discard all changes?") {
		fmt.Println("Cancelled.")
		return nil
	}

	// Execute git reset --hard
	if err := git.ResetHard("HEAD"); err != nil {
		return fmt.Errorf("failed to reset: %w", err)
	}

	// Execute git clean -fd
	if err := git.Clean(); err != nil {
		return fmt.Errorf("failed to clean untracked files: %w", err)
	}

	fmt.Println()
	fmt.Println(ui.SuccessStyle.Render("✓ All changes discarded"))

	return nil
}

// recoverResetCommits resets the last N commits
func recoverResetCommits(n int) error {
	// Get current commit count
	totalCommits, err := git.CountCommits()
	if err != nil {
		return fmt.Errorf("failed to count commits: %w", err)
	}

	if totalCommits == 0 {
		return fmt.Errorf("no commits to reset")
	}

	if n > totalCommits {
		return fmt.Errorf("cannot reset %d commits, only %d commit(s) exist", n, totalCommits)
	}

	// Get current branch
	branch, err := git.GetBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// Show what will be affected
	fmt.Println(ui.WarningStyle.Render(fmt.Sprintf("⚠ This will reset the last %d commit(s) on branch '%s'", n, branch)))
	fmt.Println()

	// Show the commits that will be reset
	if err := showCommitsToReset(n); err != nil {
		return fmt.Errorf("failed to show commits: %w", err)
	}

	fmt.Println()

	// Confirm with user
	if !confirmAction(fmt.Sprintf("Reset last %d commit(s)?", n)) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Execute git reset --hard HEAD~N
	ref := fmt.Sprintf("HEAD~%d", n)
	if err := git.ResetHard(ref); err != nil {
		return fmt.Errorf("failed to reset commits: %w", err)
	}

	fmt.Println()
	fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("✓ Reset %d commit(s)", n)))

	return nil
}

// showCommitsToReset displays the commits that will be reset
func showCommitsToReset(n int) error {
	// Use git log to show the last N commits
	cmd := exec.Command("git", "log", "--oneline", "-n", strconv.Itoa(n))
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	fmt.Println("Commits to be reset:")
	fmt.Println()

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		fmt.Printf("  %s\n", ui.MutedStyle.Render(line))
	}

	return nil
}

// confirmAction prompts the user for confirmation
func confirmAction(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("%s (y/N): ", prompt)

	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}
