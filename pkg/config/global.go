package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// GlobalConfig represents the global mvx configuration
type GlobalConfig struct {
	URLReplacements map[string]string `json:"url_replacements,omitempty" yaml:"url_replacements,omitempty"`
}

// globalConfigDirFunc is a function variable that can be overridden for testing
var globalConfigDirFunc = getGlobalConfigDirImpl

// getGlobalConfigDir returns the global configuration directory
func getGlobalConfigDir() (string, error) {
	return globalConfigDirFunc()
}

// getGlobalConfigDirImpl is the actual implementation
func getGlobalConfigDirImpl() (string, error) {
	var homeDir string
	var err error

	if runtime.GOOS == "windows" {
		homeDir = os.Getenv("USERPROFILE")
		if homeDir == "" {
			homeDir = os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		}
	} else {
		homeDir, err = os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
	}

	if homeDir == "" {
		return "", fmt.Errorf("unable to determine user home directory")
	}

	return filepath.Join(homeDir, ".mvx"), nil
}

// LoadGlobalConfig loads the global configuration
func LoadGlobalConfig() (*GlobalConfig, error) {
	configDir, err := getGlobalConfigDir()
	if err != nil {
		return nil, err
	}

	// Try different config file names in order of preference
	configFiles := []string{
		"config.json5",
		"config.yml",
		"config.yaml",
		"config.json",
	}

	for _, filename := range configFiles {
		configPath := filepath.Join(configDir, filename)
		if _, err := os.Stat(configPath); err == nil {
			return loadGlobalConfigFile(configPath)
		}
	}

	// Return empty config if no file exists (not an error)
	return &GlobalConfig{}, nil
}

// loadGlobalConfigFile loads global configuration from a specific file
func loadGlobalConfigFile(path string) (*GlobalConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read global config file %s: %w", path, err)
	}

	var config GlobalConfig

	// Determine format by file extension
	ext := filepath.Ext(path)
	switch ext {
	case ".json5":
		err = ParseJSON5(data, &config)
	case ".yml", ".yaml":
		// Import yaml package at top of file
		err = fmt.Errorf("YAML support not implemented yet for global config")
	case ".json":
		// Use JSON5 preprocessor for .json files too (allows comments)
		err = ParseJSON5(data, &config)
	default:
		return nil, fmt.Errorf("unsupported global config file format: %s", ext)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse global config file %s: %w", path, err)
	}

	return &config, nil
}

// SaveGlobalConfig saves the global configuration
func SaveGlobalConfig(cfg *GlobalConfig) error {
	configDir, err := getGlobalConfigDir()
	if err != nil {
		return err
	}

	// Ensure global config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create global config directory: %w", err)
	}

	// Use JSON5 format as the default
	configPath := filepath.Join(configDir, "config.json5")

	// Convert config to JSON5 format
	content, err := FormatGlobalAsJSON5(cfg)
	if err != nil {
		return fmt.Errorf("failed to format global configuration as JSON5: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write global configuration file: %w", err)
	}

	return nil
}

// FormatGlobalAsJSON5 formats a global configuration struct as JSON5
func FormatGlobalAsJSON5(cfg *GlobalConfig) (string, error) {
	// For now, use a simple JSON5-like format with comments
	content := "{\n"
	content += "  // Global mvx configuration\n"
	content += "  // See: https://mvx.dev/docs/url-replacements for documentation\n\n"

	if len(cfg.URLReplacements) > 0 {
		content += "  // URL replacements for enterprise networks and mirrors\n"
		content += "  url_replacements: {\n"

		for pattern, replacement := range cfg.URLReplacements {
			// Escape quotes and backslashes in JSON strings
			escapedPattern := escapeJSONString(pattern)
			escapedReplacement := escapeJSONString(replacement)
			content += fmt.Sprintf("    \"%s\": \"%s\",\n", escapedPattern, escapedReplacement)
		}

		// Remove trailing comma
		if len(cfg.URLReplacements) > 0 {
			content = content[:len(content)-2] + "\n"
		}

		content += "  }\n"
	}

	content += "}\n"
	return content, nil
}

// escapeJSONString escapes a string for use in JSON
func escapeJSONString(s string) string {
	// Replace backslashes first, then quotes
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}

// GetGlobalConfigPath returns the path to the global configuration file
func GetGlobalConfigPath() (string, error) {
	configDir, err := getGlobalConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.json5"), nil
}

// GetGlobalConfigDirFunc returns the current global config directory function (for testing)
func GetGlobalConfigDirFunc() func() (string, error) {
	return globalConfigDirFunc
}

// SetGlobalConfigDirFunc sets the global config directory function (for testing)
func SetGlobalConfigDirFunc(fn func() (string, error)) {
	globalConfigDirFunc = fn
}
