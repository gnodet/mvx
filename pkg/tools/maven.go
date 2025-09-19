package tools

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gnodet/mvx/pkg/config"
)

// MavenTool implements Tool interface for Maven management
type MavenTool struct {
	manager *Manager
}

// Name returns the tool name
func (m *MavenTool) Name() string {
	return "maven"
}

// useSystemMaven checks if system Maven should be used instead of downloading
func useSystemMaven() bool {
	return useSystemTool("maven")
}

// getSystemMavenDetector returns a system detector for Maven
func getSystemMavenDetector() SystemToolDetector {
	return &MavenSystemDetector{}
}

// Install downloads and installs the specified Maven version
func (m *MavenTool) Install(version string, cfg config.ToolConfig) error {
	installDir := m.manager.GetToolVersionDir("maven", version, "")

	// Check if we should use system Maven instead of downloading
	if useSystemMaven() {
		logVerbose("%s=true, attempting to use system Maven", getSystemToolEnvVar("maven"))

		detector := getSystemMavenDetector()
		systemMavenHome, err := detector.GetSystemHome()
		if err != nil {
			logVerbose("System Maven not available: %v", err)
			fmt.Printf("  ⚠️  System Maven not available (%v), falling back to download\n", err)
		} else {
			systemVersion, err := detector.GetSystemVersion(systemMavenHome)
			if err != nil {
				logVerbose("Could not determine system Maven version: %v", err)
				fmt.Printf("  ⚠️  Could not determine system Maven version (%v), falling back to download\n", err)
			} else if !detector.IsVersionCompatible(systemVersion, version) {
				logVerbose("System Maven version %s does not match requested version %s", systemVersion, version)
				fmt.Printf("  ⚠️  System Maven version %s does not match requested version %s, falling back to download\n", systemVersion, version)
			} else {
				// Use system Maven by creating a symlink
				fmt.Printf("  🔗 Using system Maven %s from %s\n", systemVersion, systemMavenHome)
				return detector.CreateSystemLink(systemMavenHome, installDir)
			}
		}
	}

	// Create installation directory
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}

	// Get download URL
	downloadURL := m.getDownloadURL(version)

	// Download and extract
	fmt.Printf("  ⏳ Downloading Maven %s...\n", version)
	if err := m.downloadAndExtract(downloadURL, installDir, version, cfg); err != nil {
		return fmt.Errorf("failed to download and extract: %w", err)
	}

	return nil
}

// IsInstalled checks if the specified version is installed
func (m *MavenTool) IsInstalled(version string, cfg config.ToolConfig) bool {
	// If using system Maven, check if system Maven is available and compatible
	if useSystemMaven() {
		detector := getSystemMavenDetector()
		if systemMavenHome, err := detector.GetSystemHome(); err == nil {
			if systemVersion, err := detector.GetSystemVersion(systemMavenHome); err == nil {
				if detector.IsVersionCompatible(systemVersion, version) {
					logVerbose("System Maven %s is available and compatible with requested version %s", systemVersion, version)
					return true
				}
			}
		}
		// If system Maven is not available or compatible, fall through to check downloaded version
	}

	installDir := m.manager.GetToolVersionDir("maven", version, "")
	mvnExe := filepath.Join(installDir, "bin", "mvn")
	if runtime.GOOS == "windows" {
		mvnExe += ".cmd"
	}

	// Check if mvn exists in any subdirectory (Maven archives have nested structure)
	return m.findMavenExecutable(installDir) != ""
}

// GetPath returns the installation path for the specified version
func (m *MavenTool) GetPath(version string, cfg config.ToolConfig) (string, error) {
	// If using system Maven, return system Maven home if available and compatible
	if useSystemMaven() {
		detector := getSystemMavenDetector()
		if systemMavenHome, err := detector.GetSystemHome(); err == nil {
			if systemVersion, err := detector.GetSystemVersion(systemMavenHome); err == nil {
				if detector.IsVersionCompatible(systemVersion, version) {
					logVerbose("Using system Maven %s from %s", systemVersion, systemMavenHome)
					return systemMavenHome, nil
				}
			}
		}
		// If system Maven is not available or compatible, fall through to check downloaded version
	}

	installDir := m.manager.GetToolVersionDir("maven", version, "")

	// Maven archives typically extract to apache-maven-{version}/
	entries, err := os.ReadDir(installDir)
	if err != nil {
		return "", fmt.Errorf("failed to read installation directory: %w", err)
	}

	// Look for apache-maven-* directory
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "apache-maven-") {
			subPath := filepath.Join(installDir, entry.Name())
			mvnExe := filepath.Join(subPath, "bin", "mvn")
			if runtime.GOOS == "windows" {
				mvnExe += ".cmd"
			}
			if _, err := os.Stat(mvnExe); err == nil {
				return subPath, nil
			}
		}
	}

	return installDir, nil
}

// GetBinPath returns the binary path for the specified version
func (m *MavenTool) GetBinPath(version string, cfg config.ToolConfig) (string, error) {
	mavenHome, err := m.GetPath(version, cfg)
	if err != nil {
		return "", err
	}
	return filepath.Join(mavenHome, "bin"), nil
}

// Verify checks if the installation is working correctly
func (m *MavenTool) Verify(version string, cfg config.ToolConfig) error {
	binPath, err := m.GetBinPath(version, cfg)
	if err != nil {
		return err
	}

	mvnExe := filepath.Join(binPath, "mvn")
	if runtime.GOOS == "windows" {
		mvnExe += ".cmd"
	}

	// Set up environment with JAVA_HOME if Java is available
	env := os.Environ()
	if _, err := m.manager.GetTool("java"); err == nil {
		// Find Java configuration from the manager's config
		// For now, we'll try to detect if Java is installed and use it
		if javaPath, err := m.findJavaHome(); err == nil {
			env = append(env, fmt.Sprintf("JAVA_HOME=%s", javaPath))
			logVerbose("Setting JAVA_HOME=%s for Maven verification", javaPath)
		} else {
			logVerbose("Could not find Java for Maven verification: %v", err)
		}
	}

	// Run mvn --version to verify installation
	cmd := exec.Command(mvnExe, "--version")
	cmd.Env = env
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("maven verification failed: %w\nOutput: %s", err, output)
	}

	// Check if output contains expected version
	outputStr := string(output)
	if !strings.Contains(outputStr, version) {
		return fmt.Errorf("maven version mismatch: expected %s, got %s", version, outputStr)
	}

	return nil
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

	return "", fmt.Errorf("java executable not found in %s", dir)
}

// ListVersions returns available versions for installation
func (m *MavenTool) ListVersions() ([]string, error) {
	// Use the registry to get versions
	registry := m.manager.GetRegistry()
	return registry.GetMavenVersions()
}

// getDownloadURL returns the download URL for the specified version
func (m *MavenTool) getDownloadURL(version string) string {
	// Use Apache archive for all Maven distributions
	if strings.HasPrefix(version, "4.") {
		// Maven 4.x versions are in the Maven 4 archive
		return fmt.Sprintf("https://archive.apache.org/dist/maven/maven-4/%s/binaries/apache-maven-%s-bin.zip", version, version)
	}

	// Maven 3.x versions are in the Maven 3 archive
	return fmt.Sprintf("https://archive.apache.org/dist/maven/maven-3/%s/binaries/apache-maven-%s-bin.zip", version, version)
}

// downloadAndExtract downloads and extracts a zip file with checksum verification
func (m *MavenTool) downloadAndExtract(url, destDir, version string, cfg config.ToolConfig) error {
	// Create temporary file for download
	tmpFile, err := os.CreateTemp("", "maven-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Configure robust download with checksum verification
	config := DefaultDownloadConfig(url, tmpFile.Name())
	config.ExpectedType = "application" // Accept various application types
	config.MinSize = 5 * 1024 * 1024    // Minimum 5MB for Maven distributions
	config.MaxSize = 50 * 1024 * 1024   // Maximum 50MB for Maven distributions
	config.ToolName = "maven"           // For progress reporting
	config.Version = version            // For checksum verification
	config.Config = cfg                 // Tool configuration
	config.ChecksumRegistry = m.manager.GetChecksumRegistry()

	// Perform robust download with checksum verification
	result, err := RobustDownload(config)
	if err != nil {
		return fmt.Errorf("Maven download failed: %s", DiagnoseDownloadError(url, err))
	}

	fmt.Printf("  📦 Downloaded %d bytes from %s\n", result.Size, result.FinalURL)

	// Close temp file before extraction
	tmpFile.Close()

	// Extract zip file
	return m.extractZip(tmpFile.Name(), destDir)
}

// extractZip extracts a zip file to the destination directory
func (m *MavenTool) extractZip(zipPath, destDir string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer reader.Close()

	// Extract files
	for _, file := range reader.File {
		targetPath := filepath.Join(destDir, file.Name)

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, file.FileInfo().Mode()); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", targetPath, err)
			}
			continue
		}

		if err := m.extractZipFile(file, targetPath); err != nil {
			return fmt.Errorf("failed to extract file %s: %w", targetPath, err)
		}
	}

	return nil
}

// extractZipFile extracts a single file from zip
func (m *MavenTool) extractZipFile(file *zip.File, targetPath string) error {
	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}

	// Open file in zip
	reader, err := file.Open()
	if err != nil {
		return err
	}
	defer reader.Close()

	// Create target file
	targetFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, file.FileInfo().Mode())
	if err != nil {
		return err
	}
	defer targetFile.Close()

	// Copy content
	_, err = io.Copy(targetFile, reader)
	return err
}

// findMavenExecutable searches for Maven executable in installation directory
func (m *MavenTool) findMavenExecutable(installDir string) string {
	mvnName := "mvn"
	if runtime.GOOS == "windows" {
		mvnName = "mvn.cmd"
	}

	// Walk through directory tree to find mvn executable
	var mvnPath string
	filepath.Walk(installDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && info.Name() == mvnName {
			mvnPath = path
			return filepath.SkipDir
		}
		return nil
	})

	return mvnPath
}
