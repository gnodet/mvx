package tools

import (
	"testing"

	"github.com/gnodet/mvx/pkg/config"
)

func TestPythonTool_Name(t *testing.T) {
	tool := &PythonTool{}
	if tool.Name() != "python" {
		t.Errorf("Expected tool name 'python', got '%s'", tool.Name())
	}
}

func TestPythonTool_GetDownloadURL(t *testing.T) {
	tool := &PythonTool{}

	tests := []struct {
		version string
		wantErr bool
	}{
		{"3.12.0", false},
		{"3.11.5", false},
		{"3.10.8", false},
		{"invalid", false}, // Should not error, just return a URL
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			url, err := tool.getDownloadURL(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDownloadURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && url == "" {
				t.Errorf("getDownloadURL() returned empty URL for version %s", tt.version)
			}
		})
	}
}

func TestPythonTool_BinaryNames(t *testing.T) {
	tool := &PythonTool{}

	pythonBinary := tool.pythonBinaryName()
	pipBinary := tool.pipBinaryName()

	if pythonBinary == "" {
		t.Error("pythonBinaryName() returned empty string")
	}

	if pipBinary == "" {
		t.Error("pipBinaryName() returned empty string")
	}

	// Test that binary names are platform appropriate
	// This is a basic test - in a real scenario we'd mock runtime.GOOS
	if pythonBinary != "python3" && pythonBinary != "python.exe" {
		t.Errorf("Unexpected python binary name: %s", pythonBinary)
	}

	if pipBinary != "pip3" && pipBinary != "pip.exe" {
		t.Errorf("Unexpected pip binary name: %s", pipBinary)
	}
}

func TestPythonTool_GetPath(t *testing.T) {
	manager := &Manager{
		cacheDir: "/tmp/test-cache",
		tools:    make(map[string]Tool),
		registry: NewToolRegistry(),
	}
	tool := &PythonTool{manager: manager}
	cfg := config.ToolConfig{}

	path, err := tool.GetPath("3.12.0", cfg)
	if err != nil {
		t.Errorf("GetPath() error = %v", err)
	}

	if path == "" {
		t.Error("GetPath() returned empty path")
	}

	expected := "/tmp/test-cache/tools/python/3.12.0"
	if path != expected {
		t.Errorf("GetPath() = %v, want %v", path, expected)
	}
}

func TestPythonTool_GetBinPath(t *testing.T) {
	manager := &Manager{
		cacheDir: "/tmp/test-cache",
		tools:    make(map[string]Tool),
		registry: NewToolRegistry(),
	}
	tool := &PythonTool{manager: manager}
	cfg := config.ToolConfig{}

	binPath, err := tool.GetBinPath("3.12.0", cfg)
	if err != nil {
		t.Errorf("GetBinPath() error = %v", err)
	}

	if binPath == "" {
		t.Error("GetBinPath() returned empty path")
	}

	// Should contain the base path
	if !contains(binPath, "/tmp/test-cache/tools/python/3.12.0") {
		t.Errorf("GetBinPath() = %v, should contain base path", binPath)
	}
}

func TestPythonTool_GetProjectVenvPath(t *testing.T) {
	manager := &Manager{
		cacheDir: "/tmp/test-cache",
		tools:    make(map[string]Tool),
		registry: NewToolRegistry(),
	}
	tool := &PythonTool{manager: manager}

	projectPath := "/home/user/my-project"
	venvPath := tool.GetProjectVenvPath("3.12.0", projectPath)

	if venvPath == "" {
		t.Error("GetProjectVenvPath() returned empty path")
	}

	// Should be in the project's .mvx directory
	expected := "/home/user/my-project/.mvx/venv"
	if venvPath != expected {
		t.Errorf("GetProjectVenvPath() = %v, want %v", venvPath, expected)
	}

	// Should contain .mvx/venv
	if !contains(venvPath, ".mvx/venv") {
		t.Errorf("GetProjectVenvPath() = %v, should contain .mvx/venv", venvPath)
	}
}

func TestPythonTool_GetProjectEnvironment(t *testing.T) {
	manager := &Manager{
		cacheDir: "/tmp/test-cache",
		tools:    make(map[string]Tool),
		registry: NewToolRegistry(),
	}
	tool := &PythonTool{manager: manager}
	cfg := config.ToolConfig{}

	projectPath := "/home/user/my-project"
	env, err := tool.GetProjectEnvironment("3.12.0", cfg, projectPath)
	if err != nil {
		t.Errorf("GetProjectEnvironment() error = %v", err)
	}

	// Check that important environment variables are set
	expectedVars := []string{"PYTHON_HOME", "VIRTUAL_ENV", "PATH", "PYTHONPATH", "PYTHONNOUSERSITE"}
	for _, envVar := range expectedVars {
		if _, exists := env[envVar]; !exists {
			t.Errorf("GetProjectEnvironment() missing environment variable: %s", envVar)
		}
	}

	// Check that PYTHONNOUSERSITE is set to "1" for isolation
	if env["PYTHONNOUSERSITE"] != "1" {
		t.Errorf("GetProjectEnvironment() PYTHONNOUSERSITE = %v, want 1", env["PYTHONNOUSERSITE"])
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		(len(s) > len(substr) && s[len(s)-len(substr):] == substr) ||
		(len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
