package cli

import (
	"fmt"
	"os"

	"github.com/adriancodes/gumloop/internal/memory"
	"github.com/spf13/cobra"
)

// memoryCmd represents the memory command
var memoryCmd = &cobra.Command{
	Use:   "memory",
	Short: "Manage session memory",
	Long: `Manage gumloop session memory.

Session memory persists context between runs so the agent can pick up
where the previous session left off. The memory file is stored as
.gumloop-memory.yaml in the project root.`,
}

// memoryShowCmd displays the current session memory
var memoryShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current session memory",
	Long:  `Display the contents of the session memory file (.gumloop-memory.yaml).`,
	Args:  cobra.NoArgs,
	RunE:  runMemoryShow,
}

// memoryClearCmd deletes the session memory file
var memoryClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear session memory",
	Long:  `Delete the session memory file (.gumloop-memory.yaml).`,
	Args:  cobra.NoArgs,
	RunE:  runMemoryClear,
}

func init() {
	rootCmd.AddCommand(memoryCmd)
	memoryCmd.AddCommand(memoryShowCmd)
	memoryCmd.AddCommand(memoryClearCmd)
}

func runMemoryShow(cmd *cobra.Command, args []string) error {
	mem, err := memory.Load(memory.DefaultFileName)
	if err != nil {
		return fmt.Errorf("failed to load session memory: %w", err)
	}

	if mem == nil {
		fmt.Println("No session memory found.")
		return nil
	}

	// Display formatted output
	fmt.Println("Session Memory")
	fmt.Println()
	fmt.Printf("  Started:    %s\n", mem.StartedAt.Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("  Branch:     %s\n", mem.Branch)
	fmt.Printf("  Agent:      %s\n", mem.AgentName)
	fmt.Printf("  Iterations: %d\n", mem.Iterations)
	fmt.Printf("  Commits:    %d\n", mem.Commits)
	if mem.ExitReason != "" {
		fmt.Printf("  Exit:       %s\n", mem.ExitReason)
	}

	if len(mem.CommitLog) > 0 {
		fmt.Println()
		fmt.Println("Commits:")
		for _, c := range mem.CommitLog {
			fmt.Printf("  %s  %s\n", c.Hash, c.Message)
		}
	}

	if mem.Remaining != "" {
		fmt.Println()
		fmt.Println("Remaining:")
		fmt.Printf("  %s\n", mem.Remaining)
	}

	return nil
}

func runMemoryClear(cmd *cobra.Command, args []string) error {
	if _, err := os.Stat(memory.DefaultFileName); os.IsNotExist(err) {
		fmt.Println("No session memory to clear.")
		return nil
	}

	if err := os.Remove(memory.DefaultFileName); err != nil {
		return fmt.Errorf("failed to delete session memory: %w", err)
	}

	fmt.Println("Session memory cleared.")
	return nil
}
