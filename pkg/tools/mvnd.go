package tools

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gnodet/mvx/pkg/config"
	"github.com/gnodet/mvx/pkg/version"
)

// Compile-time interface validation
var _ Tool = (*MvndTool)(nil)
var _ ToolMetadataProvider = (*MvndTool)(nil)
var _ DependencyProvider = (*MvndTool)(nil)
var _ EnvironmentProvider = (*MvndTool)(nil)

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
	// Try to fetch versions from Apache archive
	versions, err := m.fetchMvndVersionsFromApache()
	if err != nil {
		// Fallback to known versions if API is unavailable
		return m.getFallbackMvndVersions(), nil
	}
	return version.SortVersions(versions), nil
}

// GetDisplayName returns the human-readable name for Mvnd (implements ToolMetadataProvider)
func (m *MvndTool) GetDisplayName() string {
	return "Maven Daemon (mvnd)"
}

// GetEmoji returns the emoji icon for Mvnd (implements ToolMetadataProvider)
func (m *MvndTool) GetEmoji() string {
	return "🚀"
}

// GetDependencies returns the list of tools that Maven Daemon depends on (implements DependencyProvider)
func (m *MvndTool) GetDependencies() []string {
	return []string{ToolJava}
}

// SetupEnvironment sets up Maven Daemon-specific environment variables (implements EnvironmentProvider)
func (m *MvndTool) SetupEnvironment(version string, cfg config.ToolConfig, envVars map[string]string) error {
	return m.SetupHomeEnvironment(version, cfg, envVars, EnvMvndHome, m.GetPath)
}

// fetchMvndVersionsFromApache fetches mvnd versions from Apache archive
func (m *MvndTool) fetchMvndVersionsFromApache() ([]string, error) {
	registry := m.manager.GetRegistry()
	// Fetch mvnd versions from archive
	mvndVersions, err := registry.FetchVersionsFromApacheRepo(ApacheMavenBase + "/mvnd/")
	if err != nil {
		return nil, fmt.Errorf("no mvnd versions found from Apache archive: %w", err)
	}

	return mvndVersions, nil
}

// getFallbackMvndVersions returns known mvnd versions as fallback
func (m *MvndTool) getFallbackMvndVersions() []string {
	return []string{
		// Maven Daemon 2.x
		"2.0.0", "2.0.0-beta-1", "2.0.0-alpha-1",

		// Maven Daemon 1.x
		"1.0.2", "1.0.1", "1.0.0", "1.0.0-m8", "1.0.0-m7", "1.0.0-m6", "1.0.0-m5",
		"1.0.0-m4", "1.0.0-m3", "1.0.0-m2", "1.0.0-m1",

		// Maven Daemon 0.x
		"0.9.0", "0.8.2", "0.8.1", "0.8.0", "0.7.1", "0.7.0", "0.6.0", "0.5.2", "0.5.1", "0.5.0",
	}
}

// GetDownloadOptions returns download options specific to Maven Daemon
func (m *MvndTool) GetDownloadOptions() DownloadOptions {
	return DownloadOptions{
		FileExtension: ".zip",
	}
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
func (m *MvndTool) ResolveVersion(versionSpec, distribution string) (string, error) {
	availableVersions, err := m.ListVersions()
	if err != nil {
		return "", err
	}

	spec, err := version.ParseSpec(versionSpec)
	if err != nil {
		return "", fmt.Errorf("invalid version specification %s: %w", versionSpec, err)
	}

	resolved, err := spec.Resolve(availableVersions)
	if err != nil {
		return "", fmt.Errorf("failed to resolve mvnd version %s: %w", versionSpec, err)
	}

	return resolved, nil
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
		// Note: MinSize/MaxSize/ExpectedType removed - we rely on checksums for validation
		downloadConfig.ToolName = "mvnd"
		downloadConfig.Version = version
		downloadConfig.Config = cfg
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
