package runner

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/adriancodes/gumloop/internal/adapter"
	"github.com/adriancodes/gumloop/internal/agent"
	"github.com/adriancodes/gumloop/internal/git"
)

// Iteration represents a single iteration of the agent loop
type Iteration struct {
	Number    int
	Agent     *agent.Agent
	Prompt    string
	Model     string
	Verify    string
	Autonomous bool
	StartTime time.Time
	Duration  time.Duration
	Commits   int
	Modified  int
	Staged    int
	Untracked int
	Error     error
}

// RunIteration executes a single iteration of the agent
// Returns the number of commits made and any error encountered
func RunIteration(ag *agent.Agent, prompt string, model string, verify string, autonomous bool) (int, error) {
	iter := &Iteration{
		Agent:      ag,
		Prompt:     prompt,
		Model:      model,
		Verify:     verify,
		Autonomous: autonomous,
		StartTime:  time.Now(),
	}

	// Count commits before
	commitsBefore, err := git.CountCommits()
	if err != nil {
		return 0, fmt.Errorf("failed to count commits before iteration: %w", err)
	}

	// Build the command
	cmdArgs := ag.BuildCommand(prompt, model, autonomous)
	if len(cmdArgs) == 0 {
		return 0, fmt.Errorf("agent BuildCommand returned empty command")
	}

	// Create the command
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Dir, _ = os.Getwd()
	cmd.Env = os.Environ()

	// Handle prompt piping for PromptStylePipe
	if ag.PromptStyle == agent.PromptStylePipe {
		cmd.Stdin = bytes.NewBufferString(prompt)
	}

	// Set up output capture
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return 0, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return 0, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("failed to start agent: %w", err)
	}

	// Create event channel for adapter
	events := make(chan adapter.Event, 100)
	adapterDone := make(chan error, 1)

	// Select the appropriate adapter based on agent
	var adapterImpl adapter.Adapter
	switch ag.ID {
	case "claude":
		adapterImpl = &adapter.ClaudeAdapter{}
	case "codex":
		adapterImpl = &adapter.CodexAdapter{}
	default:
		// Use pass-through for gemini, opencode, cursor, ollama
		adapterImpl = &adapter.PassThroughAdapter{}
	}

	// Start processing output in a goroutine
	go func() {
		// Combine stdout and stderr
		combined := io.MultiReader(stdout, stderr)
		err := adapterImpl.Process(combined, events)
		close(events)
		adapterDone <- err
	}()

	// Display events as they arrive
	go func() {
		for event := range events {
			switch e := event.(type) {
			case adapter.ToolUse:
				fmt.Printf("🔧 %s\n", e.Name)
			case adapter.AssistantMessage:
				if e.Text != "" {
					fmt.Println(e.Text)
				}
			case adapter.Error:
				fmt.Printf("⚠️  %s\n", e.Message)
			}
		}
	}()

	// Wait for command to complete
	cmdErr := cmd.Wait()

	// Wait for adapter to finish
	adapterErr := <-adapterDone

	// Record duration
	iter.Duration = time.Since(iter.StartTime)

	// Check for errors
	if cmdErr != nil {
		// Agent exit non-zero is a warning, not a failure
		fmt.Printf("⚠️  Agent exited with code %v. Continuing...\n", cmdErr)
	}

	if adapterErr != nil {
		return 0, fmt.Errorf("adapter error: %w", adapterErr)
	}

	// Count commits after
	commitsAfter, err := git.CountCommits()
	if err != nil {
		return 0, fmt.Errorf("failed to count commits after iteration: %w", err)
	}

	commitsMade := commitsAfter - commitsBefore

	// Get changed files
	modified, staged, untracked, err := git.GetChangedFiles()
	if err != nil {
		return commitsMade, fmt.Errorf("failed to get changed files: %w", err)
	}

	iter.Commits = commitsMade
	iter.Modified = modified
	iter.Staged = staged
	iter.Untracked = untracked

	// Run verification command if specified
	if verify != "" {
		fmt.Printf("\n🧪 Running verification: %s\n", verify)
		verifyCmd := exec.Command("sh", "-c", verify)
		verifyCmd.Stdout = os.Stdout
		verifyCmd.Stderr = os.Stderr
		verifyCmd.Dir, _ = os.Getwd()

		if err := verifyCmd.Run(); err != nil {
			fmt.Printf("⚠️  Verification failed: %v\n", err)
			return commitsMade, fmt.Errorf("verification failed: %w", err)
		}
		fmt.Println("✅ Verification passed")
	}

	// Display iteration summary
	fmt.Println("\n──────────────────────────────────────")
	fmt.Printf("  Iteration complete (%s)\n", FormatDuration(iter.Duration))
	if commitsMade > 0 {
		fmt.Printf("  ✅ Commits: %d\n", commitsMade)
	} else {
		fmt.Println("  ℹ️  No commits made")
	}
	if modified > 0 || staged > 0 || untracked > 0 {
		fmt.Printf("  📝 Changes: %d modified, %d staged, %d new\n", modified, staged, untracked)
	}
	fmt.Println("──────────────────────────────────────")

	return commitsMade, nil
}
