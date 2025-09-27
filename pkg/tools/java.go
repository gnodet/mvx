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
var _ ChecksumProvider = (*JavaTool)(nil)

// getSystemJavaHome returns the system JAVA_HOME if available and valid
func getSystemJavaHome() (string, error) {
	javaHome := os.Getenv("JAVA_HOME")
	if javaHome == "" {
		return "", SystemToolError("java", fmt.Errorf("JAVA_HOME environment variable not set"))
	}

	// Check if JAVA_HOME points to a valid Java installation
	javaExe := filepath.Join(javaHome, "bin", "java")
	if runtime.GOOS == "windows" {
		javaExe += ".exe"
	}

	if _, err := os.Stat(javaExe); err != nil {
		return "", SystemToolError("java", fmt.Errorf("Java executable not found at %s: %w", javaExe, err))
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

	// Create a wrapper function that matches the expected signature
	getDownloadURLWrapper := func(v string) string {
		url, err := j.getDownloadURL(v, distribution)
		if err != nil {
			logVerbose("Failed to get download URL for Java %s (%s): %v", v, distribution, err)
			return ""
		}
		return url
	}

	// Use custom installation flow to handle distribution parameter
	return j.installWithDistribution(version, cfg, distribution, getDownloadURLWrapper)
}

// installWithDistribution provides Java-specific installation flow with distribution support
func (j *JavaTool) installWithDistribution(version string, cfg config.ToolConfig, distribution string, getDownloadURL func(string) string) error {
	// Check if we should use system tool instead of downloading
	if UseSystemTool(j.toolName) {
		logVerbose("%s=true, forcing use of system %s", getSystemToolEnvVar(j.toolName), j.toolName)
		return nil
	}

	// Create installation directory with distribution parameter
	installDir, err := j.CreateInstallDir(version, distribution)
	if err != nil {
		return InstallError(j.toolName, version, fmt.Errorf("failed to create install directory: %w", err))
	}

	// Get download URL and package ID for checksum
	downloadURL, packageID, err := j.getDownloadURLWithChecksum(version, distribution)
	if err != nil {
		return InstallError(j.toolName, version, fmt.Errorf("failed to get download URL: %w", err))
	}

	// Print download message
	j.PrintDownloadMessage(version)

	// Get checksum information if package ID is available
	var configWithChecksum config.ToolConfig = cfg
	if packageID != "" {
		if checksumInfo, err := j.getChecksumFromDiscoAPI(packageID); err == nil {
			// Add checksum to configuration for download verification
			configWithChecksum.Checksum = &config.ChecksumConfig{
				Type:  string(checksumInfo.Type),
				Value: checksumInfo.Value,
				URL:   checksumInfo.URL,
			}
			logVerbose("Added checksum to configuration: %s", checksumInfo.Value)
		} else {
			logVerbose("Failed to get checksum from Disco API: %v", err)
		}
	}

	// Get tool-specific download options
	options := j.getDownloadOptions()

	// Download the file with checksum verification
	archivePath, err := j.Download(downloadURL, version, configWithChecksum, options)
	if err != nil {
		return InstallError(j.toolName, version, err)
	}
	defer os.Remove(archivePath) // Clean up downloaded file

	// Extract the file
	if err := j.Extract(archivePath, installDir, options); err != nil {
		return InstallError(j.toolName, version, err)
	}

	// Verify installation
	if err := j.Verify(version, cfg); err != nil {
		// Clean up failed installation
		if removeErr := os.RemoveAll(installDir); removeErr != nil {
			logVerbose("Failed to clean up installation directory %s: %v", installDir, removeErr)
		}
		fmt.Printf("  🧹 Cleaning up failed installation directory...\n")
		return InstallError(j.toolName, version, fmt.Errorf("installation verification failed: %w", err))
	}

	fmt.Printf("  ✅ %s %s installation verification successful\n", j.toolName, version)
	return nil
}

// getDownloadURLWithChecksum returns download URL and package ID for checksum verification
func (j *JavaTool) getDownloadURLWithChecksum(version, distribution string) (string, string, error) {
	platformMapper := NewPlatformMapper()

	// Map Go arch to Disco API arch
	archMapping := map[string]string{
		"amd64": "x64",
		"arm64": "aarch64",
	}
	arch := platformMapper.MapArchitecture(archMapping)

	// Map OS names to Disco API format
	osMapping := map[string]string{
		"darwin": "macos",
	}
	osName := platformMapper.MapOS(osMapping)

	// Handle early access versions
	releaseStatus := "ga" // General Availability
	if strings.HasSuffix(version, "-ea") {
		releaseStatus = "ea" // Early Access
		version = strings.TrimSuffix(version, "-ea")
	}

	// Try primary distribution first
	result, err := j.tryDiscoDistributionWithChecksum(version, distribution, osName, arch, releaseStatus)
	if err == nil && result.DownloadURL != "" {
		return result.DownloadURL, result.PackageID, nil
	}

	// If primary distribution fails, try fallback distributions
	fallbackDistributions := []string{"temurin", "zulu", "microsoft", "corretto"}
	for _, fallback := range fallbackDistributions {
		if fallback == distribution {
			continue // Already tried this one
		}

		fmt.Printf("  🔄 Trying fallback distribution: %s\n", fallback)
		result, err := j.tryDiscoDistributionWithChecksum(version, fallback, osName, arch, releaseStatus)
		if err == nil && result.DownloadURL != "" {
			fmt.Printf("  ✅ Found Java %s in %s distribution\n", version, fallback)
			return result.DownloadURL, result.PackageID, nil
		}
	}

	return "", "", URLGenerationError("java", version, fmt.Errorf("Java %s not available in any supported distribution for %s/%s", version, osName, arch))
}

// IsInstalled checks if the specified version is installed
func (j *JavaTool) IsInstalled(version string, cfg config.ToolConfig) bool {
	// Use standardized installation check with Java-specific environment variables
	return j.StandardIsInstalledWithOptions(version, cfg, j.GetPath, "java", nil, []string{"JAVA_HOME"})
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
	if UseSystemTool("java") {
		if systemJavaHome, err := getSystemJavaHome(); err == nil {
			logVerbose("Using system Java from JAVA_HOME: %s (MVX_USE_SYSTEM_JAVA=true)", systemJavaHome)
			return systemJavaHome, nil
		} else {
			return "", EnvironmentError("java", version, fmt.Errorf("MVX_USE_SYSTEM_JAVA=true but system Java not available: %w", err))
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
	// Use standardized path resolution with Java-specific environment variables
	return j.StandardGetPathWithOptions(version, cfg, j.getInstalledPath, "java", nil, []string{"JAVA_HOME"})
}

// getInstalledPath returns the bin directory path for an installed Java version
func (j *JavaTool) getInstalledPath(version string, cfg config.ToolConfig) (string, error) {
	javaHome, err := j.GetJavaHome(version, cfg)
	if err != nil {
		return "", err
	}
	return filepath.Join(javaHome, "bin"), nil
}

// Verify checks if the installation is working correctly
func (j *JavaTool) Verify(version string, cfg config.ToolConfig) error {
	verifyConfig := VerificationConfig{
		BinaryName:      "java",
		VersionArgs:     []string{"-version"},
		ExpectedVersion: version,
		DebugInfo:       true, // Java needs detailed debug info
	}
	return j.StandardVerifyWithConfig(version, cfg, verifyConfig)
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
		MinSize:       10 * 1024 * 1024,  // 10MB (very permissive, rely on checksums)
		MaxSize:       800 * 1024 * 1024, // 800MB (generous upper bound)
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

	platformMapper := NewPlatformMapper()

	// Map Go arch to Disco API arch
	archMapping := map[string]string{
		"amd64": "x64",
		"arm64": "aarch64",
	}
	arch := platformMapper.MapArchitecture(archMapping)

	// Map OS names to Disco API format
	osMapping := map[string]string{
		"darwin": "macos",
	}
	osName := platformMapper.MapOS(osMapping)

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

	return "", URLGenerationError("java", version, fmt.Errorf("Java %s not available in any supported distribution for %s/%s", version, osName, arch))
}

// DiscoveryResult contains both download URL and package ID for checksum fetching
type DiscoveryResult struct {
	DownloadURL string
	PackageID   string
}

// tryDiscoDistribution attempts to get download URL from a specific distribution
func (j *JavaTool) tryDiscoDistribution(version, distribution, osName, arch, releaseStatus string) (string, error) {
	result, err := j.tryDiscoDistributionWithChecksum(version, distribution, osName, arch, releaseStatus)
	if err != nil {
		return "", err
	}
	return result.DownloadURL, nil
}

// tryDiscoDistributionWithChecksum attempts to get download URL and package ID from a specific distribution
func (j *JavaTool) tryDiscoDistributionWithChecksum(version, distribution, osName, arch, releaseStatus string) (DiscoveryResult, error) {
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
		return DiscoveryResult{}, fmt.Errorf("failed to query Disco API: %w", err)
	}
	defer resp.Body.Close()

	logVerbose("HTTP response status: %s", resp.Status)

	if resp.StatusCode != http.StatusOK {
		return DiscoveryResult{}, fmt.Errorf("Disco API request failed with status: %s", resp.Status)
	}

	var packages struct {
		Result []struct {
			ID                string `json:"id"`
			DirectDownloadURI string `json:"direct_download_uri"`
			Filename          string `json:"filename"`
			VersionNumber     string `json:"version_number"`
			LibCType          string `json:"lib_c_type"`
			Architecture      string `json:"architecture"`
			OperatingSystem   string `json:"operating_system"`
			ArchiveType       string `json:"archive_type"`
			Links             struct {
				PkgInfoURI          string `json:"pkg_info_uri"`
				PkgDownloadRedirect string `json:"pkg_download_redirect"`
			} `json:"links"`
		} `json:"result"`
	}

	// Read response body for debugging
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return DiscoveryResult{}, fmt.Errorf("failed to read response body: %w", err)
	}

	logVerbose("Raw API response: %s", string(body))

	if err := json.Unmarshal(body, &packages); err != nil {
		logVerbose("JSON parsing failed: %v", err)
		return DiscoveryResult{}, fmt.Errorf("failed to parse Disco API response: %w", err)
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
		return DiscoveryResult{}, fmt.Errorf("no packages found for Java %s (%s)", version, distribution)
	}

	// Define the package type for consistency
	type packageType struct {
		ID                string `json:"id"`
		DirectDownloadURI string `json:"direct_download_uri"`
		Filename          string `json:"filename"`
		VersionNumber     string `json:"version_number"`
		LibCType          string `json:"lib_c_type"`
		Architecture      string `json:"architecture"`
		OperatingSystem   string `json:"operating_system"`
		ArchiveType       string `json:"archive_type"`
		Links             struct {
			PkgInfoURI          string `json:"pkg_info_uri"`
			PkgDownloadRedirect string `json:"pkg_download_redirect"`
		} `json:"links"`
	}

	var selectedPkg *packageType

	// Smart selection: prefer glibc over musl on glibc systems, and tar.gz over other formats
	var glibcPkg, muslPkg, zipPkg, tarGzPkg, otherPkg *packageType

	for _, pkg := range packages.Result {
		pkgCopy := packageType{
			ID:                pkg.ID,
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
		return DiscoveryResult{}, fmt.Errorf("no suitable packages found for Java %s (%s)", version, distribution)
	}

	logVerbose("Selected package: %s", selectedPkg.Filename)
	downloadURL := selectedPkg.DirectDownloadURI
	if downloadURL == "" {
		downloadURL = selectedPkg.Links.PkgDownloadRedirect
	}

	if downloadURL == "" {
		return DiscoveryResult{}, fmt.Errorf("no download URL found for Java %s (%s)", version, distribution)
	}

	logVerbose("Selected download URL: %s", downloadURL)
	logVerbose("Package ID for checksum: %s", selectedPkg.ID)

	return DiscoveryResult{
		DownloadURL: downloadURL,
		PackageID:   selectedPkg.ID,
	}, nil
}

// getChecksumFromDiscoAPI fetches checksum information from Foojay Disco API
func (j *JavaTool) getChecksumFromDiscoAPI(packageID string) (ChecksumInfo, error) {
	if packageID == "" {
		return ChecksumInfo{}, fmt.Errorf("package ID is required")
	}

	// Build package info URL
	url := fmt.Sprintf("https://api.foojay.io/disco/v3.0/ids/%s", packageID)

	logVerbose("Fetching checksum from Disco API: %s", url)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return ChecksumInfo{}, fmt.Errorf("failed to fetch package info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ChecksumInfo{}, fmt.Errorf("Disco API returned status %d", resp.StatusCode)
	}

	var packageInfo struct {
		Result []struct {
			Filename          string `json:"filename"`
			Checksum          string `json:"checksum"`
			ChecksumType      string `json:"checksum_type"`
			ChecksumURI       string `json:"checksum_uri"`
			DirectDownloadURI string `json:"direct_download_uri"`
		} `json:"result"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&packageInfo); err != nil {
		return ChecksumInfo{}, fmt.Errorf("failed to decode package info: %w", err)
	}

	if len(packageInfo.Result) == 0 {
		return ChecksumInfo{}, fmt.Errorf("no package info found")
	}

	pkg := packageInfo.Result[0]
	if pkg.Checksum == "" {
		return ChecksumInfo{}, fmt.Errorf("no checksum available for package")
	}

	// Convert checksum type to our enum
	var checksumType ChecksumType
	switch strings.ToLower(pkg.ChecksumType) {
	case "sha256":
		checksumType = SHA256
	case "sha512":
		checksumType = SHA512
	default:
		checksumType = SHA256 // Default fallback
	}

	logVerbose("Found checksum: %s (%s)", pkg.Checksum, pkg.ChecksumType)

	return ChecksumInfo{
		Type:  checksumType,
		Value: pkg.Checksum,
		URL:   pkg.ChecksumURI,
	}, nil
}

// GetChecksum implements ChecksumProvider interface for Java
func (j *JavaTool) GetChecksum(version, filename string) (ChecksumInfo, error) {
	// Java checksums are provided via configuration from the Java tool
	// which extracts them from the Foojay Disco API during URL resolution
	fmt.Printf("  🔍 Checking for Java checksum from Disco API...\n")
	return ChecksumInfo{}, fmt.Errorf("Java checksums should be provided via configuration")
}

// ResolveVersion resolves a Java version specification to a concrete version
func (j *JavaTool) ResolveVersion(version, distribution string) (string, error) {
	registry := j.manager.GetRegistry()
	return registry.ResolveJavaVersion(version, distribution)
}

// GetDownloadURL implements Tool interface for Java
func (j *JavaTool) GetDownloadURL(version string) string {
	// Use default distribution (temurin) for URL generation
	url, err := j.getDownloadURL(version, "temurin")
	if err != nil {
		logVerbose("Failed to get download URL for Java %s: %v", version, err)
		return ""
	}
	return url
}

// GetChecksumURL implements Tool interface for Java
func (j *JavaTool) GetChecksumURL(version, filename string) string {
	// Java checksums are provided via Disco API package info
	return ""
}

// GetVersionsURL implements Tool interface for Java
func (j *JavaTool) GetVersionsURL() string {
	return FoojayDiscoAPIBase + "/major_versions?distribution=temurin&maintained=true"
}
