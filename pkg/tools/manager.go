package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/gnodet/mvx/pkg/config"
)

// Manager handles tool installation and management
type Manager struct {
	cacheDir         string
	tools            map[string]Tool
	registry         *ToolRegistry
	checksumRegistry *ChecksumRegistry
}

// InstallOptions contains options for tool installation
type InstallOptions struct {
	MaxConcurrent int  // Maximum number of concurrent downloads (default: 3)
	Parallel      bool // Whether to use parallel downloads (default: true)
	Verbose       bool // Whether to show verbose output (default: false)
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
		cacheDir:         cacheDir,
		tools:            make(map[string]Tool),
		registry:         NewToolRegistry(),
		checksumRegistry: NewChecksumRegistry(),
	}

	// Register built-in tools
	manager.RegisterTool(&JavaTool{manager: manager})
	manager.RegisterTool(&MavenTool{manager: manager})
	manager.RegisterTool(&MvndTool{manager: manager})
	manager.RegisterTool(&NodeTool{manager: manager})
	manager.RegisterTool(&GoTool{manager: manager})

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

// InstallTool installs a specific tool version with version resolution
func (m *Manager) InstallTool(name string, toolConfig config.ToolConfig) error {
	tool, err := m.GetTool(name)
	if err != nil {
		return err
	}

	// Resolve version specification to concrete version
	resolvedVersion, err := m.resolveVersion(name, toolConfig)
	if err != nil {
		return fmt.Errorf("failed to resolve version for %s: %w", name, err)
	}

	fmt.Printf("  🔍 Resolved %s version %s to %s\n", name, toolConfig.Version, resolvedVersion)

	// Update config with resolved version for installation
	resolvedConfig := toolConfig
	resolvedConfig.Version = resolvedVersion

	if tool.IsInstalled(resolvedVersion, resolvedConfig) {
		fmt.Printf("✅ %s %s already installed\n", name, resolvedVersion)
		return nil
	}

	fmt.Printf("Installing %s %s", name, resolvedVersion)
	if toolConfig.Distribution != "" {
		fmt.Printf(" (%s)", toolConfig.Distribution)
	}
	fmt.Println("...")

	if err := tool.Install(resolvedVersion, resolvedConfig); err != nil {
		return fmt.Errorf("failed to install %s %s: %w", name, resolvedVersion, err)
	}

	// Verify installation
	if err := tool.Verify(resolvedVersion, resolvedConfig); err != nil {
		return fmt.Errorf("installation verification failed for %s %s: %w", name, resolvedVersion, err)
	}

	fmt.Printf("✅ %s %s installed successfully\n", name, resolvedVersion)
	return nil
}

// GetDefaultConcurrency returns the default concurrency level from environment or default
func GetDefaultConcurrency() int {
	if concStr := os.Getenv("MVX_PARALLEL_DOWNLOADS"); concStr != "" {
		if conc, err := strconv.Atoi(concStr); err == nil && conc > 0 {
			return conc
		}
	}
	return 3 // Default to 3 concurrent downloads
}

// InstallTools installs all tools from configuration
func (m *Manager) InstallTools(cfg *config.Config) error {
	return m.InstallToolsParallel(cfg, GetDefaultConcurrency())
}

// InstallToolsParallel installs all tools from configuration with parallel downloads
func (m *Manager) InstallToolsParallel(cfg *config.Config, maxConcurrent int) error {
	if len(cfg.Tools) == 0 {
		return nil
	}

	// Get tools that actually need installation
	toolsToInstall, err := m.GetToolsNeedingInstallation(cfg)
	if err != nil {
		return err
	}

	// If no tools need installation
	if len(toolsToInstall) == 0 {
		fmt.Printf("✅ All %d tools already installed\n", len(cfg.Tools))
		return nil
	}

	// Show what's already installed vs what needs installation
	alreadyInstalled := len(cfg.Tools) - len(toolsToInstall)
	if alreadyInstalled > 0 {
		fmt.Printf("✅ %d tool(s) already installed\n", alreadyInstalled)
	}

	// If only one tool needs installation, use sequential installation
	if len(toolsToInstall) == 1 {
		for toolName, toolConfig := range toolsToInstall {
			return m.InstallTool(toolName, toolConfig)
		}
	}

	fmt.Printf("📦 Installing %d tools in parallel (max %d concurrent)...\n", len(toolsToInstall), maxConcurrent)

	// Create a semaphore to limit concurrent downloads
	semaphore := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error
	var completed int

	// Install tools in parallel
	for toolName, toolConfig := range toolsToInstall {
		wg.Add(1)
		go func(name string, config config.ToolConfig) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if err := m.InstallTool(name, config); err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("%s: %w", name, err))
				mu.Unlock()
			} else {
				mu.Lock()
				completed++
				fmt.Printf("  ✅ Progress: %d/%d tools completed\n", completed, len(toolsToInstall))
				mu.Unlock()
			}
		}(toolName, toolConfig)
	}

	// Wait for all installations to complete
	wg.Wait()

	// Check for errors
	if len(errors) > 0 {
		fmt.Printf("❌ %d tool installation(s) failed:\n", len(errors))
		for _, err := range errors {
			fmt.Printf("  • %v\n", err)
		}
		return fmt.Errorf("failed to install %d tool(s)", len(errors))
	}

	fmt.Printf("✅ All %d tools installed successfully\n", len(toolsToInstall))
	return nil
}

// InstallToolsWithOptions installs all tools from configuration with custom options
func (m *Manager) InstallToolsWithOptions(cfg *config.Config, opts *InstallOptions) error {
	if opts == nil {
		opts = &InstallOptions{
			MaxConcurrent: 3,
			Parallel:      true,
			Verbose:       false,
		}
	}

	// Set defaults
	if opts.MaxConcurrent <= 0 {
		opts.MaxConcurrent = 3
	}

	if len(cfg.Tools) == 0 {
		return nil
	}

	// Use sequential installation if parallel is disabled or only one tool
	if !opts.Parallel || len(cfg.Tools) == 1 {
		if opts.Verbose {
			fmt.Printf("📦 Installing %d tool(s) sequentially...\n", len(cfg.Tools))
		}
		for toolName, toolConfig := range cfg.Tools {
			if err := m.InstallTool(toolName, toolConfig); err != nil {
				return err
			}
		}
		return nil
	}

	return m.InstallToolsParallel(cfg, opts.MaxConcurrent)
}

// GetToolsNeedingInstallation returns a map of tools that need to be installed
func (m *Manager) GetToolsNeedingInstallation(cfg *config.Config) (map[string]config.ToolConfig, error) {
	needInstallation := make(map[string]config.ToolConfig)

	for toolName, toolConfig := range cfg.Tools {
		tool, err := m.GetTool(toolName)
		if err != nil {
			return nil, fmt.Errorf("unknown tool %s: %w", toolName, err)
		}

		// Resolve version specification to concrete version
		resolvedVersion, err := m.resolveVersion(toolName, toolConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve version for %s: %w", toolName, err)
		}

		// Update config with resolved version for checking
		resolvedConfig := toolConfig
		resolvedConfig.Version = resolvedVersion

		if !tool.IsInstalled(resolvedVersion, resolvedConfig) {
			needInstallation[toolName] = resolvedConfig
		}
	}

	return needInstallation, nil
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

// resolveVersion resolves a version specification to a concrete version
func (m *Manager) resolveVersion(toolName string, toolConfig config.ToolConfig) (string, error) {
	switch toolName {
	case "java":
		distribution := toolConfig.Distribution
		if distribution == "" {
			distribution = "temurin" // Default distribution
		}
		return m.registry.ResolveJavaVersion(toolConfig.Version, distribution)
	case "maven":
		return m.registry.ResolveMavenVersion(toolConfig.Version)
	case "mvnd":
		return m.registry.ResolveMvndVersion(toolConfig.Version)
	case "node":
		return m.registry.ResolveNodeVersion(toolConfig.Version)
	case "go":
		return m.registry.ResolveGoVersion(toolConfig.Version)
	default:
		// For unknown tools, return version as-is
		return toolConfig.Version, nil
	}
}

// GetRegistry returns the tool registry
func (m *Manager) GetRegistry() *ToolRegistry {
	return m.registry
}

// GetChecksumRegistry returns the checksum registry
func (m *Manager) GetChecksumRegistry() *ChecksumRegistry {
	return m.checksumRegistry
}
