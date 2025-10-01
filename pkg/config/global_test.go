package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGlobalConfig_LoadSaveRoundtrip(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mvx-global-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override the global config directory for testing
	originalGlobalConfigDirFunc := globalConfigDirFunc
	defer func() {
		globalConfigDirFunc = originalGlobalConfigDirFunc
	}()

	globalConfigDirFunc = func() (string, error) {
		return tempDir, nil
	}

	// Test configuration
	testConfig := &GlobalConfig{
		URLReplacements: map[string]string{
			"github.com":         "nexus.mycompany.net",
			"archive.apache.org": "apache-mirror.internal.com",
			"regex:^http://(.+)": "https://$1",
			"regex:https://github\\.com/([^/]+)/([^/]+)/releases/download/(.+)": "https://hub.corp.com/artifactory/github/$1/$2/$3",
		},
	}

	// Save the configuration
	err = SaveGlobalConfig(testConfig)
	if err != nil {
		t.Fatalf("Failed to save global config: %v", err)
	}

	// Verify the file was created
	configPath := filepath.Join(tempDir, "config.json5")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("Config file was not created at %s", configPath)
	}

	// Load the configuration back
	loadedConfig, err := LoadGlobalConfig()
	if err != nil {
		t.Fatalf("Failed to load global config: %v", err)
	}

	// Verify the loaded configuration matches
	if len(loadedConfig.URLReplacements) != len(testConfig.URLReplacements) {
		t.Errorf("Loaded config has %d URL replacements, expected %d",
			len(loadedConfig.URLReplacements), len(testConfig.URLReplacements))
	}

	for pattern, expectedReplacement := range testConfig.URLReplacements {
		if actualReplacement, exists := loadedConfig.URLReplacements[pattern]; !exists {
			t.Errorf("Pattern %s not found in loaded config", pattern)
		} else if actualReplacement != expectedReplacement {
			t.Errorf("Pattern %s: got %s, expected %s", pattern, actualReplacement, expectedReplacement)
		}
	}
}

func TestGlobalConfig_LoadNonExistent(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mvx-global-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override the global config directory for testing
	originalGlobalConfigDirFunc := globalConfigDirFunc
	defer func() {
		globalConfigDirFunc = originalGlobalConfigDirFunc
	}()

	globalConfigDirFunc = func() (string, error) {
		return tempDir, nil
	}

	// Load configuration when no file exists
	config, err := LoadGlobalConfig()
	if err != nil {
		t.Fatalf("LoadGlobalConfig should not fail when no file exists: %v", err)
	}

	// Should return empty configuration
	if config == nil {
		t.Fatal("LoadGlobalConfig returned nil config")
	}

	if len(config.URLReplacements) != 0 {
		t.Errorf("Expected empty URL replacements, got %d", len(config.URLReplacements))
	}
}

func TestGlobalConfig_FormatJSON5(t *testing.T) {
	tests := []struct {
		name     string
		config   *GlobalConfig
		contains []string
	}{
		{
			name: "Empty config",
			config: &GlobalConfig{
				URLReplacements: map[string]string{},
			},
			contains: []string{
				"// Global mvx configuration",
				"// See: https://mvx.dev/docs/url-replacements for documentation",
			},
		},
		{
			name: "Config with simple replacements",
			config: &GlobalConfig{
				URLReplacements: map[string]string{
					"github.com":         "nexus.mycompany.net",
					"archive.apache.org": "apache-mirror.internal.com",
				},
			},
			contains: []string{
				"// Global mvx configuration",
				"// URL replacements for enterprise networks and mirrors",
				"url_replacements:",
				"\"github.com\": \"nexus.mycompany.net\"",
				"\"archive.apache.org\": \"apache-mirror.internal.com\"",
			},
		},
		{
			name: "Config with regex patterns",
			config: &GlobalConfig{
				URLReplacements: map[string]string{
					"regex:^http://(.+)": "https://$1",
				},
			},
			contains: []string{
				"\"regex:^http://(.+)\": \"https://$1\"",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatGlobalAsJSON5(tt.config)
			if err != nil {
				t.Fatalf("FormatGlobalAsJSON5 failed: %v", err)
			}

			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected output to contain %q, but it didn't.\nActual output:\n%s", expected, result)
				}
			}
		})
	}
}

func TestGlobalConfig_EscapeJSONString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No escaping needed",
			input:    "simple.string",
			expected: "simple.string",
		},
		{
			name:     "Escape quotes",
			input:    "string with \"quotes\"",
			expected: "string with \\\"quotes\\\"",
		},
		{
			name:     "Escape backslashes",
			input:    "string\\with\\backslashes",
			expected: "string\\\\with\\\\backslashes",
		},
		{
			name:     "Escape both quotes and backslashes",
			input:    "regex:\\\"pattern\\\"",
			expected: "regex:\\\\\\\"pattern\\\\\\\"",
		},
		{
			name:     "Regex pattern with capture groups",
			input:    "https://$1",
			expected: "https://$1", // $ doesn't need escaping in JSON
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeJSONString(tt.input)
			if result != tt.expected {
				t.Errorf("escapeJSONString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGlobalConfig_GetGlobalConfigPath(t *testing.T) {
	path, err := GetGlobalConfigPath()
	if err != nil {
		t.Fatalf("GetGlobalConfigPath failed: %v", err)
	}

	if !strings.HasSuffix(path, "config.json5") {
		t.Errorf("Expected path to end with config.json5, got %s", path)
	}

	if !strings.Contains(path, ".mvx") {
		t.Errorf("Expected path to contain .mvx directory, got %s", path)
	}
}

func TestGlobalConfig_SaveCreatesDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mvx-global-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Use a subdirectory that doesn't exist yet
	configDir := filepath.Join(tempDir, "nonexistent", ".mvx")

	// Override the global config directory for testing
	originalGlobalConfigDirFunc := globalConfigDirFunc
	defer func() {
		globalConfigDirFunc = originalGlobalConfigDirFunc
	}()

	globalConfigDirFunc = func() (string, error) {
		return configDir, nil
	}

	// Save configuration - should create the directory
	testConfig := &GlobalConfig{
		URLReplacements: map[string]string{
			"github.com": "nexus.mycompany.net",
		},
	}

	err = SaveGlobalConfig(testConfig)
	if err != nil {
		t.Fatalf("SaveGlobalConfig failed: %v", err)
	}

	// Verify the directory was created
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Errorf("Config directory was not created at %s", configDir)
	}

	// Verify the file was created
	configPath := filepath.Join(configDir, "config.json5")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Config file was not created at %s", configPath)
	}
}
