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

// RustTool manages Rust installations (includes Cargo)
// Downloads from https://forge.rust-lang.org/infra/channel-releases.html
type RustTool struct {
	manager *Manager
}

func (r *RustTool) Name() string { return "rust" }

// Install downloads and installs the specified Rust version
func (r *RustTool) Install(version string, cfg config.ToolConfig) error {
	installDir := r.manager.GetToolVersionDir("rust", version, "")

	// Create installation directory
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}

	// Get download URL for rustup-init
	downloadURL := r.getDownloadURL()

	// Download and install using rustup-init
	fmt.Printf("  ⏳ Installing Rust %s...\n", version)
	if err := r.downloadAndInstall(downloadURL, installDir, version, cfg); err != nil {
		return fmt.Errorf("failed to download and install: %w", err)
	}

	return nil
}

// IsInstalled checks if the specified version is installed
func (r *RustTool) IsInstalled(version string, cfg config.ToolConfig) bool {
	// Check if rustc and cargo executables exist
	binPath, err := r.GetBinPath(version, cfg)
	if err != nil {
		return false
	}

	rustcExe := filepath.Join(binPath, "rustc")
	cargoExe := filepath.Join(binPath, "cargo")
	if runtime.GOOS == "windows" {
		rustcExe += ".exe"
		cargoExe += ".exe"
	}

	// Check if both rustc and cargo exist
	if _, err := os.Stat(rustcExe); err != nil {
		return false
	}
	if _, err := os.Stat(cargoExe); err != nil {
		return false
	}

	return true
}

// GetPath returns the installation path for the specified version
func (r *RustTool) GetPath(version string, cfg config.ToolConfig) (string, error) {
	installDir := r.manager.GetToolVersionDir("rust", version, "")

	// Rust toolchain is installed in a specific structure
	// ~/.mvx/tools/rust/{version}/rustup/toolchains/{version}-{arch}/
	toolchainDir := filepath.Join(installDir, "rustup", "toolchains")

	// Find the toolchain directory
	entries, err := os.ReadDir(toolchainDir)
	if err != nil {
		return "", fmt.Errorf("failed to read toolchain directory: %w", err)
	}

	// Look for the version-specific toolchain directory
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), version) {
			return filepath.Join(toolchainDir, entry.Name()), nil
		}
	}

	return "", fmt.Errorf("rust toolchain not found in %s", toolchainDir)
}

// GetBinPath returns the binary path for the specified version
func (r *RustTool) GetBinPath(version string, cfg config.ToolConfig) (string, error) {
	rustHome, err := r.GetPath(version, cfg)
	if err != nil {
		return "", err
	}
	return filepath.Join(rustHome, "bin"), nil
}

// Verify checks if the installation is working correctly
func (r *RustTool) Verify(version string, cfg config.ToolConfig) error {
	binPath, err := r.GetBinPath(version, cfg)
	if err != nil {
		return err
	}

	// Verify rustc
	rustcExe := filepath.Join(binPath, "rustc")
	if runtime.GOOS == "windows" {
		rustcExe += ".exe"
	}

	cmd := exec.Command(rustcExe, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rustc verification failed: %w\nOutput: %s", err, output)
	}

	// Verify cargo
	cargoExe := filepath.Join(binPath, "cargo")
	if runtime.GOOS == "windows" {
		cargoExe += ".exe"
	}

	cmd = exec.Command(cargoExe, "--version")
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cargo verification failed: %w\nOutput: %s", err, output)
	}

	logVerbose("Rust %s verification successful", version)
	return nil
}

// ListVersions returns available versions for installation
func (r *RustTool) ListVersions() ([]string, error) {
	return r.manager.registry.GetRustVersions()
}

// getDownloadURL returns the download URL for rustup-init
func (r *RustTool) getDownloadURL() string {
	var arch, ext string

	switch runtime.GOOS {
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			arch = "x86_64-unknown-linux-gnu"
		case "arm64":
			arch = "aarch64-unknown-linux-gnu"
		default:
			arch = "x86_64-unknown-linux-gnu"
		}
		ext = ""
	case "darwin":
		switch runtime.GOARCH {
		case "amd64":
			arch = "x86_64-apple-darwin"
		case "arm64":
			arch = "aarch64-apple-darwin"
		default:
			arch = "x86_64-apple-darwin"
		}
		ext = ""
	case "windows":
		arch = "x86_64-pc-windows-msvc"
		ext = ".exe"
	default:
		arch = "x86_64-unknown-linux-gnu"
		ext = ""
	}

	return fmt.Sprintf("https://static.rust-lang.org/rustup/dist/%s/rustup-init%s", arch, ext)
}

// downloadAndInstall downloads rustup-init and installs Rust
func (r *RustTool) downloadAndInstall(url, destDir, version string, cfg config.ToolConfig) error {
	// Create temporary file for rustup-init
	tmpFile, err := os.CreateTemp("", "rustup-init-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	fmt.Printf("  ⏳ Downloading rustup-init from %s...\n", url)

	// Configure robust download
	config := DefaultDownloadConfig(url, tmpFile.Name())
	config.ExpectedType = ""            // Accept any content type for rustup-init binary
	config.MinSize = 5 * 1024 * 1024    // Minimum 5MB for rustup-init
	config.MaxSize = 50 * 1024 * 1024   // Maximum 50MB for rustup-init
	config.ValidateMagic = false        // rustup-init is a binary, not an archive
	config.ToolName = "rust"            // For progress reporting
	config.Version = version            // For checksum verification
	config.Config = cfg                 // Tool configuration
	config.ChecksumRegistry = r.manager.GetChecksumRegistry()

	// Perform robust download
	result, err := RobustDownload(config)
	if err != nil {
		return fmt.Errorf("rustup-init download failed: %s", DiagnoseDownloadError(url, err))
	}

	fmt.Printf("  📦 Downloaded %d bytes from %s\n", result.Size, result.FinalURL)

	// Close temp file before execution
	tmpFile.Close()

	// Make rustup-init executable
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		return fmt.Errorf("failed to make rustup-init executable: %w", err)
	}

	// Set up environment for rustup installation
	env := os.Environ()
	env = append(env, fmt.Sprintf("RUSTUP_HOME=%s", filepath.Join(destDir, "rustup")))
	env = append(env, fmt.Sprintf("CARGO_HOME=%s", filepath.Join(destDir, "cargo")))

	// Install Rust using rustup-init
	fmt.Printf("  🦀 Installing Rust toolchain %s...\n", version)

	args := []string{
		"--default-toolchain", version,
		"--profile", "default",
		"--no-modify-path",
		"-y", // Accept defaults
	}

	cmd := exec.Command(tmpFile.Name(), args...)
	cmd.Env = env
	cmd.Dir = destDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rustup installation failed: %w\nOutput: %s", err, output)
	}

	fmt.Printf("  ✅ Rust %s installed successfully\n", version)
	return nil
}
