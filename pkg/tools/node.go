package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"

	"github.com/gnodet/mvx/pkg/config"
	"github.com/gnodet/mvx/pkg/version"
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
	versions, err := n.fetchNodeVersions()
	if err != nil {
		// minimal fallback
		return []string{"22.5.1", "22.4.1", "20.15.0", "18.20.4"}, nil
	}
	return version.SortVersions(versions), nil
}

func (n *NodeTool) fetchNodeVersions() ([]string, error) {
	entries, err := n.fetchNodeIndex()
	if err != nil {
		return nil, err
	}
	var versions []string
	for _, e := range entries {
		v := strings.TrimPrefix(e.Version, "v")
		versions = append(versions, v)
	}
	return versions, nil
}

type nodeIndexEntry struct {
	Version string      `json:"version"`
	LTS     interface{} `json:"lts"`
}

func (n *NodeTool) fetchNodeIndex() ([]nodeIndexEntry, error) {
	resp, err := n.manager.Get(NodeJSDistBase + "/index.json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("node index fetch failed: %d", resp.StatusCode)
	}
	var entries []nodeIndexEntry
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// fetchNodeLTSVersions fetches available Node.js LTS versions
func (n *NodeTool) fetchNodeLTSVersions() ([]string, error) {
	entries, err := n.fetchNodeIndex()
	if err != nil {
		return nil, err
	}
	var versions []string
	for _, e := range entries {
		// LTS can be false (not LTS), true (LTS but no codename), or a string (LTS with codename)
		if e.LTS != nil && e.LTS != false {
			// Check if it's a boolean true or a string (both indicate LTS)
			if ltsValue, ok := e.LTS.(bool); ok && ltsValue {
				versions = append(versions, strings.TrimPrefix(e.Version, "v"))
			} else if ltsString, ok := e.LTS.(string); ok && ltsString != "" {
				versions = append(versions, strings.TrimPrefix(e.Version, "v"))
			}
		}
	}
	return versions, nil
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
func (n *NodeTool) ResolveVersion(versionSpec, distribution string) (string, error) {
	// Special handling for "lts"
	if versionSpec == "lts" {
		lts, err := n.fetchNodeLTSVersions()
		if err != nil || len(lts) == 0 {
			return "", fmt.Errorf("failed to resolve Node LTS version")
		}
		// Return highest LTS (first element since SortVersions returns descending order)
		sorted := version.SortVersions(lts)
		return sorted[0], nil
	}

	availableVersions, err := n.ListVersions()
	if err != nil {
		return "", err
	}

	spec, err := version.ParseSpec(versionSpec)
	if err != nil {
		return "", fmt.Errorf("invalid version specification %s: %w", versionSpec, err)
	}

	resolved, err := spec.Resolve(availableVersions)
	if err != nil {
		return "", fmt.Errorf("failed to resolve Node version %s: %w", versionSpec, err)
	}

	return resolved, nil
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

	resp, err := n.manager.Get(url)
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
