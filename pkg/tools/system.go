package tools

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gnodet/mvx/pkg/util"
)

// toolHomeEnvVars maps tool names to their HOME environment variable names
var toolHomeEnvVars = map[string]string{
	ToolJava:  EnvJavaHome,
	ToolMaven: EnvMavenHome,
	ToolMvnd:  EnvMvndHome,
	ToolNode:  EnvNodeHome,
	ToolGo:    EnvGoRoot, // Go uses GOROOT instead of GO_HOME
}

// UseSystemTool checks if a system tool should be used instead of downloading
// by checking the MVX_USE_SYSTEM_<TOOL> environment variable
func UseSystemTool(toolName string) bool {
	envVar := fmt.Sprintf("MVX_USE_SYSTEM_%s", strings.ToUpper(toolName))
	return os.Getenv(envVar) == "true"
}

// getSystemToolEnvVar returns the environment variable name for a tool
func getSystemToolEnvVar(toolName string) string {
	return fmt.Sprintf("MVX_USE_SYSTEM_%s", strings.ToUpper(toolName))
}

// getSystemToolHomeEnvVar returns the HOME environment variable name for a tool, if any
func getSystemToolHomeEnvVar(toolName string) string {
	return toolHomeEnvVars[toolName]
}

// getSystemToolHome returns the system tool home directory if available and valid
// It first checks the HOME environment variable (e.g., JAVA_HOME, MAVEN_HOME), then falls back to finding the tool in PATH
// Returns the home directory and bin directory if found, or an error if not available
func getSystemToolHome(toolName, binaryName, homeEnvVar string) (homeDir string, binDir string, err error) {
	// First, try HOME environment variable if specified
	if homeEnvVar != "" {
		homeDir = os.Getenv(homeEnvVar)
		if homeDir != "" {
			// Check if HOME points to a valid installation
			binDirPath := filepath.Join(homeDir, "bin")
			binaryPath := filepath.Join(binDirPath, binaryName)
			if _, err := os.Stat(binaryPath); err == nil {
				util.LogVerbose("Found system %s via %s=%s", toolName, homeEnvVar, homeDir)
				return homeDir, binDirPath, nil
			}
			// HOME is set but doesn't point to a valid installation
			util.LogVerbose("%s is set to %s but %s executable not found at %s, trying PATH", homeEnvVar, homeDir, binaryName, binaryPath)
		}
	}

	// Fallback: try to find tool in PATH and derive home directory from it
	binaryPath, err := exec.LookPath(binaryName)
	if err != nil {
		if homeEnvVar != "" {
			return "", "", SystemToolError(toolName, fmt.Errorf("%s not found: %s environment variable not set or invalid and %s not found in PATH", toolName, homeEnvVar, binaryName))
		}
		return "", "", SystemToolError(toolName, fmt.Errorf("%s not found in PATH", binaryName))
	}

	// Resolve the absolute path (follow symlinks)
	absPath, err := filepath.EvalSymlinks(binaryPath)
	if err != nil {
		absPath = binaryPath
	}

	// Derive home directory from binary path
	// Binary is typically at $HOME/bin/binary (or bin/binary.exe on Windows)
	binDirPath := filepath.Dir(absPath)
	if filepath.Base(binDirPath) != "bin" {
		// Some tools might be installed directly without a bin directory structure
		// In this case, we can't derive a home directory, but we can still use the binary
		util.LogVerbose("Found %s at %s but not in a 'bin' directory, using directory as-is", toolName, absPath)
		return filepath.Dir(binDirPath), binDirPath, nil
	}

	homeDir = filepath.Dir(binDirPath)
	util.LogVerbose("Derived %s home directory=%s from binary at %s", toolName, homeDir, absPath)
	return homeDir, binDirPath, nil
}

// SystemToolInfo contains information about a detected system tool
type SystemToolInfo struct {
	Path    string // Full path to the tool executable
	Version string // Detected version (if available)
	Valid   bool   // Whether the tool is valid and usable
}

// DetectSystemTool performs enhanced detection of system tools
func DetectSystemTool(toolName string, alternativeNames []string) *SystemToolInfo {
	info := &SystemToolInfo{}

	// Try primary tool name first
	if path, err := exec.LookPath(toolName); err == nil {
		info.Path = path
		info.Valid = true
		return info
	}

	// Try alternative names
	for _, altName := range alternativeNames {
		if path, err := exec.LookPath(altName); err == nil {
			info.Path = path
			info.Valid = true
			return info
		}
	}

	return info
}

// ValidateSystemToolVersion checks if a system tool version is compatible
func ValidateSystemToolVersion(toolPath, expectedVersion string, versionArgs []string) bool {
	if toolPath == "" {
		return false
	}

	// Run version command
	cmd := exec.Command(toolPath, versionArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}

	// Basic version check - contains expected version
	outputStr := strings.ToLower(string(output))
	expectedLower := strings.ToLower(expectedVersion)

	return strings.Contains(outputStr, expectedLower)
}

// getToolVersionOverride checks for environment variable override for tool version
// Returns the override version if set, empty string otherwise
func getToolVersionOverride(toolName string) string {
	envVar := fmt.Sprintf("MVX_%s_VERSION", strings.ToUpper(toolName))
	return os.Getenv(envVar)
}

// getToolVersionOverrideEnvVar returns the environment variable name for tool version override
func getToolVersionOverrideEnvVar(toolName string) string {
	return fmt.Sprintf("MVX_%s_VERSION", strings.ToUpper(toolName))
}
