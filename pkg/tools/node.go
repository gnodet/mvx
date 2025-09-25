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
var _ Tool = (*NodeTool)(nil)

// useSystemNode checks if system Node should be used instead of downloading
func useSystemNode() bool {
	return UseSystemTool("node")
}

// NodeTool manages Node.js
// Downloads from https://nodejs.org/dist/
type NodeTool struct {
	*BaseTool
}

// NewNodeTool creates a new Node tool instance
func NewNodeTool(manager *Manager) *NodeTool {
	return &NodeTool{
		BaseTool: NewBaseTool(manager, "node"),
	}
}

func (n *NodeTool) Name() string { return "node" }

func (n *NodeTool) Install(version string, cfg config.ToolConfig) error {
	return n.StandardInstall(version, cfg, n.getDownloadURL)
}

func (n *NodeTool) IsInstalled(version string, cfg config.ToolConfig) bool {
	return n.StandardIsInstalled(version, cfg, n.GetPath, "node")
}

func (n *NodeTool) GetPath(version string, cfg config.ToolConfig) (string, error) {
	return n.StandardGetPath(version, cfg, n.getInstalledPath, "node")
}

// getInstalledPath returns the path for an installed Node version
func (n *NodeTool) getInstalledPath(version string, cfg config.ToolConfig) (string, error) {
	installDir := n.manager.GetToolVersionDir("node", version, "")
	entries, err := os.ReadDir(installDir)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "node-") {
			home := filepath.Join(installDir, e.Name())
			bin := filepath.Join(home, "bin")
			if info, err := os.Stat(bin); err == nil && info.IsDir() {
				return bin, nil
			}
			// On Windows, Node distributes binaries at root (no bin dir)
			return home, nil
		}
	}

	// Check if binaries are directly in install directory
	bin := filepath.Join(installDir, "bin")
	if info, err := os.Stat(bin); err == nil && info.IsDir() {
		return bin, nil
	}
	// On Windows, Node distributes binaries at root (no bin dir)
	return installDir, nil
}

func (n *NodeTool) Verify(version string, cfg config.ToolConfig) error {
	return n.StandardVerify(version, cfg, n.GetPath, "node", []string{"--version"})
}

func (n *NodeTool) ListVersions() ([]string, error) {
	registry := n.manager.GetRegistry()
	return registry.GetNodeVersions()
}

// GetDownloadOptions returns download options specific to Node.js
func (n *NodeTool) GetDownloadOptions() DownloadOptions {
	return DownloadOptions{
		FileExtension: ".tar.xz",
		ExpectedType:  "application",
		MinSize:       15 * 1024 * 1024, // 15MB (Node.js 18.17.0 is ~19.5MB)
		MaxSize:       80 * 1024 * 1024, // 80MB
		ArchiveType:   "tar.xz",
	}
}

// GetDisplayName returns the display name for Node.js
func (n *NodeTool) GetDisplayName() string {
	return "Node.js"
}

func (n *NodeTool) nodeBinaryName() string {
	if runtime.GOOS == "windows" {
		return "node.exe"
	}
	return "node"
}

func (n *NodeTool) getDownloadURL(version string) string {
	platform := ""
	switch runtime.GOOS {
	case "linux":
		if runtime.GOARCH == "arm64" {
			platform = "linux-arm64"
		} else {
			platform = "linux-x64"
		}
	case "darwin":
		if runtime.GOARCH == "arm64" {
			platform = "darwin-arm64"
		} else {
			platform = "darwin-x64"
		}
	case "windows":
		platform = "win-x64"
	}
	// Windows uses zip, others tar.xz
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("https://nodejs.org/dist/v%[1]s/node-v%[1]s-%[2]s.zip", version, platform)
	}
	return fmt.Sprintf("https://nodejs.org/dist/v%[1]s/node-v%[1]s-%[2]s.tar.xz", version, platform)
}
