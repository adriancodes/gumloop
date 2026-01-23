package git

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// IsDangerousPath checks if the given path is in the list of dangerous paths
// that should not be allowed for autonomous agent execution.
func IsDangerousPath(path string) bool {
	// Get the absolute path to normalize it
	absPath, err := filepath.Abs(path)
	if err != nil {
		// If we can't get absolute path, treat it as potentially dangerous
		return true
	}

	// Clean the path to remove any .. or . components
	absPath = filepath.Clean(absPath)

	// Get home directory
	home, err := os.UserHomeDir()
	if err != nil {
		home = ""
	}

	// Define dangerous paths based on OS
	dangerousPaths := []string{
		home, // User's home directory itself (but not subdirectories)
		"/",
		"/etc",
		"/usr",
		"/var",
		"/tmp",
		"/bin",
		"/sbin",
		"/lib",
	}

	// Add macOS-specific paths
	if runtime.GOOS == "darwin" {
		dangerousPaths = append(dangerousPaths,
			"/System",
			"/Library",
		)
	}

	// Add Windows-specific paths
	if runtime.GOOS == "windows" {
		dangerousPaths = append(dangerousPaths,
			"C:\\Windows",
			"C:\\Program Files",
			"C:\\Program Files (x86)",
		)
	}

	// Check if absPath exactly matches any dangerous path
	for _, dangerous := range dangerousPaths {
		if dangerous == "" {
			continue
		}
		// Clean the dangerous path as well for comparison
		cleanDangerous := filepath.Clean(dangerous)
		if absPath == cleanDangerous {
			return true
		}
	}

	return false
}

// IsHomeSubdirectory checks if the given path is a subdirectory of $HOME
// (but not $HOME itself, which is caught by IsDangerousPath).
func IsHomeSubdirectory(path string) bool {
	// Get home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	// Get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	// Clean both paths
	absPath = filepath.Clean(absPath)
	home = filepath.Clean(home)

	// Check if path is under home but not equal to home
	if absPath == home {
		return false
	}

	// Check if path starts with home directory
	return strings.HasPrefix(absPath, home+string(filepath.Separator))
}

// ConfirmHomeSubdirectory prompts the user to confirm running in choo-choo mode
// when the current directory is a subdirectory of $HOME.
// Returns true if the user confirms, false otherwise.
func ConfirmHomeSubdirectory() bool {
	fmt.Println()
	fmt.Println("⚠️  WARNING: You are running in choo-choo mode under your home directory.")
	fmt.Println()
	fmt.Println("   Autonomous agents will make changes without asking for permission.")
	fmt.Println("   Git is your safety net, but use caution.")
	fmt.Println()
	fmt.Print("Continue? (y/N): ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}
