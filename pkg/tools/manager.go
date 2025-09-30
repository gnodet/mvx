package tools

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gnodet/mvx/pkg/config"
	"github.com/gnodet/mvx/pkg/version"
)

// VersionCacheEntry represents a cached version resolution
type VersionCacheEntry struct {
	ResolvedVersion string    `json:"resolved_version"`
	Timestamp       time.Time `json:"timestamp"`
}

// Manager handles tool installation and management
type Manager struct {
	cacheDir       string
	tools          map[string]Tool
	registry       *ToolRegistry
	versionCache   map[string]VersionCacheEntry
	installedCache map[string]bool   // Cache for IsInstalled checks
	pathCache      map[string]string // Cache for GetPath results
	cacheMutex     sync.RWMutex
	httpClient     *http.Client
}

var (
	// Global singleton instance
	globalManager *Manager
	managerMutex  sync.Mutex
)

// InstallOptions contains options for tool installation
type InstallOptions struct {
	MaxConcurrent int  // Maximum number of concurrent downloads (default: 3)
	Parallel      bool // Whether to use parallel downloads (default: true)
	Verbose       bool // Whether to show verbose output (default: false)
}

// Tool represents a tool that can be installed and managed
type Tool interface {
	// Name returns the tool name (e.g., "java", "maven")
	GetToolName() string

	// Install downloads and installs the specified version
	Install(version string, config config.ToolConfig) error

	// IsInstalled checks if the specified version is installed
	IsInstalled(version string, config config.ToolConfig) bool

	// GetPath returns the binary path for the specified version (for PATH management)
	GetPath(version string, config config.ToolConfig) (string, error)

	// Verify checks if the installation is working correctly
	Verify(version string, config config.ToolConfig) error

	// ListVersions returns available versions for installation
	ListVersions() ([]string, error)

	// URL generation methods
	GetDownloadURL(version string) string
	GetChecksumURL(version, filename string) string
	GetVersionsURL() string

	// Checksum generation method
	GetChecksum(version, filename string) (ChecksumInfo, error)

	// SupportsChecksumVerification returns whether this tool supports checksum verification
	SupportsChecksumVerification() bool

	// GetBinaryName returns the binary name for the tool
	GetBinaryName() string

	// GetManager returns the manager instance
	GetManager() *Manager
}

// ToolInfoProvider is an optional interface for tools that can provide detailed information
type ToolInfoProvider interface {
	// GetToolInfo returns detailed information about the tool
	GetToolInfo() (map[string]interface{}, error)
}

// VersionValidator is an optional interface for tools that can validate versions
type VersionValidator interface {
	// ValidateVersion checks if a version specification is valid for this tool
	ValidateVersion(version, distribution string) error
}

// VersionResolver is an optional interface for tools that can resolve version specifications
type VersionResolver interface {
	// ResolveVersion resolves a version specification (e.g., "21", "lts") to a concrete version
	ResolveVersion(version, distribution string) (string, error)
}

// NewManager creates a new tool manager (singleton pattern)
func NewManager() (*Manager, error) {
	managerMutex.Lock()
	defer managerMutex.Unlock()

	// Return existing instance if already created
	if globalManager != nil {
		return globalManager, nil
	}

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
		cacheDir:       cacheDir,
		tools:          make(map[string]Tool),
		versionCache:   make(map[string]VersionCacheEntry),
		installedCache: make(map[string]bool),
		pathCache:      make(map[string]string),
		httpClient: &http.Client{
			Timeout: getTimeoutFromEnv("MVX_HTTP_TIMEOUT", 120*time.Second), // Default: 2 minutes for slow servers
		},
	}

	// Create registry after manager is initialized (to avoid circular dependency)
	manager.registry = NewToolRegistry(manager)

	// Load version cache from disk
	manager.loadVersionCache()

	// Auto-discover and register tools
	if err := manager.discoverAndRegisterTools(); err != nil {
		return nil, fmt.Errorf("failed to register tools: %w", err)
	}

	// Store as global singleton
	globalManager = manager
	return manager, nil
}

// ResetManager resets the global manager instance (for testing purposes)
func ResetManager() {
	managerMutex.Lock()
	defer managerMutex.Unlock()
	globalManager = nil
}

// Get performs an HTTP GET request with verbose logging
// This centralizes all HTTP requests and provides visibility into API calls
func (m *Manager) Get(url string) (*http.Response, error) {
	// Log the request if verbose mode is enabled
	if os.Getenv("MVX_VERBOSE") == "true" {
		fmt.Printf("🌐 HTTP GET: %s\n", url)
	}

	resp, err := m.httpClient.Get(url)
	if err != nil {
		if os.Getenv("MVX_VERBOSE") == "true" {
			fmt.Printf("❌ HTTP GET failed: %s - %v\n", url, err)
		}
		return nil, err
	}

	if os.Getenv("MVX_VERBOSE") == "true" {
		fmt.Printf("✅ HTTP GET %d: %s\n", resp.StatusCode, url)
	}

	return resp, nil
}

// CacheManager interface for tools that support cache management
type CacheManager interface {
	ClearPathCache()
}

// ClearAllCaches clears all tool path caches (for testing purposes)
func (m *Manager) ClearAllCaches() {
	for _, tool := range m.tools {
		if cacheManager, ok := tool.(CacheManager); ok {
			cacheManager.ClearPathCache()
		}
	}
}

// RegisterTool registers a tool with the manager
func (m *Manager) RegisterTool(tool Tool) {
	m.tools[tool.GetToolName()] = tool
}

// GetTool returns a tool by name
func (m *Manager) GetTool(name string) (Tool, error) {
	tool, exists := m.tools[name]
	if !exists {
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
	return tool, nil
}

// GetAllTools returns all registered tools
func (m *Manager) GetAllTools() map[string]Tool {
	result := make(map[string]Tool)
	for name, tool := range m.tools {
		result[name] = tool
	}
	return result
}

// GetToolNames returns the names of all registered tools
func (m *Manager) GetToolNames() []string {
	names := make([]string, 0, len(m.tools))
	for name := range m.tools {
		names = append(names, name)
	}
	return names
}

// ToolFactory represents a function that creates a tool instance
// This enables dynamic tool registration without modifying the manager code
type ToolFactory func(*Manager) Tool

// toolFactories contains all available tool factories for auto-discovery
// This registry allows tools to be registered dynamically, following the Open/Closed Principle
var toolFactories = map[string]ToolFactory{
	"java":  func(m *Manager) Tool { return NewJavaTool(m) },
	"maven": func(m *Manager) Tool { return NewMavenTool(m) },
	"mvnd":  func(m *Manager) Tool { return NewMvndTool(m) },
	"node":  func(m *Manager) Tool { return NewNodeTool(m) },
	"go":    func(m *Manager) Tool { return NewGoTool(m) },
}

// discoverAndRegisterTools automatically discovers and registers all available tools
func (m *Manager) discoverAndRegisterTools() error {
	// Register tools from the factory registry
	for toolName, factory := range toolFactories {
		tool := factory(m)
		m.RegisterTool(tool)
		logVerbose("Registered tool: %s", toolName)
	}

	// Future enhancement: could also load tools from configuration files
	// This would allow users to register custom tools without code changes

	return nil
}

// RegisterToolFactory registers a new tool factory for auto-discovery
// This allows external packages to register their own tools
func RegisterToolFactory(name string, factory ToolFactory) {
	toolFactories[name] = factory
}

// GetRegisteredToolFactories returns the names of all registered tool factories
// This is useful for debugging and introspection
func GetRegisteredToolFactories() []string {
	names := make([]string, 0, len(toolFactories))
	for name := range toolFactories {
		names = append(names, name)
	}
	return names
}

// SearchToolVersions searches for versions of a specific tool with optional filters
func (m *Manager) SearchToolVersions(toolName string, filters []string) ([]string, error) {
	tool, err := m.GetTool(toolName)
	if err != nil {
		return nil, err
	}

	versions, err := tool.ListVersions()
	if err != nil {
		return nil, fmt.Errorf("failed to get versions for %s: %w", toolName, err)
	}

	// Apply filters if provided
	if len(filters) > 0 {
		filtered := make([]string, 0)
		for _, version := range versions {
			for _, filter := range filters {
				if strings.Contains(strings.ToLower(version), strings.ToLower(filter)) {
					filtered = append(filtered, version)
					break
				}
			}
		}
		return filtered, nil
	}

	return versions, nil
}

// GetToolInfo returns detailed information about a tool
func (m *Manager) GetToolInfo(toolName string) (map[string]interface{}, error) {
	tool, err := m.GetTool(toolName)
	if err != nil {
		return nil, err
	}

	// Check if tool implements ToolInfoProvider interface
	if infoProvider, ok := tool.(ToolInfoProvider); ok {
		return infoProvider.GetToolInfo()
	}

	// Fallback to basic information
	versions, err := tool.ListVersions()
	if err != nil {
		return nil, fmt.Errorf("failed to get versions for %s: %w", toolName, err)
	}

	return map[string]interface{}{
		"name":     toolName,
		"versions": versions,
	}, nil
}

// ValidateToolVersion validates that a version exists for the given tool
func (m *Manager) ValidateToolVersion(toolName, version, distribution string) error {
	tool, err := m.GetTool(toolName)
	if err != nil {
		return err
	}

	// Check if tool implements VersionValidator interface
	if validator, ok := tool.(VersionValidator); ok {
		return validator.ValidateVersion(version, distribution)
	}

	// Fallback to checking if version exists in available versions
	versions, err := tool.ListVersions()
	if err != nil {
		return fmt.Errorf("failed to get versions for %s: %w", toolName, err)
	}

	// For version specs like "lts", "latest", etc., we need to resolve them
	resolvedVersion, err := m.resolveVersion(toolName, config.ToolConfig{
		Version:      version,
		Distribution: distribution,
	})
	if err != nil {
		return fmt.Errorf("failed to resolve version %s for %s: %w", version, toolName, err)
	}

	// Check if resolved version exists
	for _, v := range versions {
		if v == resolvedVersion {
			return nil
		}
	}

	return fmt.Errorf("version %s (resolved to %s) not found for tool %s", version, resolvedVersion, toolName)
}

// InstallTool is deprecated - use EnsureTool instead
// This method always installs, even if already installed. Use EnsureTool for smarter behavior.
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

// EnsureTools ensures all tools from configuration are installed (with parallel downloads)
// This replaces InstallTools and uses EnsureTool for automatic installation
func (m *Manager) EnsureTools(cfg *config.Config, maxConcurrent int) error {
	if len(cfg.Tools) == 0 {
		return nil
	}

	if maxConcurrent <= 0 {
		maxConcurrent = GetDefaultConcurrency()
	}

	// If only one tool, use sequential
	if len(cfg.Tools) == 1 {
		for toolName, toolConfig := range cfg.Tools {
			_, err := m.EnsureTool(toolName, toolConfig)
			if err != nil {
				return fmt.Errorf("failed to ensure %s is installed: %w", toolName, err)
			}
			fmt.Printf("✅ %s is ready\n", toolName)
		}
		return nil
	}

	fmt.Printf("📦 Ensuring %d tools are installed (max %d concurrent)...\n", len(cfg.Tools), maxConcurrent)

	// Create a semaphore to limit concurrent operations
	semaphore := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error
	var completed int

	// Ensure tools in parallel
	for toolName, toolConfig := range cfg.Tools {
		wg.Add(1)
		go func(name string, config config.ToolConfig) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if _, err := m.EnsureTool(name, config); err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("%s: %w", name, err))
				mu.Unlock()
			} else {
				mu.Lock()
				completed++
				fmt.Printf("  ✅ Progress: %d/%d tools ready\n", completed, len(cfg.Tools))
				mu.Unlock()
			}
		}(toolName, toolConfig)
	}

	// Wait for all tools to complete
	wg.Wait()

	if len(errors) > 0 {
		return fmt.Errorf("failed to ensure tools: %v", errors)
	}

	fmt.Printf("✅ All %d tools are ready\n", len(cfg.Tools))
	return nil
}

// InstallTools is deprecated - use EnsureTools instead
func (m *Manager) InstallTools(cfg *config.Config) error {
	return m.EnsureTools(cfg, GetDefaultConcurrency())
}

// InstallToolsParallel is deprecated - use EnsureTools instead
func (m *Manager) InstallToolsParallel(cfg *config.Config, maxConcurrent int) error {
	return m.EnsureTools(cfg, maxConcurrent)
}

// InstallToolsWithOptions is deprecated - use EnsureTools instead
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

	// Use sequential (maxConcurrent=1) if parallel is disabled
	maxConcurrent := opts.MaxConcurrent
	if !opts.Parallel {
		maxConcurrent = 1
	}

	return m.EnsureTools(cfg, maxConcurrent)
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
		versionDir = fmt.Sprintf("%s@%s", version, distribution)
	}
	return filepath.Join(m.GetToolDir(toolName), versionDir)
}

// getCacheKey generates a cache key for tool operations
func (m *Manager) getCacheKey(toolName, version, distribution string) string {
	return fmt.Sprintf("%s:%s:%s", toolName, version, distribution)
}

// isToolInstalled checks if a tool is installed (with caching)
func (m *Manager) isToolInstalled(toolName, version string, cfg config.ToolConfig) bool {
	cacheKey := m.getCacheKey(toolName, version, cfg.Distribution)

	m.cacheMutex.RLock()
	if installed, found := m.installedCache[cacheKey]; found {
		m.cacheMutex.RUnlock()
		return installed
	}
	m.cacheMutex.RUnlock()

	tool, err := m.GetTool(toolName)
	if err != nil {
		return false
	}

	installed := tool.IsInstalled(version, cfg)

	m.cacheMutex.Lock()
	m.installedCache[cacheKey] = installed
	m.cacheMutex.Unlock()

	return installed
}

// EnsureTool ensures a tool is installed and returns its binary path.
// This is the main entry point for tool management - it handles:
// - Version resolution (with caching)
// - Installation check (with caching)
// - Auto-installation if needed
// - Path retrieval (with caching)
// All in one atomic, cached operation.
func (m *Manager) EnsureTool(toolName string, cfg config.ToolConfig) (string, error) {
	// Resolve version
	resolvedVersion, err := m.resolveVersion(toolName, cfg)
	if err != nil {
		return "", fmt.Errorf("failed to resolve version for %s: %w", toolName, err)
	}

	resolvedConfig := cfg
	resolvedConfig.Version = resolvedVersion

	cacheKey := m.getCacheKey(toolName, resolvedVersion, cfg.Distribution)

	// Check cache first
	m.cacheMutex.RLock()
	if path, found := m.pathCache[cacheKey]; found {
		m.cacheMutex.RUnlock()
		return path, nil
	}
	m.cacheMutex.RUnlock()

	// Get tool instance
	tool, err := m.GetTool(toolName)
	if err != nil {
		return "", err
	}

	// Check if installed
	if !tool.IsInstalled(resolvedVersion, resolvedConfig) {
		// Auto-install
		logVerbose("Auto-installing %s %s...", toolName, resolvedVersion)
		if err := tool.Install(resolvedVersion, resolvedConfig); err != nil {
			return "", fmt.Errorf("failed to install %s %s: %w", toolName, resolvedVersion, err)
		}

		// Verify installation
		if err := tool.Verify(resolvedVersion, resolvedConfig); err != nil {
			return "", fmt.Errorf("failed to verify %s %s: %w", toolName, resolvedVersion, err)
		}
	}

	// Get path
	path, err := tool.GetPath(resolvedVersion, resolvedConfig)
	if err != nil {
		return "", fmt.Errorf("failed to get path for %s %s: %w", toolName, resolvedVersion, err)
	}

	// Cache the result
	m.cacheMutex.Lock()
	m.pathCache[cacheKey] = path
	m.installedCache[cacheKey] = true
	m.cacheMutex.Unlock()

	return path, nil
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
		// EnsureTool handles version resolution and checks if installed
		// It only returns a path if the tool is installed (doesn't auto-install here)
		resolvedVersion, err := m.resolveVersion(toolName, toolConfig)
		if err != nil {
			continue // Skip tools with resolution errors
		}

		resolvedConfig := toolConfig
		resolvedConfig.Version = resolvedVersion

		// Check if installed (using cache)
		if !m.isToolInstalled(toolName, resolvedVersion, resolvedConfig) {
			continue // Skip uninstalled tools
		}

		// Get tool path (using cache)
		cacheKey := m.getCacheKey(toolName, resolvedVersion, resolvedConfig.Distribution)
		m.cacheMutex.RLock()
		toolPath, found := m.pathCache[cacheKey]
		m.cacheMutex.RUnlock()

		if !found {
			// Not in cache, get it
			tool, err := m.GetTool(toolName)
			if err != nil {
				continue
			}
			toolPath, err = tool.GetPath(resolvedVersion, resolvedConfig)
			if err != nil {
				continue
			}
			m.cacheMutex.Lock()
			m.pathCache[cacheKey] = toolPath
			m.cacheMutex.Unlock()
		}

		// Set tool-specific environment variables
		switch toolName {
		case "java":
			// For Java, we need JAVA_HOME to point to the Java installation directory, not the bin directory
			tool, _ := m.GetTool(toolName)
			if javaTool, ok := tool.(*JavaTool); ok {
				if javaHome, err := javaTool.GetJavaHome(resolvedVersion, resolvedConfig); err == nil {
					env["JAVA_HOME"] = javaHome
				} else {
					logVerbose("Failed to get JAVA_HOME for Java %s: %v", resolvedVersion, err)
				}
			}
		case "maven":
			env["MAVEN_HOME"] = toolPath
		case "node":
			env["NODE_HOME"] = toolPath
		}
	}

	return env, nil
}

// SetupProjectEnvironment sets up project-specific environment for tools
func (m *Manager) SetupProjectEnvironment(cfg *config.Config, projectPath string) (map[string]string, error) {
	env := make(map[string]string)

	// Copy base environment
	baseEnv, err := m.SetupEnvironment(cfg)
	if err != nil {
		return nil, err
	}
	for key, value := range baseEnv {
		env[key] = value
	}

	return env, nil
}

// ResolveVersion resolves a version specification to a concrete version (public method)
func (m *Manager) ResolveVersion(toolName string, toolConfig config.ToolConfig) (string, error) {
	return m.resolveVersion(toolName, toolConfig)
}

// resolveVersion resolves a version specification to a concrete version
func (m *Manager) resolveVersion(toolName string, toolConfig config.ToolConfig) (string, error) {
	// Check for environment variable override first
	if overrideVersion := getToolVersionOverride(toolName); overrideVersion != "" {
		logVerbose("Using version override from %s: %s", getToolVersionOverrideEnvVar(toolName), overrideVersion)
		// Fast path: Check if override version is already concrete
		if m.isConcreteVersion(toolName, overrideVersion) {
			return overrideVersion, nil
		}
		// If override version needs resolution, use it instead of config version
		overrideConfig := toolConfig
		overrideConfig.Version = overrideVersion
		return m.resolveVersionInternal(toolName, overrideConfig)
	}

	// Fast path: Check if version is already concrete (no resolution needed)
	if m.isConcreteVersion(toolName, toolConfig.Version) {
		return toolConfig.Version, nil
	}

	return m.resolveVersionInternal(toolName, toolConfig)
}

// resolveVersionInternal performs the actual version resolution logic
func (m *Manager) resolveVersionInternal(toolName string, toolConfig config.ToolConfig) (string, error) {
	distribution := toolConfig.Distribution

	// Check cache first
	if cached, found := m.getCachedVersion(toolName, toolConfig.Version, distribution); found {
		logVerbose("Using cached version resolution: %s %s (%s) -> %s", toolName, toolConfig.Version, distribution, cached)
		return cached, nil
	}

	logVerbose("Resolving version online: %s %s (%s)", toolName, toolConfig.Version, distribution)

	// Get the tool instance
	tool, err := m.GetTool(toolName)
	if err != nil {
		return "", fmt.Errorf("unknown tool: %s", toolName)
	}

	// Check if tool implements VersionResolver interface
	var resolved string
	if resolver, ok := tool.(VersionResolver); ok {
		resolved, err = resolver.ResolveVersion(toolConfig.Version, distribution)
		if err != nil {
			return "", err
		}
	} else {
		// Fallback: return version as-is for tools that don't implement VersionResolver
		resolved = toolConfig.Version
	}

	logVerbose("Resolved %s %s (%s) -> %s (caching for 24h)", toolName, toolConfig.Version, distribution, resolved)

	// Cache the resolved version
	m.setCachedVersion(toolName, toolConfig.Version, distribution, resolved)

	return resolved, nil
}

// isConcreteVersion checks if a version specification is already concrete and doesn't need resolution
func (m *Manager) isConcreteVersion(toolName, versionSpec string) bool {
	// Handle special cases that always need resolution
	switch versionSpec {
	case "latest", "lts", "":
		return false
	}

	// Try to parse as version specification
	spec, err := version.ParseSpec(versionSpec)
	if err != nil {
		// If we can't parse it, assume it needs resolution
		return false
	}

	// Only "exact" constraint versions are concrete
	// "exact" means full major.minor.patch[-pre][+build] format
	return spec.Constraint == "exact"
}

// GetRegistry returns the tool registry
func (m *Manager) GetRegistry() *ToolRegistry {
	return m.registry
}

// loadVersionCache loads the version resolution cache from disk
func (m *Manager) loadVersionCache() {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	cacheFile := filepath.Join(m.cacheDir, "version_cache.json")
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		// Cache file doesn't exist or can't be read, start with empty cache
		return
	}

	var cache map[string]VersionCacheEntry
	if err := json.Unmarshal(data, &cache); err != nil {
		// Invalid cache file, start with empty cache
		return
	}

	// Filter out expired entries (older than 24 hours)
	now := time.Now()
	for key, entry := range cache {
		if now.Sub(entry.Timestamp) < 24*time.Hour {
			m.versionCache[key] = entry
		}
	}
}

// saveVersionCache saves the version resolution cache to disk
func (m *Manager) saveVersionCache() {
	m.cacheMutex.RLock()
	defer m.cacheMutex.RUnlock()

	cacheFile := filepath.Join(m.cacheDir, "version_cache.json")
	data, err := json.MarshalIndent(m.versionCache, "", "  ")
	if err != nil {
		return // Silently fail on cache save errors
	}

	os.WriteFile(cacheFile, data, 0644)
}

// getCachedVersion retrieves a cached version resolution
func (m *Manager) getCachedVersion(toolName, versionSpec, distribution string) (string, bool) {
	m.cacheMutex.RLock()
	defer m.cacheMutex.RUnlock()

	key := fmt.Sprintf("%s:%s:%s", toolName, versionSpec, distribution)
	entry, exists := m.versionCache[key]
	if !exists {
		return "", false
	}

	// Check if cache entry is still valid (less than 24 hours old)
	if time.Since(entry.Timestamp) > 24*time.Hour {
		return "", false
	}

	return entry.ResolvedVersion, true
}

// setCachedVersion stores a version resolution in cache
func (m *Manager) setCachedVersion(toolName, versionSpec, distribution, resolvedVersion string) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	key := fmt.Sprintf("%s:%s:%s", toolName, versionSpec, distribution)
	m.versionCache[key] = VersionCacheEntry{
		ResolvedVersion: resolvedVersion,
		Timestamp:       time.Now(),
	}

	// Save cache to disk asynchronously
	go m.saveVersionCache()
}

// InstallSpecificTools installs only the specified tools from configuration
func (m *Manager) InstallSpecificTools(cfg *config.Config, toolNames []string) error {
	if len(toolNames) == 0 {
		return nil
	}

	// Create a subset config with only the specified tools
	subsetCfg := &config.Config{
		Tools: make(map[string]config.ToolConfig),
	}

	for _, toolName := range toolNames {
		toolConfig, exists := cfg.Tools[toolName]
		if !exists {
			return fmt.Errorf("tool %s not configured", toolName)
		}
		subsetCfg.Tools[toolName] = toolConfig
	}

	// Get tools that actually need installation
	toolsToInstall, err := m.GetToolsNeedingInstallation(subsetCfg)
	if err != nil {
		return err
	}

	// If no tools need installation
	if len(toolsToInstall) == 0 {
		return nil // All specified tools already installed
	}

	// Install only the tools that need installation
	for toolName, toolConfig := range toolsToInstall {
		if err := m.InstallTool(toolName, toolConfig); err != nil {
			return err
		}
	}

	return nil
}

// EnsureToolInstalled ensures a specific tool is installed without checking others
func (m *Manager) EnsureToolInstalled(cfg *config.Config, toolName string) error {
	toolConfig, exists := cfg.Tools[toolName]
	if !exists {
		return fmt.Errorf("tool %s not configured", toolName)
	}

	tool, err := m.GetTool(toolName)
	if err != nil {
		return err
	}

	// Resolve version specification to concrete version
	resolvedVersion, err := m.resolveVersion(toolName, toolConfig)
	if err != nil {
		return fmt.Errorf("failed to resolve version for %s: %w", toolName, err)
	}

	// Update config with resolved version for checking
	resolvedConfig := toolConfig
	resolvedConfig.Version = resolvedVersion

	if tool.IsInstalled(resolvedVersion, resolvedConfig) {
		return nil // Already installed
	}

	return m.InstallTool(toolName, toolConfig)
}
