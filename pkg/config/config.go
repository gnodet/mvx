package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yosuke-furukawa/json5/encoding/json5"
	"gopkg.in/yaml.v3"
)

// Config represents the mvx project configuration
type Config struct {
	Project     ProjectConfig            `json:"project" yaml:"project"`
	Tools       map[string]ToolConfig    `json:"tools" yaml:"tools"`
	Environment map[string]string        `json:"environment" yaml:"environment"`
	Commands    map[string]CommandConfig `json:"commands" yaml:"commands"`
}

// ProjectConfig contains project metadata
type ProjectConfig struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
}

// ToolConfig represents a tool requirement
type ToolConfig struct {
	Version      string   `json:"version" yaml:"version"`
	Distribution string   `json:"distribution,omitempty" yaml:"distribution,omitempty"`
	RequiredFor  []string `json:"required_for,omitempty" yaml:"required_for,omitempty"`
}

// CommandConfig represents a command definition
type CommandConfig struct {
	Description string                 `json:"description" yaml:"description"`
	Script      string                 `json:"script" yaml:"script"`
	WorkingDir  string                 `json:"working_dir,omitempty" yaml:"working_dir,omitempty"`
	Requires    []string               `json:"requires,omitempty" yaml:"requires,omitempty"`
	Args        []CommandArgConfig     `json:"args,omitempty" yaml:"args,omitempty"`
	Environment map[string]string      `json:"environment,omitempty" yaml:"environment,omitempty"`
}

// CommandArgConfig represents a command argument
type CommandArgConfig struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
	Default     string `json:"default,omitempty" yaml:"default,omitempty"`
	Required    bool   `json:"required,omitempty" yaml:"required,omitempty"`
}

// LoadConfig loads configuration from the project directory
func LoadConfig(projectRoot string) (*Config, error) {
	mvxDir := filepath.Join(projectRoot, ".mvx")
	
	// Try different config file names in order of preference
	configFiles := []string{
		"config.json5",
		"config.yml", 
		"config.yaml",
		"config.json",
	}
	
	for _, filename := range configFiles {
		configPath := filepath.Join(mvxDir, filename)
		if _, err := os.Stat(configPath); err == nil {
			return loadConfigFile(configPath)
		}
	}
	
	return nil, fmt.Errorf("no configuration file found in %s (tried: %s)", 
		mvxDir, strings.Join(configFiles, ", "))
}

// loadConfigFile loads configuration from a specific file
func loadConfigFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}
	
	var config Config
	
	// Determine format by file extension
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json5":
		err = json5.Unmarshal(data, &config)
	case ".yml", ".yaml":
		err = yaml.Unmarshal(data, &config)
	case ".json":
		// Fall back to regular JSON for .json files
		err = json5.Unmarshal(data, &config)
	default:
		return nil, fmt.Errorf("unsupported config file format: %s", ext)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}
	
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	return &config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Project.Name == "" {
		return fmt.Errorf("project.name is required")
	}
	
	// Validate tool configurations
	for toolName, toolConfig := range c.Tools {
		if toolConfig.Version == "" {
			return fmt.Errorf("tool %s: version is required", toolName)
		}
	}
	
	// Validate command configurations
	for cmdName, cmdConfig := range c.Commands {
		if cmdConfig.Script == "" {
			return fmt.Errorf("command %s: script is required", cmdName)
		}
	}
	
	return nil
}

// GetRequiredTools returns a list of tools required for a specific command
func (c *Config) GetRequiredTools(commandName string) []string {
	if cmd, exists := c.Commands[commandName]; exists {
		return cmd.Requires
	}
	
	// If no specific requirements, return all tools
	var allTools []string
	for toolName := range c.Tools {
		allTools = append(allTools, toolName)
	}
	return allTools
}

// GetToolConfig returns the configuration for a specific tool
func (c *Config) GetToolConfig(toolName string) (ToolConfig, bool) {
	config, exists := c.Tools[toolName]
	return config, exists
}
