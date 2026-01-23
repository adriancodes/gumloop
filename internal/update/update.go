package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/adriancodes/gumloop/internal/ui"
)

const (
	githubAPIURL = "https://api.github.com/repos/adriancodes/gumloop/releases/latest"
	httpTimeout  = 30 * time.Second
)

// Release represents a GitHub release
type Release struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
	} `json:"assets"`
}

// Update performs a self-update to the latest version
func Update(currentVersion string) error {
	fmt.Println(ui.MutedStyle.Render("Checking for updates..."))

	// Fetch latest release info
	release, err := fetchLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to fetch latest release: %w", err)
	}

	// Check if already latest
	if release.TagName == currentVersion || release.TagName == "v"+currentVersion {
		fmt.Println(ui.SuccessStyle.Render("✓ Already at latest version: " + currentVersion))
		return nil
	}

	fmt.Printf("%s %s → %s\n",
		ui.MutedStyle.Render("Update available:"),
		ui.MutedStyle.Render(currentVersion),
		ui.SuccessStyle.Render(release.TagName))

	// Find appropriate binary for current OS/arch
	binaryName := getBinaryName()
	var asset *struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
	}

	for i := range release.Assets {
		if strings.Contains(release.Assets[i].Name, binaryName) {
			asset = &release.Assets[i]
			break
		}
	}

	if asset == nil {
		return fmt.Errorf("no binary found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	fmt.Printf("%s %s (%s)\n",
		ui.MutedStyle.Render("Downloading:"),
		asset.Name,
		formatBytes(asset.Size))

	// Download binary
	tmpFile, err := downloadBinary(asset.BrowserDownloadURL)
	if err != nil {
		return fmt.Errorf("failed to download binary: %w", err)
	}
	defer os.Remove(tmpFile)

	// Get current executable path
	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}

	// Resolve symlinks
	currentExe, err = filepath.EvalSymlinks(currentExe)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	// Replace current binary
	if err := replaceBinary(tmpFile, currentExe); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	fmt.Println(ui.SuccessStyle.Render("✓ Updated to " + release.TagName))
	return nil
}

// fetchLatestRelease fetches the latest release from GitHub
func fetchLatestRelease() (*Release, error) {
	client := &http.Client{Timeout: httpTimeout}

	req, err := http.NewRequest("GET", githubAPIURL, nil)
	if err != nil {
		return nil, err
	}

	// Set User-Agent (GitHub API requires it)
	req.Header.Set("User-Agent", "gumloop-updater")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

// downloadBinary downloads a file to a temporary location
func downloadBinary(url string) (string, error) {
	client := &http.Client{Timeout: 5 * time.Minute}

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Create temp file
	tmpFile, err := os.CreateTemp("", "gumloop-update-*")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	// Copy response body to temp file
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}

	// Make executable
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}

	return tmpFile.Name(), nil
}

// replaceBinary replaces the current binary with the new one
func replaceBinary(newPath, currentPath string) error {
	// Create backup
	backupPath := currentPath + ".backup"
	if err := copyFile(currentPath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Replace with new binary
	if err := copyFile(newPath, currentPath); err != nil {
		// Restore backup on failure
		copyFile(backupPath, currentPath)
		os.Remove(backupPath)
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	// Remove backup on success
	os.Remove(backupPath)

	// Ensure executable
	if err := os.Chmod(currentPath, 0755); err != nil {
		return fmt.Errorf("failed to set executable permissions: %w", err)
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	// Copy permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}

// getBinaryName returns the expected binary name for current OS/arch
func getBinaryName() string {
	// Expected format: gumloop_darwin_arm64, gumloop_linux_amd64, etc.
	return fmt.Sprintf("gumloop_%s_%s", runtime.GOOS, runtime.GOARCH)
}

// formatBytes formats byte size in human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
