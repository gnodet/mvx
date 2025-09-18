package tools

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gnodet/mvx/pkg/config"
)

// NodeTool manages Node.js
// Downloads from https://nodejs.org/dist/
type NodeTool struct {
	manager *Manager
}

func (n *NodeTool) Name() string { return "node" }

func (n *NodeTool) Install(version string, cfg config.ToolConfig) error {
	installDir := n.manager.GetToolVersionDir("node", version, "")
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}

	url := n.getDownloadURL(version)
	fmt.Printf("  ⏳ Downloading Node.js %s...\n", version)
	if err := n.downloadAndExtract(url, installDir); err != nil {
		return fmt.Errorf("failed to download and extract: %w", err)
	}
	return nil
}

func (n *NodeTool) IsInstalled(version string, cfg config.ToolConfig) bool {
	installDir := n.manager.GetToolVersionDir("node", version, "")
	bin := filepath.Join(installDir, "bin", n.nodeBinaryName())
	if _, err := os.Stat(bin); err == nil {
		return true
	}
	// also search recursively (archives include subdir)
	found := ""
	filepath.Walk(installDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.EqualFold(info.Name(), n.nodeBinaryName()) {
			found = path
			return filepath.SkipDir
		}
		return nil
	})
	return found != ""
}

func (n *NodeTool) GetPath(version string, cfg config.ToolConfig) (string, error) {
	installDir := n.manager.GetToolVersionDir("node", version, "")
	entries, err := os.ReadDir(installDir)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "node-") {
			return filepath.Join(installDir, e.Name()), nil
		}
	}
	return installDir, nil
}

func (n *NodeTool) GetBinPath(version string, cfg config.ToolConfig) (string, error) {
	home, err := n.GetPath(version, cfg)
	if err != nil {
		return "", err
	}
	bin := filepath.Join(home, "bin")
	if info, err := os.Stat(bin); err == nil && info.IsDir() {
		return bin, nil
	}
	// On Windows, Node distributes binaries at root (no bin dir)
	return home, nil
}

func (n *NodeTool) Verify(version string, cfg config.ToolConfig) error {
	bin, err := n.GetBinPath(version, cfg)
	if err != nil {
		return err
	}
	nodeExe := filepath.Join(bin, n.nodeBinaryName())
	cmd := exec.Command(nodeExe, "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("node verification failed: %w\nOutput: %s", err, out)
	}
	if !strings.Contains(string(out), version) {
		// Allow 'v' prefix in output
		if !strings.Contains(string(out), "v"+version) {
			return fmt.Errorf("node version mismatch: expected %s, got %s", version, out)
		}
	}
	return nil
}

func (n *NodeTool) ListVersions() ([]string, error) {
	return n.manager.registry.GetNodeVersions()
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

func (n *NodeTool) downloadAndExtract(url, destDir string) error {
	tmp, err := os.CreateTemp("", "node-*.pkg")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	client := &http.Client{
		Timeout: 300 * time.Second, // 5 minute timeout
	}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed: %d from %s", resp.StatusCode, url)
	}
	if _, err := io.Copy(tmp, resp.Body); err != nil {
		return err
	}
	_ = tmp.Close()

	if strings.HasSuffix(url, ".zip") {
		return extractZipFile(tmp.Name(), destDir)
	}
	// Use system tar to extract .tar.xz
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}
	cmd := exec.Command("tar", "-xJf", tmp.Name(), "-C", destDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func extractZipFile(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}
	for _, f := range r.File {
		path := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid path: %s", f.Name)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(path, f.Mode()); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}
		_, err = io.Copy(out, rc)
		rc.Close()
		out.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// ToolConfig is referenced to avoid import cycle; re-use config.ToolConfig via type alias
// but here inline to keep file self-contained for tool implementation
// In real code, we use config.ToolConfig; manager passes it in.
