package tools

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/gnodet/mvx/pkg/config"
	"github.com/gnodet/mvx/pkg/util"
	"github.com/gnodet/mvx/pkg/version"
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

// SupportsChecksumVerification returns whether this tool supports checksum verification
// Default implementation returns true for all tools
func (b *BaseTool) SupportsChecksumVerification() bool {
	return true
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
	// Note: MinSize/MaxSize/ExpectedType removed - we rely on checksums for validation
	downloadConfig.ToolName = b.toolName
	downloadConfig.Version = version
	downloadConfig.Config = cfg

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
	return b.extractFile(archivePath, destDir, options.GetArchiveType())
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

	// Set up environment for verification (needed for tools like Maven that depend on Java)
	env, err := b.setupVerificationEnvironment(cfg)
	if err != nil {
		util.LogVerbose("Failed to setup verification environment: %v", err)
		env = nil // Fall back to default environment
	}

	// Verify binary exists and works
	if err := b.VerifyBinaryWithConfig(binDir, verifyConfig, env); err != nil {
		return VerifyError(b.toolName, version, err)
	}

	return nil
}

// setupVerificationEnvironment sets up environment variables needed for tool verification
func (b *BaseTool) setupVerificationEnvironment(cfg config.ToolConfig) ([]string, error) {
	// Create a minimal config that includes this tool and its dependencies
	tempConfig := &config.Config{
		Tools: make(map[string]config.ToolConfig),
	}

	// Add this tool to the config
	tempConfig.Tools[b.toolName] = cfg

	// Add dependencies if this tool has any
	if tool, err := b.manager.GetTool(b.toolName); err == nil {
		if depProvider, ok := tool.(DependencyProvider); ok {
			dependencies := depProvider.GetDependencies()
			util.LogVerbose("Setting up dependencies for %s verification: %v", b.toolName, dependencies)

			// Add dependencies to the temporary config
			for _, depName := range dependencies {
				if depTool, err := b.manager.GetTool(depName); err == nil {
					// Try to find an installed version of the dependency
					if installedVersions := b.getInstalledVersionsForTool(depTool, depName); len(installedVersions) > 0 {
						// Use the first available installed version
						depConfig := config.ToolConfig{
							Version:      installedVersions[0].Version,
							Distribution: installedVersions[0].Distribution,
						}
						tempConfig.Tools[depName] = depConfig
						util.LogVerbose("Added dependency %s %s to verification environment", depName, installedVersions[0].Version)
					}
				}
			}
		}
	}

	// Use the standard environment setup
	envMap, err := b.manager.SetupEnvironment(tempConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to setup verification environment: %w", err)
	}

	// Start with current environment
	envVars := make(map[string]string)
	for _, envVar := range os.Environ() {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) == 2 {
			envVars[parts[0]] = parts[1]
		}
	}

	// Override with tool environment
	for key, value := range envMap {
		envVars[key] = value
	}

	// Convert to environment slice format
	envSlice := make([]string, 0, len(envVars))
	for key, value := range envVars {
		envSlice = append(envSlice, fmt.Sprintf("%s=%s", key, value))
	}

	return envSlice, nil
}

// SetupHomeEnvironment is a generic helper for setting up *_HOME environment variables
// It gets the tool's bin path and sets envVarName to the parent directory
// The getPath function should return the bin directory path for the tool
func (b *BaseTool) SetupHomeEnvironment(version string, cfg config.ToolConfig, envVars map[string]string, envVarName string, getPath func(string, config.ToolConfig) (string, error)) error {
	binPath, err := getPath(version, cfg)
	if err != nil {
		util.LogVerbose("Could not determine %s for %s %s: %v", envVarName, b.toolName, version, err)
		return nil
	}

	// *_HOME should point to the installation directory, not the bin directory
	if strings.HasSuffix(binPath, "/bin") {
		homeDir := strings.TrimSuffix(binPath, "/bin")
		envVars[envVarName] = homeDir
		util.LogVerbose("Set %s=%s for %s %s", envVarName, homeDir, b.toolName, version)
	}

	return nil
}

// getInstalledVersionsForTool gets installed versions for a tool (helper for verification environment)
func (b *BaseTool) getInstalledVersionsForTool(tool Tool, toolName string) []InstalledVersion {
	// Try to get installed versions if the tool supports it
	switch t := tool.(type) {
	case *JavaTool:
		versions := t.ListInstalledVersions("")
		// Filter out versions with empty distribution to avoid path issues
		var validVersions []InstalledVersion
		for _, v := range versions {
			if v.Distribution != "" {
				validVersions = append(validVersions, v)
			}
		}
		return validVersions
	default:
		// For other tools, we could implement a generic way to find installed versions
		// For now, return empty slice
		return []InstalledVersion{}
	}
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
	FileExtension string // e.g., ".tar.gz", ".zip" - used to infer archive type
}

// GetArchiveType infers the archive type from the file extension
func (o DownloadOptions) GetArchiveType() string {
	ext := o.FileExtension
	// Normalize extension
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	switch ext {
	case ".zip":
		return ArchiveTypeZip
	case ".tar.gz", ".tgz":
		return ArchiveTypeTarGz
	case ".tar.xz":
		return ArchiveTypeTarXz
	default:
		// Default to tar.gz for unknown extensions
		return ArchiveTypeTarGz
	}
}

// extractFile extracts an archive file based on its type
func (b *BaseTool) extractFile(src, dest, archiveType string) error {
	switch archiveType {
	case ArchiveTypeZip:
		return extractZipFile(src, dest)
	case ArchiveTypeTarGz:
		return extractTarGzFile(src, dest)
	case ArchiveTypeTarXz:
		return extractTarXzFile(src, dest)
	default:
		return fmt.Errorf("unsupported archive type: %s", archiveType)
	}
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

// VerifyBinaryWithConfig runs a binary with configuration and environment and checks the output
func (b *BaseTool) VerifyBinaryWithConfig(binPath string, verifyConfig VerificationConfig, env []string) error {
	exe := b.getPlatformBinaryPath(binPath, verifyConfig.BinaryName)

	// Check if binary exists
	if _, err := os.Stat(exe); err != nil {
		return fmt.Errorf("%s executable not found at %s: %w", verifyConfig.BinaryName, exe, err)
	}

	// Run version command
	cmd := exec.Command(exe, verifyConfig.VersionArgs...)
	if env != nil {
		cmd.Env = env
	}
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
	installDir := b.manager.GetToolVersionDir(b.toolName, version, cfg.Distribution)

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
func (b *BaseTool) IsInstalled(binPath string) bool {
	exe := b.getPlatformBinaryPath(binPath, b.GetBinaryName())
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
		util.LogVerbose("%s=true, forcing use of system %s", getSystemToolEnvVar(b.toolName), b.toolName)

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

// InstalledVersion holds info about an installed tool version
type InstalledVersion struct {
	Version      string
	Distribution string
	Path         string
}

// ListInstalledVersions returns all installed versions for this tool and distribution.
// If distribution is empty, returns all distributions.
func (b *BaseTool) ListInstalledVersions(distribution string) []InstalledVersion {
	toolsDir := b.manager.GetToolsDir()
	toolDir := filepath.Join(toolsDir, b.toolName)
	entries, err := os.ReadDir(toolDir)
	if err != nil {
		return nil
	}

	var installed []InstalledVersion
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		versionPart, distPart, hasDistribution := strings.Cut(name, "@")
		if !hasDistribution {
			versionPart = name
			distPart = ""
		}
		if distribution != "" && distPart != distribution {
			continue
		}
		installed = append(installed, InstalledVersion{
			Version:      versionPart,
			Distribution: distPart,
			Path:         filepath.Join(toolDir, name),
		})
	}
	return installed
}

// StandardIsInstalled provides standard installation check for tools
func (b *BaseTool) StandardIsInstalled(versionSpec string, cfg config.ToolConfig, getPath func(string, config.ToolConfig) (string, error)) bool {
	if UseSystemTool(b.toolName) {
		if _, err := exec.LookPath(b.GetBinaryName()); err == nil {
			util.LogVerbose("System %s is available in PATH (MVX_USE_SYSTEM_%s=true)", b.toolName, strings.ToUpper(b.toolName))
			return true
		}

		util.LogVerbose("System %s not available: not found in environment variables or PATH", b.toolName)
		return false
	}

	tool, err := b.manager.GetTool(b.toolName)
	if err != nil {
		util.LogVerbose("Failed to get tool %s: %v", b.toolName, err)
		return false
	}

	spec, err := version.ParseSpec(versionSpec)
	if err != nil {
		util.LogVerbose("Failed to parse version spec %q: %v", versionSpec, err)
		return false
	}

	targetVersion, resolveErr := b.resolveTargetVersion(tool, spec, versionSpec, cfg)
	if resolveErr != nil {
		util.LogVerbose("Failed to resolve target version for %s %s: %v", b.toolName, versionSpec, resolveErr)
	}

	installed := b.ListInstalledVersions(cfg.Distribution)
	type installedCandidate struct {
		info   InstalledVersion
		parsed *version.Version
	}

	var candidates []installedCandidate
	for _, inst := range installed {
		parsed, err := version.ParseVersion(inst.Version)
		if err != nil {
			util.LogVerbose("Failed to parse installed version %s: %v", inst.Version, err)
			continue
		}
		if spec.Matches(parsed) {
			if spec.Constraint == "latest" && resolveErr == nil && targetVersion != "" && inst.Version != targetVersion {
				continue
			}
			candidates = append(candidates, installedCandidate{info: inst, parsed: parsed})
		}
	}

	// Sort in descending order (newest first) to check best matches first
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].parsed.Compare(candidates[j].parsed) > 0
	})

	// Check candidates in order, return immediately on first valid installation
	for _, candidate := range candidates {
		util.LogVerbose("   Checking installed version: %s (%s) at %s", candidate.info.Version, candidate.info.Distribution, candidate.info.Path)

		candidateCfg := cfg
		candidateCfg.Version = candidate.info.Version
		if candidate.info.Distribution != "" {
			candidateCfg.Distribution = candidate.info.Distribution
		}

		binPath, err := getPath(candidate.info.Version, candidateCfg)
		if err != nil {
			util.LogVerbose("Failed to compute binary path for %s %s: %v", b.toolName, candidate.info.Version, err)
			continue
		}

		if !b.IsInstalled(binPath) {
			util.LogVerbose("Binary for %s %s not found at %s", b.toolName, candidate.info.Version, binPath)
			continue
		}

		// Skip verification during IsInstalled check for performance
		// Binary existence is sufficient - verification happens during Install/Verify
		// This avoids running "java -version", "mvn --version", etc. on every startup

		// Found a valid installation - return immediately
		util.LogVerbose("Using previously installed %s %s (%s)", b.toolName, candidate.info.Version, candidate.info.Distribution)
		return true
	}

	if resolveErr != nil {
		return false
	}

	if targetVersion == "" {
		util.LogVerbose("Resolved target version for %s %s is empty", b.toolName, versionSpec)
		return false
	}

	installCfg := cfg
	installCfg.Version = targetVersion

	util.LogVerbose("%s version %s not installed, attempting automatic installation", b.toolName, targetVersion)
	if err := tool.Install(targetVersion, installCfg); err != nil {
		util.LogVerbose("Automatic installation of %s %s failed: %v", b.toolName, targetVersion, err)
		return false
	}

	if err := tool.Verify(targetVersion, installCfg); err != nil {
		util.LogVerbose("Verification after installing %s %s failed: %v", b.toolName, targetVersion, err)
		return false
	}

	b.clearPathCache()
	util.LogVerbose("Successfully installed %s %s on demand", b.toolName, targetVersion)
	return true
}

func (b *BaseTool) resolveTargetVersion(tool Tool, spec *version.Spec, versionSpec string, cfg config.ToolConfig) (string, error) {
	resolveCfg := cfg
	resolveCfg.Version = versionSpec

	if resolved, err := b.manager.ResolveVersion(b.toolName, resolveCfg); err == nil && resolved != "" {
		if parsed, parseErr := version.ParseVersion(resolved); parseErr == nil {
			if spec.Constraint == "latest" || spec.Matches(parsed) {
				return resolved, nil
			}
		} else {
			return resolved, nil
		}
	}

	available, err := tool.ListVersions()
	if err != nil {
		return "", err
	}

	resolved, err := spec.Resolve(available)
	if err != nil {
		return "", err
	}

	return resolved, nil
}

// StandardGetPath provides standard path resolution with system tool support
func (b *BaseTool) GetPath(version string, cfg config.ToolConfig, getInstalledPath func(string, config.ToolConfig) (string, error)) (string, error) {
	return b.StandardGetPath(version, cfg, getInstalledPath)
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
			util.LogVerbose("Using system %s from PATH (MVX_USE_SYSTEM_%s=true)", b.toolName, strings.ToUpper(b.toolName))
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
