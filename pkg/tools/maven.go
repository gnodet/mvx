package tools

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
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

// Install downloads and installs the specified Maven version
func (m *MavenTool) Install(version string, cfg config.ToolConfig) error {
	installDir := m.manager.GetToolVersionDir("maven", version, "")
	
	// Create installation directory
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}
	
	// Get download URL
	downloadURL := m.getDownloadURL(version)
	
	// Download and extract
	fmt.Printf("  ⏳ Downloading Maven %s...\n", version)
	if err := m.downloadAndExtract(downloadURL, installDir); err != nil {
		return fmt.Errorf("failed to download and extract: %w", err)
	}
	
	return nil
}

// IsInstalled checks if the specified version is installed
func (m *MavenTool) IsInstalled(version string, cfg config.ToolConfig) bool {
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
	
	// Run mvn --version to verify installation
	cmd := exec.Command(mvnExe, "--version")
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

// ListVersions returns available versions for installation
func (m *MavenTool) ListVersions() ([]string, error) {
	// For now, return common versions
	// TODO: Implement dynamic version discovery from Maven Central
	return []string{"3.9.6", "4.0.0", "4.0.0-rc-3"}, nil
}

// getDownloadURL returns the download URL for the specified version
func (m *MavenTool) getDownloadURL(version string) string {
	// Maven download URLs follow a consistent pattern
	if strings.HasPrefix(version, "4.") {
		// Maven 4.x is still in development, use different URL structure
		if version == "4.0.0" {
			return "https://repo.maven.apache.org/maven2/org/apache/maven/apache-maven/4.0.0/apache-maven-4.0.0-bin.zip"
		}
		if version == "4.0.0-rc-3" {
			return "https://repo.maven.apache.org/maven2/org/apache/maven/apache-maven/4.0.0-rc-3/apache-maven-4.0.0-rc-3-bin.zip"
		}
	}

	// Maven 3.x versions
	return fmt.Sprintf("https://archive.apache.org/dist/maven/maven-3/%s/binaries/apache-maven-%s-bin.zip", version, version)
}

// downloadAndExtract downloads and extracts a zip file
func (m *MavenTool) downloadAndExtract(url, destDir string) error {
	// Create temporary file for download
	tmpFile, err := os.CreateTemp("", "maven-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()
	
	// Download file
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}
	
	// Copy to temporary file
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save download: %w", err)
	}
	
	// Close temp file before reading
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
