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
		logVerbose("%s=true, forcing use of system Java", getSystemToolEnvVar("java"))

		detector := getSystemJavaDetector()
		systemJavaHome, err := detector.GetSystemHome()
		if err != nil {
			return fmt.Errorf("MVX_USE_SYSTEM_JAVA=true but system Java not available: %v", err)
		}

		systemVersion, err := detector.GetSystemVersion(systemJavaHome)
		if err != nil {
			logVerbose("Could not determine system Java version: %v", err)
			fmt.Printf("  🔗 Using system Java from JAVA_HOME: %s (version detection failed)\n", systemJavaHome)
		} else {
			fmt.Printf("  🔗 Using system Java %s from JAVA_HOME: %s\n", systemVersion, systemJavaHome)
		}

		return detector.CreateSystemLink(systemJavaHome, installDir)
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
	if err := j.downloadAndExtract(downloadURL, installDir, version, cfg); err != nil {
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

	// If using system Java, check if system Java is available (no version compatibility check)
	if useSystemJava() {
		detector := getSystemJavaDetector()
		if systemJavaHome, err := detector.GetSystemHome(); err == nil {
			logVerbose("System Java is available at %s (MVX_USE_SYSTEM_JAVA=true)", systemJavaHome)
			return true
		} else {
			logVerbose("System Java not available: %v", err)
			return false
		}
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

	// If using system Java, return system JAVA_HOME if available (no version compatibility check)
	if useSystemJava() {
		detector := getSystemJavaDetector()
		if systemJavaHome, err := detector.GetSystemHome(); err == nil {
			logVerbose("Using system Java from JAVA_HOME: %s (MVX_USE_SYSTEM_JAVA=true)", systemJavaHome)
			return systemJavaHome, nil
		} else {
			return "", fmt.Errorf("MVX_USE_SYSTEM_JAVA=true but system Java not available: %v", err)
		}
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
			logVerbose("Examining subdirectory: %s", subPath)

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

			// For Alpine Linux and some other distributions, check nested subdirectories
			if nestedEntries, err := os.ReadDir(subPath); err == nil {
				for _, nestedEntry := range nestedEntries {
					if nestedEntry.IsDir() {
						nestedPath := filepath.Join(subPath, nestedEntry.Name())
						nestedJavaExe := filepath.Join(nestedPath, "bin", "java")
						if runtime.GOOS == "windows" {
							nestedJavaExe += ".exe"
						}
						logVerbose("Checking for nested Java executable at: %s", nestedJavaExe)
						if _, err := os.Stat(nestedJavaExe); err == nil {
							logVerbose("Found Java executable in nested subdirectory: %s", nestedPath)
							return nestedPath, nil
						}
					}
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
	distribution := cfg.Distribution
	if distribution == "" {
		distribution = "temurin"
	}

	binPath, err := j.GetBinPath(version, cfg)
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
	var glibcPkg, muslPkg, otherPkg *packageType

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

		// Only consider tar.gz packages for Linux
		if pkg.ArchiveType == "tar.gz" && pkg.OperatingSystem == "linux" && pkg.Architecture == "x64" {
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

		// Keep track of any package as fallback
		if otherPkg == nil {
			otherPkg = &pkgCopy
		}
	}

	// Select the best package: prefer glibc, then musl, then any other
	if glibcPkg != nil {
		selectedPkg = glibcPkg
		logVerbose("Selected glibc package: %s (lib_c_type: %s)", selectedPkg.Filename, selectedPkg.LibCType)
	} else if muslPkg != nil {
		selectedPkg = muslPkg
		logVerbose("Selected musl package: %s (lib_c_type: %s)", selectedPkg.Filename, selectedPkg.LibCType)
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

// downloadAndExtract downloads and extracts Java archives with checksum verification
func (j *JavaTool) downloadAndExtract(url, destDir, version string, cfg config.ToolConfig) error {
	// Determine file extension from URL
	fileExt := j.getFileExtension(url)

	// Create temporary file for download with appropriate extension
	tmpFile, err := os.CreateTemp("", "java-*"+fileExt)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Configure robust download with checksum verification
	config := DefaultDownloadConfig(url, tmpFile.Name())
	config.ExpectedType = "application" // Accept various application types
	config.MinSize = 50 * 1024 * 1024   // Minimum 50MB for Java distributions
	config.MaxSize = 500 * 1024 * 1024  // Maximum 500MB for Java distributions
	config.ToolName = "java"            // For progress reporting
	config.Version = version            // For checksum verification
	config.Config = cfg                 // Tool configuration
	config.ChecksumRegistry = j.manager.GetChecksumRegistry()

	// Perform robust download with checksum verification
	result, err := RobustDownload(config)
	if err != nil {
		return fmt.Errorf("Java download failed: %s", DiagnoseDownloadError(url, err))
	}

	fmt.Printf("  📦 Downloaded %d bytes from %s\n", result.Size, result.FinalURL)

	// Close temp file before extraction
	tmpFile.Close()

	// Open the downloaded file for extraction
	file, err := os.Open(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("failed to open downloaded file: %w", err)
	}
	defer file.Close()

	// Extract the downloaded archive based on file type
	return j.extractArchive(tmpFile.Name(), destDir, url)
}

// extractTarGz extracts a tar.gz file from disk
func (j *JavaTool) extractTarGz(archivePath, destDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	// Create gzip reader
	gzReader, err := gzip.NewReader(file)
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

// getFileExtension determines the file extension from URL
func (j *JavaTool) getFileExtension(url string) string {
	if strings.HasSuffix(url, ".tar.gz") {
		return ".tar.gz"
	} else if strings.HasSuffix(url, ".tgz") {
		return ".tgz"
	} else if strings.HasSuffix(url, ".dmg") {
		return ".dmg"
	} else if strings.HasSuffix(url, ".zip") {
		return ".zip"
	} else if strings.HasSuffix(url, ".tar.xz") {
		return ".tar.xz"
	}
	// Default to .tar.gz for unknown extensions
	return ".tar.gz"
}

// extractArchive extracts an archive based on its file type
func (j *JavaTool) extractArchive(archivePath, destDir, url string) error {
	if strings.HasSuffix(url, ".dmg") {
		return j.extractDmg(archivePath, destDir)
	} else if strings.HasSuffix(url, ".zip") {
		return j.extractZip(archivePath, destDir)
	} else if strings.HasSuffix(url, ".tar.xz") {
		return j.extractTarXz(archivePath, destDir)
	} else {
		// Default to tar.gz extraction for .tar.gz, .tgz, and unknown formats
		return j.extractTarGz(archivePath, destDir)
	}
}

// extractDmg extracts a DMG file on macOS using hdiutil
func (j *JavaTool) extractDmg(dmgPath, destDir string) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("DMG extraction is only supported on macOS")
	}

	logVerbose("Extracting DMG file: %s to %s", dmgPath, destDir)

	// Mount the DMG file
	mountCmd := exec.Command("hdiutil", "attach", "-quiet", "-nobrowse", "-plist", dmgPath)
	mountOutput, err := mountCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to mount DMG: %w", err)
	}

	// Parse the plist output to get mount point
	mountPoint, err := j.parseMountPoint(mountOutput)
	if err != nil {
		return fmt.Errorf("failed to parse mount point: %w", err)
	}

	logVerbose("DMG mounted at: %s", mountPoint)

	// Ensure we unmount the DMG when done
	defer func() {
		logVerbose("Unmounting DMG from: %s", mountPoint)
		unmountCmd := exec.Command("hdiutil", "detach", "-quiet", mountPoint)
		if err := unmountCmd.Run(); err != nil {
			fmt.Printf("  ⚠️  Warning: failed to unmount DMG: %v\n", err)
		}
	}()

	// Find JDK contents in the mounted volume
	return j.copyJdkFromMount(mountPoint, destDir)
}

// parseMountPoint parses hdiutil plist output to extract mount point
func (j *JavaTool) parseMountPoint(plistData []byte) (string, error) {
	// Simple parsing of hdiutil plist output
	// Look for mount-point in the plist data
	plistStr := string(plistData)

	// Find the mount-point key and extract the path
	lines := strings.Split(plistStr, "\n")
	for i, line := range lines {
		if strings.Contains(line, "<key>mount-point</key>") && i+1 < len(lines) {
			nextLine := strings.TrimSpace(lines[i+1])
			if strings.HasPrefix(nextLine, "<string>") && strings.HasSuffix(nextLine, "</string>") {
				mountPoint := strings.TrimPrefix(nextLine, "<string>")
				mountPoint = strings.TrimSuffix(mountPoint, "</string>")
				return mountPoint, nil
			}
		}
	}

	return "", fmt.Errorf("could not find mount point in hdiutil output")
}

// copyJdkFromMount copies JDK contents from mounted DMG to destination
func (j *JavaTool) copyJdkFromMount(mountPoint, destDir string) error {
	logVerbose("Searching for JDK in mount point: %s", mountPoint)

	// Look for common JDK patterns in DMG
	// 1. Direct .jdk bundle (e.g., jdk-21.jdk)
	// 2. Contents/Home structure
	// 3. Nested directory structure

	entries, err := os.ReadDir(mountPoint)
	if err != nil {
		return fmt.Errorf("failed to read mount point: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			entryPath := filepath.Join(mountPoint, entry.Name())
			logVerbose("Checking directory: %s", entryPath)

			// Check if this looks like a JDK bundle (.jdk extension or Contents/Home structure)
			if strings.HasSuffix(entry.Name(), ".jdk") || j.hasJdkStructure(entryPath) {
				logVerbose("Found JDK structure at: %s", entryPath)
				return j.copyDirectory(entryPath, destDir)
			}
		}
	}

	return fmt.Errorf("no JDK structure found in DMG")
}

// hasJdkStructure checks if a directory has JDK structure (Contents/Home/bin/java)
func (j *JavaTool) hasJdkStructure(dirPath string) bool {
	javaExe := filepath.Join(dirPath, "Contents", "Home", "bin", "java")
	_, err := os.Stat(javaExe)
	return err == nil
}

// copyDirectory recursively copies a directory
func (j *JavaTool) copyDirectory(src, dst string) error {
	logVerbose("Copying directory from %s to %s", src, dst)

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return j.copyFile(path, dstPath, info.Mode())
	})
}

// copyFile copies a single file
func (j *JavaTool) copyFile(src, dst string, mode os.FileMode) error {
	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy content
	_, err = io.Copy(dstFile, srcFile)
	return err
}

// extractZip extracts a ZIP file
func (j *JavaTool) extractZip(zipPath, destDir string) error {
	// Import archive/zip at the top of the file if not already imported
	return fmt.Errorf("ZIP extraction for Java not yet implemented - please use tar.gz version")
}

// extractTarXz extracts a tar.xz file using system tar command
func (j *JavaTool) extractTarXz(archivePath, destDir string) error {
	// Create destination directory
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Use system tar to extract .tar.xz
	cmd := exec.Command("tar", "-xJf", archivePath, "-C", destDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract tar.xz: %w", err)
	}

	return nil
}
