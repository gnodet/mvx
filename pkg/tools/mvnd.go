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

// MvndTool implements Tool interface for Maven Daemon management
type MvndTool struct {
	manager *Manager
}

// Name returns the tool name
func (m *MvndTool) Name() string {
	return "mvnd"
}

// Install downloads and installs the specified mvnd version
func (m *MvndTool) Install(version string, cfg config.ToolConfig) error {
	installDir := m.manager.GetToolVersionDir("mvnd", version, "")

	// Create installation directory
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}

	// Get download URL
	downloadURL := m.getDownloadURL(version)

	// Download and extract
	fmt.Printf("  ⏳ Downloading Maven Daemon %s...\n", version)
	if err := m.downloadAndExtract(downloadURL, installDir, version, cfg); err != nil {
		return fmt.Errorf("failed to download and extract: %w", err)
	}

	return nil
}

// IsInstalled checks if the specified version is installed
func (m *MvndTool) IsInstalled(version string, cfg config.ToolConfig) bool {
	installDir := m.manager.GetToolVersionDir("mvnd", version, "")

	// Check if mvnd exists in any subdirectory (mvnd archives have nested structure)
	return m.findMvndExecutable(installDir) != ""
}

// GetPath returns the installation path for the specified version
func (m *MvndTool) GetPath(version string, cfg config.ToolConfig) (string, error) {
	installDir := m.manager.GetToolVersionDir("mvnd", version, "")

	// mvnd archives typically extract to maven-mvnd-{version}-{platform}/
	entries, err := os.ReadDir(installDir)
	if err != nil {
		return "", fmt.Errorf("failed to read installation directory: %w", err)
	}

	// Look for maven-mvnd-* directory
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "maven-mvnd-") {
			subPath := filepath.Join(installDir, entry.Name())
			mvndExe := filepath.Join(subPath, "bin", "mvnd")
			if runtime.GOOS == "windows" {
				mvndExe += ".cmd"
			}
			if _, err := os.Stat(mvndExe); err == nil {
				return subPath, nil
			}
		}
	}

	return installDir, nil
}

// GetBinPath returns the binary path for the specified version
func (m *MvndTool) GetBinPath(version string, cfg config.ToolConfig) (string, error) {
	mvndHome, err := m.GetPath(version, cfg)
	if err != nil {
		return "", err
	}
	return filepath.Join(mvndHome, "bin"), nil
}

// Verify checks if the installation is working correctly
func (m *MvndTool) Verify(version string, cfg config.ToolConfig) error {
	binPath, err := m.GetBinPath(version, cfg)
	if err != nil {
		return err
	}

	mvndExe := filepath.Join(binPath, "mvnd")
	if runtime.GOOS == "windows" {
		mvndExe += ".cmd"
	}

	// Run mvnd --version to verify installation
	cmd := exec.Command(mvndExe, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mvnd verification failed: %w\nOutput: %s", err, output)
	}

	// Check if output contains expected version
	outputStr := string(output)
	if !strings.Contains(outputStr, version) {
		return fmt.Errorf("mvnd version mismatch: expected %s, got %s", version, outputStr)
	}

	return nil
}

// ListVersions returns available mvnd versions
func (m *MvndTool) ListVersions() ([]string, error) {
	registry := m.manager.GetRegistry()
	return registry.GetMvndVersions()
}

// getDownloadURL returns the download URL for the specified version
func (m *MvndTool) getDownloadURL(version string) string {
	// Determine platform-specific archive name
	platform := m.getPlatformString()

	// mvnd archives are in the Apache archive
	return fmt.Sprintf("https://archive.apache.org/dist/maven/mvnd/%s/maven-mvnd-%s-%s.zip", version, version, platform)
}

// getPlatformString returns the platform string for mvnd downloads
func (m *MvndTool) getPlatformString() string {
	switch runtime.GOOS {
	case "linux":
		if runtime.GOARCH == "arm64" {
			return "linux-aarch64"
		}
		return "linux-amd64"
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return "darwin-aarch64"
		}
		return "darwin-amd64"
	case "windows":
		return "windows-amd64"
	default:
		return "linux-amd64" // fallback
	}
}

// downloadAndExtract downloads and extracts a zip file with checksum verification
func (m *MvndTool) downloadAndExtract(url, destDir, version string, cfg config.ToolConfig) error {
	// Create temporary file for download
	tmpFile, err := os.CreateTemp("", "mvnd-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Configure robust download with checksum verification
	config := DefaultDownloadConfig(url, tmpFile.Name())
	config.ExpectedType = "application" // Accept various application types
	config.MinSize = 10 * 1024 * 1024   // Minimum 10MB for Maven Daemon distributions
	config.MaxSize = 100 * 1024 * 1024  // Maximum 100MB for Maven Daemon distributions
	config.ToolName = "mvnd"            // For progress reporting
	config.Version = version            // For checksum verification
	config.Config = cfg                 // Tool configuration
	config.ChecksumRegistry = m.manager.GetChecksumRegistry()

	// Perform robust download with checksum verification
	result, err := RobustDownload(config)
	if err != nil {
		return fmt.Errorf("Maven Daemon download failed: %s", DiagnoseDownloadError(url, err))
	}

	fmt.Printf("  📦 Downloaded %d bytes from %s\n", result.Size, result.FinalURL)

	// Close temp file before extraction
	tmpFile.Close()

	// Extract zip file
	return m.extractZip(tmpFile.Name(), destDir)
}

// extractZip extracts a zip file to the destination directory
func (m *MvndTool) extractZip(src, dest string) error {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	// Create destination directory
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	// Extract files
	for _, file := range reader.File {
		path := filepath.Join(dest, file.Name)

		// Security check: ensure path is within destination
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.FileInfo().Mode())
			continue
		}

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		// Extract file
		fileReader, err := file.Open()
		if err != nil {
			return err
		}

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.FileInfo().Mode())
		if err != nil {
			fileReader.Close()
			return err
		}

		_, err = io.Copy(targetFile, fileReader)
		fileReader.Close()
		targetFile.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// findMvndExecutable searches for mvnd executable in installation directory
func (m *MvndTool) findMvndExecutable(installDir string) string {
	mvndName := "mvnd"
	if runtime.GOOS == "windows" {
		mvndName = "mvnd.cmd"
	}

	// Walk through directory tree to find mvnd executable
	var mvndPath string
	filepath.Walk(installDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && info.Name() == mvndName {
			mvndPath = path
			return filepath.SkipDir
		}
		return nil
	})

	return mvndPath
}
