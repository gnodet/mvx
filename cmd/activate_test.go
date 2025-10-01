package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnodet/mvx/pkg/shell"
)

func TestGetMvxBinaryPath(t *testing.T) {
	path, err := getMvxBinaryPath()
	if err != nil {
		t.Fatalf("getMvxBinaryPath() failed: %v", err)
	}

	if path == "" {
		t.Error("getMvxBinaryPath() returned empty path")
	}

	// Check that the path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("getMvxBinaryPath() returned non-existent path: %s", path)
	}

	t.Logf("mvx binary path: %s", path)
}

func TestActivateCommand(t *testing.T) {
	// Save original args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	tests := []struct {
		name           string
		shell          string
		expectedInHook []string
		shouldError    bool
	}{
		{
			name:  "bash activation",
			shell: "bash",
			expectedInHook: []string{
				"# mvx shell integration for bash",
				"_mvx_hook()",
				"PROMPT_COMMAND",
				"mvx_deactivate()",
				".mvx",
			},
			shouldError: false,
		},
		{
			name:  "zsh activation",
			shell: "zsh",
			expectedInHook: []string{
				"# mvx shell integration for zsh",
				"_mvx_hook()",
				"precmd",
				"mvx_deactivate()",
				".mvx",
			},
			shouldError: false,
		},
		{
			name:  "fish activation",
			shell: "fish",
			expectedInHook: []string{
				"# mvx shell integration for fish",
				"function _mvx_hook",
				"--on-variable PWD",
				"mvx_deactivate",
				".mvx",
			},
			shouldError: false,
		},
		{
			name:  "powershell activation",
			shell: "powershell",
			expectedInHook: []string{
				"# mvx shell integration for PowerShell",
				"function global:_mvx_hook",
				"prompt",
				"mvx-deactivate",
				".mvx",
			},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Set up command args
			os.Args = []string{"mvx", "activate", tt.shell}

			// Execute activate command
			activateCmd.Run(activateCmd, []string{tt.shell})

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout

			// Read captured output
			buf := make([]byte, 10000)
			n, _ := r.Read(buf)
			output := string(buf[:n])

			// Check for expected content
			for _, expected := range tt.expectedInHook {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected hook to contain '%s', but it didn't.\nOutput:\n%s", expected, output)
				}
			}

			// Check that the hook contains the mvx binary path
			mvxPath, _ := getMvxBinaryPath()
			if !strings.Contains(output, mvxPath) {
				t.Errorf("Expected hook to contain mvx binary path '%s', but it didn't.\nOutput:\n%s", mvxPath, output)
			}

			t.Logf("Generated hook for %s:\n%s", tt.shell, output)
		})
	}
}

func TestActivateCommandUnsupportedShell(t *testing.T) {
	// Test that GenerateHook returns error for unsupported shell
	_, err := shell.GenerateHook("unsupported-shell", "/usr/local/bin/mvx")
	if err == nil {
		t.Error("Expected error for unsupported shell, got nil")
	}

	if !strings.Contains(err.Error(), "unsupported shell") {
		t.Errorf("Expected error message about unsupported shell, got: %v", err)
	}
}

func TestDeactivateCommand(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute deactivate command
	deactivateCmd.Run(deactivateCmd, []string{})

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	buf := make([]byte, 2000)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// Check for expected content
	expectedStrings := []string{
		"deactivate mvx",
		"Bash/Zsh",
		"Fish",
		"PowerShell",
		"mvx_deactivate",
		"mvx-deactivate",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected deactivate output to contain '%s', but it didn't.\nOutput:\n%s", expected, output)
		}
	}

	t.Logf("Deactivate command output:\n%s", output)
}

func TestActivateCommandIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Build a test binary
	tempDir := t.TempDir()
	mvxBinary := filepath.Join(tempDir, "mvx-test")

	// Note: This assumes we're running from the project root
	// In a real test environment, you might need to adjust the path
	t.Logf("Building test binary at %s", mvxBinary)

	// For now, we'll skip the actual build and just test the command structure
	// A full integration test would build the binary and test the actual shell hooks
	t.Skip("Full integration test requires building binary - implement if needed")
}
