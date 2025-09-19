package tools

import (
	"archive/tar"
	"compress/gzip"
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

// isVerbose checks if verbose logging is enabled
func isVerbose() bool {
	return os.Getenv("MVX_VERBOSE") == "true"
}

// logVerbose prints verbose log messages
func logVerbose(format string, args ...interface{}) {
	if isVerbose() {
		fmt.Printf("[VERBOSE] "+format+"\n", args...)
	}
}

// useSystemJava checks if system Java should be used instead of downloading
func useSystemJava() bool {
	return useSystemTool("java")
}

// getSystemJavaDetector returns a system detector for Java
func getSystemJavaDetector() SystemToolDetector {
	return &JavaSystemDetector{}
}

// JavaTool implements Tool interface for Java/JDK management
type JavaTool struct {
	manager *Manager
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
		logVerbose("%s=true, attempting to use system Java", getSystemToolEnvVar("java"))

		detector := getSystemJavaDetector()
		systemJavaHome, err := detector.GetSystemHome()
		if err != nil {
			logVerbose("System Java not available: %v", err)
			fmt.Printf("  ⚠️  System Java not available (%v), falling back to download\n", err)
		} else {
			systemVersion, err := detector.GetSystemVersion(systemJavaHome)
			if err != nil {
				logVerbose("Could not determine system Java version: %v", err)
				fmt.Printf("  ⚠️  Could not determine system Java version (%v), falling back to download\n", err)
			} else if !detector.IsVersionCompatible(systemVersion, version) {
				logVerbose("System Java version %s does not match requested version %s", systemVersion, version)
				fmt.Printf("  ⚠️  System Java version %s does not match requested version %s, falling back to download\n", systemVersion, version)
			} else {
				// Use system Java by creating a symlink
				fmt.Printf("  🔗 Using system Java %s from JAVA_HOME: %s\n", systemVersion, systemJavaHome)
				return detector.CreateSystemLink(systemJavaHome, installDir)
			}
		}
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
	if err := j.downloadAndExtract(downloadURL, installDir); err != nil {
		return fmt.Errorf("failed to download and extract: %w", err)
	}

	return nil
}

// IsInstalled checks if the specified version is installed
func (j *JavaTool) IsInstalled(version string, cfg config.ToolConfig) bool {
	distribution := cfg.Distribution
	if distribution == "" {
		distribution = "temurin"
	}

	// If using system Java, check if system Java is available and compatible
	if useSystemJava() {
		detector := getSystemJavaDetector()
		if systemJavaHome, err := detector.GetSystemHome(); err == nil {
			if systemVersion, err := detector.GetSystemVersion(systemJavaHome); err == nil {
				if detector.IsVersionCompatible(systemVersion, version) {
					logVerbose("System Java %s is available and compatible with requested version %s", systemVersion, version)
					return true
				}
			}
		}
		// If system Java is not available or compatible, fall through to check downloaded version
	}

	// Try to get the actual Java path (which handles nested directories)
	javaPath, err := j.GetPath(version, cfg)
	if err != nil {
		logVerbose("Java %s (%s) not installed: %v", version, distribution, err)
		return false
	}

	logVerbose("Java %s (%s) found at: %s", version, distribution, javaPath)
	return true
}

// GetPath returns the installation path for the specified version
func (j *JavaTool) GetPath(version string, cfg config.ToolConfig) (string, error) {
	distribution := cfg.Distribution
	if distribution == "" {
		distribution = "temurin"
	}

	// If using system Java, return system JAVA_HOME if available and compatible
	if useSystemJava() {
		detector := getSystemJavaDetector()
		if systemJavaHome, err := detector.GetSystemHome(); err == nil {
			if systemVersion, err := detector.GetSystemVersion(systemJavaHome); err == nil {
				if detector.IsVersionCompatible(systemVersion, version) {
					logVerbose("Using system Java %s from JAVA_HOME: %s", systemVersion, systemJavaHome)
					return systemJavaHome, nil
				}
			}
		}
		// If system Java is not available or compatible, fall through to check downloaded version
	}

	installDir := j.manager.GetToolVersionDir("java", version, distribution)
	logVerbose("Checking Java installation in: %s", installDir)

	// Check if there's a nested directory (common with JDK archives)
	entries, err := os.ReadDir(installDir)
	if err != nil {
		logVerbose("Failed to read installation directory %s: %v", installDir, err)
		return "", fmt.Errorf("failed to read installation directory: %w", err)
	}

	logVerbose("Found %d entries in installation directory", len(entries))

	// Look for a subdirectory that looks like a JDK
	for _, entry := range entries {
		if entry.IsDir() {
			subPath := filepath.Join(installDir, entry.Name())

			// Check standard location first
			javaExe := filepath.Join(subPath, "bin", "java")
			if runtime.GOOS == "windows" {
				javaExe += ".exe"
			}
			logVerbose("Checking for Java executable at: %s", javaExe)
			if _, err := os.Stat(javaExe); err == nil {
				logVerbose("Found Java executable in subdirectory: %s", subPath)
				return subPath, nil
			}

			// On macOS, also check Contents/Home/bin/java (common with JDK packages)
			if runtime.GOOS == "darwin" {
				macOSJavaExe := filepath.Join(subPath, "Contents", "Home", "bin", "java")
				logVerbose("Checking for macOS Java executable at: %s", macOSJavaExe)
				if _, err := os.Stat(macOSJavaExe); err == nil {
					macOSJavaHome := filepath.Join(subPath, "Contents", "Home")
					logVerbose("Found macOS Java executable, using JAVA_HOME: %s", macOSJavaHome)
					return macOSJavaHome, nil
				}
			}
		}
	}

	// Also check if java is directly in the install directory (some distributions)
	javaExe := filepath.Join(installDir, "bin", "java")
	if runtime.GOOS == "windows" {
		javaExe += ".exe"
	}
	logVerbose("Checking for Java executable directly at: %s", javaExe)
	if _, err := os.Stat(javaExe); err == nil {
		logVerbose("Found Java executable directly in install directory: %s", installDir)
		return installDir, nil
	}

	// On macOS, also check Contents/Home/bin/java directly in install directory
	if runtime.GOOS == "darwin" {
		macOSJavaExe := filepath.Join(installDir, "Contents", "Home", "bin", "java")
		logVerbose("Checking for macOS Java executable directly at: %s", macOSJavaExe)
		if _, err := os.Stat(macOSJavaExe); err == nil {
			macOSJavaHome := filepath.Join(installDir, "Contents", "Home")
			logVerbose("Found macOS Java executable directly, using JAVA_HOME: %s", macOSJavaHome)
			return macOSJavaHome, nil
		}
	}

	logVerbose("Java executable not found anywhere in %s", installDir)
	return "", fmt.Errorf("Java executable not found in %s", installDir)
}

// GetBinPath returns the binary path for the specified version
func (j *JavaTool) GetBinPath(version string, cfg config.ToolConfig) (string, error) {
	javaHome, err := j.GetPath(version, cfg)
	if err != nil {
		return "", err
	}
	return filepath.Join(javaHome, "bin"), nil
}

// Verify checks if the installation is working correctly
func (j *JavaTool) Verify(version string, cfg config.ToolConfig) error {
	binPath, err := j.GetBinPath(version, cfg)
	if err != nil {
		return err
	}

	javaExe := filepath.Join(binPath, "java")
	if runtime.GOOS == "windows" {
		javaExe += ".exe"
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

	// Prefer tar.gz files over other formats (pkg, dmg, zip)
	var selectedPkg *struct {
		DirectDownloadURI string `json:"direct_download_uri"`
		Filename          string `json:"filename"`
		VersionNumber     string `json:"version_number"`
		Links             struct {
			PkgDownloadRedirect string `json:"pkg_download_redirect"`
		} `json:"links"`
	}

	// First, look for tar.gz files
	for _, pkg := range packages.Result {
		if strings.HasSuffix(pkg.Filename, ".tar.gz") {
			selectedPkg = &pkg
			break
		}
	}

	// If no tar.gz found, use the first available package
	if selectedPkg == nil {
		selectedPkg = &packages.Result[0]
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

// downloadAndExtract downloads and extracts a tar.gz file
func (j *JavaTool) downloadAndExtract(url, destDir string) error {
	// Download file with timeout
	client := &http.Client{
		Timeout: 300 * time.Second, // 5 minute timeout
	}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	// Create gzip reader
	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	// Create tar reader
	tarReader := tar.NewReader(gzReader)

	// Extract files
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar entry: %w", err)
		}

		targetPath := filepath.Join(destDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", targetPath, err)
			}
		case tar.TypeReg:
			// Ensure we have write permissions for the file
			mode := os.FileMode(header.Mode)
			if mode&0200 == 0 {
				mode |= 0200 // Add write permission for owner
			}
			if err := j.extractFile(tarReader, targetPath, mode); err != nil {
				return fmt.Errorf("failed to extract file %s: %w", targetPath, err)
			}
		}
	}

	return nil
}

// extractFile extracts a single file from tar reader
func (j *JavaTool) extractFile(tarReader *tar.Reader, targetPath string, mode os.FileMode) error {
	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}

	// Create file
	file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer file.Close()

	// Copy content
	_, err = io.Copy(file, tarReader)
	return err
}
