package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	Version      string            `json:"version" yaml:"version"`
	Distribution string            `json:"distribution,omitempty" yaml:"distribution,omitempty"`
	RequiredFor  []string          `json:"required_for,omitempty" yaml:"required_for,omitempty"`
	Options      map[string]string `json:"options,omitempty" yaml:"options,omitempty"`
	Checksum     *ChecksumConfig   `json:"checksum,omitempty" yaml:"checksum,omitempty"`
}

// ChecksumConfig represents checksum verification configuration
type ChecksumConfig struct {
	Type     string `json:"type,omitempty" yaml:"type,omitempty"`         // sha256, etc.
	Value    string `json:"value,omitempty" yaml:"value,omitempty"`       // direct checksum value
	URL      string `json:"url,omitempty" yaml:"url,omitempty"`           // URL to fetch checksum from
	Filename string `json:"filename,omitempty" yaml:"filename,omitempty"` // filename to look for in checksum file
	Required bool   `json:"required,omitempty" yaml:"required,omitempty"` // whether checksum verification is required
}

// CommandConfig represents a command definition
type CommandConfig struct {
	Description string             `json:"description" yaml:"description"`
	Script      string             `json:"script" yaml:"script"`
	WorkingDir  string             `json:"working_dir,omitempty" yaml:"working_dir,omitempty"`
	Requires    []string           `json:"requires,omitempty" yaml:"requires,omitempty"`
	Args        []CommandArgConfig `json:"args,omitempty" yaml:"args,omitempty"`
	Environment map[string]string  `json:"environment,omitempty" yaml:"environment,omitempty"`

	// Hooks for built-in commands
	Pre  string `json:"pre,omitempty" yaml:"pre,omitempty"`   // Script to run before built-in command
	Post string `json:"post,omitempty" yaml:"post,omitempty"` // Script to run after built-in command

	// Override built-in command behavior
	Override bool `json:"override,omitempty" yaml:"override,omitempty"` // If true, replace built-in command entirely
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
		err = ParseJSON5(data, &config)
	case ".yml", ".yaml":
		err = yaml.Unmarshal(data, &config)
	case ".json":
		// Use JSON5 preprocessor for .json files too (allows comments)
		err = ParseJSON5(data, &config)
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

// SaveConfig saves configuration to the project directory in JSON5 format
func SaveConfig(cfg *Config, projectRoot string) error {
	mvxDir := filepath.Join(projectRoot, ".mvx")

	// Ensure .mvx directory exists
	if err := os.MkdirAll(mvxDir, 0755); err != nil {
		return fmt.Errorf("failed to create .mvx directory: %w", err)
	}

	// Use JSON5 format as the default
	configPath := filepath.Join(mvxDir, "config.json5")

	// Convert config to JSON5 format
	content, err := FormatAsJSON5(cfg)
	if err != nil {
		return fmt.Errorf("failed to format configuration as JSON5: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	return nil
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
		// For override commands or custom commands, script is required
		if cmdConfig.Override || !isBuiltinCommand(cmdName) {
			if cmdConfig.Script == "" {
				return fmt.Errorf("command %s: script is required", cmdName)
			}
		}
		// For built-in commands with hooks, at least one of pre/post/script should be present
		if isBuiltinCommand(cmdName) && !cmdConfig.Override {
			if cmdConfig.Script == "" && cmdConfig.Pre == "" && cmdConfig.Post == "" {
				return fmt.Errorf("command %s: at least one of script, pre, or post is required for built-in command hooks", cmdName)
			}
		}
	}

	return nil
}

// isBuiltinCommand checks if a command name is a built-in mvx command
func isBuiltinCommand(commandName string) bool {
	builtinCommands := map[string]bool{
		"build":            true,
		"test":             true,
		"setup":            true,
		"init":             true,
		"tools":            true,
		"version":          true,
		"help":             true,
		"completion":       true,
		"info":             true,
		"update-bootstrap": true,
	}
	return builtinCommands[commandName]
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

// HasCommandOverride checks if a built-in command is overridden
func (c *Config) HasCommandOverride(commandName string) bool {
	if cmd, exists := c.Commands[commandName]; exists {
		return cmd.Override && isBuiltinCommand(commandName)
	}
	return false
}

// HasCommandHooks checks if a built-in command has pre/post hooks
func (c *Config) HasCommandHooks(commandName string) bool {
	if cmd, exists := c.Commands[commandName]; exists {
		return isBuiltinCommand(commandName) && !cmd.Override && (cmd.Pre != "" || cmd.Post != "")
	}
	return false
}

// GetCommandConfig returns the configuration for a specific command
func (c *Config) GetCommandConfig(commandName string) (CommandConfig, bool) {
	config, exists := c.Commands[commandName]
	return config, exists
}
