package tools

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
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

// useSystemTool checks if a system tool should be used instead of downloading
// by checking the MVX_USE_SYSTEM_<TOOL> environment variable
func useSystemTool(toolName string) bool {
	envVar := fmt.Sprintf("MVX_USE_SYSTEM_%s", strings.ToUpper(toolName))
	return os.Getenv(envVar) == "true"
}

// getSystemToolEnvVar returns the environment variable name for a tool
func getSystemToolEnvVar(toolName string) string {
	return fmt.Sprintf("MVX_USE_SYSTEM_%s", strings.ToUpper(toolName))
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

// SystemToolDetector provides methods for detecting and validating system tools
type SystemToolDetector interface {
	// GetSystemHome returns the system installation path for the tool
	GetSystemHome() (string, error)

	// GetSystemVersion returns the version of the system tool
	GetSystemVersion(toolHome string) (string, error)

	// IsVersionCompatible checks if the system version is compatible with requested version
	IsVersionCompatible(systemVersion, requestedVersion string) bool

	// CreateSystemLink creates a symlink to the system tool installation
	CreateSystemLink(systemHome, installDir string) error
}

// JavaSystemDetector implements SystemToolDetector for Java
type JavaSystemDetector struct{}

// GetSystemHome returns the system JAVA_HOME if available and valid
func (d *JavaSystemDetector) GetSystemHome() (string, error) {
	javaHome := os.Getenv("JAVA_HOME")
	if javaHome == "" {
		return "", fmt.Errorf("JAVA_HOME environment variable not set")
	}

	// Check if JAVA_HOME points to a valid Java installation
	javaExe := filepath.Join(javaHome, "bin", "java")
	if runtime.GOOS == "windows" {
		javaExe += ".exe"
	}

	if _, err := os.Stat(javaExe); err != nil {
		return "", fmt.Errorf("Java executable not found at %s", javaExe)
	}

	return javaHome, nil
}

// GetSystemVersion returns the version of the system Java installation
func (d *JavaSystemDetector) GetSystemVersion(javaHome string) (string, error) {
	javaExe := filepath.Join(javaHome, "bin", "java")
	if runtime.GOOS == "windows" {
		javaExe += ".exe"
	}

	cmd := exec.Command(javaExe, "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get Java version: %w", err)
	}

	// Parse version from output (e.g., "openjdk version "21.0.1" 2023-10-17")
	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("no version output from Java")
	}

	// Look for version pattern in first line
	versionLine := lines[0]
	// Extract version number (handles both old and new format)
	if strings.Contains(versionLine, "\"") {
		start := strings.Index(versionLine, "\"")
		if start != -1 {
			end := strings.Index(versionLine[start+1:], "\"")
			if end != -1 {
				version := versionLine[start+1 : start+1+end]
				// Extract major version (e.g., "21.0.1" -> "21", "1.8.0_391" -> "8")
				if strings.HasPrefix(version, "1.") {
					// Old format (Java 8 and below): "1.8.0_391" -> "8"
					parts := strings.Split(version, ".")
					if len(parts) >= 2 {
						return parts[1], nil
					}
				} else {
					// New format (Java 9+): "21.0.1" -> "21"
					parts := strings.Split(version, ".")
					if len(parts) >= 1 {
						return parts[0], nil
					}
				}
			}
		}
	}

	return "", fmt.Errorf("could not parse Java version from: %s", versionLine)
}

// IsVersionCompatible checks if the system Java version is compatible with the requested version
func (d *JavaSystemDetector) IsVersionCompatible(systemVersion, requestedVersion string) bool {
	// For now, we require exact major version match
	// This could be made more flexible in the future
	return systemVersion == requestedVersion
}

// CreateSystemLink creates a symlink to the system Java installation
func (d *JavaSystemDetector) CreateSystemLink(systemHome, installDir string) error {
	// Create installation directory
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}

	// Create a symlink to the system Java
	linkPath := filepath.Join(installDir, "jdk")

	// Remove existing link/directory if it exists
	if _, err := os.Lstat(linkPath); err == nil {
		if err := os.RemoveAll(linkPath); err != nil {
			return fmt.Errorf("failed to remove existing installation: %w", err)
		}
	}

	// Create symlink
	if err := os.Symlink(systemHome, linkPath); err != nil {
		return fmt.Errorf("failed to create symlink to system tool: %w", err)
	}

	logVerbose("Created symlink from %s to %s", linkPath, systemHome)
	return nil
}

// MavenSystemDetector implements SystemToolDetector for Maven
type MavenSystemDetector struct{}

// GetSystemHome returns the system MAVEN_HOME if available and valid
func (d *MavenSystemDetector) GetSystemHome() (string, error) {
	// Try MAVEN_HOME first
	mavenHome := os.Getenv("MAVEN_HOME")
	if mavenHome != "" {
		mvnExe := filepath.Join(mavenHome, "bin", "mvn")
		if runtime.GOOS == "windows" {
			mvnExe += ".cmd"
		}
		if _, err := os.Stat(mvnExe); err == nil {
			return mavenHome, nil
		}
	}

	// Try M2_HOME as fallback
	m2Home := os.Getenv("M2_HOME")
	if m2Home != "" {
		mvnExe := filepath.Join(m2Home, "bin", "mvn")
		if runtime.GOOS == "windows" {
			mvnExe += ".cmd"
		}
		if _, err := os.Stat(mvnExe); err == nil {
			return m2Home, nil
		}
	}

	// Try to find mvn in PATH
	mvnExe := "mvn"
	if runtime.GOOS == "windows" {
		mvnExe = "mvn.cmd"
	}

	if mvnPath, err := exec.LookPath(mvnExe); err == nil {
		// Try to determine MAVEN_HOME from mvn path
		// mvn is typically at $MAVEN_HOME/bin/mvn
		binDir := filepath.Dir(mvnPath)
		mavenHome := filepath.Dir(binDir)

		// Verify this looks like a Maven installation
		if _, err := os.Stat(filepath.Join(mavenHome, "lib")); err == nil {
			return mavenHome, nil
		}
	}

	return "", fmt.Errorf("Maven not found in MAVEN_HOME, M2_HOME, or PATH")
}

// GetSystemVersion returns the version of the system Maven installation
func (d *MavenSystemDetector) GetSystemVersion(mavenHome string) (string, error) {
	mvnExe := filepath.Join(mavenHome, "bin", "mvn")
	if runtime.GOOS == "windows" {
		mvnExe += ".cmd"
	}

	cmd := exec.Command(mvnExe, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get Maven version: %w", err)
	}

	// Parse version from output (e.g., "Apache Maven 3.9.6 (bc0240f3c744dd6b6ec2920b3cd08dcc295161ae)")
	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("no version output from Maven")
	}

	// Look for "Apache Maven" in the first line
	versionLine := lines[0]
	if strings.Contains(versionLine, "Apache Maven") {
		parts := strings.Fields(versionLine)
		if len(parts) >= 3 {
			version := parts[2]
			// Remove ANSI escape sequences (e.g., "3.6.3\x1b[m" -> "3.6.3")
			ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
			version = ansiRegex.ReplaceAllString(version, "")
			// Also remove any remaining bracket sequences
			if idx := strings.Index(version, "["); idx != -1 {
				version = version[:idx]
			}
			return version, nil // "Apache Maven 3.9.6" -> "3.9.6"
		}
	}

	return "", fmt.Errorf("could not parse Maven version from: %s", versionLine)
}

// IsVersionCompatible checks if the system Maven version is compatible with the requested version
func (d *MavenSystemDetector) IsVersionCompatible(systemVersion, requestedVersion string) bool {
	// For Maven, we can be more flexible - allow compatible versions
	// For now, exact match, but this could be enhanced to support semantic versioning
	return systemVersion == requestedVersion
}

// CreateSystemLink creates a symlink to the system Maven installation
func (d *MavenSystemDetector) CreateSystemLink(systemHome, installDir string) error {
	// Create installation directory
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}

	// For Maven, we typically create a direct symlink to the Maven home
	linkPath := filepath.Join(installDir, "maven")

	// Remove existing link/directory if it exists
	if _, err := os.Lstat(linkPath); err == nil {
		if err := os.RemoveAll(linkPath); err != nil {
			return fmt.Errorf("failed to remove existing installation: %w", err)
		}
	}

	// Create symlink
	if err := os.Symlink(systemHome, linkPath); err != nil {
		return fmt.Errorf("failed to create symlink to system tool: %w", err)
	}

	logVerbose("Created symlink from %s to %s", linkPath, systemHome)
	return nil
}
