package tools

import (
	"archive/tar"
	"compress/gzip"
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

	// Try to get the actual Java path (which handles nested directories)
	_, err := j.GetPath(version, cfg)
	return err == nil
}

// GetPath returns the installation path for the specified version
func (j *JavaTool) GetPath(version string, cfg config.ToolConfig) (string, error) {
	distribution := cfg.Distribution
	if distribution == "" {
		distribution = "temurin"
	}
	
	installDir := j.manager.GetToolVersionDir("java", version, distribution)
	
	// Check if there's a nested directory (common with JDK archives)
	entries, err := os.ReadDir(installDir)
	if err != nil {
		return "", fmt.Errorf("failed to read installation directory: %w", err)
	}
	
	// Look for a subdirectory that looks like a JDK
	for _, entry := range entries {
		if entry.IsDir() {
			subPath := filepath.Join(installDir, entry.Name())
			javaExe := filepath.Join(subPath, "bin", "java")
			if runtime.GOOS == "windows" {
				javaExe += ".exe"
			}
			if _, err := os.Stat(javaExe); err == nil {
				return subPath, nil
			}
		}
	}
	
	return installDir, nil
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

// ListVersions returns available versions for installation
func (j *JavaTool) ListVersions() ([]string, error) {
	// For now, return common LTS versions + latest
	// TODO: Implement dynamic version discovery from Adoptium API
	return []string{"8", "11", "17", "21", "22"}, nil
}

// getDownloadURL returns the download URL for the specified version and distribution
func (j *JavaTool) getDownloadURL(version, distribution string) (string, error) {
	switch distribution {
	case "temurin":
		return j.getTemurinURL(version)
	default:
		return "", fmt.Errorf("unsupported Java distribution: %s", distribution)
	}
}

// getTemurinURL returns the download URL for Eclipse Temurin
func (j *JavaTool) getTemurinURL(version string) (string, error) {
	// For now, let's use a simpler approach with known working URLs
	// TODO: Use Adoptium API for dynamic URL discovery

	osName := runtime.GOOS
	arch := runtime.GOARCH

	// Map Go arch to JDK arch
	switch arch {
	case "amd64":
		arch = "x64"
	case "arm64":
		arch = "aarch64"
	}

	// Map OS names
	switch osName {
	case "darwin":
		osName = "mac"
	case "windows":
		osName = "windows"
	case "linux":
		osName = "linux"
	}

	// Use Adoptium API endpoint for latest releases
	// This is more reliable than hardcoded URLs
	switch version {
	case "22":
		return fmt.Sprintf("https://api.adoptium.net/v3/binary/latest/22/ga/%s/%s/jdk/hotspot/normal/eclipse", osName, arch), nil
	case "21":
		return fmt.Sprintf("https://api.adoptium.net/v3/binary/latest/21/ga/%s/%s/jdk/hotspot/normal/eclipse", osName, arch), nil
	case "17":
		return fmt.Sprintf("https://api.adoptium.net/v3/binary/latest/17/ga/%s/%s/jdk/hotspot/normal/eclipse", osName, arch), nil
	case "11":
		return fmt.Sprintf("https://api.adoptium.net/v3/binary/latest/11/ga/%s/%s/jdk/hotspot/normal/eclipse", osName, arch), nil
	case "8":
		return fmt.Sprintf("https://api.adoptium.net/v3/binary/latest/8/ga/%s/%s/jdk/hotspot/normal/eclipse", osName, arch), nil
	default:
		return "", fmt.Errorf("unsupported Java version: %s (supported: 8, 11, 17, 21, 22)", version)
	}
}

// downloadAndExtract downloads and extracts a tar.gz file
func (j *JavaTool) downloadAndExtract(url, destDir string) error {
	// Download file
	resp, err := http.Get(url)
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
