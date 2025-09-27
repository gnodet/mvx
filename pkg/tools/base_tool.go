package tools

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gnodet/mvx/pkg/config"
)

// getUserFriendlyURL converts long redirect URLs to user-friendly display URLs
func getUserFriendlyURL(inputURL string) string {
	// Handle GitHub release asset URLs
	if strings.Contains(inputURL, "release-assets.githubusercontent.com") {
		// Extract filename from response-content-disposition parameter
		if strings.Contains(inputURL, "response-content-disposition=attachment") {
			if start := strings.Index(inputURL, "filename%3D"); start != -1 {
				start += len("filename%3D")
				if end := strings.Index(inputURL[start:], "&"); end != -1 {
					filename := inputURL[start : start+end]
					// URL decode the filename
					if decoded, err := url.QueryUnescape(filename); err == nil {
						return fmt.Sprintf("github.com/.../releases/%s", decoded)
					}
				}
			}
		}
		return "github.com/.../releases/[asset]"
	}

	// Handle other long URLs with query parameters
	if len(inputURL) > 80 && strings.Contains(inputURL, "?") {
		baseURL := strings.Split(inputURL, "?")[0]
		if len(baseURL) > 50 {
			// Extract domain and path
			if parsed, err := url.Parse(baseURL); err == nil {
				if len(parsed.Path) > 30 {
					// Show domain + shortened path
					pathParts := strings.Split(parsed.Path, "/")
					if len(pathParts) > 3 {
						return fmt.Sprintf("%s/.../%s", parsed.Host, pathParts[len(pathParts)-1])
					}
				}
				return fmt.Sprintf("%s%s", parsed.Host, parsed.Path)
			}
		}
		return baseURL
	}

	// For normal URLs, return as-is
	return inputURL
}

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

// getPlatformBinaryPath returns the platform-specific binary path
func (b *BaseTool) getPlatformBinaryPath(binPath, binaryName string) string {
	return b.getPlatformBinaryPathWithExtension(binPath, binaryName, ".exe")
}

// getPlatformBinaryPathWithExtension returns the platform-specific binary path with custom extension
func (b *BaseTool) getPlatformBinaryPathWithExtension(binPath, binaryName, windowsExt string) string {
	platformMapper := NewPlatformMapper()

	// Add platform-specific extension if needed
	if platformMapper.IsWindows() && binaryName != "" {
		if !strings.HasSuffix(binaryName, ".exe") && !strings.HasSuffix(binaryName, ".cmd") {
			binaryName = binaryName + windowsExt
		}
	}

	return filepath.Join(binPath, binaryName)
}

// checkSystemBinaryExists checks if a system binary exists with platform-specific extensions
func (b *BaseTool) checkSystemBinaryExists(basePath, binaryName string) (bool, string) {
	return b.checkSystemBinaryExistsWithExtensions(basePath, binaryName, []string{".exe", ".cmd"})
}

// checkSystemBinaryExistsWithExtensions checks if a system binary exists with custom extensions
func (b *BaseTool) checkSystemBinaryExistsWithExtensions(basePath, binaryName string, windowsExtensions []string) (bool, string) {
	platformMapper := NewPlatformMapper()

	// Try the exact binary name first
	fullPath := filepath.Join(basePath, binaryName)
	if _, err := os.Stat(fullPath); err == nil {
		return true, fullPath
	}

	// On Windows, try specified extensions
	if platformMapper.IsWindows() && !strings.HasSuffix(binaryName, ".exe") && !strings.HasSuffix(binaryName, ".cmd") {
		for _, ext := range windowsExtensions {
			extPath := fullPath + ext
			if _, err := os.Stat(extPath); err == nil {
				return true, extPath
			}
		}
	}

	return false, ""
}

// wrapError wraps an error with tool context using standardized error types
func (b *BaseTool) wrapError(operation, version string, err error) error {
	return WrapError(b.toolName, version, operation, err)
}

// CreateInstallDir creates the installation directory for a tool version
func (b *BaseTool) CreateInstallDir(version, distribution string) (string, error) {
	pathManager := NewInstallationPathManager(b.manager.GetToolsDir())
	return pathManager.CreateToolInstallDir(b.toolName, version, distribution)
}

// Download performs a robust download with checksum verification
func (b *BaseTool) Download(url, version string, cfg config.ToolConfig, options DownloadOptions) (string, error) {
	// Create temporary file for download
	tmpFile, err := os.CreateTemp("", fmt.Sprintf("%s-*%s", b.toolName, options.FileExtension))
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}
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

	// Get the tool instance for checksum verification
	if tool, err := b.manager.GetTool(b.toolName); err == nil {
		downloadConfig.Tool = tool
	}

	// Perform robust download with checksum verification
	result, err := RobustDownload(downloadConfig)
	if err != nil {
		os.Remove(tmpFile.Name()) // Clean up on failure
		return "", fmt.Errorf("%s download failed: %s", strings.Title(b.toolName), DiagnoseDownloadError(url, err))
	}

	// Show user-friendly URL instead of long redirect URLs
	displayURL := getUserFriendlyURL(result.FinalURL)
	fmt.Printf("  📦 Downloaded %d bytes from %s\n", result.Size, displayURL)

	// Return the path to the downloaded file
	return tmpFile.Name(), nil
}

// Extract extracts an archive file to the destination directory
func (b *BaseTool) Extract(archivePath, destDir string, options DownloadOptions) error {
	return b.extractFile(archivePath, destDir, options.ArchiveType)
}

// VerificationConfig contains configuration for tool verification
type VerificationConfig struct {
	BinaryName      string   // Primary binary name to verify
	VersionArgs     []string // Arguments to get version (e.g., ["--version"])
	ExpectedVersion string   // Expected version string in output
	DebugInfo       bool     // Whether to show debug information on failure
}

// Verify performs post-installation verification of a tool
func (b *BaseTool) Verify(installDir, version string, cfg config.ToolConfig) error {
	// Default verification: check if installation directory exists and is not empty
	if _, err := os.Stat(installDir); os.IsNotExist(err) {
		return fmt.Errorf("installation directory does not exist: %s", installDir)
	}

	// Check if directory is not empty
	entries, err := os.ReadDir(installDir)
	if err != nil {
		return fmt.Errorf("failed to read installation directory: %w", err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("installation directory is empty: %s", installDir)
	}

	return nil
}

// VerifyWithConfig performs comprehensive verification using configuration
func (b *BaseTool) VerifyWithConfig(version string, cfg config.ToolConfig, verifyConfig VerificationConfig) error {
	// Get tool path
	tool, err := b.manager.GetTool(b.toolName)
	if err != nil {
		return VerifyError(b.toolName, version, fmt.Errorf("failed to get tool: %w", err))
	}

	// Get path using tool's GetPath method
	var binPath string
	if pathProvider, hasGetPath := tool.(interface {
		GetPath(string, config.ToolConfig) (string, error)
	}); hasGetPath {
		binPath, err = pathProvider.GetPath(version, cfg)
		if err != nil {
			if verifyConfig.DebugInfo {
				b.printVerificationDebugInfo(version, cfg, err)
			}
			return VerifyError(b.toolName, version, fmt.Errorf("failed to get binary path: %w", err))
		}
	} else {
		return VerifyError(b.toolName, version, fmt.Errorf("tool does not implement GetPath method"))
	}

	// Verify binary exists and works
	if err := b.VerifyBinaryWithConfig(binPath, verifyConfig); err != nil {
		return VerifyError(b.toolName, version, err)
	}

	return nil
}

// DownloadAndExtract performs a robust download with checksum verification and extraction
// This is a convenience method that combines Download and Extract
func (b *BaseTool) DownloadAndExtract(url, destDir, version string, cfg config.ToolConfig, options DownloadOptions) error {
	// Download the file
	archivePath, err := b.Download(url, version, cfg, options)
	if err != nil {
		return err
	}
	defer os.Remove(archivePath) // Clean up downloaded file after extraction

	// Extract the file
	return b.Extract(archivePath, destDir, options)
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
	exe := b.getPlatformBinaryPath(binPath, binaryName)

	// Run version command
	cmd := exec.Command(exe, versionArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s verification failed: %w\nOutput: %s", b.toolName, err, output)
	}

	return nil
}

// VerifyBinaryWithConfig runs a binary with configuration and checks the output
func (b *BaseTool) VerifyBinaryWithConfig(binPath string, verifyConfig VerificationConfig) error {
	exe := b.getPlatformBinaryPath(binPath, verifyConfig.BinaryName)

	// Check if binary exists
	if _, err := os.Stat(exe); err != nil {
		return fmt.Errorf("%s executable not found at %s: %w", verifyConfig.BinaryName, exe, err)
	}

	// Run version command
	cmd := exec.Command(exe, verifyConfig.VersionArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s verification failed: %w\nOutput: %s", b.toolName, err, output)
	}

	// Check version if expected version is provided
	if verifyConfig.ExpectedVersion != "" {
		outputStr := string(output)
		if !strings.Contains(outputStr, verifyConfig.ExpectedVersion) {
			return fmt.Errorf("%s version mismatch: expected %s, got %s", b.toolName, verifyConfig.ExpectedVersion, outputStr)
		}
	}

	return nil
}

// printVerificationDebugInfo prints detailed debug information for verification failures
func (b *BaseTool) printVerificationDebugInfo(version string, cfg config.ToolConfig, pathErr error) {
	fmt.Printf("  🔍 Debug: %s installation verification failed\n", b.toolName)

	// Try to determine install directory
	installDir := ""
	if cfg.Distribution != "" {
		installDir = b.manager.GetToolVersionDir(b.toolName, version, cfg.Distribution)
	} else {
		installDir = b.manager.GetToolVersionDir(b.toolName, version, "")
	}

	fmt.Printf("     Install directory: %s\n", installDir)
	fmt.Printf("     Error getting bin path: %v\n", pathErr)

	// List contents of install directory for debugging
	if entries, readErr := os.ReadDir(installDir); readErr == nil {
		fmt.Printf("     Install directory contents:\n")
		for _, entry := range entries {
			fmt.Printf("       - %s (dir: %t)\n", entry.Name(), entry.IsDir())
		}
	}
}

// IsInstalled checks if a binary exists at the expected path
func (b *BaseTool) IsInstalled(binPath, binaryName string) bool {
	exe := b.getPlatformBinaryPath(binPath, binaryName)
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
		MinSize:       100 * 1024,        // 100KB (very permissive, rely on checksums)
		MaxSize:       500 * 1024 * 1024, // 500MB (generous upper bound)
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
				binPath := filepath.Join(envValue, "bin")
				if exists, _ := b.checkSystemBinaryExists(binPath, b.toolName); exists {
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

	// Get tool-specific download options
	options := b.getDownloadOptions()

	// Download the file
	archivePath, err := b.Download(downloadURL, version, cfg, options)
	if err != nil {
		return InstallError(b.toolName, version, err)
	}
	defer os.Remove(archivePath) // Clean up downloaded file

	// Extract the file
	if err := b.Extract(archivePath, installDir, options); err != nil {
		return InstallError(b.toolName, version, err)
	}

	// Verify installation was successful using tool's Verify method
	if tool, err := b.manager.GetTool(b.toolName); err == nil {
		if verifier, hasVerify := tool.(interface {
			Verify(string, config.ToolConfig) error
		}); hasVerify {
			if err := verifier.Verify(version, cfg); err != nil {
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
	}

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

// StandardVerifyWithConfig provides standard verification using VerificationConfig
func (b *BaseTool) StandardVerifyWithConfig(version string, cfg config.ToolConfig, verifyConfig VerificationConfig) error {
	return b.VerifyWithConfig(version, cfg, verifyConfig)
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

		return "", SystemToolError(b.toolName, fmt.Errorf("MVX_USE_SYSTEM_%s=true but system %s not available", strings.ToUpper(b.toolName), b.toolName))
	}

	// Use mvx-managed tool path
	return getInstalledPath(version, cfg)
}
