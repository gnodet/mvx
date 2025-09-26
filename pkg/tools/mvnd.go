package tools

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gnodet/mvx/pkg/config"
)

// Compile-time interface validation
var _ Tool = (*MvndTool)(nil)

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
	pathResolver := NewPathResolver(m.manager.GetToolsDir())

	options := DirectorySearchOptions{
		DirectoryPrefix:           "maven-mvnd-",
		BinSubdirectory:           "bin",
		BinaryName:                "mvnd",
		UsePlatformExtensions:     true,
		PreferredWindowsExtension: ".cmd", // Mvnd uses .cmd on Windows
		FallbackToParent:          false,  // Mvnd always has bin subdirectory
	}

	binPath, err := pathResolver.FindToolBinaryPath(installDir, options)
	if err != nil {
		// Fallback to default bin directory if search fails
		return filepath.Join(installDir, "bin"), nil
	}

	return binPath, nil
}

// Verify checks if the installation is working correctly
func (m *MvndTool) Verify(version string, cfg config.ToolConfig) error {
	verifyConfig := VerificationConfig{
		BinaryName:  "mvnd",
		VersionArgs: []string{"--version"},
		DebugInfo:   false,
	}
	return m.StandardVerifyWithConfig(version, cfg, verifyConfig)
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
	platformMapper := NewPlatformMapper()

	switch platformMapper.GetOS() {
	case "linux":
		if platformMapper.GetArch() == "arm64" {
			return "linux-aarch64"
		}
		return "linux-amd64"
	case "darwin":
		if platformMapper.GetArch() == "arm64" {
			return "darwin-aarch64"
		}
		return "darwin-amd64"
	case "windows":
		return "windows-amd64"
	default:
		return "linux-amd64" // fallback
	}
}

// ResolveVersion resolves a Mvnd version specification to a concrete version
func (m *MvndTool) ResolveVersion(version, distribution string) (string, error) {
	registry := m.manager.GetRegistry()
	return registry.ResolveMvndVersion(version)
}

// GetDownloadURL implements Tool interface for Maven Daemon
func (m *MvndTool) GetDownloadURL(version string) string {
	return m.getDownloadURL(version)
}

// GetChecksumURL implements Tool interface for Maven Daemon
func (m *MvndTool) GetChecksumURL(version, filename string) string {
	return fmt.Sprintf("%s/mvnd/%s/%s.sha512",
		ApacheMavenBase, version, filename)
}

// GetVersionsURL implements Tool interface for Maven Daemon
func (m *MvndTool) GetVersionsURL() string {
	return ApacheMavenBase + "/mvnd/"
}

// GetChecksum implements Tool interface for Maven Daemon
func (m *MvndTool) GetChecksum(version, filename string) (ChecksumInfo, error) {
	fmt.Printf("  🔍 Fetching Maven Daemon checksum from Apache archive...\n")

	checksumURL := m.GetChecksumURL(version, filename)
	if checksumURL == "" {
		return ChecksumInfo{}, fmt.Errorf("no checksum URL available for Maven Daemon %s", version)
	}

	// Fetch checksum from Apache archive (reuse Maven's implementation)
	checksum, err := m.fetchChecksumFromURL(checksumURL)
	if err != nil {
		fmt.Printf("  ⚠️  Failed to get Maven Daemon checksum: %v\n", err)
		return ChecksumInfo{}, err
	}

	fmt.Printf("  ✅ Found Maven Daemon checksum from Apache archive\n")
	return ChecksumInfo{
		Type:  SHA512,
		Value: checksum,
	}, nil
}

// fetchChecksumFromURL fetches checksum from a URL (same as Maven)
func (m *MvndTool) fetchChecksumFromURL(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch checksum: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("checksum request returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read checksum response: %w", err)
	}

	// Apache checksum files contain just the checksum value
	checksum := strings.TrimSpace(string(body))
	if len(checksum) == 0 {
		return "", fmt.Errorf("empty checksum response")
	}

	return checksum, nil
}
