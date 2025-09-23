package tools

import (
	"os"
	"testing"

	"github.com/gnodet/mvx/pkg/config"
)

func TestGetToolVersionOverride(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		envVar   string
		envValue string
		expected string
	}{
		{
			name:     "Java version override",
			toolName: "java",
			envVar:   "MVX_JAVA_VERSION",
			envValue: "21",
			expected: "21",
		},
		{
			name:     "Maven version override",
			toolName: "maven",
			envVar:   "MVX_MAVEN_VERSION",
			envValue: "3.9.6",
			expected: "3.9.6",
		},
		{
			name:     "No override set",
			toolName: "java",
			envVar:   "MVX_JAVA_VERSION",
			envValue: "",
			expected: "",
		},
		{
			name:     "Go version override",
			toolName: "go",
			envVar:   "MVX_GO_VERSION",
			envValue: "1.21.0",
			expected: "1.21.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment
			os.Unsetenv(tt.envVar)

			// Set environment variable if value is provided
			if tt.envValue != "" {
				os.Setenv(tt.envVar, tt.envValue)
				defer os.Unsetenv(tt.envVar)
			}

			result := getToolVersionOverride(tt.toolName)
			if result != tt.expected {
				t.Errorf("getToolVersionOverride(%s) = %s, want %s", tt.toolName, result, tt.expected)
			}
		})
	}
}

func TestGetToolVersionOverrideEnvVar(t *testing.T) {
	tests := []struct {
		toolName string
		expected string
	}{
		{"java", "MVX_JAVA_VERSION"},
		{"maven", "MVX_MAVEN_VERSION"},
		{"go", "MVX_GO_VERSION"},
		{"python", "MVX_PYTHON_VERSION"},
		{"node", "MVX_NODE_VERSION"},
	}

	for _, tt := range tests {
		t.Run(tt.toolName, func(t *testing.T) {
			result := getToolVersionOverrideEnvVar(tt.toolName)
			if result != tt.expected {
				t.Errorf("getToolVersionOverrideEnvVar(%s) = %s, want %s", tt.toolName, result, tt.expected)
			}
		})
	}
}

func TestResolveVersionWithOverride(t *testing.T) {
	// Create a test manager
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	tests := []struct {
		name           string
		toolName       string
		configVersion  string
		overrideVar    string
		overrideValue  string
		expectOverride bool
	}{
		{
			name:           "Java version override",
			toolName:       "java",
			configVersion:  "17",
			overrideVar:    "MVX_JAVA_VERSION",
			overrideValue:  "21",
			expectOverride: true,
		},
		{
			name:           "No override",
			toolName:       "java",
			configVersion:  "17",
			overrideVar:    "MVX_JAVA_VERSION",
			overrideValue:  "",
			expectOverride: false,
		},
		{
			name:           "Maven version override",
			toolName:       "maven",
			configVersion:  "3.9",
			overrideVar:    "MVX_MAVEN_VERSION",
			overrideValue:  "4.0.0-rc-4",
			expectOverride: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment
			os.Unsetenv(tt.overrideVar)

			// Set override if provided
			if tt.overrideValue != "" {
				os.Setenv(tt.overrideVar, tt.overrideValue)
				defer os.Unsetenv(tt.overrideVar)
			}

			// Test that the override is detected
			override := getToolVersionOverride(tt.toolName)
			if tt.expectOverride {
				if override != tt.overrideValue {
					t.Errorf("Expected override %s, got %s", tt.overrideValue, override)
				}
			} else {
				if override != "" {
					t.Errorf("Expected no override, got %s", override)
				}
			}

			// Test version resolution (this will use concrete versions to avoid network calls)
			if tt.expectOverride && tt.overrideValue != "" {
				// Use concrete version to avoid network resolution
				concreteConfig := config.ToolConfig{
					Version: tt.overrideValue,
				}
				resolved, err := manager.resolveVersion(tt.toolName, concreteConfig)
				if err != nil {
					t.Errorf("Failed to resolve version: %v", err)
				}
				if resolved != tt.overrideValue {
					t.Errorf("Expected resolved version %s, got %s", tt.overrideValue, resolved)
				}
			}
		})
	}
}
