package tools

import (
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/gnodet/mvx/pkg/config"
)

// Compile-time interface validation
var _ Tool = (*NodeTool)(nil)

// NodeTool manages Node.js
// Downloads from https://nodejs.org/dist/
type NodeTool struct {
	*BaseTool
}

func getNodeBinaryName() string {
	if NewPlatformMapper().IsWindows() {
		return BinaryNode + ExtExe
	}
	return BinaryNode
}

func getNodeToolName() string {
	return ToolNode
}

// NewNodeTool creates a new Node tool instance
func NewNodeTool(manager *Manager) *NodeTool {
	return &NodeTool{
		BaseTool: NewBaseTool(manager, getNodeToolName(), getNodeBinaryName()),
	}
}

func (n *NodeTool) Name() string { return getNodeToolName() }

func (n *NodeTool) Install(version string, cfg config.ToolConfig) error {
	return n.StandardInstall(version, cfg, n.getDownloadURL)
}

func (n *NodeTool) IsInstalled(version string, cfg config.ToolConfig) bool {
	return n.StandardIsInstalled(version, cfg, n.GetPath)
}

func (n *NodeTool) GetPath(version string, cfg config.ToolConfig) (string, error) {
	return n.StandardGetPath(version, cfg, n.getInstalledPath)
}

func (n *NodeTool) GetBinaryName() string {
	return getNodeBinaryName()
}

// getInstalledPath returns the path for an installed Node version
func (n *NodeTool) getInstalledPath(version string, cfg config.ToolConfig) (string, error) {
	installDir := n.manager.GetToolVersionDir(n.Name(), version, "")
	pathResolver := NewPathResolver(n.manager.GetToolsDir())
	binDir, err := pathResolver.FindBinaryParentDir(installDir, n.GetBinaryName())
	if err != nil {
		return "", err
	}
	return binDir, nil
}

func (n *NodeTool) Verify(version string, cfg config.ToolConfig) error {
	verifyConfig := VerificationConfig{
		BinaryName:  n.GetBinaryName(),
		VersionArgs: []string{"--version"},
		DebugInfo:   false,
	}
	return n.StandardVerifyWithConfig(version, cfg, verifyConfig)
}

func (n *NodeTool) ListVersions() ([]string, error) {
	registry := n.manager.GetRegistry()
	return registry.GetNodeVersions()
}

// GetDownloadOptions returns download options specific to Node.js
func (n *NodeTool) GetDownloadOptions() DownloadOptions {
	return DownloadOptions{
		FileExtension: ExtTarXz,
	}
}

// GetDisplayName returns the display name for Node.js
func (n *NodeTool) GetDisplayName() string {
	return "Node.js"
}

func (n *NodeTool) getDownloadURL(version string) string {
	platformMapper := NewPlatformMapper()

	// Generate Node.js platform string
	var platform string
	switch platformMapper.GetOS() {
	case "linux":
		if platformMapper.GetArch() == "arm64" {
			platform = "linux-arm64"
		} else {
			platform = "linux-x64"
		}
	case "darwin":
		if platformMapper.GetArch() == "arm64" {
			platform = "darwin-arm64"
		} else {
			platform = "darwin-x64"
		}
	case "windows":
		platform = "win-x64"
	default:
		platform = "linux-x64" // fallback
	}

	// Determine file extension
	var fileExt string
	if platformMapper.IsWindows() {
		fileExt = ExtZip
	} else {
		fileExt = ExtTarGz
	}

	return fmt.Sprintf(NodeJSDistBase+"/v%[1]s/node-v%[1]s-%[2]s%[3]s", version, platform, fileExt)
}

// ResolveVersion resolves a Node version specification to a concrete version
func (n *NodeTool) ResolveVersion(version, distribution string) (string, error) {
	registry := n.manager.GetRegistry()
	return registry.ResolveNodeVersion(version)
}

// GetChecksum implements ChecksumProvider interface for Node.js
func (n *NodeTool) GetChecksum(version, filename string) (ChecksumInfo, error) {
	fmt.Printf("  🔍 Fetching Node.js checksum from SHASUMS256.txt...\n")

	checksum, err := n.fetchNodeChecksum(version, filename)
	if err != nil {
		fmt.Printf("  ⚠️  Failed to get Node.js checksum from SHASUMS256.txt: %v\n", err)
		return ChecksumInfo{}, err
	}

	fmt.Printf("  ✅ Found Node.js checksum from SHASUMS256.txt\n")
	return ChecksumInfo{
		Type:  SHA256,
		Value: checksum,
	}, nil
}

// fetchNodeChecksum fetches Node.js checksum from SHASUMS256.txt
func (n *NodeTool) fetchNodeChecksum(version, filename string) (string, error) {
	url := fmt.Sprintf("%s/v%s/SHASUMS256.txt", NodeJSDistBase, version)

	client := &http.Client{
		Timeout: 120 * time.Second, // 2 minutes for slow servers
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch Node.js checksums: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Node.js checksums returned status %d", resp.StatusCode)
	}

	// Read and parse the checksum file content
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read Node.js checksums: %w", err)
	}

	content := string(body)

	// Parse the checksum file to find the checksum for our filename
	nodeFilename := n.getNodeFilename(version)
	return n.parseNodeChecksumFile(content, nodeFilename)
}

// parseNodeChecksumFile parses Node.js SHASUMS256.txt content to find checksum for specific filename
func (n *NodeTool) parseNodeChecksumFile(content, filename string) (string, error) {
	lines := strings.Split(content, "\n")
	var candidateFiles []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on whitespace - Node.js format is: checksum  filename
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		checksum := parts[0]
		fileInLine := parts[1]

		// Store candidate files for debugging
		candidateFiles = append(candidateFiles, fileInLine)

		// Check if this line matches our filename (exact match)
		if fileInLine == filename {
			return checksum, nil
		}

		// Also try matching just the basename
		if strings.Contains(fileInLine, "/") {
			parts := strings.Split(fileInLine, "/")
			basename := parts[len(parts)-1]
			if basename == filename {
				return checksum, nil
			}
		}
	}

	// If we get here, no match was found - provide helpful debug info
	fmt.Printf("⚠️  Debug: Available Node.js files in SHASUMS256.txt: %v\n", candidateFiles)
	fmt.Printf("   Looking for: %s\n", filename)
	return "", fmt.Errorf("checksum not found for Node.js file %s", filename)
}

// getNodeFilename determines the correct Node.js filename based on version and platform
func (n *NodeTool) getNodeFilename(version string) string {
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
		return fmt.Sprintf("node-v%s-%s.zip", version, platform)
	}
	return fmt.Sprintf("node-v%s-%s.tar.xz", version, platform)
}

// GetDownloadURL implements URLProvider interface for Node.js
func (n *NodeTool) GetDownloadURL(version string) string {
	return n.getDownloadURL(version)
}

// GetChecksumURL implements URLProvider interface for Node.js
func (n *NodeTool) GetChecksumURL(version, filename string) string {
	return fmt.Sprintf("%s/v%s/SHASUMS256.txt", NodeJSDistBase, version)
}

// GetVersionsURL implements URLProvider interface for Node.js
func (n *NodeTool) GetVersionsURL() string {
	return NodeJSDistBase + "/index.json"
}
