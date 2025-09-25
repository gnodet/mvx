package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gnodet/mvx/pkg/config"
)

// Compile-time interface validation
var _ Tool = (*MvndTool)(nil)

// useSystemMvnd checks if system Mvnd should be used instead of downloading
func useSystemMvnd() bool {
	return UseSystemTool("mvnd")
}

// MvndTool implements Tool interface for Maven Daemon management
type MvndTool struct {
	*BaseTool
}

// NewMvndTool creates a new Mvnd tool instance
func NewMvndTool(manager *Manager) *MvndTool {
	return &MvndTool{
		BaseTool: NewBaseTool(manager, "mvnd"),
	}
}

// Name returns the tool name
func (m *MvndTool) Name() string {
	return "mvnd"
}

// Install downloads and installs the specified mvnd version
func (m *MvndTool) Install(version string, cfg config.ToolConfig) error {
	return m.StandardInstall(version, cfg, m.getDownloadURL)
}

// IsInstalled checks if the specified version is installed
func (m *MvndTool) IsInstalled(version string, cfg config.ToolConfig) bool {
	return m.StandardIsInstalled(version, cfg, m.GetPath, "mvnd")
}

// GetPath returns the binary path for the specified version (for PATH management)
func (m *MvndTool) GetPath(version string, cfg config.ToolConfig) (string, error) {
	return m.StandardGetPath(version, cfg, m.getInstalledPath, "mvnd")
}

// getInstalledPath returns the path for an installed Mvnd version
func (m *MvndTool) getInstalledPath(version string, cfg config.ToolConfig) (string, error) {
	installDir := m.manager.GetToolVersionDir("mvnd", version, "")

	// mvnd archives typically extract to maven-mvnd-{version}-{platform}/
	entries, err := os.ReadDir(installDir)
	if err != nil {
		return "", fmt.Errorf("failed to read installation directory: %w", err)
	}

	// Look for maven-mvnd-* directory
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "maven-mvnd-") {
			subPath := filepath.Join(installDir, entry.Name())
			mvndExe := filepath.Join(subPath, "bin", "mvnd")
			if runtime.GOOS == "windows" {
				mvndExe += ".cmd"
			}
			if _, err := os.Stat(mvndExe); err == nil {
				return filepath.Join(subPath, "bin"), nil
			}
		}
	}

	return filepath.Join(installDir, "bin"), nil
}

// Verify checks if the installation is working correctly
func (m *MvndTool) Verify(version string, cfg config.ToolConfig) error {
	return m.StandardVerify(version, cfg, m.GetPath, "mvnd", []string{"--version"})
}

// ListVersions returns available mvnd versions
func (m *MvndTool) ListVersions() ([]string, error) {
	registry := m.manager.GetRegistry()
	return registry.GetMvndVersions()
}

// GetDownloadOptions returns download options specific to Maven Daemon
func (m *MvndTool) GetDownloadOptions() DownloadOptions {
	return DownloadOptions{
		FileExtension: ".zip",
		ExpectedType:  "application",
		MinSize:       5 * 1024 * 1024,   // 5MB
		MaxSize:       100 * 1024 * 1024, // 100MB
		ArchiveType:   "zip",
	}
}

// GetDisplayName returns the display name for Maven Daemon
func (m *MvndTool) GetDisplayName() string {
	return "Maven Daemon"
}

// getDownloadURL returns the download URL for the specified version
func (m *MvndTool) getDownloadURL(version string) string {
	// Determine platform-specific archive name
	platform := m.getPlatformString()

	// Try dist first for recent releases (CDN-backed)
	return fmt.Sprintf("https://dist.apache.org/repos/dist/release/maven/mvnd/%s/maven-mvnd-%s-%s.zip", version, version, platform)
}

// getArchiveDownloadURL returns the fallback archive URL for the specified version
func (m *MvndTool) getArchiveDownloadURL(version string) string {
	// Determine platform-specific archive name
	platform := m.getPlatformString()

	// mvnd archives are in the Apache archive
	return fmt.Sprintf("https://archive.apache.org/dist/maven/mvnd/%s/maven-mvnd-%s-%s.zip", version, version, platform)
}

// getPlatformString returns the platform string for mvnd downloads
func (m *MvndTool) getPlatformString() string {
	switch runtime.GOOS {
	case "linux":
		if runtime.GOARCH == "arm64" {
			return "linux-aarch64"
		}
		return "linux-amd64"
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return "darwin-aarch64"
		}
		return "darwin-amd64"
	case "windows":
		return "windows-amd64"
	default:
		return "linux-amd64" // fallback
	}
}
