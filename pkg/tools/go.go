package tools

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gnodet/mvx/pkg/config"
	"github.com/gnodet/mvx/pkg/version"
)

// Compile-time interface validation
var _ Tool = (*GoTool)(nil)

// GoTool implements Tool interface for Go toolchain management
type GoTool struct {
	*BaseTool
}

func getGoBinaryName() string {
	platformMapper := NewPlatformMapper()
	if platformMapper.IsWindows() {
		return "go.exe"
	}
	return "go"
}

// NewGoTool creates a new Go tool instance
func NewGoTool(manager *Manager) *GoTool {
	return &GoTool{
		BaseTool: NewBaseTool(manager, ToolGo, getGoBinaryName()),
	}
}

// Name returns the tool name
func (g *GoTool) Name() string {
	return ToolGo
}

// Install downloads and installs the specified Go version
func (g *GoTool) Install(version string, cfg config.ToolConfig) error {
	return g.StandardInstall(version, cfg, g.getDownloadURL)
}

// IsInstalled checks if the specified version is installed
func (g *GoTool) IsInstalled(version string, cfg config.ToolConfig) bool {
	return g.StandardIsInstalled(version, cfg, g.GetPath)
}

// GetPath returns the binary path for the specified version (for PATH management)
func (g *GoTool) GetPath(version string, cfg config.ToolConfig) (string, error) {
	return g.StandardGetPath(version, cfg, g.getInstalledPath)
}

func (g *GoTool) GetBinaryName() string {
	return getGoBinaryName()
}

// getInstalledPath returns the path for an installed Go version
func (g *GoTool) getInstalledPath(version string, cfg config.ToolConfig) (string, error) {
	installDir := g.manager.GetToolVersionDir(g.Name(), version, "")
	pathResolver := NewPathResolver(g.manager.GetToolsDir())
	binDir, err := pathResolver.FindBinaryParentDir(installDir, g.GetBinaryName())
	if err != nil {
		return "", err
	}
	return binDir, nil
}

// Verify checks if the installation is working correctly
func (g *GoTool) Verify(version string, cfg config.ToolConfig) error {
	verifyConfig := VerificationConfig{
		BinaryName:  g.GetBinaryName(),
		VersionArgs: []string{"version"},
		DebugInfo:   false,
	}
	return g.StandardVerifyWithConfig(version, cfg, verifyConfig)
}

// ListVersions returns available versions for installation
func (g *GoTool) ListVersions() ([]string, error) {
	// Try to fetch versions from Go releases API
	versions, err := g.fetchGoVersions()
	if err != nil {
		// Fallback to known versions if API is unavailable
		return g.getFallbackGoVersions(), nil
	}

	// If API returned empty results, use fallback
	if len(versions) == 0 {
		return g.getFallbackGoVersions(), nil
	}

	return version.SortVersions(versions), nil
}

// fetchGoVersions fetches Go versions from GitHub releases API
func (g *GoTool) fetchGoVersions() ([]string, error) {
	registry := g.manager.GetRegistry()
	resp, err := registry.GetHTTPClient().Get(GoGithubAPIBase + "/tags?per_page=100")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch Go versions: status %d", resp.StatusCode)
	}

	var tags []struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, err
	}

	var versions []string
	for _, tag := range tags {
		// Go tags are like "go1.21.0", "go1.20.5", etc.
		if strings.HasPrefix(tag.Name, "go") && g.isValidGoVersion(tag.Name[2:]) {
			versions = append(versions, tag.Name[2:]) // Remove "go" prefix
		}
	}

	return versions, nil
}

// isValidGoVersion checks if a version string looks like a valid Go version
func (g *GoTool) isValidGoVersion(version string) bool {
	// Simple validation: should contain dots and numbers
	return strings.Contains(version, ".") && len(version) > 2
}

// getFallbackGoVersions returns known Go versions as fallback
func (g *GoTool) getFallbackGoVersions() []string {
	return []string{
		"1.24.2", "1.24.1", "1.24.0",
		"1.23.4", "1.23.3", "1.23.2", "1.23.1", "1.23.0",
		"1.22.10", "1.22.9", "1.22.8", "1.22.7", "1.22.6", "1.22.5", "1.22.4", "1.22.3", "1.22.2", "1.22.1", "1.22.0",
		"1.21.13", "1.21.12", "1.21.11", "1.21.10", "1.21.9", "1.21.8", "1.21.7", "1.21.6", "1.21.5", "1.21.4", "1.21.3", "1.21.2", "1.21.1", "1.21.0",
		"1.20.14", "1.20.13", "1.20.12", "1.20.11", "1.20.10", "1.20.9", "1.20.8", "1.20.7", "1.20.6", "1.20.5", "1.20.4", "1.20.3", "1.20.2", "1.20.1", "1.20.0",
	}
}

// GetDownloadOptions returns download options specific to Go
func (g *GoTool) GetDownloadOptions() DownloadOptions {
	return DownloadOptions{
		FileExtension: ExtTarGz,
	}
}

// GetDisplayName returns the display name for Go
func (g *GoTool) GetDisplayName() string {
	return "Go"
}

// getDownloadURL returns the download URL for the specified version
func (g *GoTool) getDownloadURL(version string) string {
	platformMapper := NewPlatformMapper()
	osName := platformMapper.GetOS()
	arch := platformMapper.GetArch()

	// Determine file extension
	var fileExt string
	if platformMapper.IsWindows() {
		fileExt = ExtZip
	} else {
		fileExt = ExtTarGz
	}

	// Construct filename
	filename := fmt.Sprintf("go%s.%s-%s%s", version, osName, arch, fileExt)

	return fmt.Sprintf("%s/%s", GoDevAPIBase, filename)
}

// ResolveVersion resolves a Go version specification to a concrete version
func (g *GoTool) ResolveVersion(versionSpec, distribution string) (string, error) {
	availableVersions, err := g.ListVersions()
	if err != nil {
		return "", err
	}

	spec, err := version.ParseSpec(versionSpec)
	if err != nil {
		return "", fmt.Errorf("invalid version specification %s: %w", versionSpec, err)
	}

	resolved, err := spec.Resolve(availableVersions)
	if err != nil {
		return "", fmt.Errorf("failed to resolve Go version %s: %w", versionSpec, err)
	}

	return resolved, nil
}

// GetChecksum implements ChecksumProvider interface for Go
func (g *GoTool) GetChecksum(version, filename string) (ChecksumInfo, error) {
	fmt.Printf("  🔍 Fetching Go checksum from go.dev API...\n")

	checksum, err := g.fetchGoChecksum(version, filename)
	if err != nil {
		fmt.Printf("  ⚠️  Failed to get Go checksum from go.dev API: %v\n", err)
		return ChecksumInfo{}, err
	}

	fmt.Printf("  ✅ Found Go checksum from go.dev API\n")
	return ChecksumInfo{
		Type:  SHA256,
		Value: checksum,
	}, nil
}

// GoRelease represents a Go release from go.dev API
type GoRelease struct {
	Version string   `json:"version"`
	Stable  bool     `json:"stable"`
	Files   []GoFile `json:"files"`
}

// GoFile represents a Go file from go.dev API
type GoFile struct {
	Filename string `json:"filename"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Version  string `json:"version"`
	SHA256   string `json:"sha256"`
	Size     int64  `json:"size"`
	Kind     string `json:"kind"`
}

// fetchGoChecksum fetches Go checksum from go.dev API
func (g *GoTool) fetchGoChecksum(version, filename string) (string, error) {
	url := GoDevAPIBase + "/?mode=json&include=all"

	client := &http.Client{
		Timeout: 120 * time.Second, // 2 minutes for slow servers
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch Go checksums: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Go checksums returned status %d", resp.StatusCode)
	}

	// Parse the JSON response
	var releases []GoRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return "", fmt.Errorf("failed to decode Go API response: %w", err)
	}

	// Find the matching version and file
	for _, release := range releases {
		if release.Version == "go"+version {
			for _, file := range release.Files {
				if file.Filename == filename && file.Kind == "archive" {
					return file.SHA256, nil
				}
			}
		}
	}

	return "", fmt.Errorf("no matching Go file found for version %s, filename %s", version, filename)
}

// GetDownloadURL implements URLProvider interface for Go
func (g *GoTool) GetDownloadURL(version string) string {
	return g.getDownloadURL(version)
}

// GetChecksumURL implements URLProvider interface for Go
func (g *GoTool) GetChecksumURL(version, filename string) string {
	return GoDevAPIBase + "/?mode=json&include=all"
}

// GetVersionsURL implements URLProvider interface for Go
func (g *GoTool) GetVersionsURL() string {
	return GitHubAPIBase + "/repos/golang/go/tags?per_page=100"
}
