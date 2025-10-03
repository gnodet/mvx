package tools

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gnodet/mvx/pkg/config"
	"github.com/gnodet/mvx/pkg/util"
	"github.com/gnodet/mvx/pkg/version"
)

// Compile-time interface validation
var _ Tool = (*MavenTool)(nil)
var _ ToolMetadataProvider = (*MavenTool)(nil)
var _ DependencyProvider = (*MavenTool)(nil)
var _ EnvironmentProvider = (*MavenTool)(nil)

// MavenTool implements Tool interface for Maven management
type MavenTool struct {
	*BaseTool
}

// NewMavenTool creates a new Maven tool instance
func NewMavenTool(manager *Manager) Tool {
	var binaryName = BinaryMaven
	if NewPlatformMapper().IsWindows() {
		binaryName = BinaryMaven + ExtCmd
	}
	return &MavenTool{
		BaseTool: NewBaseTool(manager, ToolMaven, binaryName),
	}
}

// Install downloads and installs the specified Maven version
func (m *MavenTool) Install(version string, cfg config.ToolConfig) error {
	return m.installWithFallback(version, cfg)
}

// installWithFallback tries primary URL first, then fallback archive URL
func (m *MavenTool) installWithFallback(version string, cfg config.ToolConfig) error {
	// Check if we should use system tool instead of downloading (use standardized approach)
	if UseSystemTool(m.GetToolName()) {
		// Use standardized system tool detection
		return m.StandardInstall(version, cfg, m.getDownloadURL)
	}

	// Create installation directory
	installDir, err := m.CreateInstallDir(version, "")
	if err != nil {
		return InstallError(m.GetToolName(), version, fmt.Errorf("failed to create install directory: %w", err))
	}

	// Try both URLs with reduced retries instead of exhausting retries on first URL
	primaryURL := m.getDownloadURL(version)
	archiveURL := m.getArchiveDownloadURL(version)
	m.PrintDownloadMessage(version)

	options := m.GetDownloadOptions()

	// Try to download with alternating URLs and reduced retries per URL
	archivePath, err := m.downloadWithAlternatingURLs(primaryURL, archiveURL, version, cfg, options)
	if err != nil {
		return InstallError(m.GetToolName(), version, fmt.Errorf("download failed from both primary and archive URLs: %w", err))
	}
	defer os.Remove(archivePath) // Clean up downloaded file

	// Extract the downloaded file
	if err := m.Extract(archivePath, installDir, options); err != nil {
		return InstallError(m.toolName, version, err)
	}

	// Verify installation
	if err := m.Verify(version, cfg); err != nil {
		// Installation verification failed, clean up the installation directory
		fmt.Printf("  ❌ Maven installation verification failed: %v\n", err)
		fmt.Printf("  🧹 Cleaning up failed installation directory...\n")
		if removeErr := os.RemoveAll(installDir); removeErr != nil {
			fmt.Printf("  ⚠️  Warning: failed to clean up installation directory: %v\n", removeErr)
		}
		return InstallError("maven", version, fmt.Errorf("installation verification failed: %w", err))
	}
	fmt.Printf("  ✅ Maven %s installation verification successful\n", version)

	return nil
}

// IsInstalled checks if the specified version is installed
func (m *MavenTool) IsInstalled(version string, cfg config.ToolConfig) bool {
	return m.StandardIsInstalled(version, cfg, m.GetPath)
}

// GetPath returns the binary path for the specified version (for PATH management)
func (m *MavenTool) GetPath(version string, cfg config.ToolConfig) (string, error) {
	return m.StandardGetPath(version, cfg, m.getInstalledPath)
}

// getInstalledPath returns the path for an installed Maven version
func (m *MavenTool) getInstalledPath(version string, cfg config.ToolConfig) (string, error) {
	installDir := m.manager.GetToolVersionDir(m.GetToolName(), version, "")
	pathResolver := NewPathResolver(m.manager.GetToolsDir())
	binDir, err := pathResolver.FindBinaryParentDir(installDir, m.GetBinaryName())
	if err != nil {
		return "", err
	}
	return binDir, nil
}

// Verify checks if the installation is working correctly
func (m *MavenTool) Verify(version string, cfg config.ToolConfig) error {
	verifyConfig := VerificationConfig{
		BinaryName:  m.GetBinaryName(),
		VersionArgs: []string{"--version"},
		DebugInfo:   false,
	}
	return m.StandardVerifyWithConfig(version, cfg, verifyConfig)
}

// ListVersions returns available versions for installation
func (m *MavenTool) ListVersions() ([]string, error) {
	// Try to fetch versions from Apache distribution repos
	versions, err := m.fetchMavenVersionsFromApache()
	if err != nil {
		// Fallback to known versions if API is unavailable
		return m.getFallbackMavenVersions(), nil
	}
	return version.SortVersions(versions), nil
}

// GetDisplayName returns the human-readable name for Maven (implements ToolMetadataProvider)
func (m *MavenTool) GetDisplayName() string {
	return "Apache Maven"
}

// GetEmoji returns the emoji icon for Maven (implements ToolMetadataProvider)
func (m *MavenTool) GetEmoji() string {
	return "📦"
}

// GetDependencies returns the list of tools that Maven depends on (implements DependencyProvider)
func (m *MavenTool) GetDependencies() []string {
	return []string{ToolJava}
}

// SetupEnvironment sets up Maven-specific environment variables (implements EnvironmentProvider)
func (m *MavenTool) SetupEnvironment(version string, cfg config.ToolConfig, envManager *EnvironmentManager) error {
	binPath, err := m.GetPath(version, cfg)
	if err != nil {
		util.LogVerbose("Could not determine %s for %s %s: %v", EnvMavenHome, m.toolName, version, err)
		return nil
	}

	if strings.HasSuffix(binPath, "/bin") {
		homeDir := strings.TrimSuffix(binPath, "/bin")
		envManager.SetEnv(EnvMavenHome, homeDir)
		util.LogVerbose("Set %s=%s for %s %s", EnvMavenHome, homeDir, m.toolName, version)
	}

	return nil
}

// fetchMavenVersionsFromApache fetches Maven versions from Apache archive repositories
func (m *MavenTool) fetchMavenVersionsFromApache() ([]string, error) {
	var allVersions []string

	// Fetch Maven 3.x versions from archive
	maven3Versions, err := m.fetchVersionsFromApacheRepo(ApacheMavenBase + "/maven-3/")
	if err == nil {
		allVersions = append(allVersions, maven3Versions...)
	}

	// Fetch Maven 4.x versions from archive
	maven4Versions, err := m.fetchVersionsFromApacheRepo(ApacheMavenBase + "/maven-4/")
	if err == nil {
		allVersions = append(allVersions, maven4Versions...)
	}

	if len(allVersions) == 0 {
		return nil, fmt.Errorf("no versions found from Apache repositories")
	}

	return allVersions, nil
}

// getFallbackMavenVersions returns known Maven versions as fallback
func (m *MavenTool) getFallbackMavenVersions() []string {
	return []string{
		// Maven 4.x (pre-release versions)
		"4.0.0", "4.0.0-rc-4", "4.0.0-rc-3", "4.0.0-rc-2", "4.0.0-rc-1",
		"4.0.0-beta-5", "4.0.0-beta-4", "4.0.0-beta-3", "4.0.0-beta-2", "4.0.0-beta-1",
		"4.0.0-alpha-13", "4.0.0-alpha-12", "4.0.0-alpha-11", "4.0.0-alpha-10",

		// Maven 3.9.x (latest stable)
		"3.9.11", "3.9.10", "3.9.9", "3.9.8", "3.9.7", "3.9.6", "3.9.5", "3.9.4", "3.9.3", "3.9.2", "3.9.1", "3.9.0",

		// Maven 3.8.x (previous stable)
		"3.8.8", "3.8.7", "3.8.6", "3.8.5", "3.8.4", "3.8.3", "3.8.2", "3.8.1",

		// Maven 3.6.x (older stable)
		"3.6.3", "3.6.2", "3.6.1", "3.6.0",
	}
}

// fetchVersionsFromApacheRepo fetches version directories from Apache repository
func (m *MavenTool) fetchVersionsFromApacheRepo(repoURL string) ([]string, error) {
	registry := m.manager.GetRegistry()
	return registry.FetchVersionsFromApacheRepo(repoURL)
}

// GetDownloadOptions returns download options specific to Maven
func (m *MavenTool) GetDownloadOptions() DownloadOptions {
	return DownloadOptions{
		FileExtension: ExtZip,
	}
}

// getDownloadURL returns the download URL for the specified version
func (m *MavenTool) getDownloadURL(version string) string {
	// For recent releases, use dist.apache.org (CDN-backed)
	if strings.HasPrefix(version, "4.") {
		// Maven 4.x versions - try dist first for recent releases
		return fmt.Sprintf(ApacheDistBase+"/maven-4/%s/binaries/apache-maven-%s-bin.zip", version, version)
	}

	// Maven 3.x versions - try dist first for recent releases
	return fmt.Sprintf(ApacheDistBase+"/maven-3/%s/binaries/apache-maven-%s-bin.zip", version, version)
}

// getArchiveDownloadURL returns the fallback archive URL for the specified version
func (m *MavenTool) getArchiveDownloadURL(version string) string {
	if strings.HasPrefix(version, "4.") {
		// Maven 4.x versions are in the Maven 4 archive
		return fmt.Sprintf(ApacheMavenBase+"/maven-4/%s/binaries/apache-maven-%s-bin.zip", version, version)
	}

	// Maven 3.x versions are in the Maven 3 archive
	return fmt.Sprintf(ApacheMavenBase+"/maven-3/%s/binaries/apache-maven-%s-bin.zip", version, version)
}

// ResolveVersion resolves a Maven version specification to a concrete version
func (m *MavenTool) ResolveVersion(versionSpec, distribution string) (string, error) {
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
		return "", fmt.Errorf("failed to resolve Maven version %s: %w", versionSpec, err)
	}

	return resolved, nil
}

// GetDownloadURL implements URLProvider interface for Maven
func (m *MavenTool) GetDownloadURL(version string) string {
	return m.getDownloadURL(version)
}

// GetChecksumURL implements URLProvider interface for Maven
func (m *MavenTool) GetChecksumURL(version, filename string) string {
	if strings.HasPrefix(version, "4.") {
		return fmt.Sprintf("%s/maven-4/%s/binaries/%s.sha512",
			ApacheMavenBase, version, filename)
	}
	return fmt.Sprintf("%s/maven-3/%s/binaries/%s.sha512",
		ApacheMavenBase, version, filename)
}

// GetVersionsURL implements Tool interface for Maven
func (m *MavenTool) GetVersionsURL() string {
	return ApacheMavenBase + "/maven-3/"
}

// GetChecksum implements Tool interface for Maven
func (m *MavenTool) GetChecksum(version, filename string) (ChecksumInfo, error) {
	fmt.Printf("  🔍 Fetching Maven checksum from Apache archive...\n")

	checksumURL := m.GetChecksumURL(version, filename)
	if checksumURL == "" {
		return ChecksumInfo{}, fmt.Errorf("no checksum URL available for Maven %s", version)
	}

	// Fetch checksum from Apache archive
	checksum, err := m.fetchChecksumFromURL(checksumURL)
	if err != nil {
		fmt.Printf("  ⚠️  Failed to get Maven checksum: %v\n", err)
		return ChecksumInfo{}, err
	}

	fmt.Printf("  ✅ Found Maven checksum from Apache archive\n")
	return ChecksumInfo{
		Type:  SHA512,
		Value: checksum,
	}, nil
}

// fetchChecksumFromURL fetches checksum from a URL
func (m *MavenTool) fetchChecksumFromURL(url string) (string, error) {
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

// downloadWithAlternatingURLs tries downloading from both URLs with reduced retries per URL
// instead of exhausting all retries on the first URL before trying the second
func (m *MavenTool) downloadWithAlternatingURLs(primaryURL, archiveURL, version string, cfg config.ToolConfig, options DownloadOptions) (string, error) {
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
		downloadConfig.ToolName = m.toolName
		downloadConfig.Version = version
		downloadConfig.Config = cfg
		downloadConfig.Tool = m

		// Create temporary file for download
		tmpFile, err := os.CreateTemp("", fmt.Sprintf("maven-*%s", options.FileExtension))
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
