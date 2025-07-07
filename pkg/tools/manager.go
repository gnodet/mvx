package tools

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gnodet/mvx/pkg/config"
)

// Manager handles tool installation and management
type Manager struct {
	cacheDir string
	tools    map[string]Tool
}

// Tool represents a tool that can be installed and managed
type Tool interface {
	// Name returns the tool name (e.g., "java", "maven")
	Name() string
	
	// Install downloads and installs the specified version
	Install(version string, config config.ToolConfig) error
	
	// IsInstalled checks if the specified version is installed
	IsInstalled(version string, config config.ToolConfig) bool
	
	// GetPath returns the installation path for the specified version
	GetPath(version string, config config.ToolConfig) (string, error)
	
	// GetBinPath returns the binary path for the specified version
	GetBinPath(version string, config config.ToolConfig) (string, error)
	
	// Verify checks if the installation is working correctly
	Verify(version string, config config.ToolConfig) error
	
	// ListVersions returns available versions for installation
	ListVersions() ([]string, error)
}

// NewManager creates a new tool manager
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}
	
	cacheDir := filepath.Join(homeDir, ".mvx")
	
	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory %s: %w", cacheDir, err)
	}
	
	manager := &Manager{
		cacheDir: cacheDir,
		tools:    make(map[string]Tool),
	}
	
	// Register built-in tools
	manager.RegisterTool(&JavaTool{manager: manager})
	manager.RegisterTool(&MavenTool{manager: manager})
	
	return manager, nil
}

// RegisterTool registers a tool with the manager
func (m *Manager) RegisterTool(tool Tool) {
	m.tools[tool.Name()] = tool
}

// GetTool returns a tool by name
func (m *Manager) GetTool(name string) (Tool, error) {
	tool, exists := m.tools[name]
	if !exists {
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
	return tool, nil
}

// InstallTool installs a specific tool version
func (m *Manager) InstallTool(name string, toolConfig config.ToolConfig) error {
	tool, err := m.GetTool(name)
	if err != nil {
		return err
	}
	
	if tool.IsInstalled(toolConfig.Version, toolConfig) {
		return nil // Already installed
	}
	
	fmt.Printf("Installing %s %s...\n", name, toolConfig.Version)
	if err := tool.Install(toolConfig.Version, toolConfig); err != nil {
		return fmt.Errorf("failed to install %s %s: %w", name, toolConfig.Version, err)
	}
	
	// Verify installation
	if err := tool.Verify(toolConfig.Version, toolConfig); err != nil {
		return fmt.Errorf("installation verification failed for %s %s: %w", name, toolConfig.Version, err)
	}
	
	fmt.Printf("✅ %s %s installed successfully\n", name, toolConfig.Version)
	return nil
}

// InstallTools installs all tools from configuration
func (m *Manager) InstallTools(cfg *config.Config) error {
	for toolName, toolConfig := range cfg.Tools {
		if err := m.InstallTool(toolName, toolConfig); err != nil {
			return err
		}
	}
	return nil
}

// GetCacheDir returns the cache directory path
func (m *Manager) GetCacheDir() string {
	return m.cacheDir
}

// GetToolsDir returns the tools directory path
func (m *Manager) GetToolsDir() string {
	return filepath.Join(m.cacheDir, "tools")
}

// GetToolDir returns the directory for a specific tool
func (m *Manager) GetToolDir(toolName string) string {
	return filepath.Join(m.GetToolsDir(), toolName)
}

// GetToolVersionDir returns the directory for a specific tool version
func (m *Manager) GetToolVersionDir(toolName, version string, distribution string) string {
	versionDir := version
	if distribution != "" {
		versionDir = fmt.Sprintf("%s-%s", version, distribution)
	}
	return filepath.Join(m.GetToolDir(toolName), versionDir)
}

// SetupEnvironment sets up environment variables for installed tools
func (m *Manager) SetupEnvironment(cfg *config.Config) (map[string]string, error) {
	env := make(map[string]string)
	
	// Copy existing environment
	for key, value := range cfg.Environment {
		env[key] = value
	}
	
	// Add tool-specific environment variables
	for toolName, toolConfig := range cfg.Tools {
		tool, err := m.GetTool(toolName)
		if err != nil {
			continue // Skip unknown tools
		}
		
		if !tool.IsInstalled(toolConfig.Version, toolConfig) {
			continue // Skip uninstalled tools
		}
		
		// Get tool path
		toolPath, err := tool.GetPath(toolConfig.Version, toolConfig)
		if err != nil {
			continue
		}
		
		// Set tool-specific environment variables
		switch toolName {
		case "java":
			env["JAVA_HOME"] = toolPath
		case "maven":
			env["MAVEN_HOME"] = toolPath
		case "node":
			env["NODE_HOME"] = toolPath
		}
	}
	
	return env, nil
}
