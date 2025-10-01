package tools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gnodet/mvx/pkg/config"
)

func TestURLReplacer_SimpleReplacements(t *testing.T) {
	replacements := map[string]string{
		"github.com":         "nexus.mycompany.net",
		"archive.apache.org": "apache-mirror.internal.com",
		"dist.apache.org":    "apache-dist-mirror.internal.com",
	}

	replacer := NewURLReplacer(replacements)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "GitHub hostname replacement",
			input:    "https://github.com/owner/repo/archive/v1.0.0.tar.gz",
			expected: "https://nexus.mycompany.net/owner/repo/archive/v1.0.0.tar.gz",
		},
		{
			name:     "Apache archive replacement",
			input:    "https://archive.apache.org/dist/maven/maven-3/3.9.6/binaries/apache-maven-3.9.6-bin.zip",
			expected: "https://apache-mirror.internal.com/dist/maven/maven-3/3.9.6/binaries/apache-maven-3.9.6-bin.zip",
		},
		{
			name:     "Apache dist replacement",
			input:    "https://dist.apache.org/repos/dist/release/maven/maven-4/4.0.0-rc-4/binaries/apache-maven-4.0.0-rc-4-bin.zip",
			expected: "https://apache-dist-mirror.internal.com/repos/dist/release/maven/maven-4/4.0.0-rc-4/binaries/apache-maven-4.0.0-rc-4-bin.zip",
		},
		{
			name:     "No replacement needed",
			input:    "https://nodejs.org/dist/v18.17.0/node-v18.17.0-linux-x64.tar.xz",
			expected: "https://nodejs.org/dist/v18.17.0/node-v18.17.0-linux-x64.tar.xz",
		},
		{
			name:     "Partial match replaces substring",
			input:    "https://api.github.com/repos/owner/repo",
			expected: "https://api.nexus.mycompany.net/repos/owner/repo", // github.com is replaced
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := replacer.ApplyReplacements(tt.input)
			if result != tt.expected {
				t.Errorf("ApplyReplacements() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestURLReplacer_RegexReplacements(t *testing.T) {
	replacements := map[string]string{
		"regex:^http://(.+)": "https://$1",
		"regex:https://github\\.com/([^/]+)/([^/]+)/releases/download/(.+)": "https://hub.corp.com/artifactory/github/$1/$2/$3",
		"regex:https://([^.]+)\\.cdn\\.example\\.com/(.+)":                  "https://unified-cdn.com/$1/$2",
	}

	replacer := NewURLReplacer(replacements)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "HTTP to HTTPS conversion",
			input:    "http://example.com/file.zip",
			expected: "https://example.com/file.zip",
		},
		{
			name:     "GitHub releases restructuring",
			input:    "https://github.com/microsoft/vscode/releases/download/1.85.0/vscode-linux-x64.tar.gz",
			expected: "https://hub.corp.com/artifactory/github/microsoft/vscode/1.85.0/vscode-linux-x64.tar.gz",
		},
		{
			name:     "Subdomain to path conversion",
			input:    "https://assets.cdn.example.com/downloads/file.tar.gz",
			expected: "https://unified-cdn.com/assets/downloads/file.tar.gz",
		},
		{
			name:     "No regex match",
			input:    "https://nodejs.org/dist/v18.17.0/node-v18.17.0-linux-x64.tar.xz",
			expected: "https://nodejs.org/dist/v18.17.0/node-v18.17.0-linux-x64.tar.xz",
		},
		{
			name:     "HTTPS URL not affected by HTTP regex",
			input:    "https://example.com/file.zip",
			expected: "https://example.com/file.zip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := replacer.ApplyReplacements(tt.input)
			if result != tt.expected {
				t.Errorf("ApplyReplacements() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestURLReplacer_MixedReplacements(t *testing.T) {
	// Test with both simple and regex replacements
	replacements := map[string]string{
		"github.com":         "nexus.mycompany.net",
		"regex:^http://(.+)": "https://$1",
		"regex:https://github\\.com/([^/]+)/([^/]+)/releases/download/(.+)": "https://hub.corp.com/artifactory/github/$1/$2/$3",
		"archive.apache.org": "apache-mirror.internal.com",
	}

	replacer := NewURLReplacer(replacements)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple replacement takes precedence (first match wins)",
			input:    "https://github.com/owner/repo/releases/download/v1.0.0/file.tar.gz",
			expected: "https://nexus.mycompany.net/owner/repo/releases/download/v1.0.0/file.tar.gz", // Simple replacement wins
		},
		{
			name:     "HTTP conversion works",
			input:    "http://example.com/file.zip",
			expected: "https://example.com/file.zip",
		},
		{
			name:     "Apache replacement works",
			input:    "https://archive.apache.org/dist/maven/maven-3/3.9.6/binaries/apache-maven-3.9.6-bin.zip",
			expected: "https://apache-mirror.internal.com/dist/maven/maven-3/3.9.6/binaries/apache-maven-3.9.6-bin.zip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := replacer.ApplyReplacements(tt.input)
			if result != tt.expected {
				t.Errorf("ApplyReplacements() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestURLReplacer_Validation(t *testing.T) {
	tests := []struct {
		name         string
		replacements map[string]string
		expectErrors bool
		errorCount   int
	}{
		{
			name: "Valid patterns",
			replacements: map[string]string{
				"github.com":         "nexus.example.com",
				"regex:^http://(.+)": "https://$1",
			},
			expectErrors: false,
			errorCount:   0,
		},
		{
			name: "Invalid regex pattern",
			replacements: map[string]string{
				"regex:[invalid": "replacement",
			},
			expectErrors: true,
			errorCount:   1,
		},
		{
			name: "Multiple invalid patterns",
			replacements: map[string]string{
				"regex:[invalid": "replacement1",
				"regex:*invalid": "replacement2",
				"valid.com":      "mirror.com",
			},
			expectErrors: true,
			errorCount:   2,
		},
		{
			name:         "Empty replacements",
			replacements: map[string]string{},
			expectErrors: false,
			errorCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			replacer := NewURLReplacer(tt.replacements)
			errors := replacer.ValidateReplacements()

			if tt.expectErrors && len(errors) == 0 {
				t.Errorf("Expected validation errors but got none")
			}

			if !tt.expectErrors && len(errors) > 0 {
				t.Errorf("Expected no validation errors but got %d: %v", len(errors), errors)
			}

			if tt.expectErrors && len(errors) != tt.errorCount {
				t.Errorf("Expected %d validation errors but got %d: %v", tt.errorCount, len(errors), errors)
			}
		})
	}
}

func TestURLReplacer_GetMethods(t *testing.T) {
	replacements := map[string]string{
		"github.com":         "nexus.mycompany.net",
		"regex:^http://(.+)": "https://$1",
	}

	replacer := NewURLReplacer(replacements)

	// Test GetReplacementCount
	if count := replacer.GetReplacementCount(); count != 2 {
		t.Errorf("GetReplacementCount() = %d, want 2", count)
	}

	// Test GetReplacements
	result := replacer.GetReplacements()
	if len(result) != 2 {
		t.Errorf("GetReplacements() returned %d items, want 2", len(result))
	}

	// Verify it's a copy (modifying result shouldn't affect original)
	result["test"] = "value"
	if replacer.GetReplacementCount() != 2 {
		t.Errorf("GetReplacements() should return a copy, but original was modified")
	}
}

func TestURLReplacer_EmptyReplacements(t *testing.T) {
	replacer := NewURLReplacer(map[string]string{})

	testURL := "https://github.com/owner/repo/archive/v1.0.0.tar.gz"
	result := replacer.ApplyReplacements(testURL)

	if result != testURL {
		t.Errorf("ApplyReplacements() with empty replacements = %v, want %v", result, testURL)
	}

	if count := replacer.GetReplacementCount(); count != 0 {
		t.Errorf("GetReplacementCount() with empty replacements = %d, want 0", count)
	}
}

func TestURLReplacer_NilReplacements(t *testing.T) {
	replacer := NewURLReplacer(nil)

	testURL := "https://github.com/owner/repo/archive/v1.0.0.tar.gz"
	result := replacer.ApplyReplacements(testURL)

	if result != testURL {
		t.Errorf("ApplyReplacements() with nil replacements = %v, want %v", result, testURL)
	}

	if count := replacer.GetReplacementCount(); count != 0 {
		t.Errorf("GetReplacementCount() with nil replacements = %d, want 0", count)
	}
}

func TestLoadURLReplacer(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mvx-url-replacer-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override the global config directory for testing
	originalGlobalConfigDirFunc := config.GetGlobalConfigDirFunc()
	defer func() {
		config.SetGlobalConfigDirFunc(originalGlobalConfigDirFunc)
	}()

	config.SetGlobalConfigDirFunc(func() (string, error) {
		return tempDir, nil
	})

	tests := []struct {
		name          string
		globalConfig  *config.GlobalConfig
		expectError   bool
		expectedCount int
	}{
		{
			name: "Load with global config",
			globalConfig: &config.GlobalConfig{
				URLReplacements: map[string]string{
					"github.com":         "nexus.mycompany.net",
					"archive.apache.org": "apache-mirror.internal.com",
				},
			},
			expectError:   false,
			expectedCount: 2,
		},
		{
			name:          "Load with no global config",
			globalConfig:  nil, // No config file will be created
			expectError:   false,
			expectedCount: 0,
		},
		{
			name: "Load with empty global config",
			globalConfig: &config.GlobalConfig{
				URLReplacements: map[string]string{},
			},
			expectError:   false,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing config file
			configPath := filepath.Join(tempDir, "config.json5")
			os.Remove(configPath)

			// Create global config if specified
			if tt.globalConfig != nil {
				err := config.SaveGlobalConfig(tt.globalConfig)
				if err != nil {
					t.Fatalf("Failed to save test global config: %v", err)
				}
			}

			// Load URL replacer
			replacer, err := LoadURLReplacer()

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if replacer == nil {
				t.Fatalf("LoadURLReplacer returned nil replacer")
			}

			if count := replacer.GetReplacementCount(); count != tt.expectedCount {
				t.Errorf("Expected %d replacements, got %d", tt.expectedCount, count)
			}

			// Test that replacements work if they exist
			if tt.expectedCount > 0 {
				testURL := "https://github.com/owner/repo/archive/v1.0.0.tar.gz"
				result := replacer.ApplyReplacements(testURL)
				if result == testURL {
					t.Errorf("Expected URL to be replaced but it wasn't")
				}
			}
		})
	}
}
