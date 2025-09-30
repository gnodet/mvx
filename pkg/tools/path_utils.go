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
		installDir = filepath.Join(ipm.toolsDir, toolName, version+"@"+distribution)
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

// FindBinaryParentDir recursively searches for any binary in binaryNames under rootDir.
// Returns the parent directory of the first found binary.
func (r *PathResolver) FindBinaryParentDir(rootDir string, binaryName string) (string, error) {
	var foundPath string
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if info.Name() == binaryName {
				foundPath = filepath.Dir(path)
				return filepath.SkipDir // Stop walking once found
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if foundPath == "" {
		return "", os.ErrNotExist
	}
	return foundPath, nil
}
