package tools

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gnodet/mvx/pkg/config"
)

func TestRustTool_Name(t *testing.T) {
	tool := &RustTool{}
	if tool.Name() != "rust" {
		t.Errorf("Expected name 'rust', got '%s'", tool.Name())
	}
}

func TestRustTool_GetDownloadURL(t *testing.T) {
	tool := &RustTool{}
	
	url := tool.getDownloadURL()
	
	// Check that URL contains expected components
	if !strings.Contains(url, "static.rust-lang.org") {
		t.Errorf("Expected URL to contain 'static.rust-lang.org', got %s", url)
	}
	
	if !strings.Contains(url, "rustup-init") {
		t.Errorf("Expected URL to contain 'rustup-init', got %s", url)
	}

	// Check platform-specific URL components
	switch runtime.GOOS {
	case "windows":
		if !strings.Contains(url, "x86_64-pc-windows-msvc") {
			t.Errorf("Expected Windows URL to contain 'x86_64-pc-windows-msvc', got %s", url)
		}
		if !strings.Contains(url, ".exe") {
			t.Errorf("Expected Windows URL to contain '.exe', got %s", url)
		}
	case "darwin":
		if runtime.GOARCH == "arm64" {
			if !strings.Contains(url, "aarch64-apple-darwin") {
				t.Errorf("Expected macOS ARM64 URL to contain 'aarch64-apple-darwin', got %s", url)
			}
		} else {
			if !strings.Contains(url, "x86_64-apple-darwin") {
				t.Errorf("Expected macOS x64 URL to contain 'x86_64-apple-darwin', got %s", url)
			}
		}
	case "linux":
		if runtime.GOARCH == "arm64" {
			if !strings.Contains(url, "aarch64-unknown-linux-gnu") {
				t.Errorf("Expected Linux ARM64 URL to contain 'aarch64-unknown-linux-gnu', got %s", url)
			}
		} else {
			if !strings.Contains(url, "x86_64-unknown-linux-gnu") {
				t.Errorf("Expected Linux x64 URL to contain 'x86_64-unknown-linux-gnu', got %s", url)
			}
		}
	}
}

func TestRustTool_IsInstalled(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "rust-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock manager
	manager := &Manager{
		cacheDir: tempDir,
	}
	tool := &RustTool{manager: manager}
	cfg := config.ToolConfig{Version: "1.84.0"}

	// Test when not installed
	if tool.IsInstalled("1.84.0", cfg) {
		t.Error("Expected IsInstalled to return false for non-existent installation")
	}

	// Create a mock installation structure
	installDir := manager.GetToolVersionDir("rust", "1.84.0", "")
	toolchainDir := filepath.Join(installDir, "toolchains", "stable-x86_64-unknown-linux-gnu")
	binDir := filepath.Join(toolchainDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("Failed to create mock installation: %v", err)
	}

	// Create mock rust executables
	rustcExe := filepath.Join(binDir, "rustc")
	cargoExe := filepath.Join(binDir, "cargo")
	if runtime.GOOS == "windows" {
		rustcExe += ".exe"
		cargoExe += ".exe"
	}

	if err := os.WriteFile(rustcExe, []byte("#!/bin/bash\necho rustc"), 0755); err != nil {
		t.Fatalf("Failed to create mock rustc executable: %v", err)
	}
	if err := os.WriteFile(cargoExe, []byte("#!/bin/bash\necho cargo"), 0755); err != nil {
		t.Fatalf("Failed to create mock cargo executable: %v", err)
	}

	// Test when installed
	if !tool.IsInstalled("1.84.0", cfg) {
		t.Error("Expected IsInstalled to return true for existing installation")
	}
}

func TestRustTool_GetPath(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "rust-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock manager
	manager := &Manager{
		cacheDir: tempDir,
	}
	tool := &RustTool{manager: manager}
	cfg := config.ToolConfig{Version: "1.84.0"}

	// Create a mock installation structure
	installDir := manager.GetToolVersionDir("rust", "1.84.0", "")
	toolchainDir := filepath.Join(installDir, "toolchains", "stable-x86_64-unknown-linux-gnu")
	if err := os.MkdirAll(toolchainDir, 0755); err != nil {
		t.Fatalf("Failed to create mock installation: %v", err)
	}

	// Test GetPath
	path, err := tool.GetPath("1.84.0", cfg)
	if err != nil {
		t.Fatalf("GetPath failed: %v", err)
	}

	expectedPath := toolchainDir
	if path != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, path)
	}
}

func TestRustTool_GetBinPath(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "rust-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock manager
	manager := &Manager{
		cacheDir: tempDir,
	}
	tool := &RustTool{manager: manager}
	cfg := config.ToolConfig{Version: "1.84.0"}

	// Create a mock installation structure
	installDir := manager.GetToolVersionDir("rust", "1.84.0", "")
	toolchainDir := filepath.Join(installDir, "toolchains", "stable-x86_64-unknown-linux-gnu")
	binDir := filepath.Join(toolchainDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("Failed to create mock installation: %v", err)
	}

	// Test GetBinPath
	binPath, err := tool.GetBinPath("1.84.0", cfg)
	if err != nil {
		t.Fatalf("GetBinPath failed: %v", err)
	}

	expectedBinPath := binDir
	if binPath != expectedBinPath {
		t.Errorf("Expected bin path %s, got %s", expectedBinPath, binPath)
	}
}
