package tools

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
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

// PythonTool implements Tool interface for Python management
type PythonTool struct {
	manager *Manager
}

// Name returns the tool name
func (p *PythonTool) Name() string {
	return "python"
}

// Install downloads and installs the specified Python version
func (p *PythonTool) Install(version string, cfg config.ToolConfig) error {
	installDir := p.manager.GetToolVersionDir("python", version, "")

	// Create installation directory
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}

	// Get download URL
	downloadURL, err := p.getDownloadURL(version)
	if err != nil {
		return fmt.Errorf("failed to get download URL: %w", err)
	}

	// Download and extract
	fmt.Printf("  ⏳ Downloading Python %s...\n", version)
	if err := p.downloadAndExtract(downloadURL, installDir); err != nil {
		return fmt.Errorf("failed to download and extract: %w", err)
	}

	// Verify installation
	if err := p.Verify(version, cfg); err != nil {
		return fmt.Errorf("installation verification failed: %w", err)
	}

	// Create project-specific virtual environment directory structure
	if err := p.setupProjectEnvironments(version, cfg); err != nil {
		return fmt.Errorf("failed to setup project environments: %w", err)
	}

	fmt.Printf("  ✅ Python %s installed successfully\n", version)
	return nil
}

// IsInstalled checks if the specified version is installed
func (p *PythonTool) IsInstalled(version string, cfg config.ToolConfig) bool {
	binPath, err := p.GetBinPath(version, cfg)
	if err != nil {
		return false
	}

	pythonExe := filepath.Join(binPath, p.pythonBinaryName())
	_, err = os.Stat(pythonExe)
	return err == nil
}

// GetPath returns the installation path for the specified version
func (p *PythonTool) GetPath(version string, cfg config.ToolConfig) (string, error) {
	return p.manager.GetToolVersionDir("python", version, ""), nil
}

// GetBinPath returns the binary path for the specified version
func (p *PythonTool) GetBinPath(version string, cfg config.ToolConfig) (string, error) {
	installPath, err := p.GetPath(version, cfg)
	if err != nil {
		return "", err
	}

	// Different platforms have different directory structures
	switch runtime.GOOS {
	case "windows":
		return installPath, nil
	case "darwin", "linux":
		return filepath.Join(installPath, "bin"), nil
	default:
		return filepath.Join(installPath, "bin"), nil
	}
}

// Verify checks if the installation is working correctly
func (p *PythonTool) Verify(version string, cfg config.ToolConfig) error {
	binPath, err := p.GetBinPath(version, cfg)
	if err != nil {
		return err
	}

	pythonExe := filepath.Join(binPath, p.pythonBinaryName())

	// Run python --version to verify installation
	cmd := exec.Command(pythonExe, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("python verification failed: %w\nOutput: %s", err, output)
	}

	// Check if output contains expected version
	outputStr := string(output)
	if !strings.Contains(outputStr, version) {
		return fmt.Errorf("python version mismatch: expected %s, got %s", version, outputStr)
	}

	// Also verify pip is available
	pipExe := filepath.Join(binPath, p.pipBinaryName())
	cmd = exec.Command(pipExe, "--version")
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pip verification failed: %w\nOutput: %s", err, output)
	}

	return nil
}

// ListVersions returns available versions for installation
func (p *PythonTool) ListVersions() ([]string, error) {
	registry := p.manager.GetRegistry()
	return registry.GetPythonVersions()
}

// pythonBinaryName returns the Python binary name for the current platform
func (p *PythonTool) pythonBinaryName() string {
	if runtime.GOOS == "windows" {
		return "python.exe"
	}
	return "python3"
}

// pipBinaryName returns the pip binary name for the current platform
func (p *PythonTool) pipBinaryName() string {
	if runtime.GOOS == "windows" {
		return "pip.exe"
	}
	return "pip3"
}

// getDownloadURL returns the download URL for the specified version
func (p *PythonTool) getDownloadURL(version string) (string, error) {
	// Determine platform and architecture
	osName := runtime.GOOS
	arch := runtime.GOARCH

	// Map architecture names to Python's naming convention
	var archStr string
	switch arch {
	case "amd64":
		if osName == "windows" {
			archStr = "amd64"
		} else {
			archStr = "x86_64"
		}
	case "arm64":
		if osName == "darwin" {
			archStr = "arm64"
		} else {
			archStr = "aarch64"
		}
	case "386":
		archStr = "i386"
	default:
		return "", fmt.Errorf("unsupported architecture: %s", arch)
	}

	// Use archStr in URL construction to avoid unused variable error
	_ = archStr

	// Construct download URL based on platform
	switch osName {
	case "windows":
		if arch == "amd64" {
			return fmt.Sprintf("https://www.python.org/ftp/python/%s/python-%s-embed-amd64.zip", version, version), nil
		} else {
			return fmt.Sprintf("https://www.python.org/ftp/python/%s/python-%s-embed-win32.zip", version, version), nil
		}
	case "darwin":
		// For macOS, we'll use the universal2 installer when available, otherwise x86_64
		if arch == "arm64" {
			return fmt.Sprintf("https://www.python.org/ftp/python/%s/python-%s-macos11.pkg", version, version), nil
		} else {
			return fmt.Sprintf("https://www.python.org/ftp/python/%s/python-%s-macosx10.9.pkg", version, version), nil
		}
	case "linux":
		// For Linux, we'll build from source or use precompiled binaries
		return fmt.Sprintf("https://www.python.org/ftp/python/%s/Python-%s.tgz", version, version), nil
	default:
		return "", fmt.Errorf("unsupported platform: %s", osName)
	}
}

// downloadAndExtract downloads and extracts a Python archive
func (p *PythonTool) downloadAndExtract(url, destDir string) error {
	// Create temporary file for download
	tmpFile, err := os.CreateTemp("", "python-*.archive")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Download file with timeout
	client := &http.Client{
		Timeout: 600 * time.Second, // 10 minute timeout for larger Python downloads
	}
	resp, err := client.Get(url)
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

	// Extract archive based on file extension
	if strings.HasSuffix(url, ".zip") {
		return p.extractZip(tmpFile.Name(), destDir)
	} else if strings.HasSuffix(url, ".tgz") || strings.HasSuffix(url, ".tar.gz") {
		return p.extractTarGz(tmpFile.Name(), destDir)
	} else if strings.HasSuffix(url, ".pkg") {
		return p.installPkg(tmpFile.Name(), destDir)
	} else {
		return fmt.Errorf("unsupported archive format: %s", url)
	}
}

// extractZip extracts a ZIP file
func (p *PythonTool) extractZip(archivePath, destDir string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		path := filepath.Join(destDir, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.FileInfo().Mode())
			continue
		}

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Extract file
		fileReader, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in zip: %w", err)
		}

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.FileInfo().Mode())
		if err != nil {
			fileReader.Close()
			return fmt.Errorf("failed to create target file: %w", err)
		}

		_, err = io.Copy(targetFile, fileReader)
		fileReader.Close()
		targetFile.Close()

		if err != nil {
			return fmt.Errorf("failed to extract file: %w", err)
		}
	}

	return nil
}

// extractTarGz extracts a tar.gz file
func (p *PythonTool) extractTarGz(archivePath, destDir string) error {
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

		// Skip the top-level directory (e.g., "Python-3.11.0/")
		parts := strings.Split(header.Name, "/")
		if len(parts) <= 1 {
			continue
		}
		relativePath := strings.Join(parts[1:], "/")
		if relativePath == "" {
			continue
		}

		path := filepath.Join(destDir, relativePath)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		case tar.TypeReg:
			// Create parent directories
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory: %w", err)
			}

			// Create file
			outFile, err := os.Create(path)
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}

			// Copy file content
			_, err = io.Copy(outFile, tarReader)
			outFile.Close()
			if err != nil {
				return fmt.Errorf("failed to extract file: %w", err)
			}

			// Set file permissions
			if err := os.Chmod(path, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to set file permissions: %w", err)
			}
		}
	}

	return nil
}

// installPkg handles macOS .pkg installer files
func (p *PythonTool) installPkg(pkgPath, destDir string) error {
	// For now, we'll return an error as .pkg installation requires more complex handling
	// In a full implementation, we would need to extract the .pkg and handle the installation
	return fmt.Errorf("macOS .pkg installation not yet implemented - please install Python manually")
}

// setupProjectEnvironments creates the directory structure for project-specific Python environments
func (p *PythonTool) setupProjectEnvironments(version string, cfg config.ToolConfig) error {
	// No global setup needed since venvs are stored per-project
	return nil
}

// GetProjectVenvPath returns the path to a project-specific virtual environment
func (p *PythonTool) GetProjectVenvPath(version string, projectPath string) string {
	// Store virtual environment in the project's .mvx directory
	return filepath.Join(projectPath, ".mvx", "venv")
}

// CreateProjectVenv creates a virtual environment for a specific project
func (p *PythonTool) CreateProjectVenv(version string, cfg config.ToolConfig, projectPath string) error {
	venvPath := p.GetProjectVenvPath(version, projectPath)

	// Check if virtual environment already exists
	if _, err := os.Stat(venvPath); err == nil {
		return nil // Already exists
	}

	// Get Python binary path
	binPath, err := p.GetBinPath(version, cfg)
	if err != nil {
		return fmt.Errorf("failed to get Python binary path: %w", err)
	}

	pythonExe := filepath.Join(binPath, p.pythonBinaryName())

	// Create virtual environment
	fmt.Printf("  🐍 Creating Python virtual environment for project...\n")
	cmd := exec.Command(pythonExe, "-m", "venv", venvPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create virtual environment: %w\nOutput: %s", err, output)
	}

	// Upgrade pip in the virtual environment
	venvPython := p.getVenvPythonPath(venvPath)
	cmd = exec.Command(venvPython, "-m", "pip", "install", "--upgrade", "pip")
	output, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("  ⚠️  Warning: failed to upgrade pip in virtual environment: %v\n", err)
	}

	fmt.Printf("  ✅ Virtual environment created at: %s\n", venvPath)
	return nil
}

// getVenvPythonPath returns the Python executable path within a virtual environment
func (p *PythonTool) getVenvPythonPath(venvPath string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(venvPath, "Scripts", "python.exe")
	}
	return filepath.Join(venvPath, "bin", "python")
}

// getVenvPipPath returns the pip executable path within a virtual environment
func (p *PythonTool) getVenvPipPath(venvPath string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(venvPath, "Scripts", "pip.exe")
	}
	return filepath.Join(venvPath, "bin", "pip")
}

// GetProjectEnvironment returns environment variables for a project-specific Python setup
func (p *PythonTool) GetProjectEnvironment(version string, cfg config.ToolConfig, projectPath string) (map[string]string, error) {
	env := make(map[string]string)

	// Get base Python installation path
	pythonHome, err := p.GetPath(version, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get Python path: %w", err)
	}

	// Get project-specific virtual environment path
	venvPath := p.GetProjectVenvPath(version, projectPath)

	// Set Python-specific environment variables
	env["PYTHON_HOME"] = pythonHome
	env["VIRTUAL_ENV"] = venvPath

	// Set up PATH to prioritize virtual environment binaries
	venvBinPath := filepath.Join(venvPath, "bin")
	if runtime.GOOS == "windows" {
		venvBinPath = filepath.Join(venvPath, "Scripts")
	}

	// Get current PATH and prepend virtual environment bin path
	currentPath := os.Getenv("PATH")
	if currentPath != "" {
		env["PATH"] = venvBinPath + string(os.PathListSeparator) + currentPath
	} else {
		env["PATH"] = venvBinPath
	}

	// Set PYTHONPATH to ensure project isolation
	env["PYTHONPATH"] = filepath.Join(venvPath, "lib", "python"+version[:3], "site-packages")
	if runtime.GOOS == "windows" {
		env["PYTHONPATH"] = filepath.Join(venvPath, "Lib", "site-packages")
	}

	// Disable user site packages to ensure isolation
	env["PYTHONNOUSERSITE"] = "1"

	// Set pip configuration for project isolation
	env["PIP_PREFIX"] = venvPath
	env["PIP_DISABLE_PIP_VERSION_CHECK"] = "1"

	return env, nil
}

// InstallProjectRequirements installs requirements from a requirements file in the project's virtual environment
func (p *PythonTool) InstallProjectRequirements(version string, cfg config.ToolConfig, projectPath, requirementsFile string) error {
	venvPath := p.GetProjectVenvPath(version, projectPath)

	// Check if virtual environment exists
	if _, err := os.Stat(venvPath); os.IsNotExist(err) {
		if err := p.CreateProjectVenv(version, cfg, projectPath); err != nil {
			return fmt.Errorf("failed to create virtual environment: %w", err)
		}
	}

	// Check if requirements file exists
	reqPath := filepath.Join(projectPath, requirementsFile)
	if _, err := os.Stat(reqPath); os.IsNotExist(err) {
		return nil // No requirements file, nothing to install
	}

	// Install requirements using pip in the virtual environment
	venvPip := p.getVenvPipPath(venvPath)
	fmt.Printf("  📦 Installing Python requirements from %s...\n", requirementsFile)

	cmd := exec.Command(venvPip, "install", "-r", reqPath)
	cmd.Dir = projectPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install requirements: %w\nOutput: %s", err, output)
	}

	fmt.Printf("  ✅ Requirements installed successfully\n")
	return nil
}
