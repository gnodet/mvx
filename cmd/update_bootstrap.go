package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// updateBootstrapCmd represents the update-bootstrap command
var updateBootstrapCmd = &cobra.Command{
	Use:   "update-bootstrap",
	Short: "Update the mvx bootstrap scripts to the latest version",
	Long: `Update the mvx bootstrap scripts (mvx and mvx.cmd) to the latest version.

This command will:
  - Check for the latest mvx release on GitHub
  - Download the latest bootstrap scripts
  - Update the local mvx and mvx.cmd files
  - Update version configuration files

Examples:
  mvx update-bootstrap           # Update to latest version
  mvx update-bootstrap --check   # Only check for updates, don't update`,

	Run: func(cmd *cobra.Command, args []string) {
		if err := updateBootstrap(); err != nil {
			printError("%v", err)
			os.Exit(1)
		}
	},
}

var (
	checkOnly bool
)

func init() {
	updateBootstrapCmd.Flags().BoolVar(&checkOnly, "check", false, "only check for updates, don't update")
}

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	HTMLURL string `json:"html_url"`
}

// getLatestRelease fetches the latest release information from GitHub
func getLatestRelease() (*GitHubRelease, error) {
	url := "https://api.github.com/repos/gnodet/mvx/releases/latest"
	
	printVerbose("Fetching latest release from: %s", url)
	
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release information: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("failed to parse release information: %w", err)
	}

	return &release, nil
}

// getCurrentVersion reads the current version from .mvx/mvx.properties file
func getCurrentVersion() (string, error) {
	propertiesFile := ".mvx/mvx.properties"
	if _, err := os.Stat(propertiesFile); os.IsNotExist(err) {
		return "", nil // No properties file means no current version
	}

	content, err := os.ReadFile(propertiesFile)
	if err != nil {
		return "", fmt.Errorf("failed to read properties file: %w", err)
	}

	// Parse the properties file to find mvxVersion
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "mvxVersion=") {
			return strings.TrimSpace(strings.TrimPrefix(line, "mvxVersion=")), nil
		}
	}

	return "", nil // No mvxVersion found
}

// downloadFile downloads a file from the given URL and saves it to the specified path
func downloadFile(url, filepath string) error {
	printVerbose("Downloading %s to %s", url, filepath)
	
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download %s: HTTP %d", url, resp.StatusCode)
	}

	// Create temporary file
	tempFile := filepath + ".tmp"
	out, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer out.Close()

	// Copy content
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Move temporary file to final location
	if err := os.Rename(tempFile, filepath); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("failed to move file to final location: %w", err)
	}

	return nil
}

// updatePropertiesFile updates the mvx.properties file with the new version
func updatePropertiesFile(propertiesFile, version, baseURL string) error {
	if _, err := os.Stat(propertiesFile); os.IsNotExist(err) {
		// Download properties file if it doesn't exist
		printVerbose("Downloading mvx.properties...")
		if err := downloadFile(baseURL+"/.mvx/mvx.properties", propertiesFile); err != nil {
			// Create a minimal properties file if download fails
			printVerbose("Download failed, creating minimal properties file")
			content := fmt.Sprintf("# mvx Configuration\nmvxVersion=%s\n", version)
			return os.WriteFile(propertiesFile, []byte(content), 0644)
		}
	}

	// Read existing properties file
	content, err := os.ReadFile(propertiesFile)
	if err != nil {
		return fmt.Errorf("failed to read properties file: %w", err)
	}

	// Update the mvxVersion line
	lines := strings.Split(string(content), "\n")
	updated := false
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "mvxVersion=") {
			lines[i] = fmt.Sprintf("mvxVersion=%s", version)
			updated = true
			break
		}
	}

	// If mvxVersion line wasn't found, add it
	if !updated {
		lines = append(lines, fmt.Sprintf("mvxVersion=%s", version))
	}

	// Write back the updated content
	updatedContent := strings.Join(lines, "\n")
	return os.WriteFile(propertiesFile, []byte(updatedContent), 0644)
}

// updateBootstrap performs the bootstrap update
func updateBootstrap() error {
	printInfo("🔍 Checking for mvx bootstrap updates...")

	// Get latest release
	release, err := getLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to get latest release: %w", err)
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	printVerbose("Latest version: %s", latestVersion)

	// Get current version
	currentVersion, err := getCurrentVersion()
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if currentVersion == "" {
		printInfo("No current version found, will update to latest")
	} else {
		printVerbose("Current version: %s", currentVersion)
		if currentVersion == latestVersion {
			printInfo("✅ Bootstrap scripts are already up to date (version %s)", currentVersion)
			return nil
		}
		printInfo("📦 Update available: %s → %s", currentVersion, latestVersion)
	}

	if checkOnly {
		if currentVersion != latestVersion {
			printInfo("🆕 New version available: %s", latestVersion)
			printInfo("Run 'mvx update-bootstrap' to update")
		}
		return nil
	}

	// Perform the update
	printInfo("⬇️  Downloading bootstrap scripts...")

	baseURL := fmt.Sprintf("https://raw.githubusercontent.com/gnodet/mvx/%s", release.TagName)

	// Download mvx script
	if err := downloadFile(baseURL+"/mvx", "mvx"); err != nil {
		return fmt.Errorf("failed to download mvx script: %w", err)
	}

	// Make mvx executable
	if err := os.Chmod("mvx", 0755); err != nil {
		return fmt.Errorf("failed to make mvx executable: %w", err)
	}

	// Download mvx.cmd script
	if err := downloadFile(baseURL+"/mvx.cmd", "mvx.cmd"); err != nil {
		return fmt.Errorf("failed to download mvx.cmd script: %w", err)
	}

	// Create .mvx directory if it doesn't exist
	mvxDir := ".mvx"
	if err := os.MkdirAll(mvxDir, 0755); err != nil {
		return fmt.Errorf("failed to create .mvx directory: %w", err)
	}

	// Update mvx.properties with new version
	propertiesFile := filepath.Join(mvxDir, "mvx.properties")
	if err := updatePropertiesFile(propertiesFile, latestVersion, baseURL); err != nil {
		return fmt.Errorf("failed to update properties file: %w", err)
	}

	printInfo("✅ Bootstrap scripts updated successfully to version %s", latestVersion)
	printInfo("📝 Files updated:")
	printInfo("  - mvx (Unix/Linux/macOS script)")
	printInfo("  - mvx.cmd (Windows script)")
	printInfo("  - .mvx/mvx.properties (version specification)")

	return nil
}
