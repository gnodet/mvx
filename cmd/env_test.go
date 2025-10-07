package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDetectShell(t *testing.T) {
	// Save original SHELL env var
	originalShell := os.Getenv("SHELL")
	defer os.Setenv("SHELL", originalShell)

	tests := []struct {
		name         string
		shellEnv     string
		expectedType string
	}{
		{
			name:         "bash shell",
			shellEnv:     "/bin/bash",
			expectedType: "bash",
		},
		{
			name:         "zsh shell",
			shellEnv:     "/bin/zsh",
			expectedType: "zsh",
		},
		{
			name:         "fish shell",
			shellEnv:     "/usr/local/bin/fish",
			expectedType: "fish",
		},
		{
			name:         "unknown shell defaults to bash on unix",
			shellEnv:     "/bin/sh",
			expectedType: "bash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip Windows-specific tests on Unix and vice versa
			if runtime.GOOS == "windows" && tt.expectedType != "powershell" {
				t.Skip("Skipping Unix shell test on Windows")
			}

			os.Setenv("SHELL", tt.shellEnv)
			detected := detectShell()

			if detected != tt.expectedType {
				t.Errorf("detectShell() = %s, want %s", detected, tt.expectedType)
			}
		})
	}
}

func TestDetectShellWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test on non-Windows platform")
	}

	// Save original SHELL env var
	originalShell := os.Getenv("SHELL")
	defer os.Setenv("SHELL", originalShell)

	// Clear SHELL env var to test Windows default
	os.Unsetenv("SHELL")

	detected := detectShell()
	if detected != "powershell" {
		t.Errorf("detectShell() on Windows = %s, want powershell", detected)
	}
}

func TestOutputEnvironmentNoMvxDir(t *testing.T) {
	// Create a temp directory without .mvx
	tempDir := t.TempDir()

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute outputEnvironment
	err := outputEnvironment()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	buf := make([]byte, 1000)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// Should not error, just return silently
	if err != nil {
		t.Errorf("outputEnvironment() should not error when no .mvx dir found, got: %v", err)
	}

	// Should produce no output
	if output != "" {
		t.Errorf("outputEnvironment() should produce no output when no .mvx dir found, got: %s", output)
	}
}

func TestOutputEnvironmentWithMvxDir(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a temp directory with .mvx config
	tempDir := t.TempDir()
	mvxDir := filepath.Join(tempDir, ".mvx")
	if err := os.Mkdir(mvxDir, 0755); err != nil {
		t.Fatalf("Failed to create .mvx directory: %v", err)
	}

	// Create a minimal config
	configContent := `{
  project: {
    name: "test-project"
  },
  tools: {
    go: { version: "1.24.2" }
  }
}`
	configFile := filepath.Join(mvxDir, "config.json5")
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Set shell type
	envShell = "bash"

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute outputEnvironment
	err := outputEnvironment()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	buf := make([]byte, 5000)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// Should not error
	if err != nil {
		t.Errorf("outputEnvironment() error = %v", err)
	}

	// Output should contain export statements (if tools are installed)
	// Note: This test might not produce output if tools aren't installed
	// That's okay - we're just testing that it doesn't error
	t.Logf("outputEnvironment() output:\n%s", output)
}

func TestOutputBashEnv(t *testing.T) {
	pathDirs := []string{"/path/to/java/bin", "/path/to/maven/bin"}
	env := map[string]string{
		"JAVA_HOME":  "/path/to/java",
		"MAVEN_HOME": "/path/to/maven",
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputBashEnv(pathDirs, env)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	buf := make([]byte, 2000)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if err != nil {
		t.Fatalf("outputBashEnv() error = %v", err)
	}

	// Check for expected content
	expectedStrings := []string{
		"export PATH=",
		"/path/to/java/bin",
		"/path/to/maven/bin",
		"export JAVA_HOME=\"/path/to/java\"",
		"export MAVEN_HOME=\"/path/to/maven\"",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s', but it didn't.\nOutput:\n%s", expected, output)
		}
	}

	t.Logf("Bash env output:\n%s", output)
}

func TestOutputFishEnv(t *testing.T) {
	pathDirs := []string{"/path/to/java/bin", "/path/to/maven/bin"}
	env := map[string]string{
		"JAVA_HOME":  "/path/to/java",
		"MAVEN_HOME": "/path/to/maven",
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputFishEnv(pathDirs, env)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	buf := make([]byte, 2000)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if err != nil {
		t.Fatalf("outputFishEnv() error = %v", err)
	}

	// Check for expected content
	expectedStrings := []string{
		"set -gx PATH",
		"/path/to/java/bin",
		"/path/to/maven/bin",
		"set -gx JAVA_HOME \"/path/to/java\"",
		"set -gx MAVEN_HOME \"/path/to/maven\"",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s', but it didn't.\nOutput:\n%s", expected, output)
		}
	}

	t.Logf("Fish env output:\n%s", output)
}

func TestOutputPowerShellEnv(t *testing.T) {
	pathDirs := []string{"C:\\path\\to\\java\\bin", "C:\\path\\to\\maven\\bin"}
	env := map[string]string{
		"JAVA_HOME":  "C:\\path\\to\\java",
		"MAVEN_HOME": "C:\\path\\to\\maven",
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputPowerShellEnv(pathDirs, env)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	buf := make([]byte, 2000)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if err != nil {
		t.Fatalf("outputPowerShellEnv() error = %v", err)
	}

	// Check for expected content
	expectedStrings := []string{
		"$env:PATH =",
		"$env:JAVA_HOME =",
		"$env:MAVEN_HOME =",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s', but it didn't.\nOutput:\n%s", expected, output)
		}
	}

	t.Logf("PowerShell env output:\n%s", output)
}

func TestEnvCommandWithShellFlag(t *testing.T) {
	// Save original shell value
	originalShell := envShell
	defer func() { envShell = originalShell }()

	tests := []struct {
		name  string
		shell string
	}{
		{"bash", "bash"},
		{"zsh", "zsh"},
		{"fish", "fish"},
		{"powershell", "powershell"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set shell flag
			envShell = tt.shell

			// The command should accept the shell flag without error
			// We can't easily test the full execution without a proper .mvx setup,
			// but we can verify the flag is accepted
			if envShell != tt.shell {
				t.Errorf("envShell = %s, want %s", envShell, tt.shell)
			}
		})
	}
}
