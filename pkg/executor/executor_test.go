package executor

import (
	"strings"
	"testing"

	"github.com/gnodet/mvx/pkg/config"
	"github.com/gnodet/mvx/pkg/tools"
)

func TestExecutor_SetupEnvironment(t *testing.T) {
	// Reset manager for test isolation
	tools.ResetManager()

	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a mock config with tools
	cfg := &config.Config{
		Tools: map[string]config.ToolConfig{
			"go": {
				Version: "1.24.2",
			},
			"java": {
				Version:      "17",
				Distribution: "zulu",
			},
		},
		Environment: map[string]string{
			"TEST_VAR": "test_value",
		},
		Commands: map[string]config.CommandConfig{
			"test-cmd": {
				Description: "Test command",
				Script:      "echo test",
				Environment: map[string]string{
					"CMD_VAR": "cmd_value",
				},
			},
		},
	}

	// Create tool manager
	manager, err := tools.NewManager()
	if err != nil {
		t.Fatalf("Failed to create tool manager: %v", err)
	}

	// Create executor
	executor := NewExecutor(cfg, manager, tempDir)

	// Test environment setup
	cmdConfig := cfg.Commands["test-cmd"]
	env, err := executor.setupEnvironment(cmdConfig)
	if err != nil {
		t.Fatalf("setupEnvironment() error = %v", err)
	}

	// Convert environment slice to map for easier testing
	envMap := make(map[string]string)
	for _, envVar := range env {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	// Test that global environment variables are set
	if envMap["TEST_VAR"] != "test_value" {
		t.Errorf("Expected TEST_VAR=test_value, got %s", envMap["TEST_VAR"])
	}

	// Test that command-specific environment variables are set
	if envMap["CMD_VAR"] != "cmd_value" {
		t.Errorf("Expected CMD_VAR=cmd_value, got %s", envMap["CMD_VAR"])
	}

	// Test that PATH is set (should contain current PATH)
	if _, exists := envMap["PATH"]; !exists {
		t.Error("PATH environment variable should be set")
	}
}

func TestExecutor_ProcessScriptString(t *testing.T) {
	// Reset manager for test isolation
	tools.ResetManager()

	tempDir := t.TempDir()
	cfg := &config.Config{}
	manager, _ := tools.NewManager()
	executor := NewExecutor(cfg, manager, tempDir)

	tests := []struct {
		name     string
		script   string
		args     []string
		expected string
	}{
		{
			name:     "script without args",
			script:   "echo hello",
			args:     []string{},
			expected: "echo hello",
		},
		{
			name:     "script with args",
			script:   "echo hello",
			args:     []string{"world", "test"},
			expected: "echo hello world test",
		},
		{
			name:     "complex script with args",
			script:   "go test -v",
			args:     []string{"./...", "-timeout=30s"},
			expected: "go test -v ./... -timeout=30s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.processScriptString(tt.script, tt.args)
			if result != tt.expected {
				t.Errorf("processScriptString() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestExecutor_ListCommands(t *testing.T) {
	// Reset manager for test isolation
	tools.ResetManager()

	tempDir := t.TempDir()
	cfg := &config.Config{
		Commands: map[string]config.CommandConfig{
			"build": {
				Description: "Build the project",
				Script:      "go build",
			},
			"test": {
				Description: "Run tests",
				Script:      "go test",
			},
			"fmt": {
				Description: "Format code",
				Script:      "go fmt",
			},
		},
	}
	manager, _ := tools.NewManager()
	executor := NewExecutor(cfg, manager, tempDir)

	commands := executor.ListCommands()

	expectedCommands := map[string]string{
		"build": "Build the project",
		"test":  "Run tests",
		"fmt":   "Format code",
	}

	if len(commands) != len(expectedCommands) {
		t.Errorf("Expected %d commands, got %d", len(expectedCommands), len(commands))
	}

	for name, description := range expectedCommands {
		if commands[name] != description {
			t.Errorf("Expected command %s with description %s, got %s", name, description, commands[name])
		}
	}
}

func TestExecutor_GetCommandInfo(t *testing.T) {
	// Reset manager for test isolation
	tools.ResetManager()

	tempDir := t.TempDir()
	cfg := &config.Config{
		Commands: map[string]config.CommandConfig{
			"build": {
				Description: "Build the project",
				Script:      "go build",
				Interpreter: "native",
			},
		},
	}
	manager, _ := tools.NewManager()
	executor := NewExecutor(cfg, manager, tempDir)

	// Test existing command
	cmdInfo, err := executor.GetCommandInfo("build")
	if err != nil {
		t.Errorf("GetCommandInfo() error = %v", err)
	}
	if cmdInfo.Description != "Build the project" {
		t.Errorf("Expected description 'Build the project', got %s", cmdInfo.Description)
	}
	if cmdInfo.Script != "go build" {
		t.Errorf("Expected script 'go build', got %s", cmdInfo.Script)
	}

	// Test non-existing command
	_, err = executor.GetCommandInfo("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent command")
	}
}

func TestExecutor_ValidateCommand(t *testing.T) {
	// Reset manager for test isolation
	tools.ResetManager()

	tempDir := t.TempDir()
	cfg := &config.Config{
		Tools: map[string]config.ToolConfig{
			"go": {
				Version: "1.24.2",
			},
		},
		Commands: map[string]config.CommandConfig{
			"valid-cmd": {
				Description: "Valid command",
				Script:      "echo test",
			},
			"requires-go": {
				Description: "Command requiring Go",
				Script:      "go version",
				Requires:    []string{"go"},
			},
			"requires-missing": {
				Description: "Command requiring missing tool",
				Script:      "missing-tool version",
				Requires:    []string{"missing-tool"},
			},
		},
	}
	manager, _ := tools.NewManager()
	executor := NewExecutor(cfg, manager, tempDir)

	// Test valid command without requirements
	err := executor.ValidateCommand("valid-cmd")
	if err != nil {
		t.Errorf("ValidateCommand() error = %v for valid command", err)
	}

	// Test command with missing tool requirement
	// Note: ValidateCommand is now deprecated and only checks command existence
	// Tools are auto-installed via EnsureTool, so this should NOT error
	err = executor.ValidateCommand("requires-missing")
	if err != nil {
		t.Errorf("ValidateCommand() error = %v (should not error for missing tools, they are auto-installed)", err)
	}

	// Test non-existent command
	err = executor.ValidateCommand("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent command")
	}
}

func TestExecutor_InterpreterSelection(t *testing.T) {
	// Reset manager for test isolation
	tools.ResetManager()

	cfg := &config.Config{}
	manager, _ := tools.NewManager()
	_ = NewExecutor(cfg, manager, "")

	tests := []struct {
		name                string
		script              interface{}
		defaultInterpreter  string
		expectedInterpreter string
		description         string
	}{
		{
			name:                "simple string defaults to mvx-shell",
			script:              "echo hello",
			defaultInterpreter:  "",
			expectedInterpreter: "mvx-shell",
			description:         "Simple string scripts should default to mvx-shell",
		},
		{
			name:                "explicit native interpreter",
			script:              "echo hello",
			defaultInterpreter:  "native",
			expectedInterpreter: "native",
			description:         "Explicit interpreter should override default",
		},
		{
			name:                "explicit mvx-shell interpreter",
			script:              "echo hello",
			defaultInterpreter:  "mvx-shell",
			expectedInterpreter: "mvx-shell",
			description:         "Explicit mvx-shell interpreter should be used",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script, interpreter, err := config.ResolvePlatformScriptWithInterpreter(tt.script, tt.defaultInterpreter)
			if err != nil {
				t.Errorf("ResolvePlatformScriptWithInterpreter() error = %v", err)
				return
			}

			if interpreter != tt.expectedInterpreter {
				t.Errorf("Expected interpreter %q, got %q: %s", tt.expectedInterpreter, interpreter, tt.description)
			}

			if script != "echo hello" {
				t.Errorf("Expected script 'echo hello', got %q", script)
			}
		})
	}
}
