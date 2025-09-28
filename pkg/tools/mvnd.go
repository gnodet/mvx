package tools

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gnodet/mvx/pkg/config"
)

// Compile-time interface validation
var _ Tool = (*MvndTool)(nil)

// MvndTool implements Tool interface for Maven Daemon management
type MvndTool struct {
	*BaseTool
}

func getMvndBinaryName() string {
	if NewPlatformMapper().IsWindows() {
		return "mvnd.exe"
	}
	return "mvnd"
}

// NewMvndTool creates a new Mvnd tool instance
func NewMvndTool(manager *Manager) *MvndTool {
	return &MvndTool{
		BaseTool: NewBaseTool(manager, "mvnd", getMvndBinaryName()),
	}
}

// Name returns the tool name
func (m *MvndTool) Name() string {
	return "mvnd"
}

// Install downloads and installs the specified mvnd version
func (m *MvndTool) Install(version string, cfg config.ToolConfig) error {
	return m.installWithFallback(version, cfg)
}

// IsInstalled checks if the specified version is installed
func (m *MvndTool) IsInstalled(version string, cfg config.ToolConfig) bool {
	return m.StandardIsInstalled(version, cfg, m.GetPath)
}

// GetPath returns the binary path for the specified version (for PATH management)
func (m *MvndTool) GetPath(version string, cfg config.ToolConfig) (string, error) {
	return m.StandardGetPath(version, cfg, m.getInstalledPath)
}

func (m *MvndTool) GetBinaryName() string {
	return getMvndBinaryName()
}

// getInstalledPath returns the path for an installed Mvnd version
func (m *MvndTool) getInstalledPath(version string, cfg config.ToolConfig) (string, error) {
	installDir := m.manager.GetToolVersionDir("mvnd", version, "")
	pathResolver := NewPathResolver(m.manager.GetToolsDir())
	binDir, err := pathResolver.FindBinaryParentDir(installDir, m.GetBinaryName())
	if err != nil {
		return "", err
	}
	return binDir, nil
}

// Verify checks if the installation is working correctly
func (m *MvndTool) Verify(version string, cfg config.ToolConfig) error {
	verifyConfig := VerificationConfig{
		BinaryName:  m.GetBinaryName(),
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

// installWithFallback tries primary URL first, then fallback archive URL with improved retry strategy
func (m *MvndTool) installWithFallback(version string, cfg config.ToolConfig) error {
	// Check if we should use system tool instead of downloading
	if UseSystemTool("mvnd") {
		return m.StandardInstall(version, cfg, m.getDownloadURL)
	}

	// Create installation directory
	installDir, err := m.CreateInstallDir(version, "")
	if err != nil {
		return InstallError("mvnd", version, fmt.Errorf("failed to create install directory: %w", err))
	}

	// Try both URLs with reduced retries instead of exhausting retries on first URL
	primaryURL := m.getDownloadURL(version)
	archiveURL := m.getArchiveDownloadURL(version)
	m.PrintDownloadMessage(version)

	options := m.GetDownloadOptions()

	// Try to download with alternating URLs and reduced retries per URL
	archivePath, err := m.downloadWithAlternatingURLs(primaryURL, archiveURL, version, cfg, options)
	if err != nil {
		return InstallError("mvnd", version, fmt.Errorf("download failed from both primary and archive URLs: %w", err))
	}

	// Extract archive
	if err := m.Extract(archivePath, installDir, options); err != nil {
		return InstallError("mvnd", version, fmt.Errorf("failed to extract archive: %w", err))
	}

	// Clean up downloaded archive
	if err := os.Remove(archivePath); err != nil {
		fmt.Printf("  ⚠️  Warning: failed to remove archive file: %v\n", err)
	}

	fmt.Printf("  ✅ Maven Daemon %s installed successfully\n", version)
	return nil
}

// downloadWithAlternatingURLs tries downloading from both URLs with reduced retries per URL
// instead of exhausting all retries on the first URL before trying the second
func (m *MvndTool) downloadWithAlternatingURLs(primaryURL, archiveURL, version string, cfg config.ToolConfig, options DownloadOptions) (string, error) {
	urls := []struct {
		url  string
		name string
	}{
		{primaryURL, "primary"},
		{archiveURL, "archive"},
	}

	maxRetries := 3    // Total retries across both URLs
	retriesPerURL := 1 // Reduced retries per URL to allow trying both

	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		urlIndex := attempt % len(urls)
		currentURL := urls[urlIndex]

		if attempt > 0 && urlIndex == 0 {
			fmt.Printf("  🔄 Trying %s URL again (attempt %d)...\n", currentURL.name, (attempt/len(urls))+1)
		} else if urlIndex == 1 {
			fmt.Printf("  🔄 Switching to %s URL...\n", currentURL.name)
		}

		// Create download config with reduced retries per URL
		downloadConfig := DefaultDownloadConfig(currentURL.url, "")
		downloadConfig.MaxRetries = retriesPerURL
		downloadConfig.ExpectedType = options.ExpectedType
		downloadConfig.MinSize = options.MinSize
		downloadConfig.MaxSize = options.MaxSize
		downloadConfig.ToolName = "mvnd"
		downloadConfig.Version = version
		downloadConfig.Config = cfg
		downloadConfig.ChecksumRegistry = m.manager.GetChecksumRegistry()
		downloadConfig.Tool = m

		// Create temporary file for download
		tmpFile, err := os.CreateTemp("", fmt.Sprintf("mvnd-*%s", options.FileExtension))
		if err != nil {
			lastErr = fmt.Errorf("failed to create temporary file: %w", err)
			continue
		}
		downloadConfig.DestPath = tmpFile.Name()
		tmpFile.Close()

		_, err = RobustDownload(downloadConfig)
		if err == nil {
			fmt.Printf("  ✅ Successfully downloaded from %s URL\n", currentURL.name)
			return downloadConfig.DestPath, nil
		}

		lastErr = err
		fmt.Printf("  ⚠️  Download from %s URL failed: %v\n", currentURL.name, err)
		// Clean up failed download
		os.Remove(downloadConfig.DestPath)
	}

	return "", fmt.Errorf("all download attempts failed, last error: %w", lastErr)
}
