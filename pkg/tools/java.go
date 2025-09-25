package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gnodet/mvx/pkg/config"
)

// Compile-time interface validation
var _ Tool = (*JavaTool)(nil)

// useSystemJava checks if system Java should be used instead of downloading
func useSystemJava() bool {
	return UseSystemTool("java")
}

// getSystemJavaHome returns the system JAVA_HOME if available and valid
func getSystemJavaHome() (string, error) {
	javaHome := os.Getenv("JAVA_HOME")
	if javaHome == "" {
		return "", fmt.Errorf("JAVA_HOME environment variable not set")
	}

	// Check if JAVA_HOME points to a valid Java installation
	javaExe := filepath.Join(javaHome, "bin", "java")
	if runtime.GOOS == "windows" {
		javaExe += ".exe"
	}

	if _, err := os.Stat(javaExe); err != nil {
		return "", fmt.Errorf("Java executable not found at %s", javaExe)
	}

	return javaHome, nil
}

// getSystemJavaVersion returns the version of the system Java installation
func getSystemJavaVersion(javaHome string) (string, error) {
	javaExe := filepath.Join(javaHome, "bin", "java")
	if runtime.GOOS == "windows" {
		javaExe += ".exe"
	}

	cmd := exec.Command(javaExe, "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get Java version: %w", err)
	}

	// Parse version from output (e.g., "openjdk version "21.0.1" 2023-10-17")
	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("no version output from Java")
	}

	// Look for version pattern in first line
	versionLine := lines[0]
	// Extract version number (handles both old and new format)
	if strings.Contains(versionLine, "\"") {
		start := strings.Index(versionLine, "\"")
		if start != -1 {
			end := strings.Index(versionLine[start+1:], "\"")
			if end != -1 {
				version := versionLine[start+1 : start+1+end]
				// Extract major version (e.g., "21.0.1" -> "21", "1.8.0_391" -> "8")
				if strings.HasPrefix(version, "1.") {
					// Old format (Java 8 and below): "1.8.0_391" -> "8"
					parts := strings.Split(version, ".")
					if len(parts) >= 2 {
						return parts[1], nil
					}
				} else {
					// New format (Java 9+): "21.0.1" -> "21"
					parts := strings.Split(version, ".")
					if len(parts) >= 1 {
						return parts[0], nil
					}
				}
			}
		}
	}

	return "", fmt.Errorf("could not parse Java version from: %s", versionLine)
}

// isJavaVersionCompatible checks if the system Java version is compatible with the requested version
func isJavaVersionCompatible(systemVersion, requestedVersion string) bool {
	// For now, we require exact major version match
	// This could be made more flexible in the future
	return systemVersion == requestedVersion
}

// JavaTool implements Tool interface for Java/JDK management
type JavaTool struct {
	*BaseTool
}

// NewJavaTool creates a new Java tool instance
func NewJavaTool(manager *Manager) *JavaTool {
	return &JavaTool{
		BaseTool: NewBaseTool(manager, "java"),
	}
}

// Name returns the tool name
func (j *JavaTool) Name() string {
	return "java"
}

// Install downloads and installs the specified Java version
func (j *JavaTool) Install(version string, cfg config.ToolConfig) error {
	distribution := cfg.Distribution
	if distribution == "" {
		distribution = "temurin" // Default to Eclipse Temurin
	}

	installDir := j.manager.GetToolVersionDir("java", version, distribution)

	// Check if we should use system Java instead of downloading
	if useSystemJava() {
		logVerbose("%s=true, forcing use of system Java", getSystemToolEnvVar("java"))

		systemJavaHome, err := getSystemJavaHome()
		if err != nil {
			return fmt.Errorf("MVX_USE_SYSTEM_JAVA=true but system Java not available: %v", err)
		}

		systemVersion, err := getSystemJavaVersion(systemJavaHome)
		if err != nil {
			logVerbose("Could not determine system Java version: %v", err)
			fmt.Printf("  🔗 Using system Java from JAVA_HOME: %s (version detection failed)\n", systemJavaHome)
		} else {
			fmt.Printf("  🔗 Using system Java %s from JAVA_HOME: %s\n", systemVersion, systemJavaHome)
		}

		fmt.Printf("  ✅ System Java configured (mvx will use system PATH)\n")
		return nil
	}

	// Create installation directory
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}

	// Get download URL
	downloadURL, err := j.getDownloadURL(version, distribution)
	if err != nil {
		return fmt.Errorf("failed to get download URL: %w", err)
	}

	// Download and extract
	fmt.Printf("  ⏳ Downloading Java %s (%s)...\n", version, distribution)
	options := j.GetDownloadOptions()
	if err := j.DownloadAndExtract(downloadURL, installDir, version, cfg, options); err != nil {
		return fmt.Errorf("failed to download and extract: %w", err)
	}

	return nil
}

// IsInstalled checks if the specified version is installed
func (j *JavaTool) IsInstalled(version string, cfg config.ToolConfig) bool {
	// Check if we should use system Java instead of mvx-managed Java
	if useSystemJava() {
		systemJavaHome, err := getSystemJavaHome()
		if err != nil {
			logVerbose("System Java not available: %v", err)
			return false
		}

		// For system Java, we need to check if the version matches
		systemVersion, err := getSystemJavaVersion(systemJavaHome)
		if err != nil {
			logVerbose("Could not determine system Java version: %v", err)
			return false
		}

		// Check if system Java version is compatible with requested version
		if !isJavaVersionCompatible(systemVersion, version) {
			logVerbose("System Java version %s is not compatible with requested version %s", systemVersion, version)
			return false
		}

		logVerbose("System Java version %s is compatible with requested version %s", systemVersion, version)
		return true
	}

	// Use standard mvx-managed installation check
	return j.StandardIsInstalled(version, cfg, j.GetPath, "java")
}

// isVersionCompatible checks if the system Java version is compatible with the requested version
func (j *JavaTool) isVersionCompatible(systemVersion, requestedVersion string) bool {
	// Extract major version numbers for comparison
	systemMajor := extractMajorVersion(systemVersion)
	requestedMajor := extractMajorVersion(requestedVersion)

	// For Java, we require exact major version match
	return systemMajor == requestedMajor
}

// extractMajorVersion extracts the major version number from a Java version string
func extractMajorVersion(version string) string {
	// Handle different Java version formats:
	// - "11.0.16" -> "11"
	// - "17.0.2" -> "17"
	// - "1.8.0_345" -> "8"
	// - "21" -> "21"

	// Remove any leading/trailing whitespace
	version = strings.TrimSpace(version)

	// Handle legacy format (1.x.y -> x)
	if strings.HasPrefix(version, "1.") {
		parts := strings.Split(version, ".")
		if len(parts) >= 2 {
			return parts[1]
		}
	}

	// Handle modern format (x.y.z -> x)
	parts := strings.Split(version, ".")
	if len(parts) > 0 {
		return parts[0]
	}

	return version
}

// GetJavaHome returns the JAVA_HOME path for the specified version
func (j *JavaTool) GetJavaHome(version string, cfg config.ToolConfig) (string, error) {
	distribution := cfg.Distribution
	if distribution == "" {
		distribution = "temurin"
	}

	// If using system Java, return system JAVA_HOME if available (no version compatibility check)
	if useSystemJava() {
		if systemJavaHome, err := getSystemJavaHome(); err == nil {
			logVerbose("Using system Java from JAVA_HOME: %s (MVX_USE_SYSTEM_JAVA=true)", systemJavaHome)
			return systemJavaHome, nil
		} else {
			return "", fmt.Errorf("MVX_USE_SYSTEM_JAVA=true but system Java not available: %v", err)
		}
	}

	installDir := j.manager.GetToolVersionDir("java", version, distribution)

	// Check if installation directory exists
	if _, err := os.Stat(installDir); os.IsNotExist(err) {
		return "", fmt.Errorf("Java %s is not installed", version)
	}

	logVerbose("Checking Java installation in: %s", installDir)

	// Search for java executable recursively and determine JAVA_HOME from its location
	javaExePath, err := j.findJavaExecutable(installDir)
	if err != nil {
		logVerbose("Java executable not found in %s: %v", installDir, err)
		return "", fmt.Errorf("Java executable not found in %s", installDir)
	}

	// Determine JAVA_HOME from java executable path
	// java executable is typically at $JAVA_HOME/bin/java
	javaHome := filepath.Dir(filepath.Dir(javaExePath))
	logVerbose("Found Java executable at: %s, using JAVA_HOME: %s", javaExePath, javaHome)
	return javaHome, nil
}

// findJavaExecutable recursively searches for the java executable in the given directory
func (j *JavaTool) findJavaExecutable(dir string) (string, error) {
	javaExeName := "java"
	if runtime.GOOS == "windows" {
		javaExeName = "java.exe"
	}

	var foundPath string
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Continue walking even if we can't read a directory
		}

		// Check if this is the java executable we're looking for
		if !d.IsDir() && d.Name() == javaExeName {
			// Verify it's in a bin directory (to avoid false positives)
			if filepath.Base(filepath.Dir(path)) == "bin" {
				logVerbose("Found Java executable at: %s", path)
				foundPath = path
				return filepath.SkipAll // Stop walking once we find it
			}
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	if foundPath == "" {
		return "", fmt.Errorf("java executable not found")
	}

	return foundPath, nil
}

// GetPath returns the binary path for the specified version (for PATH management)
func (j *JavaTool) GetPath(version string, cfg config.ToolConfig) (string, error) {
	javaHome, err := j.GetJavaHome(version, cfg)
	if err != nil {
		return "", err
	}
	return filepath.Join(javaHome, "bin"), nil
}

// Verify checks if the installation is working correctly
func (j *JavaTool) Verify(version string, cfg config.ToolConfig) error {
	distribution := cfg.Distribution
	if distribution == "" {
		distribution = "temurin"
	}

	binPath, err := j.GetPath(version, cfg)
	if err != nil {
		// Provide detailed debugging information
		installDir := j.manager.GetToolVersionDir("java", version, distribution)
		fmt.Printf("  🔍 Debug: Java installation verification failed\n")
		fmt.Printf("     Install directory: %s\n", installDir)
		fmt.Printf("     Error getting bin path: %v\n", err)

		// List contents of install directory for debugging
		if entries, readErr := os.ReadDir(installDir); readErr == nil {
			fmt.Printf("     Install directory contents:\n")
			for _, entry := range entries {
				fmt.Printf("       - %s (dir: %t)\n", entry.Name(), entry.IsDir())
			}
		}

		return fmt.Errorf("installation verification failed for java %s: %w", version, err)
	}

	javaExe := filepath.Join(binPath, "java")
	if runtime.GOOS == "windows" {
		javaExe += ".exe"
	}

	// Check if java executable exists
	if _, err := os.Stat(javaExe); err != nil {
		return fmt.Errorf("java verification failed: java executable not found at %s: %w", javaExe, err)
	}

	// Run java -version to verify installation
	cmd := exec.Command(javaExe, "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("java verification failed: %w\nOutput: %s", err, output)
	}

	// Check if output contains expected version
	outputStr := string(output)
	if !strings.Contains(outputStr, version) {
		return fmt.Errorf("java version mismatch: expected %s, got %s", version, outputStr)
	}

	return nil
}

// ListVersions returns available versions for installation using Disco API
func (j *JavaTool) ListVersions() ([]string, error) {
	// Use the registry to get versions from Disco API
	registry := j.manager.GetRegistry()
	return registry.GetJavaVersions("temurin") // Default to Temurin
}

// GetDownloadOptions returns download options specific to Java
func (j *JavaTool) GetDownloadOptions() DownloadOptions {
	return DownloadOptions{
		FileExtension: ".tar.gz",
		ExpectedType:  "application",
		MinSize:       100 * 1024 * 1024, // 100MB
		MaxSize:       500 * 1024 * 1024, // 500MB
		ArchiveType:   "tar.gz",
	}
}

// GetDisplayName returns the display name for Java
func (j *JavaTool) GetDisplayName() string {
	return "Java"
}

// getDownloadURL returns the download URL for the specified version and distribution using Disco API
func (j *JavaTool) getDownloadURL(version, distribution string) (string, error) {
	return j.getDiscoURL(version, distribution)
}

// getDiscoURL returns the download URL using Foojay Disco API
func (j *JavaTool) getDiscoURL(version, distribution string) (string, error) {
	if distribution == "" {
		distribution = "temurin" // Default to Temurin
	}

	osName := runtime.GOOS
	arch := runtime.GOARCH

	// Map Go arch to Disco API arch
	switch arch {
	case "amd64":
		arch = "x64"
	case "arm64":
		arch = "aarch64"
	}

	// Map OS names to Disco API format
	switch osName {
	case "darwin":
		osName = "macos"
	case "windows":
		osName = "windows"
	case "linux":
		osName = "linux"
	}

	// Handle early access versions
	releaseStatus := "ga" // General Availability
	if strings.HasSuffix(version, "-ea") {
		releaseStatus = "ea" // Early Access
		version = strings.TrimSuffix(version, "-ea")
	}

	// Try primary distribution first
	downloadURL, err := j.tryDiscoDistribution(version, distribution, osName, arch, releaseStatus)
	if err == nil && downloadURL != "" {
		return downloadURL, nil
	}

	// If primary distribution fails, try fallback distributions
	fallbackDistributions := []string{"temurin", "zulu", "microsoft", "corretto"}
	for _, fallback := range fallbackDistributions {
		if fallback == distribution {
			continue // Already tried this one
		}

		fmt.Printf("  🔄 Trying fallback distribution: %s\n", fallback)
		downloadURL, err := j.tryDiscoDistribution(version, fallback, osName, arch, releaseStatus)
		if err == nil && downloadURL != "" {
			fmt.Printf("  ✅ Found Java %s in %s distribution\n", version, fallback)
			return downloadURL, nil
		}
	}

	return "", fmt.Errorf("Java %s not available in any supported distribution for %s/%s", version, osName, arch)
}

// tryDiscoDistribution attempts to get download URL from a specific distribution
func (j *JavaTool) tryDiscoDistribution(version, distribution, osName, arch, releaseStatus string) (string, error) {
	// Build Disco API URL for package search
	url := fmt.Sprintf("https://api.foojay.io/disco/v3.0/packages?version=%s&distribution=%s&operating_system=%s&architecture=%s&package_type=jdk&release_status=%s&latest=available",
		version, distribution, osName, arch, releaseStatus)

	// Add verbose logging for debugging
	logVerbose("Disco API URL: %s", url)
	logVerbose("Query parameters: version=%s, distribution=%s, os=%s, arch=%s, release_status=%s",
		version, distribution, osName, arch, releaseStatus)

	// Get package information
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		logVerbose("HTTP request failed: %v", err)
		return "", fmt.Errorf("failed to query Disco API: %w", err)
	}
	defer resp.Body.Close()

	logVerbose("HTTP response status: %s", resp.Status)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Disco API request failed with status: %s", resp.Status)
	}

	var packages struct {
		Result []struct {
			DirectDownloadURI string `json:"direct_download_uri"`
			Filename          string `json:"filename"`
			VersionNumber     string `json:"version_number"`
			LibCType          string `json:"lib_c_type"`
			Architecture      string `json:"architecture"`
			OperatingSystem   string `json:"operating_system"`
			ArchiveType       string `json:"archive_type"`
			Links             struct {
				PkgDownloadRedirect string `json:"pkg_download_redirect"`
			} `json:"links"`
		} `json:"result"`
	}

	// Read response body for debugging
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	logVerbose("Raw API response: %s", string(body))

	if err := json.Unmarshal(body, &packages); err != nil {
		logVerbose("JSON parsing failed: %v", err)
		return "", fmt.Errorf("failed to parse Disco API response: %w", err)
	}

	logVerbose("Found %d packages in response", len(packages.Result))
	for i, pkg := range packages.Result {
		downloadURI := pkg.DirectDownloadURI
		if downloadURI == "" {
			downloadURI = pkg.Links.PkgDownloadRedirect
		}
		logVerbose("Package %d: filename=%s, version=%s, download_uri=%s",
			i+1, pkg.Filename, pkg.VersionNumber, downloadURI)
	}

	if len(packages.Result) == 0 {
		return "", fmt.Errorf("no packages found for Java %s (%s)", version, distribution)
	}

	// Define the package type for consistency
	type packageType struct {
		DirectDownloadURI string `json:"direct_download_uri"`
		Filename          string `json:"filename"`
		VersionNumber     string `json:"version_number"`
		LibCType          string `json:"lib_c_type"`
		Architecture      string `json:"architecture"`
		OperatingSystem   string `json:"operating_system"`
		ArchiveType       string `json:"archive_type"`
		Links             struct {
			PkgDownloadRedirect string `json:"pkg_download_redirect"`
		} `json:"links"`
	}

	var selectedPkg *packageType

	// Smart selection: prefer glibc over musl on glibc systems, and tar.gz over other formats
	var glibcPkg, muslPkg, zipPkg, tarGzPkg, otherPkg *packageType

	for _, pkg := range packages.Result {
		pkgCopy := packageType{
			DirectDownloadURI: pkg.DirectDownloadURI,
			Filename:          pkg.Filename,
			VersionNumber:     pkg.VersionNumber,
			LibCType:          pkg.LibCType,
			Architecture:      pkg.Architecture,
			OperatingSystem:   pkg.OperatingSystem,
			ArchiveType:       pkg.ArchiveType,
			Links:             pkg.Links,
		}

		// Check architecture compatibility
		archMatch := false
		if pkg.Architecture == "x64" || pkg.Architecture == "amd64" {
			archMatch = true
		} else if pkg.Architecture == "aarch64" || pkg.Architecture == "arm64" {
			archMatch = true
		}

		if !archMatch {
			continue
		}

		// Linux-specific package selection with libc preference
		if pkg.OperatingSystem == "linux" && pkg.ArchiveType == "tar.gz" {
			if pkg.LibCType == "musl" {
				if muslPkg == nil {
					muslPkg = &pkgCopy
					logVerbose("Found musl candidate: %s", pkg.Filename)
				}
			} else if pkg.LibCType == "glibc" {
				if glibcPkg == nil {
					glibcPkg = &pkgCopy
					logVerbose("Found glibc candidate: %s", pkg.Filename)
				}
			}
		}

		// For all platforms: prefer tar.gz over zip (tar.gz is smaller and more standard)
		if pkg.ArchiveType == "tar.gz" && tarGzPkg == nil {
			tarGzPkg = &pkgCopy
			logVerbose("Found TAR.GZ candidate: %s", pkg.Filename)
		} else if pkg.ArchiveType == "zip" && zipPkg == nil {
			zipPkg = &pkgCopy
			logVerbose("Found ZIP candidate: %s", pkg.Filename)
		}

		// Keep track of any package as fallback
		if otherPkg == nil {
			otherPkg = &pkgCopy
		}
	}

	// Select the best package with priority order:
	// 1. glibc packages (Linux tar.gz with glibc)
	// 2. musl packages (Linux tar.gz with musl)
	// 3. tar.gz packages (all platforms - preferred for size/compatibility)
	// 4. zip packages (all platforms - fallback)
	// 5. other packages (final fallback)
	if glibcPkg != nil {
		selectedPkg = glibcPkg
		logVerbose("Selected glibc package: %s (lib_c_type: %s)", selectedPkg.Filename, selectedPkg.LibCType)
	} else if muslPkg != nil {
		selectedPkg = muslPkg
		logVerbose("Selected musl package: %s (lib_c_type: %s)", selectedPkg.Filename, selectedPkg.LibCType)
	} else if tarGzPkg != nil {
		selectedPkg = tarGzPkg
		logVerbose("Selected TAR.GZ package: %s", selectedPkg.Filename)
	} else if zipPkg != nil {
		selectedPkg = zipPkg
		logVerbose("Selected ZIP package: %s", selectedPkg.Filename)
	} else if otherPkg != nil {
		selectedPkg = otherPkg
		logVerbose("Selected fallback package: %s", selectedPkg.Filename)
	} else {
		return "", fmt.Errorf("no suitable packages found for Java %s (%s)", version, distribution)
	}

	logVerbose("Selected package: %s", selectedPkg.Filename)
	downloadURL := selectedPkg.DirectDownloadURI
	if downloadURL == "" {
		downloadURL = selectedPkg.Links.PkgDownloadRedirect
	}

	if downloadURL == "" {
		return "", fmt.Errorf("no download URL found for Java %s (%s)", version, distribution)
	}

	logVerbose("Selected download URL: %s", downloadURL)
	return downloadURL, nil
}
