package config

import (
	"runtime"
	"testing"
)

func TestResolvePlatformScript(t *testing.T) {
	tests := []struct {
		name     string
		script   interface{}
		expected string
		wantErr  bool
	}{
		{
			name:     "simple string script",
			script:   "echo hello",
			expected: "echo hello",
			wantErr:  false,
		},
		{
			name: "platform-specific map with current platform",
			script: map[string]interface{}{
				"windows": "echo windows",
				"linux":   "echo linux",
				"darwin":  "echo darwin",
				"default": "echo default",
			},
			expected: getPlatformExpected(),
			wantErr:  false,
		},
		{
			name: "platform-specific map with unix fallback",
			script: map[string]interface{}{
				"unix":    "echo unix",
				"default": "echo default",
			},
			expected: getUnixExpected(),
			wantErr:  false,
		},
		{
			name: "platform-specific map with default fallback",
			script: map[string]interface{}{
				"unsupported": "echo unsupported",
				"default":     "echo default",
			},
			expected: "echo default",
			wantErr:  false,
		},
		{
			name: "platform-specific map with no matching platform",
			script: map[string]interface{}{
				"unsupported": "echo unsupported",
			},
			expected: "",
			wantErr:  true,
		},
		{
			name: "PlatformScript struct",
			script: PlatformScript{
				Windows: "echo windows",
				Linux:   "echo linux",
				Darwin:  "echo darwin",
				Default: "echo default",
			},
			expected: getPlatformExpected(),
			wantErr:  false,
		},
		{
			name:     "invalid script type",
			script:   123,
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ResolvePlatformScript(tt.script)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolvePlatformScript() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("ResolvePlatformScript() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestHasValidScript(t *testing.T) {
	tests := []struct {
		name     string
		script   interface{}
		expected bool
	}{
		{
			name:     "non-empty string",
			script:   "echo hello",
			expected: true,
		},
		{
			name:     "empty string",
			script:   "",
			expected: false,
		},
		{
			name: "map with valid scripts",
			script: map[string]interface{}{
				"windows": "echo windows",
				"linux":   "echo linux",
			},
			expected: true,
		},
		{
			name: "map with empty scripts",
			script: map[string]interface{}{
				"windows": "",
				"linux":   "",
			},
			expected: false,
		},
		{
			name: "PlatformScript with valid scripts",
			script: PlatformScript{
				Windows: "echo windows",
			},
			expected: true,
		},
		{
			name:     "PlatformScript with empty scripts",
			script:   PlatformScript{},
			expected: false,
		},
		{
			name:     "invalid type",
			script:   123,
			expected: false,
		},
		{
			name:     "nil",
			script:   nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasValidScript(tt.script)
			if result != tt.expected {
				t.Errorf("HasValidScript() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Helper functions to get expected results based on current platform
func getPlatformExpected() string {
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

func getUnixExpected() string {
	switch runtime.GOOS {
	case "windows":
		return "echo default"
	default:
		return "echo unix"
	}
}
