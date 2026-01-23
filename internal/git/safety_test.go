package git

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsDangerousPath(t *testing.T) {
	home, err := os.UserHomeDir()
	assert.NoError(t, err)

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		// Home directory itself is dangerous
		{
			name:     "home directory",
			path:     home,
			expected: true,
		},
		// Root paths are dangerous
		{
			name:     "root directory",
			path:     "/",
			expected: true,
		},
		{
			name:     "etc directory",
			path:     "/etc",
			expected: true,
		},
		{
			name:     "usr directory",
			path:     "/usr",
			expected: true,
		},
		{
			name:     "var directory",
			path:     "/var",
			expected: true,
		},
		{
			name:     "tmp directory",
			path:     "/tmp",
			expected: true,
		},
		{
			name:     "bin directory",
			path:     "/bin",
			expected: true,
		},
		{
			name:     "sbin directory",
			path:     "/sbin",
			expected: true,
		},
		{
			name:     "lib directory",
			path:     "/lib",
			expected: true,
		},
		// Home subdirectories are NOT dangerous (handled by separate check)
		{
			name:     "home subdirectory",
			path:     filepath.Join(home, "projects"),
			expected: false,
		},
		{
			name:     "deep home subdirectory",
			path:     filepath.Join(home, "dev", "gumloop"),
			expected: false,
		},
		// Safe project directories
		{
			name:     "opt directory",
			path:     "/opt",
			expected: false,
		},
		{
			name:     "srv directory",
			path:     "/srv",
			expected: false,
		},
	}

	// Add macOS-specific tests
	if runtime.GOOS == "darwin" {
		tests = append(tests, []struct {
			name     string
			path     string
			expected bool
		}{
			{
				name:     "System directory (macOS)",
				path:     "/System",
				expected: true,
			},
			{
				name:     "Library directory (macOS)",
				path:     "/Library",
				expected: true,
			},
		}...)
	}

	// Add Windows-specific tests
	if runtime.GOOS == "windows" {
		tests = append(tests, []struct {
			name     string
			path     string
			expected bool
		}{
			{
				name:     "Windows directory",
				path:     "C:\\Windows",
				expected: true,
			},
			{
				name:     "Program Files directory",
				path:     "C:\\Program Files",
				expected: true,
			},
			{
				name:     "Program Files x86 directory",
				path:     "C:\\Program Files (x86)",
				expected: true,
			},
		}...)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsDangerousPath(tt.path)
			assert.Equal(t, tt.expected, result, "IsDangerousPath(%q) = %v, want %v", tt.path, result, tt.expected)
		})
	}
}

func TestIsDangerousPath_WithTrailingSlash(t *testing.T) {
	// Test that paths with trailing slashes are handled correctly
	result := IsDangerousPath("/etc/")
	assert.True(t, result, "path with trailing slash should be recognized as dangerous")
}

func TestIsDangerousPath_WithDotComponents(t *testing.T) {
	// Test that paths with .. or . components are cleaned properly
	result := IsDangerousPath("/usr/../etc")
	assert.True(t, result, "path with .. should be cleaned and recognized as dangerous")
}

func TestIsHomeSubdirectory(t *testing.T) {
	home, err := os.UserHomeDir()
	assert.NoError(t, err)

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		// Home itself is NOT a subdirectory of home
		{
			name:     "home directory itself",
			path:     home,
			expected: false,
		},
		// Direct subdirectories of home
		{
			name:     "home subdirectory - projects",
			path:     filepath.Join(home, "projects"),
			expected: true,
		},
		{
			name:     "home subdirectory - Documents",
			path:     filepath.Join(home, "Documents"),
			expected: true,
		},
		// Deep subdirectories of home
		{
			name:     "deep home subdirectory",
			path:     filepath.Join(home, "dev", "gumloop"),
			expected: true,
		},
		// Paths outside home
		{
			name:     "opt directory",
			path:     "/opt",
			expected: false,
		},
		{
			name:     "tmp directory",
			path:     "/tmp",
			expected: false,
		},
		{
			name:     "srv directory",
			path:     "/srv/myproject",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsHomeSubdirectory(tt.path)
			assert.Equal(t, tt.expected, result, "IsHomeSubdirectory(%q) = %v, want %v", tt.path, result, tt.expected)
		})
	}
}

func TestIsHomeSubdirectory_RelativePath(t *testing.T) {
	// Test with a relative path - it should be resolved to absolute
	// Create a temp directory under home for testing
	home, err := os.UserHomeDir()
	assert.NoError(t, err)

	tempDir := filepath.Join(home, "temp-test-dir-gumloop")
	err = os.MkdirAll(tempDir, 0755)
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to that directory
	originalDir, err := os.Getwd()
	assert.NoError(t, err)
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	assert.NoError(t, err)

	// Now test with current directory "."
	result := IsHomeSubdirectory(".")
	assert.True(t, result, "current directory (.) inside home should be recognized as home subdirectory")
}

func TestIsHomeSubdirectory_EdgeCases(t *testing.T) {
	home, err := os.UserHomeDir()
	assert.NoError(t, err)

	// Test with path that has the home prefix but isn't actually under home
	// This is a contrived example but tests the implementation
	notHomeDir := home + "x" // e.g., /Users/adrianx if home is /Users/adrian
	result := IsHomeSubdirectory(notHomeDir)
	assert.False(t, result, "path with home prefix but not under home should return false")
}

// TestSafetyChecksIntegration tests the interaction between IsDangerousPath and IsHomeSubdirectory
func TestSafetyChecksIntegration(t *testing.T) {
	home, err := os.UserHomeDir()
	assert.NoError(t, err)

	projectDir := filepath.Join(home, "projects", "myapp")

	// Project directory under home should:
	// - NOT be dangerous (IsDangerousPath = false)
	// - BE a home subdirectory (IsHomeSubdirectory = true)
	isDangerous := IsDangerousPath(projectDir)
	isHomeSub := IsHomeSubdirectory(projectDir)

	assert.False(t, isDangerous, "project under home should not be dangerous")
	assert.True(t, isHomeSub, "project under home should be recognized as home subdirectory")
}

// Note: ConfirmHomeSubdirectory is not tested here because it requires user input.
// It would need integration tests with mocked stdin or manual testing.
