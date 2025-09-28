package tools

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gnodet/mvx/pkg/config"
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
		BaseTool: NewBaseTool(manager, "go", getGoBinaryName()),
	}
}

// Name returns the tool name
func (g *GoTool) Name() string {
	return "go"
}

// Install downloads and installs the specified Go version
func (g *GoTool) Install(version string, cfg config.ToolConfig) error {
	return g.StandardInstall(version, cfg, g.getDownloadURL)
}

// IsInstalled checks if the specified version is installed
func (g *GoTool) IsInstalled(version string, cfg config.ToolConfig) bool {
	return g.StandardIsInstalled(version, cfg, g.GetPath, g.GetBinaryName())
}

// GetPath returns the binary path for the specified version (for PATH management)
func (g *GoTool) GetPath(version string, cfg config.ToolConfig) (string, error) {
	return g.StandardGetPath(version, cfg, g.getInstalledPath, g.GetBinaryName())
}

func (g *GoTool) GetBinaryName() string {
	return getGoBinaryName()
}

// getInstalledPath returns the path for an installed Go version
func (g *GoTool) getInstalledPath(version string, cfg config.ToolConfig) (string, error) {
	installDir := g.manager.GetToolVersionDir("go", version, "")
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
	registry := g.manager.GetRegistry()
	return registry.GetGoVersions()
}

// GetDownloadOptions returns download options specific to Go
func (g *GoTool) GetDownloadOptions() DownloadOptions {
	return DownloadOptions{
		FileExtension: ".tar.gz",
		ExpectedType:  "application",
		MinSize:       50 * 1024 * 1024,  // 50MB
		MaxSize:       200 * 1024 * 1024, // 200MB
		ArchiveType:   "tar.gz",
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
		fileExt = ".zip"
	} else {
		fileExt = ".tar.gz"
	}

	// Construct filename
	filename := fmt.Sprintf("go%s.%s-%s%s", version, osName, arch, fileExt)

	return fmt.Sprintf("https://go.dev/dl/%s", filename)
}

// ResolveVersion resolves a Go version specification to a concrete version
func (g *GoTool) ResolveVersion(version, distribution string) (string, error) {
	registry := g.manager.GetRegistry()
	return registry.ResolveGoVersion(version)
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
	url := "https://go.dev/dl/?mode=json&include=all"

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
