package tools

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/gnodet/mvx/pkg/config"
)

// Compile-time interface validation
var _ Tool = (*MavenTool)(nil)

// MavenTool implements Tool interface for Maven management
type MavenTool struct {
	*BaseTool
}

// NewMavenTool creates a new Maven tool instance
func NewMavenTool(manager *Manager) *MavenTool {
	return &MavenTool{
		BaseTool: NewBaseTool(manager, "maven"),
	}
}

// Name returns the tool name
func (m *MavenTool) Name() string {
	return "maven"
}

// useSystemMaven checks if system Maven should be used instead of downloading
func useSystemMaven() bool {
	return UseSystemTool("maven")
}

// getSystemMavenHome returns the system MAVEN_HOME if available and valid
func getSystemMavenHome() (string, error) {
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

	return "", fmt.Errorf("Maven not found in MAVEN_HOME or PATH")
}

// getSystemMavenVersion returns the version of the system Maven installation
func getSystemMavenVersion(mavenHome string) (string, error) {
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

// isMavenVersionCompatible checks if the system Maven version is compatible with the requested version
func isMavenVersionCompatible(systemVersion, requestedVersion string) bool {
	// For Maven, we can be more flexible - allow compatible versions
	// For now, exact match, but this could be enhanced to support semantic versioning
	return systemVersion == requestedVersion
}

// Install downloads and installs the specified Maven version
func (m *MavenTool) Install(version string, cfg config.ToolConfig) error {
	return m.installWithFallback(version, cfg)
}

// installWithFallback tries primary URL first, then fallback archive URL
func (m *MavenTool) installWithFallback(version string, cfg config.ToolConfig) error {
	// Check if we should use system tool instead of downloading
	if UseSystemTool("maven") {
		logVerbose("MVX_USE_SYSTEM_MAVEN=true, forcing use of system maven")

		// Try environment variables first
		if mavenHome := os.Getenv("MAVEN_HOME"); mavenHome != "" {
			toolPath := filepath.Join(mavenHome, "bin", "mvn")
			if runtime.GOOS == "windows" {
				toolPath += ".cmd"
			}
			if _, err := os.Stat(toolPath); err == nil {
				fmt.Printf("  🔗 Using system Maven from MAVEN_HOME: %s\n", mavenHome)
				fmt.Printf("  ✅ System Maven configured (mvx will use MAVEN_HOME)\n")
				return nil
			}
		}

		// Try PATH
		binaryName := "mvn"
		if runtime.GOOS == "windows" {
			binaryName = "mvn.cmd"
		}
		if toolPath, err := exec.LookPath(binaryName); err == nil {
			fmt.Printf("  🔗 Using system Maven from PATH: %s\n", toolPath)
			fmt.Printf("  ✅ System Maven configured (mvx will use system PATH)\n")
			return nil
		}

		return fmt.Errorf("MVX_USE_SYSTEM_MAVEN=true but system Maven not available")
	}

	// Create installation directory
	installDir, err := m.CreateInstallDir(version, "")
	if err != nil {
		return InstallError("maven", version, fmt.Errorf("failed to create install directory: %w", err))
	}

	// Try primary URL first
	primaryURL := m.getDownloadURL(version)
	m.PrintDownloadMessage(version)

	options := m.GetDownloadOptions()
	err = m.DownloadAndExtract(primaryURL, installDir, version, cfg, options)
	if err == nil {
		return nil
	}

	// If primary URL fails, try archive URL
	fmt.Printf("  🔄 Primary download failed, trying archive URL...\n")
	archiveURL := m.getArchiveDownloadURL(version)
	err = m.DownloadAndExtract(archiveURL, installDir, version, cfg, options)
	if err != nil {
		return InstallError("maven", version, fmt.Errorf("both primary and archive downloads failed: %w", err))
	}

	return nil
}

// IsInstalled checks if the specified version is installed
func (m *MavenTool) IsInstalled(version string, cfg config.ToolConfig) bool {
	return m.StandardIsInstalledWithOptions(version, cfg, m.GetPath, "mvn", nil, []string{"MAVEN_HOME"})
}

// GetPath returns the binary path for the specified version (for PATH management)
func (m *MavenTool) GetPath(version string, cfg config.ToolConfig) (string, error) {
	return m.StandardGetPathWithOptions(version, cfg, m.getInstalledPath, "mvn", nil, []string{"MAVEN_HOME"})
}

// getInstalledPath returns the path for an installed Maven version
func (m *MavenTool) getInstalledPath(version string, cfg config.ToolConfig) (string, error) {
	installDir := m.manager.GetToolVersionDir("maven", version, "")

	// Maven archives typically extract to apache-maven-{version}/
	entries, err := os.ReadDir(installDir)
	if err != nil {
		return "", fmt.Errorf("failed to read installation directory: %w", err)
	}

	// Look for apache-maven-* directory
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "apache-maven-") {
			subPath := filepath.Join(installDir, entry.Name())
			mvnExe := filepath.Join(subPath, "bin", "mvn")
			if runtime.GOOS == "windows" {
				mvnExe += ".cmd"
			}
			if _, err := os.Stat(mvnExe); err == nil {
				return filepath.Join(subPath, "bin"), nil
			}
		}
	}

	return filepath.Join(installDir, "bin"), nil
}

// Verify checks if the installation is working correctly
func (m *MavenTool) Verify(version string, cfg config.ToolConfig) error {
	return m.StandardVerify(version, cfg, m.GetPath, "mvn", []string{"--version"})
}

// findJavaHome attempts to find an installed Java home directory
func (m *MavenTool) findJavaHome() (string, error) {
	// Check if JAVA_HOME is already set
	if javaHome := os.Getenv("JAVA_HOME"); javaHome != "" {
		return javaHome, nil
	}

	// Try to find Java installations in the tools directory
	javaToolsDir := filepath.Join(m.manager.GetToolsDir(), "java")
	if entries, err := os.ReadDir(javaToolsDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				javaVersionDir := filepath.Join(javaToolsDir, entry.Name())

				// Try to find Java executable in this version directory
				if javaPath, err := m.findJavaInDirectory(javaVersionDir); err == nil {
					logVerbose("Found Java installation at: %s", javaPath)
					return javaPath, nil
				}
			}
		}
	}

	return "", fmt.Errorf("no Java installation found")
}

// findJavaInDirectory searches for Java executable in a directory and returns JAVA_HOME
func (m *MavenTool) findJavaInDirectory(dir string) (string, error) {
	// Check if there are subdirectories (common with JDK archives)
	if entries, err := os.ReadDir(dir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				subPath := filepath.Join(dir, entry.Name())
				javaExe := filepath.Join(subPath, "bin", "java")
				if runtime.GOOS == "windows" {
					javaExe += ".exe"
				}
				if _, err := os.Stat(javaExe); err == nil {
					return subPath, nil
				}
			}
		}
	}

	// Also check if java is directly in the directory
	javaExe := filepath.Join(dir, "bin", "java")
	if runtime.GOOS == "windows" {
		javaExe += ".exe"
	}
	if _, err := os.Stat(javaExe); err == nil {
		return dir, nil
	}

	return "", fmt.Errorf("java executable not found in %s", dir)
}

// ListVersions returns available versions for installation
func (m *MavenTool) ListVersions() ([]string, error) {
	registry := m.manager.GetRegistry()
	return registry.GetMavenVersions()
}

// GetDownloadOptions returns download options specific to Maven
func (m *MavenTool) GetDownloadOptions() DownloadOptions {
	return DownloadOptions{
		FileExtension: ".zip",
		ExpectedType:  "application",
		MinSize:       5 * 1024 * 1024,   // 5MB
		MaxSize:       100 * 1024 * 1024, // 100MB
		ArchiveType:   "zip",
	}
}

// GetDisplayName returns the display name for Maven
func (m *MavenTool) GetDisplayName() string {
	return "Maven"
}

// getDownloadURL returns the download URL for the specified version
func (m *MavenTool) getDownloadURL(version string) string {
	// For recent releases, use dist.apache.org (CDN-backed)
	if strings.HasPrefix(version, "4.") {
		// Maven 4.x versions - try dist first for recent releases
		return fmt.Sprintf("https://dist.apache.org/repos/dist/release/maven/maven-4/%s/binaries/apache-maven-%s-bin.zip", version, version)
	}

	// Maven 3.x versions - try dist first for recent releases
	return fmt.Sprintf("https://dist.apache.org/repos/dist/release/maven/maven-3/%s/binaries/apache-maven-%s-bin.zip", version, version)
}

// getArchiveDownloadURL returns the fallback archive URL for the specified version
func (m *MavenTool) getArchiveDownloadURL(version string) string {
	if strings.HasPrefix(version, "4.") {
		// Maven 4.x versions are in the Maven 4 archive
		return fmt.Sprintf("https://archive.apache.org/dist/maven/maven-4/%s/binaries/apache-maven-%s-bin.zip", version, version)
	}

	// Maven 3.x versions are in the Maven 3 archive
	return fmt.Sprintf("https://archive.apache.org/dist/maven/maven-3/%s/binaries/apache-maven-%s-bin.zip", version, version)
}
