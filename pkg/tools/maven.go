package tools

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/gnodet/mvx/pkg/config"
)

// Compile-time interface validation
var _ Tool = (*MavenTool)(nil)

// MavenTool implements Tool interface for Maven management
type MavenTool struct {
	*BaseTool
}

// NewMavenTool creates a new Maven tool instance
func NewMavenTool(manager *Manager) *MavenTool {
	return &MavenTool{
		BaseTool: NewBaseTool(manager, "maven"),
	}
}

// Name returns the tool name
func (m *MavenTool) Name() string {
	return "maven"
}

// getSystemMavenHome returns the system MAVEN_HOME if available and valid
func getSystemMavenHome() (string, error) {
	// Try MAVEN_HOME first
	mavenHome := os.Getenv("MAVEN_HOME")
	if mavenHome != "" {
		mvnExe := filepath.Join(mavenHome, "bin", "mvn")
		if runtime.GOOS == "windows" {
			mvnExe += ".cmd"
		}
		if _, err := os.Stat(mvnExe); err == nil {
			return mavenHome, nil
		}
	}

	// Try to find mvn in PATH
	mvnExe := "mvn"
	if runtime.GOOS == "windows" {
		mvnExe = "mvn.cmd"
	}

	if mvnPath, err := exec.LookPath(mvnExe); err == nil {
		// Try to determine MAVEN_HOME from mvn path
		// mvn is typically at $MAVEN_HOME/bin/mvn
		binDir := filepath.Dir(mvnPath)
		mavenHome := filepath.Dir(binDir)

		// Verify this looks like a Maven installation
		if _, err := os.Stat(filepath.Join(mavenHome, "lib")); err == nil {
			return mavenHome, nil
		}
	}

	return "", SystemToolError("maven", fmt.Errorf("Maven not found in MAVEN_HOME or PATH"))
}

// getSystemMavenVersion returns the version of the system Maven installation
func getSystemMavenVersion(mavenHome string) (string, error) {
	mvnExe := filepath.Join(mavenHome, "bin", "mvn")
	if runtime.GOOS == "windows" {
		mvnExe += ".cmd"
	}

	cmd := exec.Command(mvnExe, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get Maven version: %w", err)
	}

	// Parse version from output (e.g., "Apache Maven 3.9.6 (bc0240f3c744dd6b6ec2920b3cd08dcc295161ae)")
	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("no version output from Maven")
	}

	// Look for "Apache Maven" in the first line
	versionLine := lines[0]
	if strings.Contains(versionLine, "Apache Maven") {
		parts := strings.Fields(versionLine)
		if len(parts) >= 3 {
			version := parts[2]
			// Remove ANSI escape sequences (e.g., "3.6.3\x1b[m" -> "3.6.3")
			ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
			version = ansiRegex.ReplaceAllString(version, "")
			// Also remove any remaining bracket sequences
			if idx := strings.Index(version, "["); idx != -1 {
				version = version[:idx]
			}
			return version, nil // "Apache Maven 3.9.6" -> "3.9.6"
		}
	}

	return "", fmt.Errorf("could not parse Maven version from: %s", versionLine)
}

// isMavenVersionCompatible checks if the system Maven version is compatible with the requested version
func isMavenVersionCompatible(systemVersion, requestedVersion string) bool {
	// For Maven, we can be more flexible - allow compatible versions
	// For now, exact match, but this could be enhanced to support semantic versioning
	return systemVersion == requestedVersion
}

// Install downloads and installs the specified Maven version
func (m *MavenTool) Install(version string, cfg config.ToolConfig) error {
	return m.installWithFallback(version, cfg)
}

// installWithFallback tries primary URL first, then fallback archive URL
func (m *MavenTool) installWithFallback(version string, cfg config.ToolConfig) error {
	// Check if we should use system tool instead of downloading (use standardized approach)
	if UseSystemTool("maven") {
		// Use standardized system tool detection
		return m.StandardInstallWithOptions(version, cfg, m.getDownloadURL, []string{"mvn.cmd"}, []string{"MAVEN_HOME"})
	}

	// Create installation directory
	installDir, err := m.CreateInstallDir(version, "")
	if err != nil {
		return InstallError("maven", version, fmt.Errorf("failed to create install directory: %w", err))
	}

	// Try both URLs with reduced retries instead of exhausting retries on first URL
	primaryURL := m.getDownloadURL(version)
	archiveURL := m.getArchiveDownloadURL(version)
	m.PrintDownloadMessage(version)

	options := m.GetDownloadOptions()

	// Try to download with alternating URLs and reduced retries per URL
	archivePath, err := m.downloadWithAlternatingURLs(primaryURL, archiveURL, version, cfg, options)
	if err != nil {
		return InstallError("maven", version, fmt.Errorf("download failed from both primary and archive URLs: %w", err))
	}
	defer os.Remove(archivePath) // Clean up downloaded file

	// Extract the downloaded file
	if err := m.Extract(archivePath, installDir, options); err != nil {
		return InstallError("maven", version, err)
	}

	// Verify installation
	if err := m.Verify(version, cfg); err != nil {
		// Installation verification failed, clean up the installation directory
		fmt.Printf("  ❌ Maven installation verification failed: %v\n", err)
		fmt.Printf("  🧹 Cleaning up failed installation directory...\n")
		if removeErr := os.RemoveAll(installDir); removeErr != nil {
			fmt.Printf("  ⚠️  Warning: failed to clean up installation directory: %v\n", removeErr)
		}
		return InstallError("maven", version, fmt.Errorf("installation verification failed: %w", err))
	}
	fmt.Printf("  ✅ Maven %s installation verification successful\n", version)

	return nil
}

// IsInstalled checks if the specified version is installed
func (m *MavenTool) IsInstalled(version string, cfg config.ToolConfig) bool {
	return m.StandardIsInstalledWithOptions(version, cfg, m.GetPath, "mvn", nil, []string{"MAVEN_HOME"})
}

// GetPath returns the binary path for the specified version (for PATH management)
func (m *MavenTool) GetPath(version string, cfg config.ToolConfig) (string, error) {
	return m.StandardGetPathWithOptions(version, cfg, m.getInstalledPath, "mvn", nil, []string{"MAVEN_HOME"})
}

// getInstalledPath returns the path for an installed Maven version
func (m *MavenTool) getInstalledPath(version string, cfg config.ToolConfig) (string, error) {
	installDir := m.manager.GetToolVersionDir("maven", version, "")
	pathResolver := NewPathResolver(m.manager.GetToolsDir())

	// Use robust directory walking approach
	options := DirectorySearchOptions{
		BinSubdirectory:           "bin",
		BinaryName:                "mvn",
		UsePlatformExtensions:     true,
		PreferredWindowsExtension: ".cmd", // Maven uses .cmd on Windows
		FallbackToParent:          false,  // Maven always has bin subdirectory
	}

	binPath, err := pathResolver.FindToolBinaryPath(installDir, options)
	if err != nil {
		// Final fallback to default bin directory
		return filepath.Join(installDir, "bin"), nil
	}

	return binPath, nil
}

// Verify checks if the installation is working correctly
func (m *MavenTool) Verify(version string, cfg config.ToolConfig) error {
	verifyConfig := VerificationConfig{
		BinaryName:  "mvn",
		VersionArgs: []string{"--version"},
		DebugInfo:   false,
	}
	return m.StandardVerifyWithConfig(version, cfg, verifyConfig)
}

// findJavaHome attempts to find an installed Java home directory
func (m *MavenTool) findJavaHome() (string, error) {
	// Check if JAVA_HOME is already set
	if javaHome := os.Getenv("JAVA_HOME"); javaHome != "" {
		return javaHome, nil
	}

	// Try to find Java installations in the tools directory
	javaToolsDir := filepath.Join(m.manager.GetToolsDir(), "java")
	if entries, err := os.ReadDir(javaToolsDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				javaVersionDir := filepath.Join(javaToolsDir, entry.Name())

				// Try to find Java executable in this version directory
				if javaPath, err := m.findJavaInDirectory(javaVersionDir); err == nil {
					logVerbose("Found Java installation at: %s", javaPath)
					return javaPath, nil
				}
			}
		}
	}

	return "", fmt.Errorf("no Java installation found")
}

// findJavaInDirectory searches for Java executable in a directory and returns JAVA_HOME
func (m *MavenTool) findJavaInDirectory(dir string) (string, error) {
	// Check if there are subdirectories (common with JDK archives)
	if entries, err := os.ReadDir(dir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				subPath := filepath.Join(dir, entry.Name())
				javaExe := filepath.Join(subPath, "bin", "java")
				if runtime.GOOS == "windows" {
					javaExe += ".exe"
				}
				if _, err := os.Stat(javaExe); err == nil {
					return subPath, nil
				}
			}
		}
	}

	// Also check if java is directly in the directory
	javaExe := filepath.Join(dir, "bin", "java")
	if runtime.GOOS == "windows" {
		javaExe += ".exe"
	}
	if _, err := os.Stat(javaExe); err == nil {
		return dir, nil
	}

	return "", PathError("java", "", fmt.Errorf("java executable not found in %s", dir))
}

// ListVersions returns available versions for installation
func (m *MavenTool) ListVersions() ([]string, error) {
	registry := m.manager.GetRegistry()
	return registry.GetMavenVersions()
}

// GetDownloadOptions returns download options specific to Maven
func (m *MavenTool) GetDownloadOptions() DownloadOptions {
	return DownloadOptions{
		FileExtension: ".zip",
		ExpectedType:  "application",
		MinSize:       5 * 1024 * 1024,   // 5MB
		MaxSize:       100 * 1024 * 1024, // 100MB
		ArchiveType:   "zip",
	}
}

// GetDisplayName returns the display name for Maven
func (m *MavenTool) GetDisplayName() string {
	return "Maven"
}

// getDownloadURL returns the download URL for the specified version
func (m *MavenTool) getDownloadURL(version string) string {
	// For recent releases, use dist.apache.org (CDN-backed)
	if strings.HasPrefix(version, "4.") {
		// Maven 4.x versions - try dist first for recent releases
		return fmt.Sprintf("https://dist.apache.org/repos/dist/release/maven/maven-4/%s/binaries/apache-maven-%s-bin.zip", version, version)
	}

	// Maven 3.x versions - try dist first for recent releases
	return fmt.Sprintf("https://dist.apache.org/repos/dist/release/maven/maven-3/%s/binaries/apache-maven-%s-bin.zip", version, version)
}

// getArchiveDownloadURL returns the fallback archive URL for the specified version
func (m *MavenTool) getArchiveDownloadURL(version string) string {
	if strings.HasPrefix(version, "4.") {
		// Maven 4.x versions are in the Maven 4 archive
		return fmt.Sprintf("https://archive.apache.org/dist/maven/maven-4/%s/binaries/apache-maven-%s-bin.zip", version, version)
	}

	// Maven 3.x versions are in the Maven 3 archive
	return fmt.Sprintf("https://archive.apache.org/dist/maven/maven-3/%s/binaries/apache-maven-%s-bin.zip", version, version)
}

// ResolveVersion resolves a Maven version specification to a concrete version
func (m *MavenTool) ResolveVersion(version, distribution string) (string, error) {
	registry := m.manager.GetRegistry()
	return registry.ResolveMavenVersion(version)
}

// GetDownloadURL implements URLProvider interface for Maven
func (m *MavenTool) GetDownloadURL(version string) string {
	return m.getDownloadURL(version)
}

// GetChecksumURL implements URLProvider interface for Maven
func (m *MavenTool) GetChecksumURL(version, filename string) string {
	if strings.HasPrefix(version, "4.") {
		return fmt.Sprintf("%s/maven-4/%s/binaries/%s.sha512",
			ApacheMavenBase, version, filename)
	}
	return fmt.Sprintf("%s/maven-3/%s/binaries/%s.sha512",
		ApacheMavenBase, version, filename)
}

// GetVersionsURL implements Tool interface for Maven
func (m *MavenTool) GetVersionsURL() string {
	return ApacheMavenBase + "/maven-3/"
}

// GetChecksum implements Tool interface for Maven
func (m *MavenTool) GetChecksum(version, filename string) (ChecksumInfo, error) {
	fmt.Printf("  🔍 Fetching Maven checksum from Apache archive...\n")

	checksumURL := m.GetChecksumURL(version, filename)
	if checksumURL == "" {
		return ChecksumInfo{}, fmt.Errorf("no checksum URL available for Maven %s", version)
	}

	// Fetch checksum from Apache archive
	checksum, err := m.fetchChecksumFromURL(checksumURL)
	if err != nil {
		fmt.Printf("  ⚠️  Failed to get Maven checksum: %v\n", err)
		return ChecksumInfo{}, err
	}

	fmt.Printf("  ✅ Found Maven checksum from Apache archive\n")
	return ChecksumInfo{
		Type:  SHA512,
		Value: checksum,
	}, nil
}

// fetchChecksumFromURL fetches checksum from a URL
func (m *MavenTool) fetchChecksumFromURL(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch checksum: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("checksum request returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read checksum response: %w", err)
	}

	// Apache checksum files contain just the checksum value
	checksum := strings.TrimSpace(string(body))
	if len(checksum) == 0 {
		return "", fmt.Errorf("empty checksum response")
	}

	return checksum, nil
}

// downloadWithAlternatingURLs tries downloading from both URLs with reduced retries per URL
// instead of exhausting all retries on the first URL before trying the second
func (m *MavenTool) downloadWithAlternatingURLs(primaryURL, archiveURL, version string, cfg config.ToolConfig, options DownloadOptions) (string, error) {
	urls := []struct {
		url  string
		name string
	}{
		{primaryURL, "primary"},
		{archiveURL, "archive"},
	}

	maxRetries := 3    // Total retries across both URLs
	retriesPerURL := 1 // Reduced retries per URL to allow trying both

	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		urlIndex := attempt % len(urls)
		currentURL := urls[urlIndex]

		if attempt > 0 && urlIndex == 0 {
			fmt.Printf("  🔄 Trying %s URL again (attempt %d)...\n", currentURL.name, (attempt/len(urls))+1)
		} else if urlIndex == 1 {
			fmt.Printf("  🔄 Switching to %s URL...\n", currentURL.name)
		}

		// Create download config with reduced retries per URL
		downloadConfig := DefaultDownloadConfig(currentURL.url, "")
		downloadConfig.MaxRetries = retriesPerURL
		downloadConfig.ExpectedType = options.ExpectedType
		downloadConfig.MinSize = options.MinSize
		downloadConfig.MaxSize = options.MaxSize
		downloadConfig.ToolName = "maven"
		downloadConfig.Version = version
		downloadConfig.Config = cfg
		downloadConfig.ChecksumRegistry = m.manager.GetChecksumRegistry()
		downloadConfig.Tool = m

		// Create temporary file for download
		tmpFile, err := os.CreateTemp("", fmt.Sprintf("maven-*%s", options.FileExtension))
		if err != nil {
			lastErr = fmt.Errorf("failed to create temporary file: %w", err)
			continue
		}
		downloadConfig.DestPath = tmpFile.Name()
		tmpFile.Close()

		_, err = RobustDownload(downloadConfig)
		if err == nil {
			fmt.Printf("  ✅ Successfully downloaded from %s URL\n", currentURL.name)
			return downloadConfig.DestPath, nil
		}

		lastErr = err
		fmt.Printf("  ⚠️  Download from %s URL failed: %v\n", currentURL.name, err)
		// Clean up failed download
		os.Remove(downloadConfig.DestPath)
	}

	return "", fmt.Errorf("all download attempts failed, last error: %w", lastErr)
}
