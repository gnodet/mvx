package tools

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gnodet/mvx/pkg/config"
)

// GradleTool manages Gradle installations
// Downloads from https://services.gradle.org/distributions/
type GradleTool struct {
	manager *Manager
}

func (g *GradleTool) Name() string { return "gradle" }

// Install downloads and installs the specified Gradle version
func (g *GradleTool) Install(version string, cfg config.ToolConfig) error {
	installDir := g.manager.GetToolVersionDir("gradle", version, "")

	// Create installation directory
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}

	// Get download URL
	downloadURL := g.getDownloadURL(version)

	// Download and extract
	fmt.Printf("  ⏳ Downloading Gradle %s...\n", version)
	if err := g.downloadAndExtract(downloadURL, installDir, version, cfg); err != nil {
		return fmt.Errorf("failed to download and extract: %w", err)
	}

	return nil
}

// IsInstalled checks if the specified version is installed
func (g *GradleTool) IsInstalled(version string, cfg config.ToolConfig) bool {
	installDir := g.manager.GetToolVersionDir("gradle", version, "")

	// Check if gradle executable exists in any subdirectory (Gradle archives have nested structure)
	return g.findGradleExecutable(installDir) != ""
}

// GetPath returns the installation path for the specified version
func (g *GradleTool) GetPath(version string, cfg config.ToolConfig) (string, error) {
	installDir := g.manager.GetToolVersionDir("gradle", version, "")

	// Find the actual Gradle home directory (gradle-{version})
	entries, err := os.ReadDir(installDir)
	if err != nil {
		return "", fmt.Errorf("failed to read installation directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "gradle-") {
			return filepath.Join(installDir, entry.Name()), nil
		}
	}

	return "", fmt.Errorf("gradle installation not found in %s", installDir)
}

// GetBinPath returns the binary path for the specified version
func (g *GradleTool) GetBinPath(version string, cfg config.ToolConfig) (string, error) {
	gradleHome, err := g.GetPath(version, cfg)
	if err != nil {
		return "", err
	}
	return filepath.Join(gradleHome, "bin"), nil
}

// Verify checks if the installation is working correctly
func (g *GradleTool) Verify(version string, cfg config.ToolConfig) error {
	binPath, err := g.GetBinPath(version, cfg)
	if err != nil {
		return err
	}

	gradleExe := filepath.Join(binPath, "gradle")
	if runtime.GOOS == "windows" {
		gradleExe += ".bat"
	}

	// Set up environment with JAVA_HOME if Java is available
	env := os.Environ()
	if javaPath, err := g.findJavaHome(); err == nil {
		env = append(env, fmt.Sprintf("JAVA_HOME=%s", javaPath))
		logVerbose("Setting JAVA_HOME=%s for Gradle verification", javaPath)
	} else {
		logVerbose("Could not find Java for Gradle verification: %v", err)
	}

	// Run gradle --version to verify installation
	cmd := exec.Command(gradleExe, "--version")
	cmd.Env = env
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gradle verification failed: %w\nOutput: %s", err, output)
	}

	// Check if output contains expected version
	outputStr := string(output)
	if !strings.Contains(outputStr, version) {
		return fmt.Errorf("gradle version mismatch: expected %s, got output: %s", version, outputStr)
	}

	logVerbose("Gradle %s verification successful", version)
	return nil
}

// ListVersions returns available versions for installation
func (g *GradleTool) ListVersions() ([]string, error) {
	return g.manager.registry.GetGradleVersions()
}

// getDownloadURL returns the download URL for the specified version
func (g *GradleTool) getDownloadURL(version string) string {
	return fmt.Sprintf("https://services.gradle.org/distributions/gradle-%s-bin.zip", version)
}

// findGradleExecutable searches for gradle executable in the installation directory
func (g *GradleTool) findGradleExecutable(installDir string) string {
	gradleExe := "gradle"
	if runtime.GOOS == "windows" {
		gradleExe = "gradle.bat"
	}

	// Walk through the installation directory to find gradle executable
	var foundPath string
	filepath.Walk(installDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && info.Name() == gradleExe {
			foundPath = path
			return filepath.SkipDir
		}
		return nil
	})

	return foundPath
}

// findJavaHome attempts to find Java installation for Gradle
func (g *GradleTool) findJavaHome() (string, error) {
	// Try to get Java from environment first
	if javaHome := os.Getenv("JAVA_HOME"); javaHome != "" {
		return javaHome, nil
	}

	// Try to find Java tool in manager
	javaTool, err := g.manager.GetTool("java")
	if err != nil {
		return "", fmt.Errorf("java tool not available: %w", err)
	}

	// We need to find the Java configuration from the manager's current context
	// For verification, we'll try to find any installed Java version
	// This is a simplified approach - in practice, we'd get the version from the current config

	// Try to find any Java installation in the tools directory
	javaToolsDir := filepath.Join(g.manager.cacheDir, "tools", "java")
	if _, err := os.Stat(javaToolsDir); os.IsNotExist(err) {
		return "", fmt.Errorf("no Java installations found in %s", javaToolsDir)
	}

	// Look for any Java version directory
	entries, err := os.ReadDir(javaToolsDir)
	if err != nil {
		return "", fmt.Errorf("failed to read Java tools directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Try to get the path for this Java version
			// We'll use a dummy config since we just need the path
			dummyConfig := config.ToolConfig{Version: entry.Name()}
			if javaPath, err := javaTool.GetPath(entry.Name(), dummyConfig); err == nil {
				return javaPath, nil
			}
		}
	}

	return "", fmt.Errorf("JAVA_HOME not set and no usable Java installation found")
}

// downloadAndExtract downloads and extracts Gradle
func (g *GradleTool) downloadAndExtract(url, destDir, version string, cfg config.ToolConfig) error {
	// Create temporary file for download
	tmpFile, err := os.CreateTemp("", "gradle-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	fmt.Printf("  ⏳ Downloading Gradle %s from %s...\n", version, url)

	// Configure robust download with checksum verification
	config := DefaultDownloadConfig(url, tmpFile.Name())
	config.ExpectedType = "application" // Accept various application types
	config.MinSize = 50 * 1024 * 1024   // Minimum 50MB for Gradle distributions
	config.MaxSize = 200 * 1024 * 1024  // Maximum 200MB for Gradle distributions
	config.ToolName = "gradle"          // For progress reporting
	config.Version = version            // For checksum verification
	config.Config = cfg                 // Tool configuration
	config.ChecksumRegistry = g.manager.GetChecksumRegistry()

	// Perform robust download with checksum verification
	result, err := RobustDownload(config)
	if err != nil {
		return fmt.Errorf("Gradle download failed: %s", DiagnoseDownloadError(url, err))
	}

	fmt.Printf("  📦 Downloaded %d bytes from %s\n", result.Size, result.FinalURL)

	// Close temp file before extraction
	tmpFile.Close()

	// Extract ZIP file
	if err := extractZipFile(tmpFile.Name(), destDir); err != nil {
		return fmt.Errorf("failed to extract Gradle archive: %w", err)
	}

	fmt.Printf("  ✅ Gradle %s extracted successfully\n", version)
	return nil
}
