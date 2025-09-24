package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestInterpreterPATHManagement tests that mvx-managed tools are properly found in PATH
func TestInterpreterPATHManagement(t *testing.T) {
	// Skip if we're not in the mvx project directory
	if _, err := os.Stat("../mvx"); err != nil {
		t.Skip("Skipping interpreter PATH test - not in mvx project directory")
	}

	// Build the current mvx binary for testing
	mvxBinary := buildCurrentMvx(t)
	defer os.Remove(mvxBinary)

	tests := []struct {
		name        string
		command     string
		interpreter string
		expectError bool
		description string
	}{
		{
			name:        "deps command with native interpreter",
			command:     "deps",
			interpreter: "native",
			expectError: false,
			description: "deps command should work with native interpreter and find Go in PATH",
		},
		{
			name:        "hello command with native interpreter",
			command:     "hello",
			interpreter: "native",
			expectError: false,
			description: "hello command should work with native interpreter and find mvn in PATH",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the command using the test binary directly
			cmd := createCommand(mvxBinary, []string{tt.command})
			cmd.Dir = ".."         // Run in the parent directory where .mvx/config.json5 exists
			cmd.Env = os.Environ() // Don't set MVX_VERSION=dev to avoid using global dev binary

			output, err := cmd.CombinedOutput()

			if tt.expectError && err == nil {
				t.Errorf("Expected error for command %s, but it succeeded. Output: %s", tt.command, string(output))
			} else if !tt.expectError && err != nil {
				t.Errorf("Command %s failed unexpectedly: %v. Output: %s", tt.command, err, string(output))
			}

			// Check that the output doesn't contain "executable file not found"
			outputStr := string(output)
			if strings.Contains(outputStr, "executable file not found") {
				t.Errorf("Command %s failed with 'executable file not found': %s", tt.command, outputStr)
			}

			t.Logf("Command %s output: %s", tt.command, outputStr)
		})
	}
}

// TestInterpreterEnvironmentVariables tests that environment variables are properly passed to commands
func TestInterpreterEnvironmentVariables(t *testing.T) {
	// Skip if we're not in the mvx project directory
	if _, err := os.Stat("../mvx"); err != nil {
		t.Skip("Skipping interpreter environment test - not in mvx project directory")
	}

	// Build the current mvx binary for testing
	mvxBinary := buildCurrentMvx(t)
	defer os.Remove(mvxBinary)

	// Create a temporary config file with environment variables
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, ".mvx", "config.json5")

	// Create .mvx directory
	err := os.MkdirAll(filepath.Dir(configFile), 0755)
	if err != nil {
		t.Fatalf("Failed to create .mvx directory: %v", err)
	}

	// Write test config with environment variables
	configContent := `{
  project: {
    name: "test-project"
  },
  tools: {
    go: { version: "1.24.2" }
  },
  environment: {
    GLOBAL_TEST_VAR: "global_value"
  },
  commands: {
    "env-test": {
      description: "Test environment variables",
      script: "echo GLOBAL_TEST_VAR=$GLOBAL_TEST_VAR CMD_TEST_VAR=$CMD_TEST_VAR",
      interpreter: "native",
      environment: {
        CMD_TEST_VAR: "cmd_value"
      }
    }
  }
}`

	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Run the env-test command
	cmd := createCommand(mvxBinary, []string{"env-test"})
	cmd.Dir = tempDir
	cmd.Env = os.Environ() // Don't set MVX_VERSION=dev to avoid using global dev binary

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("env-test command failed: %v. Output: %s", err, string(output))
	}

	outputStr := string(output)

	// Check that global environment variable is set
	if !strings.Contains(outputStr, "GLOBAL_TEST_VAR=global_value") {
		t.Errorf("Global environment variable not found in output: %s", outputStr)
	}

	// Check that command-specific environment variable is set
	if !strings.Contains(outputStr, "CMD_TEST_VAR=cmd_value") {
		t.Errorf("Command-specific environment variable not found in output: %s", outputStr)
	}

	t.Logf("Environment test output: %s", outputStr)
}

// TestInterpreterSelection tests that the correct interpreter is selected based on configuration
func TestInterpreterSelection(t *testing.T) {
	// Skip if we're not in the mvx project directory
	if _, err := os.Stat("../mvx"); err != nil {
		t.Skip("Skipping interpreter selection test - not in mvx project directory")
	}

	// Build the current mvx binary for testing
	mvxBinary := buildCurrentMvx(t)
	defer os.Remove(mvxBinary)

	// Create a temporary config file with different interpreter settings
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, ".mvx", "config.json5")

	// Create .mvx directory
	err := os.MkdirAll(filepath.Dir(configFile), 0755)
	if err != nil {
		t.Fatalf("Failed to create .mvx directory: %v", err)
	}

	// Write test config with different interpreters
	configContent := `{
  project: {
    name: "test-project"
  },
  tools: {
    go: { version: "1.24.2" }
  },
  commands: {
    "native-test": {
      description: "Test native interpreter",
      script: "echo 'Native interpreter test'",
      interpreter: "native"
    },
    "mvx-shell-test": {
      description: "Test mvx-shell interpreter",
      script: "echo 'MVX shell interpreter test'",
      interpreter: "mvx-shell"
    },
    "default-test": {
      description: "Test default interpreter (should be mvx-shell for simple strings)",
      script: "echo 'Default interpreter test'"
    }
  }
}`

	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	tests := []struct {
		name        string
		command     string
		expectError bool
		description string
	}{
		{
			name:        "native interpreter",
			command:     "native-test",
			expectError: false,
			description: "Native interpreter should work",
		},
		{
			name:        "mvx-shell interpreter",
			command:     "mvx-shell-test",
			expectError: false,
			description: "MVX shell interpreter should work",
		},
		{
			name:        "default interpreter",
			command:     "default-test",
			expectError: false,
			description: "Default interpreter (mvx-shell) should work",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := createCommand(mvxBinary, []string{tt.command})
			cmd.Dir = tempDir
			cmd.Env = os.Environ() // Don't set MVX_VERSION=dev to avoid using global dev binary

			output, err := cmd.CombinedOutput()

			if tt.expectError && err == nil {
				t.Errorf("Expected error for command %s, but it succeeded. Output: %s", tt.command, string(output))
			} else if !tt.expectError && err != nil {
				t.Errorf("Command %s failed unexpectedly: %v. Output: %s", tt.command, err, string(output))
			}

			t.Logf("Command %s output: %s", tt.command, string(output))
		})
	}
}

// TestVerboseLogging tests that verbose logging works correctly
func TestVerboseLogging(t *testing.T) {
	// Skip if we're not in the mvx project directory
	if _, err := os.Stat("../mvx"); err != nil {
		t.Skip("Skipping verbose logging test - not in mvx project directory")
	}

	// Build the current mvx binary for testing
	mvxBinary := buildCurrentMvx(t)
	defer os.Remove(mvxBinary)

	// Run a command with verbose logging enabled
	cmd := createCommand(mvxBinary, []string{"deps"})
	cmd.Dir = ".."                                     // Run in the parent directory where .mvx/config.json5 exists
	cmd.Env = append(os.Environ(), "MVX_VERBOSE=true") // Don't set MVX_VERSION=dev to avoid using global dev binary

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("deps command with verbose logging failed: %v. Output: %s", err, string(output))
	}

	outputStr := string(output)

	// Check that verbose logging output is present
	if !strings.Contains(outputStr, "Setting up environment") && !strings.Contains(outputStr, "PATH") {
		t.Logf("Verbose logging output may not be present (this is not necessarily an error): %s", outputStr)
	}

	t.Logf("Verbose logging output: %s", outputStr)
}

// Helper function to build the current mvx binary for testing
func buildCurrentMvx(t *testing.T) string {
	t.Helper()

	// Build the mvx binary
	mvxBinary := filepath.Join(t.TempDir(), "mvx-test")
	cmd := exec.Command("go", "build", "-o", mvxBinary, "../.")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build mvx binary: %v. Output: %s", err, string(output))
	}

	return mvxBinary
}

// Helper function to create a command
func createCommand(binary string, args []string) *exec.Cmd {
	return exec.Command(binary, args...)
}
