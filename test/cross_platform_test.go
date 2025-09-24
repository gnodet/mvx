package test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gnodet/mvx/pkg/config"
	"github.com/gnodet/mvx/pkg/executor"
	"github.com/gnodet/mvx/pkg/tools"
)

func TestCrossPlatformScripts(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a test configuration with cross-platform scripts
	cfg := &config.Config{
		Project: config.ProjectConfig{
			Name:        "test-project",
			Description: "Test project for cross-platform scripts",
		},
		Commands: map[string]config.CommandConfig{
			"platform-specific": {
				Description: "Platform-specific script test",
				Script: map[string]interface{}{
					"windows": "echo Windows script",
					"linux":   "echo Linux script",
					"darwin":  "echo macOS script",
					"unix":    "echo Unix script",
					"default": "echo Default script",
				},
			},
			"mvx-shell-test": {
				Description: "Cross-platform mvx-shell test",
				Script:      "mkdir testdir && cd testdir && echo Created directory && cd ..",
				Interpreter: "mvx-shell",
			},
			"native-shell-test": {
				Description: "Native shell test",
				Script:      "echo Native shell test",
				Interpreter: "native",
			},
			"default-interpreter": {
				Description: "Default interpreter test",
				Script:      "echo Default interpreter test",
				// No interpreter specified - should default to native
			},
		},
	}

	// Create tool manager
	toolManager, err := tools.NewManager()
	if err != nil {
		t.Fatalf("Failed to create tool manager: %v", err)
	}

	// Create executor
	exec := executor.NewExecutor(cfg, toolManager, tempDir)

	t.Run("PlatformSpecificScript", func(t *testing.T) {
		err := exec.ExecuteCommand("platform-specific", []string{})
		if err != nil {
			t.Errorf("Failed to execute platform-specific command: %v", err)
		}
	})

	t.Run("MVXShellScript", func(t *testing.T) {
		err := exec.ExecuteCommand("mvx-shell-test", []string{})
		if err != nil {
			t.Errorf("Failed to execute mvx-shell command: %v", err)
		}

		// Check if directory was created
		testDirPath := filepath.Join(tempDir, "testdir")
		if _, err := os.Stat(testDirPath); os.IsNotExist(err) {
			t.Errorf("mvx-shell command did not create expected directory: %s", testDirPath)
		}
	})

	t.Run("NativeShellScript", func(t *testing.T) {
		err := exec.ExecuteCommand("native-shell-test", []string{})
		if err != nil {
			t.Errorf("Failed to execute native shell command: %v", err)
		}
	})

	t.Run("DefaultInterpreter", func(t *testing.T) {
		err := exec.ExecuteCommand("default-interpreter", []string{})
		if err != nil {
			t.Errorf("Failed to execute default interpreter command: %v", err)
		}
	})
}

func TestPlatformScriptResolution(t *testing.T) {
	tests := []struct {
		name           string
		script         interface{}
		expectedResult string
		shouldError    bool
	}{
		{
			name:           "simple string script",
			script:         "echo hello",
			expectedResult: "echo hello",
			shouldError:    false,
		},
		{
			name: "platform-specific with current platform",
			script: map[string]interface{}{
				"windows": "echo windows",
				"linux":   "echo linux",
				"darwin":  "echo darwin",
				"default": "echo default",
			},
			expectedResult: getExpectedForCurrentPlatform(),
			shouldError:    false,
		},
		{
			name: "platform-specific with unix fallback",
			script: map[string]interface{}{
				"unix":    "echo unix",
				"default": "echo default",
			},
			expectedResult: getExpectedForUnix(),
			shouldError:    false,
		},
		{
			name: "platform-specific with only default",
			script: map[string]interface{}{
				"unsupported": "echo unsupported",
				"default":     "echo default",
			},
			expectedResult: "echo default",
			shouldError:    false,
		},
		{
			name: "platform-specific with no match",
			script: map[string]interface{}{
				"unsupported": "echo unsupported",
			},
			expectedResult: "",
			shouldError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := config.ResolvePlatformScript(tt.script)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expectedResult {
				t.Errorf("Expected %q, got %q", tt.expectedResult, result)
			}
		})
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.Config
		shouldError bool
		errorMsg    string
	}{
		{
			name: "valid config with platform-specific scripts",
			config: &config.Config{
				Project: config.ProjectConfig{Name: "test"},
				Commands: map[string]config.CommandConfig{
					"test-cmd": {
						Description: "Test command",
						Script: map[string]interface{}{
							"windows": "echo windows",
							"unix":    "echo unix",
						},
					},
				},
			},
			shouldError: false,
		},
		{
			name: "valid config with mvx-shell interpreter",
			config: &config.Config{
				Project: config.ProjectConfig{Name: "test"},
				Commands: map[string]config.CommandConfig{
					"test-cmd": {
						Description: "Test command",
						Script:      "echo hello",
						Interpreter: "mvx-shell",
					},
				},
			},
			shouldError: false,
		},
		{
			name: "invalid interpreter",
			config: &config.Config{
				Project: config.ProjectConfig{Name: "test"},
				Commands: map[string]config.CommandConfig{
					"test-cmd": {
						Description: "Test command",
						Script:      "echo hello",
						Interpreter: "invalid-interpreter",
					},
				},
			},
			shouldError: true,
			errorMsg:    "invalid interpreter",
		},
		{
			name: "empty platform-specific script",
			config: &config.Config{
				Project: config.ProjectConfig{Name: "test"},
				Commands: map[string]config.CommandConfig{
					"test-cmd": {
						Description: "Test command",
						Script: map[string]interface{}{
							"windows": "",
							"unix":    "",
						},
					},
				},
			},
			shouldError: true,
			errorMsg:    "script is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestIntelligentInterpreterDefaults(t *testing.T) {
	tests := []struct {
		name                string
		script              interface{}
		explicitInterpreter string
		expectedInterpreter string
		description         string
	}{
		{
			name:                "simple string script defaults to mvx-shell",
			script:              "echo hello",
			explicitInterpreter: "",
			expectedInterpreter: "mvx-shell",
			description:         "Simple scripts should default to cross-platform mvx-shell",
		},
		{
			name: "platform-specific script defaults to native",
			script: map[string]interface{}{
				"windows": "echo windows",
				"unix":    "echo unix",
			},
			explicitInterpreter: "",
			expectedInterpreter: "native",
			description:         "Platform-specific scripts should default to native",
		},
		{
			name: "nested platform script defaults to native",
			script: map[string]interface{}{
				"windows": map[string]interface{}{
					"script": "echo windows",
				},
				"unix": map[string]interface{}{
					"script": "echo unix",
				},
			},
			explicitInterpreter: "",
			expectedInterpreter: "native",
			description:         "Nested platform scripts should default to native",
		},
		{
			name:                "explicit interpreter overrides default",
			script:              "echo hello",
			explicitInterpreter: "native",
			expectedInterpreter: "native",
			description:         "Explicit interpreter should override intelligent default",
		},
		{
			name: "explicit interpreter overrides platform default",
			script: map[string]interface{}{
				"windows": "echo windows",
				"unix":    "echo unix",
			},
			explicitInterpreter: "mvx-shell",
			expectedInterpreter: "mvx-shell",
			description:         "Explicit interpreter should override platform default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, interpreter, err := config.ResolvePlatformScriptWithInterpreter(tt.script, tt.explicitInterpreter)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if interpreter != tt.expectedInterpreter {
				t.Errorf("Expected interpreter %q, got %q: %s", tt.expectedInterpreter, interpreter, tt.description)
			}
		})
	}
}

// Helper functions
func getExpectedForCurrentPlatform() string {
	switch runtime.GOOS {
	case "windows":
		return "echo windows"
	case "linux":
		return "echo linux"
	case "darwin":
		return "echo darwin"
	default:
		return "echo default"
	}
}

func getExpectedForUnix() string {
	switch runtime.GOOS {
	case "windows":
		return "echo default"
	default:
		return "echo unix"
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
