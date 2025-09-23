package config

import (
	"encoding/json"
	"fmt"

	"github.com/adhocore/jsonc"
)

// ParseJSON5 parses JSON5 content and unmarshals it into the target
func ParseJSON5(data []byte, target interface{}) error {
	// Use jsonc library for native JSON5 support (supports comments, trailing commas, unquoted keys, etc.)
	j := jsonc.New()
	return j.Unmarshal(data, target)
}

// FormatAsJSON5 formats a configuration struct as JSON5 with proper formatting
func FormatAsJSON5(cfg *Config) (string, error) {
	// For now, just use regular JSON formatting
	// TODO: Add proper JSON5 formatting with comments
	jsonData, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal config to JSON: %w", err)
	}
	return string(jsonData), nil
}
