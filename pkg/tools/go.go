package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/gnodet/mvx/pkg/config"
)

// Compile-time interface validation
var _ Tool = (*GoTool)(nil)

// useSystemGo checks if system Go should be used instead of downloading
func useSystemGo() bool {
	return UseSystemTool("go")
}

// GoTool implements Tool interface for Go toolchain management
type GoTool struct {
	*BaseTool
}

// NewGoTool creates a new Go tool instance
func NewGoTool(manager *Manager) *GoTool {
	return &GoTool{
		BaseTool: NewBaseTool(manager, "go"),
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
	return g.StandardIsInstalled(version, cfg, g.GetPath, "go")
}

// GetPath returns the binary path for the specified version (for PATH management)
func (g *GoTool) GetPath(version string, cfg config.ToolConfig) (string, error) {
	return g.StandardGetPath(version, cfg, g.getInstalledPath, "go")
}

// getInstalledPath returns the path for an installed Go version
func (g *GoTool) getInstalledPath(version string, cfg config.ToolConfig) (string, error) {
	installDir := g.manager.GetToolVersionDir("go", version, "")

	// Go archives extract to a "go" subdirectory
	goPath := filepath.Join(installDir, "go")
	if _, err := os.Stat(goPath); err == nil {
		return filepath.Join(goPath, "bin"), nil
	}

	return filepath.Join(installDir, "bin"), nil
}

// Verify checks if the installation is working correctly
func (g *GoTool) Verify(version string, cfg config.ToolConfig) error {
	return g.StandardVerify(version, cfg, g.GetPath, "go", []string{"version"})
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
	// Determine platform string
	osName := runtime.GOOS
	arch := runtime.GOARCH

	// Map architecture names to Go's naming convention
	switch arch {
	case "amd64":
		arch = "amd64"
	case "arm64":
		arch = "arm64"
	case "386":
		arch = "386"
	}

	// Construct filename
	var filename string
	if osName == "windows" {
		filename = fmt.Sprintf("go%s.%s-%s.zip", version, osName, arch)
	} else {
		filename = fmt.Sprintf("go%s.%s-%s.tar.gz", version, osName, arch)
	}

	return fmt.Sprintf("https://go.dev/dl/%s", filename)
}
