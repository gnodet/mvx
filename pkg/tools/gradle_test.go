package tools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gnodet/mvx/pkg/config"
)

func TestGradleTool_Name(t *testing.T) {
	tool := &GradleTool{}
	if tool.Name() != "gradle" {
		t.Errorf("Expected name 'gradle', got '%s'", tool.Name())
	}
}

func TestGradleTool_GetDownloadURL(t *testing.T) {
	tool := &GradleTool{}
	
	tests := []struct {
		version     string
		expectedURL string
	}{
		{"8.5", "https://services.gradle.org/distributions/gradle-8.5-bin.zip"},
		{"7.6.4", "https://services.gradle.org/distributions/gradle-7.6.4-bin.zip"},
		{"6.9.4", "https://services.gradle.org/distributions/gradle-6.9.4-bin.zip"},
	}

	for _, test := range tests {
		url := tool.getDownloadURL(test.version)
		if url != test.expectedURL {
			t.Errorf("For version %s, expected URL %s, got %s", 
				test.version, test.expectedURL, url)
		}
	}
}

func TestGradleTool_IsInstalled(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gradle-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock manager
	manager := &Manager{
		cacheDir: tempDir,
	}
	tool := &GradleTool{manager: manager}
	cfg := config.ToolConfig{Version: "8.5"}

	// Test when not installed
	if tool.IsInstalled("8.5", cfg) {
		t.Error("Expected IsInstalled to return false for non-existent installation")
	}

	// Create a mock installation structure
	installDir := manager.GetToolVersionDir("gradle", "8.5", "")
	gradleDir := filepath.Join(installDir, "gradle-8.5")
	binDir := filepath.Join(gradleDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("Failed to create mock installation: %v", err)
	}

	// Create mock gradle executable
	gradleExe := filepath.Join(binDir, "gradle")
	if err := os.WriteFile(gradleExe, []byte("#!/bin/bash\necho gradle"), 0755); err != nil {
		t.Fatalf("Failed to create mock gradle executable: %v", err)
	}

	// Test when installed
	if !tool.IsInstalled("8.5", cfg) {
		t.Error("Expected IsInstalled to return true for existing installation")
	}
}

func TestGradleTool_GetPath(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gradle-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock manager
	manager := &Manager{
		cacheDir: tempDir,
	}
	tool := &GradleTool{manager: manager}
	cfg := config.ToolConfig{Version: "8.5"}

	// Create a mock installation structure
	installDir := manager.GetToolVersionDir("gradle", "8.5", "")
	gradleDir := filepath.Join(installDir, "gradle-8.5")
	if err := os.MkdirAll(gradleDir, 0755); err != nil {
		t.Fatalf("Failed to create mock installation: %v", err)
	}

	// Test GetPath
	path, err := tool.GetPath("8.5", cfg)
	if err != nil {
		t.Fatalf("GetPath failed: %v", err)
	}

	expectedPath := gradleDir
	if path != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, path)
	}
}

func TestGradleTool_GetBinPath(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gradle-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock manager
	manager := &Manager{
		cacheDir: tempDir,
	}
	tool := &GradleTool{manager: manager}
	cfg := config.ToolConfig{Version: "8.5"}

	// Create a mock installation structure
	installDir := manager.GetToolVersionDir("gradle", "8.5", "")
	gradleDir := filepath.Join(installDir, "gradle-8.5")
	binDir := filepath.Join(gradleDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("Failed to create mock installation: %v", err)
	}

	// Test GetBinPath
	binPath, err := tool.GetBinPath("8.5", cfg)
	if err != nil {
		t.Fatalf("GetBinPath failed: %v", err)
	}

	expectedBinPath := binDir
	if binPath != expectedBinPath {
		t.Errorf("Expected bin path %s, got %s", expectedBinPath, binPath)
	}
}

func TestGradleTool_FindGradleExecutable(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gradle-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tool := &GradleTool{}

	// Test when executable doesn't exist
	foundPath := tool.findGradleExecutable(tempDir)
	if foundPath != "" {
		t.Errorf("Expected empty path for non-existent executable, got %s", foundPath)
	}

	// Create mock gradle executable
	binDir := filepath.Join(tempDir, "gradle-8.5", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("Failed to create bin directory: %v", err)
	}

	gradleExe := filepath.Join(binDir, "gradle")
	if err := os.WriteFile(gradleExe, []byte("#!/bin/bash\necho gradle"), 0755); err != nil {
		t.Fatalf("Failed to create mock gradle executable: %v", err)
	}

	// Test when executable exists
	foundPath = tool.findGradleExecutable(tempDir)
	if foundPath != gradleExe {
		t.Errorf("Expected path %s, got %s", gradleExe, foundPath)
	}
}
