package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestShellCommand tests the mvx shell command functionality
func TestShellCommand(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a test configuration with Java and Go
	configContent := `{
  project: {
    name: "shell-test",
    description: "Test configuration for shell command"
  },
  tools: {
    java: {
      version: "17",
      distribution: "zulu"
    },
    go: {
      version: "1.24.2"
    }
  }
}`

	// Create .mvx directory and config file
	mvxDir := filepath.Join(tempDir, ".mvx")
	if err := os.MkdirAll(mvxDir, 0755); err != nil {
		t.Fatalf("Failed to create .mvx directory: %v", err)
	}

	configPath := filepath.Join(mvxDir, "config.json5")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Get the mvx binary path
	mvxBinary := findMvxBinary(t)

	tests := []struct {
		name        string
		command     string
		expectError bool
		contains    []string // Strings that should be present in output
	}{
		{
			name:        "echo simple message",
			command:     `echo "Hello from mvx shell"`,
			expectError: false,
			contains:    []string{"Hello from mvx shell"},
		},
		{
			name:        "show java home",
			command:     `echo $JAVA_HOME`,
			expectError: false,
			contains:    []string{".mvx/tools/java"},
		},
		{
			name:        "run java version",
			command:     `java -version`,
			expectError: false,
			contains:    []string{"openjdk version", "17.0"},
		},
		{
			name:        "run go version",
			command:     `go version`,
			expectError: false,
			contains:    []string{"go version", "1.24.2"},
		},
		{
			name:        "check PATH contains mvx tools",
			command:     `echo $PATH`,
			expectError: false,
			contains:    []string{".mvx/tools/java", ".mvx/tools/go"},
		},
		{
			name:        "create and list directory",
			command:     `mkdir test-shell-dir && ls -la`,
			expectError: false,
			contains:    []string{"test-shell-dir"},
		},
		{
			name:        "change directory and show path",
			command:     `cd test-shell-dir && pwd`,
			expectError: false,
			contains:    []string{"test-shell-dir"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(mvxBinary, "shell", tt.command)
			cmd.Dir = tempDir
			cmd.Env = os.Environ() // Don't set MVX_VERSION=dev to avoid using global dev binary

			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for command %s, but it succeeded. Output: %s", tt.command, outputStr)
			} else if !tt.expectError && err != nil {
				t.Errorf("Command %s failed unexpectedly: %v. Output: %s", tt.command, err, outputStr)
			}

			// Check that expected strings are present in output
			for _, expected := range tt.contains {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Command %s output does not contain expected string '%s'. Output: %s", tt.command, expected, outputStr)
				}
			}

			t.Logf("Command %s output: %s", tt.command, outputStr)
		})
	}
}

// TestShellCommandHelp tests the shell command help functionality
func TestShellCommandHelp(t *testing.T) {
	tempDir := t.TempDir()
	mvxBinary := findMvxBinary(t)

	cmd := exec.Command(mvxBinary, "shell", "--help")
	cmd.Dir = tempDir
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Shell help command failed: %v. Output: %s", err, string(output))
	}

	outputStr := string(output)
	expectedStrings := []string{
		"Execute shell commands",
		"mvx-managed tools",
		"Examples:",
		"mvx shell echo",
		"Usage:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Help output does not contain expected string '%s'. Output: %s", expected, outputStr)
		}
	}
}

// TestShellCommandNoArgs tests shell command without arguments
func TestShellCommandNoArgs(t *testing.T) {
	tempDir := t.TempDir()
	mvxBinary := findMvxBinary(t)

	cmd := exec.Command(mvxBinary, "shell")
	cmd.Dir = tempDir
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Errorf("Expected error when running shell command without arguments, but it succeeded. Output: %s", string(output))
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "no command specified") {
		t.Errorf("Error message should mention 'no command specified'. Output: %s", outputStr)
	}
}
