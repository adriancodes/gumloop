package runner

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/adriancodes/gumloop/internal/agent"
	"github.com/adriancodes/gumloop/internal/config"
	"github.com/adriancodes/gumloop/internal/git"
	"github.com/adriancodes/gumloop/internal/memory"
)

// ExitCode represents the exit code returned by the runner
type ExitCode int

const (
	// ExitSuccess indicates work is complete (no changes detected)
	ExitSuccess ExitCode = 0

	// ExitError indicates a general error (config, validation, runtime)
	ExitError ExitCode = 1

	// ExitSafety indicates a safety refusal (dangerous path, no git)
	ExitSafety ExitCode = 2

	// ExitMaxIterations indicates max iterations reached
	ExitMaxIterations ExitCode = 3

	// ExitStuck indicates stuck (changes but no commits for N iterations)
	ExitStuck ExitCode = 4

	// ExitInterrupt indicates user interrupted (Ctrl+C)
	ExitInterrupt ExitCode = 130
)

// Runner manages the execution loop
type Runner struct {
	config  *config.Config
	prompt  string
	agent   *agent.Agent
	maxIters int  // 0 means unlimited (loop until complete)
	singleRun bool // true if not in choo-choo mode
	metrics *Metrics
	memory  *memory.SessionMemory // nil if memory disabled

	// For stuck detection
	iterationsWithoutCommit int
}

// New creates a new Runner instance
func New(cfg *config.Config, prompt string, ag *agent.Agent, chooChoo bool, maxIters int, mem *memory.SessionMemory) *Runner {
	return &Runner{
		config:    cfg,
		prompt:    prompt,
		agent:     ag,
		maxIters:  maxIters,
		singleRun: !chooChoo,
		metrics:   NewMetrics(),
		memory:    mem,
	}
}

// Run executes the main loop and returns the exit code
func (r *Runner) Run() ExitCode {
	// Set up signal handling for Ctrl+C
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n⚠️  Interrupted by user")
		cancel()
	}()

	// Main loop
	for {
		// Check if context was cancelled (Ctrl+C)
		select {
		case <-ctx.Done():
			r.metrics.ExitReason = ExitReasonString(ExitInterrupt)
			r.saveMemory(ExitInterrupt)
			return ExitInterrupt
		default:
		}

		// Check if we've reached max iterations
		if r.maxIters > 0 && r.metrics.Iterations >= r.maxIters {
			r.metrics.ExitReason = ExitReasonString(ExitMaxIterations)
			r.saveMemory(ExitMaxIterations)
			return ExitMaxIterations
		}

		// Increment iteration counter
		r.metrics.Iterations++

		// Display iteration header
		if r.maxIters > 0 {
			fmt.Printf("\n══════════════════════════════════════\n")
			fmt.Printf("  🚂 ITERATION %d of %d\n", r.metrics.Iterations, r.maxIters)
			fmt.Printf("  %s | %s\n", time.Now().Format("15:04:05"), r.agent.Name)
			fmt.Printf("══════════════════════════════════════\n\n")
		} else {
			fmt.Printf("\n══════════════════════════════════════\n")
			fmt.Printf("  🚂 ITERATION %d\n", r.metrics.Iterations)
			fmt.Printf("  %s | %s\n", time.Now().Format("15:04:05"), r.agent.Name)
			fmt.Printf("══════════════════════════════════════\n\n")
		}

		// Run the iteration
		commitsMade, err := RunIteration(
			r.agent,
			r.prompt,
			r.config.Model,
			r.config.Verify,
			!r.singleRun, // autonomous mode = choo-choo mode
		)

		if err != nil {
			fmt.Printf("⚠️  Iteration error: %v\n", err)
			// Continue to next iteration on error (don't fail the whole loop)
		}

		r.metrics.Commits += commitsMade

		// Update session memory with iteration results
		r.recordMemory(commitsMade)

		// Push if commits were made and auto_push is enabled
		if commitsMade > 0 && r.config.AutoPush {
			branch, err := git.GetBranch()
			if err != nil {
				fmt.Printf("⚠️  Warning: failed to get branch name: %v\n", err)
			} else {
				fmt.Printf("☁️  Pushing to origin/%s...\n", branch)
				if err := git.Push(branch); err != nil {
					fmt.Printf("⚠️  Push failed: %v. Continuing without push.\n", err)
				} else {
					fmt.Printf("✅ Pushed to origin/%s\n", branch)
				}
			}
		}

		// Check for changes
		hasChanges, err := git.HasChanges()
		if err != nil {
			fmt.Printf("⚠️  Warning: failed to check for changes: %v\n", err)
			hasChanges = false
		}

		// Exit condition: no changes (complete)
		if !hasChanges && commitsMade == 0 {
			r.metrics.ExitReason = ExitReasonString(ExitSuccess)
			r.saveMemory(ExitSuccess)
			return ExitSuccess
		}

		// Stuck detection: changes but no commits
		if hasChanges && commitsMade == 0 {
			r.iterationsWithoutCommit++
			if r.iterationsWithoutCommit >= r.config.StuckThreshold {
				r.metrics.ExitReason = ExitReasonString(ExitStuck)
				r.saveMemory(ExitStuck)
				return ExitStuck
			}
		} else if commitsMade > 0 {
			// Reset stuck counter if commits were made
			r.iterationsWithoutCommit = 0
		}

		// Exit after first iteration if single-run mode
		if r.singleRun {
			r.metrics.ExitReason = ExitReasonString(ExitSuccess)
			r.saveMemory(ExitSuccess)
			return ExitSuccess
		}

		// Continue to next iteration
	}
}

// recordMemory updates the session memory with results from the latest iteration.
// Silently no-ops if memory is disabled.
func (r *Runner) recordMemory(commitsMade int) {
	if r.memory == nil {
		return
	}

	// Get commit details if commits were made
	var newCommits []memory.CommitRecord
	if commitsMade > 0 {
		gitCommits, err := git.GetRecentCommits(commitsMade)
		if err == nil {
			for _, c := range gitCommits {
				newCommits = append(newCommits, memory.CommitRecord{
					Hash:    c.Hash,
					Message: c.Message,
				})
			}
		}
	}

	r.memory.RecordIteration(commitsMade, newCommits)

	// Save after each iteration so Ctrl+C doesn't lose state
	if err := r.memory.Save(memory.DefaultFileName); err != nil {
		fmt.Printf("⚠️  Warning: failed to save session memory: %v\n", err)
	}
}

// saveMemory records the exit reason and saves the memory file.
// Silently no-ops if memory is disabled.
func (r *Runner) saveMemory(exitCode ExitCode) {
	if r.memory == nil {
		return
	}

	r.memory.SetExit(ExitReasonString(exitCode))
	if err := r.memory.Save(memory.DefaultFileName); err != nil {
		fmt.Printf("⚠️  Warning: failed to save session memory: %v\n", err)
	}
}

// GetMetrics returns current runner metrics
func (r *Runner) GetMetrics() *Metrics {
	return r.metrics
}
