package tools

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// isVerbose checks if verbose logging is enabled
func isVerbose() bool {
	return os.Getenv("MVX_VERBOSE") == "true"
}

// logVerbose prints verbose log messages
func logVerbose(format string, args ...interface{}) {
	if isVerbose() {
		fmt.Printf("[VERBOSE] "+format+"\n", args...)
	}
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
