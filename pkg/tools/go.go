package tools

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gnodet/mvx/pkg/config"
)

// GoTool implements Tool interface for Go toolchain management
type GoTool struct {
	manager *Manager
}

// Name returns the tool name
func (g *GoTool) Name() string {
	return "go"
}

// Install downloads and installs the specified Go version
func (g *GoTool) Install(version string, cfg config.ToolConfig) error {
	installDir := g.manager.GetToolVersionDir("go", version, "")

	// Create installation directory
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}

	// Get download URL
	downloadURL := g.getDownloadURL(version)

	// Download and extract
	fmt.Printf("  ⏳ Downloading Go %s...\n", version)
	if err := g.downloadAndExtract(downloadURL, installDir); err != nil {
		return fmt.Errorf("failed to download and extract: %w", err)
	}

	return nil
}

// IsInstalled checks if the specified version is installed
func (g *GoTool) IsInstalled(version string, cfg config.ToolConfig) bool {
	installDir := g.manager.GetToolVersionDir("go", version, "")
	goExe := filepath.Join(installDir, "go", "bin", "go")
	if runtime.GOOS == "windows" {
		goExe += ".exe"
	}

	_, err := os.Stat(goExe)
	return err == nil
}

// GetPath returns the installation path for the specified version
func (g *GoTool) GetPath(version string, cfg config.ToolConfig) (string, error) {
	installDir := g.manager.GetToolVersionDir("go", version, "")

	// Go archives extract to a "go" subdirectory
	goPath := filepath.Join(installDir, "go")
	if _, err := os.Stat(goPath); err == nil {
		return goPath, nil
	}

	return installDir, nil
}

// GetBinPath returns the binary path for the specified version
func (g *GoTool) GetBinPath(version string, cfg config.ToolConfig) (string, error) {
	goHome, err := g.GetPath(version, cfg)
	if err != nil {
		return "", err
	}
	return filepath.Join(goHome, "bin"), nil
}

// Verify checks if the installation is working correctly
func (g *GoTool) Verify(version string, cfg config.ToolConfig) error {
	binPath, err := g.GetBinPath(version, cfg)
	if err != nil {
		return err
	}

	goExe := filepath.Join(binPath, "go")
	if runtime.GOOS == "windows" {
		goExe += ".exe"
	}

	// Run go version to verify installation
	cmd := exec.Command(goExe, "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go verification failed: %w\nOutput: %s", err, output)
	}

	// Check if output contains expected version
	outputStr := string(output)
	if !strings.Contains(outputStr, version) {
		return fmt.Errorf("go version mismatch: expected %s, got %s", version, outputStr)
	}

	return nil
}

// ListVersions returns available versions for installation
func (g *GoTool) ListVersions() ([]string, error) {
	registry := g.manager.GetRegistry()
	return registry.GetGoVersions()
}

// getDownloadURL returns the download URL for the specified version
func (g *GoTool) getDownloadURL(version string) string {
	// Determine platform string
	osName := runtime.GOOS
	arch := runtime.GOARCH

	// Map architecture names to Go's naming convention
	switch arch {
	case "amd64":
		arch = "amd64"
	case "arm64":
		arch = "arm64"
	case "386":
		arch = "386"
	}

	// Construct filename
	var filename string
	if osName == "windows" {
		filename = fmt.Sprintf("go%s.%s-%s.zip", version, osName, arch)
	} else {
		filename = fmt.Sprintf("go%s.%s-%s.tar.gz", version, osName, arch)
	}

	return fmt.Sprintf("https://go.dev/dl/%s", filename)
}

// downloadAndExtract downloads and extracts a Go archive
func (g *GoTool) downloadAndExtract(url, destDir string) error {
	// Create temporary file for download
	tmpFile, err := os.CreateTemp("", "go-*.archive")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Configure robust download
	config := DefaultDownloadConfig(url, tmpFile.Name())
	config.ExpectedType = "application" // Accept various application types
	config.MinSize = 50 * 1024 * 1024   // Minimum 50MB for Go distributions
	config.MaxSize = 200 * 1024 * 1024  // Maximum 200MB for Go distributions
	config.ToolName = "go"              // For progress reporting

	// Perform robust download
	result, err := RobustDownload(config)
	if err != nil {
		return fmt.Errorf("Go download failed: %s", DiagnoseDownloadError(url, err))
	}

	fmt.Printf("  📦 Downloaded %d bytes from %s\n", result.Size, result.FinalURL)

	// Close temp file before extraction
	tmpFile.Close()

	// Extract archive based on file extension
	if strings.HasSuffix(url, ".zip") {
		return g.extractZip(tmpFile.Name(), destDir)
	} else {
		return g.extractTarGz(tmpFile.Name(), destDir)
	}
}

// extractTarGz extracts a tar.gz file
func (g *GoTool) extractTarGz(archivePath, destDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar entry: %w", err)
		}

		targetPath := filepath.Join(destDir, header.Name)

		// Security check: ensure path is within destination directory
		if !strings.HasPrefix(targetPath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory: %w", err)
			}

			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return fmt.Errorf("failed to extract file: %w", err)
			}
			outFile.Close()
		}
	}

	return nil
}

// extractZip extracts a zip file (for Windows)
func (g *GoTool) extractZip(archivePath, destDir string) error {
	// Import archive/zip at the top of the file
	return fmt.Errorf("zip extraction not yet implemented for Go on Windows - use tar.gz version")
}
