package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PathResolver provides common path resolution utilities for tools
type PathResolver struct {
	toolsDir string
}

// NewPathResolver creates a new path resolver
func NewPathResolver(toolsDir string) *PathResolver {
	return &PathResolver{
		toolsDir: toolsDir,
	}
}

// DirectorySearchOptions configures directory search behavior
type DirectorySearchOptions struct {
	// Prefix to match directory names (e.g., "apache-maven-", "node-")
	DirectoryPrefix string
	// Subdirectory to look for within matched directories (e.g., "bin")
	BinSubdirectory string
	// Binary name to verify exists (e.g., "mvn", "node")
	BinaryName string
	// Whether to use platform-specific binary extensions (.exe, .cmd)
	UsePlatformExtensions bool
	// Preferred Windows extension (e.g., ".cmd" for Maven tools, ".exe" for others)
	PreferredWindowsExtension string
	// Fallback to parent directory if bin subdirectory not found
	FallbackToParent bool
}

// FindToolBinaryPath searches for a tool's binary path in an installation directory
func (pr *PathResolver) FindToolBinaryPath(installDir string, options DirectorySearchOptions) (string, error) {
	entries, err := os.ReadDir(installDir)
	if err != nil {
		return "", fmt.Errorf("failed to read installation directory: %w", err)
	}

	// Search for directories matching the prefix
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), options.DirectoryPrefix) {
			toolHome := filepath.Join(installDir, entry.Name())

			// Check for bin subdirectory
			if options.BinSubdirectory != "" {
				binPath := filepath.Join(toolHome, options.BinSubdirectory)
				if pr.verifyBinaryExists(binPath, options.BinaryName, options) {
					return binPath, nil
				}
			}

			// Fallback to parent directory if enabled
			if options.FallbackToParent {
				if pr.verifyBinaryExists(toolHome, options.BinaryName, options) {
					return toolHome, nil
				}
			}
		}
	}

	// If no prefixed directory found, check direct installation
	if options.BinSubdirectory != "" {
		binPath := filepath.Join(installDir, options.BinSubdirectory)
		if pr.verifyBinaryExists(binPath, options.BinaryName, options) {
			return binPath, nil
		}
	}

	// Final fallback to install directory itself
	if options.FallbackToParent {
		if pr.verifyBinaryExists(installDir, options.BinaryName, options) {
			return installDir, nil
		}
	}

	return "", fmt.Errorf("binary %s not found in installation directory", options.BinaryName)
}

// verifyBinaryExists checks if a binary exists in the given directory
func (pr *PathResolver) verifyBinaryExists(binPath, binaryName string, options DirectorySearchOptions) bool {
	if binaryName == "" {
		// If no binary name specified, just check if directory exists
		if info, err := os.Stat(binPath); err == nil && info.IsDir() {
			return true
		}
		return false
	}

	platformMapper := NewPlatformMapper()

	// Try exact binary name first
	fullPath := filepath.Join(binPath, binaryName)
	if _, err := os.Stat(fullPath); err == nil {
		return true
	}

	// Try platform-specific extensions if enabled
	if options.UsePlatformExtensions && platformMapper.IsWindows() {
		if !strings.HasSuffix(binaryName, ".exe") && !strings.HasSuffix(binaryName, ".cmd") {
			// Try preferred extension first if specified
			if options.PreferredWindowsExtension != "" {
				preferredPath := fullPath + options.PreferredWindowsExtension
				if _, err := os.Stat(preferredPath); err == nil {
					return true
				}
			}

			// Try .exe as fallback
			if options.PreferredWindowsExtension != ".exe" {
				exePath := fullPath + ".exe"
				if _, err := os.Stat(exePath); err == nil {
					return true
				}
			}
		}
	}

	return false
}

// PathEnvironmentManager provides utilities for managing PATH environment variable
type PathEnvironmentManager struct{}

// NewPathEnvironmentManager creates a new PATH environment manager
func NewPathEnvironmentManager() *PathEnvironmentManager {
	return &PathEnvironmentManager{}
}

// PrependToPATH prepends directories to the PATH environment variable
func (pem *PathEnvironmentManager) PrependToPATH(pathDirs []string, currentPATH string) string {
	if len(pathDirs) == 0 {
		return currentPATH
	}

	newPath := strings.Join(pathDirs, string(os.PathListSeparator))
	if currentPATH != "" {
		newPath = newPath + string(os.PathListSeparator) + currentPATH
	}

	return newPath
}

// BuildToolPATH builds PATH entries for configured tools
func (pem *PathEnvironmentManager) BuildToolPATH(tools map[string]ToolPathInfo) []string {
	var pathDirs []string

	for toolName, info := range tools {
		if info.BinPath != "" {
			logVerbose("Adding %s bin path to PATH: %s", toolName, info.BinPath)
			pathDirs = append(pathDirs, info.BinPath)
		}
	}

	return pathDirs
}

// ToolPathInfo contains path information for a tool
type ToolPathInfo struct {
	BinPath      string
	Version      string
	IsSystemTool bool
}

// InstallationPathManager provides utilities for managing tool installation paths
type InstallationPathManager struct {
	toolsDir string
}

// NewInstallationPathManager creates a new installation path manager
func NewInstallationPathManager(toolsDir string) *InstallationPathManager {
	return &InstallationPathManager{
		toolsDir: toolsDir,
	}
}

// CreateToolInstallDir creates an installation directory for a tool version
func (ipm *InstallationPathManager) CreateToolInstallDir(toolName, version, distribution string) (string, error) {
	var installDir string
	if distribution != "" {
		installDir = filepath.Join(ipm.toolsDir, toolName, version+"-"+distribution)
	} else {
		installDir = filepath.Join(ipm.toolsDir, toolName, version)
	}

	if err := os.MkdirAll(installDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create installation directory: %w", err)
	}

	return installDir, nil
}

// GetToolVersionDir returns the installation directory for a tool version
func (ipm *InstallationPathManager) GetToolVersionDir(toolName, version, distribution string) string {
	if distribution != "" {
		return filepath.Join(ipm.toolsDir, toolName, version+"-"+distribution)
	}
	return filepath.Join(ipm.toolsDir, toolName, version)
}

// ListInstalledVersions returns a list of installed versions for a tool
func (ipm *InstallationPathManager) ListInstalledVersions(toolName string) ([]string, error) {
	toolDir := filepath.Join(ipm.toolsDir, toolName)
	entries, err := os.ReadDir(toolDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil // No versions installed
		}
		return nil, fmt.Errorf("failed to read tool directory: %w", err)
	}

	var versions []string
	for _, entry := range entries {
		if entry.IsDir() {
			versions = append(versions, entry.Name())
		}
	}

	return versions, nil
}

// CleanupFailedInstallation removes a failed installation directory
func (ipm *InstallationPathManager) CleanupFailedInstallation(installDir string) error {
	if installDir == "" || installDir == "/" || installDir == ipm.toolsDir {
		return fmt.Errorf("refusing to remove critical directory: %s", installDir)
	}

	// Additional safety check - ensure it's within tools directory
	if !strings.HasPrefix(installDir, ipm.toolsDir) {
		return fmt.Errorf("installation directory is outside tools directory: %s", installDir)
	}

	return os.RemoveAll(installDir)
}
