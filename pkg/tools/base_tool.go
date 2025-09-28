package tools

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

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

// pathCacheEntry represents a cached path result
type pathCacheEntry struct {
	path string
	err  error
}

// BaseTool provides common functionality for all tools
type BaseTool struct {
	manager    *Manager
	toolName   string
	binaryName string
	pathCache  map[string]pathCacheEntry
	cacheMux   sync.RWMutex
}

// NewBaseTool creates a new base tool instance
func NewBaseTool(manager *Manager, toolName string, binaryName string) *BaseTool {
	return &BaseTool{
		manager:    manager,
		toolName:   toolName,
		binaryName: binaryName,
		pathCache:  make(map[string]pathCacheEntry),
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

func (b *BaseTool) GetBinaryName() string {
	return b.binaryName
}

// getCacheKey generates a cache key for path operations
func (b *BaseTool) getCacheKey(version string, cfg config.ToolConfig, operation string) string {
	return fmt.Sprintf("%s:%s:%s:%s", operation, version, cfg.Distribution, b.toolName)
}

// getCachedPath retrieves a cached path result
func (b *BaseTool) getCachedPath(cacheKey string) (string, error, bool) {
	b.cacheMux.RLock()
	defer b.cacheMux.RUnlock()

	if entry, exists := b.pathCache[cacheKey]; exists {
		return entry.path, entry.err, true
	}
	return "", nil, false
}

// setCachedPath stores a path result in cache
func (b *BaseTool) setCachedPath(cacheKey string, path string, err error) {
	b.cacheMux.Lock()
	defer b.cacheMux.Unlock()

	b.pathCache[cacheKey] = pathCacheEntry{
		path: path,
		err:  err,
	}
}

// clearPathCache clears the path cache (useful for testing)
func (b *BaseTool) clearPathCache() {
	b.cacheMux.Lock()
	defer b.cacheMux.Unlock()

	b.pathCache = make(map[string]pathCacheEntry)
}

// ClearPathCache clears the path cache (public method for CacheManager interface)
func (b *BaseTool) ClearPathCache() {
	b.clearPathCache()
}

// getPlatformBinaryPath returns the platform-specific binary path
func (b *BaseTool) getPlatformBinaryPath(binPath, binaryName string) string {
	return b.getPlatformBinaryPathWithExtension(binPath, binaryName)
}

// getPlatformBinaryPathWithExtension returns the platform-specific binary path with custom extension
func (b *BaseTool) getPlatformBinaryPathWithExtension(binPath, binaryName string) string {
	return filepath.Join(binPath, binaryName)
}

// checkSystemBinaryExists checks if a system binary exists with platform-specific extensions
func (b *BaseTool) checkSystemBinaryExists(basePath, binaryName string) (bool, string) {
	// Try the exact binary name first
	fullPath := filepath.Join(basePath, binaryName)
	if _, err := os.Stat(fullPath); err == nil {
		return true, fullPath
	}

	return false, ""
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
	BinaryName      string   // Binary name to verify
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
	_, err := b.manager.GetTool(b.toolName)
	if err != nil {
		return VerifyError(b.toolName, version, fmt.Errorf("failed to get tool: %w", err))
	}

	// Use FindBinaryParentDir to locate the binary directory
	installDir := b.manager.GetToolVersionDir(b.toolName, version, cfg.Distribution)
	binaryName := verifyConfig.BinaryName
	pathResolver := NewPathResolver(b.manager.GetToolsDir())
	binDir, err := pathResolver.FindBinaryParentDir(installDir, binaryName)
	if err != nil {
		if verifyConfig.DebugInfo {
			b.printVerificationDebugInfo(version, cfg, err)
		}
		return VerifyError(b.toolName, version, fmt.Errorf("failed to get binary path: %w", err))
	}

	// Verify binary exists and works
	if err := b.VerifyBinaryWithConfig(binDir, verifyConfig); err != nil {
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
		FileExtension: ExtTarGz,
		ExpectedType:  "application",
		MinSize:       100 * 1024,        // 100KB (very permissive, rely on checksums)
		MaxSize:       500 * 1024 * 1024, // 500MB (generous upper bound)
		ArchiveType:   ArchiveTypeTarGz,
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
	// Check if we should use system tool instead of downloading
	if UseSystemTool(b.toolName) {
		logVerbose("%s=true, forcing use of system %s", getSystemToolEnvVar(b.toolName), b.toolName)

		// Try primary binary name in PATH
		if toolPath, err := exec.LookPath(b.binaryName); err == nil {
			fmt.Printf("  🔗 Using system %s from PATH: %s\n", b.toolName, toolPath)
			fmt.Printf("  ✅ System %s configured (mvx will use system PATH)\n", b.toolName)
			return nil
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
				// if removeErr := os.RemoveAll(installDir); removeErr != nil {
				// 	fmt.Printf("  ⚠️  Warning: failed to clean up installation directory: %v\n", removeErr)
				// }
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
func (b *BaseTool) StandardIsInstalled(version string, cfg config.ToolConfig, getPath func(string, config.ToolConfig) (string, error)) bool {
	// Check if we should use system tool instead of mvx-managed tool
	if UseSystemTool(b.toolName) {
		// Try primary binary name in PATH
		if _, err := exec.LookPath(b.GetBinaryName()); err == nil {
			logVerbose("System %s is available in PATH (MVX_USE_SYSTEM_%s=true)", b.toolName, strings.ToUpper(b.toolName))
			return true
		}

		logVerbose("System %s not available: not found in environment variables or PATH", b.toolName)
		return false
	}

	binPath, err := getPath(version, cfg)
	if err != nil {
		logVerbose("Failed to get %s binary path: %v", b.toolName, err)
		return false
	}
	logVerbose("Checking for %s binary at: %s", b.toolName, binPath)
	return b.IsInstalled(binPath, b.GetBinaryName())
}

// StandardGetPath provides standard path resolution with system tool support
func (b *BaseTool) StandardGetPath(version string, cfg config.ToolConfig, getInstalledPath func(string, config.ToolConfig) (string, error)) (string, error) {
	// Check cache first
	cacheKey := b.getCacheKey(version, cfg, "getPath")
	if cachedPath, cachedErr, found := b.getCachedPath(cacheKey); found {
		return cachedPath, cachedErr
	}
	// Check if we should use system tool instead of mvx-managed tool
	if UseSystemTool(b.toolName) {
		// Try primary binary name in PATH
		if _, err := exec.LookPath(b.GetBinaryName()); err == nil {
			logVerbose("Using system %s from PATH (MVX_USE_SYSTEM_%s=true)", b.toolName, strings.ToUpper(b.toolName))
			b.setCachedPath(cacheKey, "", nil)
			return "", nil
		}

		systemErr := SystemToolError(b.toolName, fmt.Errorf("MVX_USE_SYSTEM_%s=true but system %s not available", strings.ToUpper(b.toolName), b.toolName))
		b.setCachedPath(cacheKey, "", systemErr)
		return "", systemErr
	}

	// Use mvx-managed tool path
	path, err := getInstalledPath(version, cfg)
	b.setCachedPath(cacheKey, path, err)
	return path, err
}
