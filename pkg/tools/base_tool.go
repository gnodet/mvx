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

// BaseTool provides common functionality for all tools
type BaseTool struct {
	manager  *Manager
	toolName string
}

// NewBaseTool creates a new base tool instance
func NewBaseTool(manager *Manager, toolName string) *BaseTool {
	return &BaseTool{
		manager:  manager,
		toolName: toolName,
	}
}

// GetManager returns the tool manager
func (b *BaseTool) GetManager() *Manager {
	return b.manager
}

// GetToolName returns the tool name
func (b *BaseTool) GetToolName() string {
	return b.toolName
}

// CreateInstallDir creates the installation directory for a tool version
func (b *BaseTool) CreateInstallDir(version, distribution string) (string, error) {
	installDir := b.manager.GetToolVersionDir(b.toolName, version, distribution)
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create installation directory: %w", err)
	}
	return installDir, nil
}

// DownloadAndExtract performs a robust download with checksum verification
func (b *BaseTool) DownloadAndExtract(url, destDir, version string, cfg config.ToolConfig, options DownloadOptions) error {
	// Create temporary file for download
	tmpFile, err := os.CreateTemp("", fmt.Sprintf("%s-*%s", b.toolName, options.FileExtension))
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Configure robust download with checksum verification
	downloadConfig := DefaultDownloadConfig(url, tmpFile.Name())
	downloadConfig.ExpectedType = options.ExpectedType
	downloadConfig.MinSize = options.MinSize
	downloadConfig.MaxSize = options.MaxSize
	downloadConfig.ToolName = b.toolName
	downloadConfig.Version = version
	downloadConfig.Config = cfg
	downloadConfig.ChecksumRegistry = b.manager.GetChecksumRegistry()

	// Perform robust download with checksum verification
	result, err := RobustDownload(downloadConfig)
	if err != nil {
		return fmt.Errorf("%s download failed: %s", strings.Title(b.toolName), DiagnoseDownloadError(url, err))
	}

	fmt.Printf("  📦 Downloaded %d bytes from %s\n", result.Size, result.FinalURL)

	// Close temp file before extraction
	tmpFile.Close()

	// Extract based on file type
	return b.extractFile(tmpFile.Name(), destDir, options.ArchiveType)
}

// DownloadOptions contains options for downloading and extracting files
type DownloadOptions struct {
	FileExtension string // e.g., ".tar.gz", ".zip"
	ExpectedType  string // e.g., "application"
	MinSize       int64  // Minimum expected file size
	MaxSize       int64  // Maximum expected file size
	ArchiveType   string // "tar.gz", "zip", "tar.xz"
}

// extractFile extracts an archive file based on its type
func (b *BaseTool) extractFile(src, dest, archiveType string) error {
	switch archiveType {
	case "zip":
		return b.extractZip(src, dest)
	case "tar.gz":
		return b.extractTarGz(src, dest)
	case "tar.xz":
		return b.extractTarXz(src, dest)
	default:
		return fmt.Errorf("unsupported archive type: %s", archiveType)
	}
}

// extractZip extracts a zip file to the destination directory
func (b *BaseTool) extractZip(src, dest string) error {
	return extractZipFile(src, dest)
}

// extractTarGz extracts a tar.gz file to the destination directory
func (b *BaseTool) extractTarGz(src, dest string) error {
	return extractTarGzFile(src, dest)
}

// extractTarXz extracts a tar.xz file to the destination directory
func (b *BaseTool) extractTarXz(src, dest string) error {
	return extractTarXzFile(src, dest)
}

// VerifyBinary runs a binary with version flag and checks the output
func (b *BaseTool) VerifyBinary(binPath, binaryName, version string, versionArgs []string) error {
	exe := filepath.Join(binPath, binaryName)
	if runtime.GOOS == "windows" && !strings.HasSuffix(exe, ".exe") {
		exe += ".exe"
	}

	// Run version command
	cmd := exec.Command(exe, versionArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s verification failed: %w\nOutput: %s", b.toolName, err, output)
	}

	// Check if output contains expected version
	outputStr := string(output)
	if !strings.Contains(outputStr, version) {
		// Allow 'v' prefix in output for some tools
		if !strings.Contains(outputStr, "v"+version) {
			return fmt.Errorf("%s version mismatch: expected %s, got %s", b.toolName, version, outputStr)
		}
	}

	return nil
}

// IsInstalled checks if a binary exists at the expected path
func (b *BaseTool) IsInstalled(binPath, binaryName string) bool {
	exe := filepath.Join(binPath, binaryName)
	if runtime.GOOS == "windows" && !strings.HasSuffix(exe, ".exe") {
		exe += ".exe"
	}
	_, err := os.Stat(exe)
	return err == nil
}

// GetDisplayName returns the display name for this tool
// Tools can implement GetDisplayName() to provide their own display name
func (b *BaseTool) GetDisplayName() string {
	// Use reflection to check if the concrete tool implements GetDisplayName
	if tool, err := b.manager.GetTool(b.toolName); err == nil {
		if nameProvider, hasMethod := tool.(interface{ GetDisplayName() string }); hasMethod {
			return nameProvider.GetDisplayName()
		}
	}
	// Fall back to title case of tool name
	return strings.Title(b.toolName)
}

// PrintDownloadMessage prints a standardized download message
func (b *BaseTool) PrintDownloadMessage(version string) {
	toolDisplayName := b.GetDisplayName()
	fmt.Printf("  ⏳ Downloading %s %s...\n", toolDisplayName, version)
}

// GetDefaultDownloadOptions returns default download options for tools that don't specify their own
func (b *BaseTool) GetDefaultDownloadOptions() DownloadOptions {
	return DownloadOptions{
		FileExtension: ".tar.gz",
		ExpectedType:  "application",
		MinSize:       1 * 1024 * 1024,   // 1MB
		MaxSize:       100 * 1024 * 1024, // 100MB
		ArchiveType:   "tar.gz",
	}
}

// getDownloadOptions returns download options for this tool
// Tools can implement GetDownloadOptions() to provide their own options
func (b *BaseTool) getDownloadOptions() DownloadOptions {
	// Use reflection to check if the concrete tool implements GetDownloadOptions
	if tool, err := b.manager.GetTool(b.toolName); err == nil {
		if optionsProvider, hasMethod := tool.(interface{ GetDownloadOptions() DownloadOptions }); hasMethod {
			return optionsProvider.GetDownloadOptions()
		}
	}
	// Fall back to default options
	return b.GetDefaultDownloadOptions()
}

// StandardInstall provides a standard installation flow for most tools
func (b *BaseTool) StandardInstall(version string, cfg config.ToolConfig, getDownloadURL func(string) string) error {
	return b.StandardInstallWithOptions(version, cfg, getDownloadURL, nil, nil)
}

// StandardInstallWithOptions provides a standard installation flow with system tool support
func (b *BaseTool) StandardInstallWithOptions(version string, cfg config.ToolConfig, getDownloadURL func(string) string, alternativeNames []string, envVars []string) error {
	// Check if we should use system tool instead of downloading
	if UseSystemTool(b.toolName) {
		logVerbose("%s=true, forcing use of system %s", getSystemToolEnvVar(b.toolName), b.toolName)

		// Try environment variables first (for tools like Maven that need MAVEN_HOME)
		for _, envVar := range envVars {
			if envValue := os.Getenv(envVar); envValue != "" {
				// Verify the tool exists in this environment path
				toolPath := filepath.Join(envValue, "bin", b.toolName)
				if runtime.GOOS == "windows" && !strings.HasSuffix(b.toolName, ".exe") && !strings.HasSuffix(b.toolName, ".cmd") {
					if _, err := os.Stat(toolPath + ".cmd"); err == nil {
						fmt.Printf("  🔗 Using system %s from %s: %s\n", b.toolName, envVar, envValue)
						fmt.Printf("  ✅ System %s configured (mvx will use system PATH)\n", b.toolName)
						return nil
					}
					if _, err := os.Stat(toolPath + ".exe"); err == nil {
						fmt.Printf("  🔗 Using system %s from %s: %s\n", b.toolName, envVar, envValue)
						fmt.Printf("  ✅ System %s configured (mvx will use system PATH)\n", b.toolName)
						return nil
					}
				}
				if _, err := os.Stat(toolPath); err == nil {
					fmt.Printf("  🔗 Using system %s from %s: %s\n", b.toolName, envVar, envValue)
					fmt.Printf("  ✅ System %s configured (mvx will use system PATH)\n", b.toolName)
					return nil
				}
			}
		}

		// Try primary binary name in PATH
		if toolPath, err := exec.LookPath(b.toolName); err == nil {
			fmt.Printf("  🔗 Using system %s from PATH: %s\n", b.toolName, toolPath)
			fmt.Printf("  ✅ System %s configured (mvx will use system PATH)\n", b.toolName)
			return nil
		}

		// Try alternative binary names in PATH
		for _, altName := range alternativeNames {
			if toolPath, err := exec.LookPath(altName); err == nil {
				fmt.Printf("  🔗 Using system %s from PATH as %s: %s\n", b.toolName, altName, toolPath)
				fmt.Printf("  ✅ System %s configured (mvx will use system PATH)\n", b.toolName)
				return nil
			}
		}

		return fmt.Errorf("MVX_USE_SYSTEM_%s=true but system %s not available", strings.ToUpper(b.toolName), b.toolName)
	}

	// Create installation directory
	installDir, err := b.CreateInstallDir(version, "")
	if err != nil {
		return InstallError(b.toolName, version, fmt.Errorf("failed to create install directory: %w", err))
	}

	// Get download URL
	downloadURL := getDownloadURL(version)

	// Print download message
	b.PrintDownloadMessage(version)

	// Download and extract with tool-specific options
	options := b.getDownloadOptions()
	if err := b.DownloadAndExtract(downloadURL, installDir, version, cfg, options); err != nil {
		return InstallError(b.toolName, version, err)
	}

	// Verify installation was successful (if verification function is available)
	if verifyFunc := b.getVerificationFunction(); verifyFunc != nil {
		if err := verifyFunc(version, cfg); err != nil {
			// Installation verification failed, clean up the installation directory
			fmt.Printf("  ❌ %s installation verification failed: %v\n", b.toolName, err)
			fmt.Printf("  🧹 Cleaning up failed installation directory...\n")
			if removeErr := os.RemoveAll(installDir); removeErr != nil {
				fmt.Printf("  ⚠️  Warning: failed to clean up installation directory: %v\n", removeErr)
			}
			return InstallError(b.toolName, version, fmt.Errorf("installation verification failed: %w", err))
		}
		fmt.Printf("  ✅ %s %s installation verification successful\n", b.toolName, version)
	}

	return nil
}

// getVerificationFunction returns a tool-specific verification function if available
func (b *BaseTool) getVerificationFunction() func(string, config.ToolConfig) error {
	// This is a placeholder - individual tools can override this by implementing their own verification
	// For now, we return nil to indicate no verification is available at the base level
	return nil
}

// StandardVerify provides standard verification for tools with simple version commands
func (b *BaseTool) StandardVerify(version string, cfg config.ToolConfig, getPath func(string, config.ToolConfig) (string, error), binaryName string, versionArgs []string) error {
	binPath, err := getPath(version, cfg)
	if err != nil {
		return VerifyError(b.toolName, version, fmt.Errorf("failed to get binary path: %w", err))
	}
	if err := b.VerifyBinary(binPath, binaryName, version, versionArgs); err != nil {
		return VerifyError(b.toolName, version, err)
	}
	return nil
}

// StandardIsInstalled provides standard installation check for tools
func (b *BaseTool) StandardIsInstalled(version string, cfg config.ToolConfig, getPath func(string, config.ToolConfig) (string, error), binaryName string) bool {
	return b.StandardIsInstalledWithAlternatives(version, cfg, getPath, binaryName, nil)
}

// StandardIsInstalledWithAlternatives provides standard installation check for tools with alternative binary names
func (b *BaseTool) StandardIsInstalledWithAlternatives(version string, cfg config.ToolConfig, getPath func(string, config.ToolConfig) (string, error), binaryName string, alternativeNames []string) bool {
	return b.StandardIsInstalledWithOptions(version, cfg, getPath, binaryName, alternativeNames, nil)
}

// StandardIsInstalledWithOptions provides standard installation check for tools with full customization options
func (b *BaseTool) StandardIsInstalledWithOptions(version string, cfg config.ToolConfig, getPath func(string, config.ToolConfig) (string, error), binaryName string, alternativeNames []string, envVars []string) bool {
	// Check if we should use system tool instead of mvx-managed tool
	if UseSystemTool(b.toolName) {
		// Try environment variables first (for tools like Maven that need MAVEN_HOME)
		for _, envVar := range envVars {
			if envValue := os.Getenv(envVar); envValue != "" {
				// Verify the tool exists in this environment path
				toolPath := filepath.Join(envValue, "bin", binaryName)
				if runtime.GOOS == "windows" && !strings.HasSuffix(binaryName, ".exe") && !strings.HasSuffix(binaryName, ".cmd") {
					if _, err := os.Stat(toolPath + ".cmd"); err == nil {
						logVerbose("System %s found via %s=%s (MVX_USE_SYSTEM_%s=true)", b.toolName, envVar, envValue, strings.ToUpper(b.toolName))
						return true
					}
					if _, err := os.Stat(toolPath + ".exe"); err == nil {
						logVerbose("System %s found via %s=%s (MVX_USE_SYSTEM_%s=true)", b.toolName, envVar, envValue, strings.ToUpper(b.toolName))
						return true
					}
				}
				if _, err := os.Stat(toolPath); err == nil {
					logVerbose("System %s found via %s=%s (MVX_USE_SYSTEM_%s=true)", b.toolName, envVar, envValue, strings.ToUpper(b.toolName))
					return true
				}
			}
		}

		// Try primary binary name in PATH
		if _, err := exec.LookPath(binaryName); err == nil {
			logVerbose("System %s is available in PATH (MVX_USE_SYSTEM_%s=true)", b.toolName, strings.ToUpper(b.toolName))
			return true
		}

		// Try alternative binary names in PATH
		for _, altName := range alternativeNames {
			if _, err := exec.LookPath(altName); err == nil {
				logVerbose("System %s is available in PATH as %s (MVX_USE_SYSTEM_%s=true)", b.toolName, altName, strings.ToUpper(b.toolName))
				return true
			}
		}

		logVerbose("System %s not available: not found in environment variables or PATH", b.toolName)
		return false
	}

	binPath, err := getPath(version, cfg)
	if err != nil {
		return false
	}
	return b.IsInstalled(binPath, binaryName)
}

// StandardGetPath provides standard path resolution with system tool support
func (b *BaseTool) StandardGetPath(version string, cfg config.ToolConfig, getInstalledPath func(string, config.ToolConfig) (string, error), binaryName string) (string, error) {
	return b.StandardGetPathWithOptions(version, cfg, getInstalledPath, binaryName, nil, nil)
}

// StandardGetPathWithOptions provides standard path resolution with system tool support and options
func (b *BaseTool) StandardGetPathWithOptions(version string, cfg config.ToolConfig, getInstalledPath func(string, config.ToolConfig) (string, error), binaryName string, alternativeNames []string, envVars []string) (string, error) {
	// Check if we should use system tool instead of mvx-managed tool
	if UseSystemTool(b.toolName) {
		// Try environment variables first (for tools like Maven that need MAVEN_HOME)
		for _, envVar := range envVars {
			if envValue := os.Getenv(envVar); envValue != "" {
				// Verify the tool exists in this environment path
				toolPath := filepath.Join(envValue, "bin", binaryName)
				if runtime.GOOS == "windows" && !strings.HasSuffix(binaryName, ".exe") && !strings.HasSuffix(binaryName, ".cmd") {
					if _, err := os.Stat(toolPath + ".cmd"); err == nil {
						logVerbose("Using system %s from %s (MVX_USE_SYSTEM_%s=true)", b.toolName, envVar, strings.ToUpper(b.toolName))
						return "", nil
					}
					if _, err := os.Stat(toolPath + ".exe"); err == nil {
						logVerbose("Using system %s from %s (MVX_USE_SYSTEM_%s=true)", b.toolName, envVar, strings.ToUpper(b.toolName))
						return "", nil
					}
				}
				if _, err := os.Stat(toolPath); err == nil {
					logVerbose("Using system %s from %s (MVX_USE_SYSTEM_%s=true)", b.toolName, envVar, strings.ToUpper(b.toolName))
					return "", nil
				}
			}
		}

		// Try primary binary name in PATH
		if _, err := exec.LookPath(binaryName); err == nil {
			logVerbose("Using system %s from PATH (MVX_USE_SYSTEM_%s=true)", b.toolName, strings.ToUpper(b.toolName))
			return "", nil
		}

		// Try alternative binary names in PATH
		for _, altName := range alternativeNames {
			if _, err := exec.LookPath(altName); err == nil {
				logVerbose("Using system %s from PATH as %s (MVX_USE_SYSTEM_%s=true)", b.toolName, altName, strings.ToUpper(b.toolName))
				return "", nil
			}
		}

		return "", fmt.Errorf("MVX_USE_SYSTEM_%s=true but system %s not available", strings.ToUpper(b.toolName), b.toolName)
	}

	// Use mvx-managed tool path
	return getInstalledPath(version, cfg)
}
