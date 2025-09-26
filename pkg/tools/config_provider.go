package tools

import (
	"os"
	"strconv"
	"time"
)

// ConfigProvider interface for providing configuration values
type ConfigProvider interface {
	GetTimeout(key string, defaultValue time.Duration) time.Duration
	GetInt(key string, defaultValue int) int
	GetString(key string, defaultValue string) string
	GetBool(key string, defaultValue bool) bool
}

// EnvironmentConfigProvider provides configuration from environment variables
type EnvironmentConfigProvider struct{}

// NewEnvironmentConfigProvider creates a new environment-based config provider
func NewEnvironmentConfigProvider() *EnvironmentConfigProvider {
	return &EnvironmentConfigProvider{}
}

// GetTimeout returns a timeout value from environment or default
func (p *EnvironmentConfigProvider) GetTimeout(key string, defaultValue time.Duration) time.Duration {
	return getTimeoutFromEnv(key, defaultValue)
}

// GetInt returns an integer value from environment or default
func (p *EnvironmentConfigProvider) GetInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetString returns a string value from environment or default
func (p *EnvironmentConfigProvider) GetString(key string, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetBool returns a boolean value from environment or default
func (p *EnvironmentConfigProvider) GetBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true"
	}
	return defaultValue
}

// DownloadConfigProvider provides download-specific configuration
type DownloadConfigProvider struct {
	configProvider ConfigProvider
}

// NewDownloadConfigProvider creates a new download config provider
func NewDownloadConfigProvider(configProvider ConfigProvider) *DownloadConfigProvider {
	return &DownloadConfigProvider{
		configProvider: configProvider,
	}
}

// GetDownloadTimeout returns the download timeout
func (p *DownloadConfigProvider) GetDownloadTimeout() time.Duration {
	return p.configProvider.GetTimeout(EnvDownloadTimeout, DefaultDownloadTimeout)
}

// GetRegistryTimeout returns the registry timeout
func (p *DownloadConfigProvider) GetRegistryTimeout() time.Duration {
	return p.configProvider.GetTimeout(EnvRegistryTimeout, DefaultRegistryTimeout)
}

// GetChecksumTimeout returns the checksum timeout
func (p *DownloadConfigProvider) GetChecksumTimeout() time.Duration {
	return p.configProvider.GetTimeout(EnvChecksumTimeout, DefaultChecksumTimeout)
}

// GetTLSTimeout returns the TLS timeout
func (p *DownloadConfigProvider) GetTLSTimeout() time.Duration {
	return p.configProvider.GetTimeout(EnvTLSTimeout, DefaultTLSTimeout)
}

// GetResponseTimeout returns the response timeout
func (p *DownloadConfigProvider) GetResponseTimeout() time.Duration {
	return p.configProvider.GetTimeout(EnvResponseTimeout, DefaultResponseTimeout)
}

// GetIdleTimeout returns the idle timeout
func (p *DownloadConfigProvider) GetIdleTimeout() time.Duration {
	return p.configProvider.GetTimeout(EnvIdleTimeout, DefaultIdleTimeout)
}

// GetMaxRetries returns the maximum number of retries
func (p *DownloadConfigProvider) GetMaxRetries() int {
	return p.configProvider.GetInt(EnvMaxRetries, DefaultMaxRetries)
}

// GetRetryDelay returns the retry delay
func (p *DownloadConfigProvider) GetRetryDelay() time.Duration {
	return p.configProvider.GetTimeout(EnvRetryDelay, DefaultRetryDelay)
}

// GetMaxConcurrent returns the maximum concurrent downloads
func (p *DownloadConfigProvider) GetMaxConcurrent() int {
	return p.configProvider.GetInt(EnvParallelDownloads, DefaultMaxConcurrent)
}

// GetMinFileSize returns the minimum file size
func (p *DownloadConfigProvider) GetMinFileSize() int64 {
	return int64(p.configProvider.GetInt("MVX_MIN_FILE_SIZE", DefaultMinFileSize))
}

// GetMaxFileSize returns the maximum file size
func (p *DownloadConfigProvider) GetMaxFileSize() int64 {
	return int64(p.configProvider.GetInt("MVX_MAX_FILE_SIZE", DefaultMaxFileSize))
}

// IsVerbose returns whether verbose logging is enabled
func (p *DownloadConfigProvider) IsVerbose() bool {
	return p.configProvider.GetBool(EnvVerbose, false)
}

// IsColorDisabled returns whether color output is disabled
func (p *DownloadConfigProvider) IsColorDisabled() bool {
	return p.configProvider.GetBool(EnvNoColor, false)
}

// ToolConfigProvider provides tool-specific configuration
type ToolConfigProvider struct {
	configProvider ConfigProvider
	toolName       string
}

// NewToolConfigProvider creates a new tool config provider
func NewToolConfigProvider(configProvider ConfigProvider, toolName string) *ToolConfigProvider {
	return &ToolConfigProvider{
		configProvider: configProvider,
		toolName:       toolName,
	}
}

// ShouldUseSystemTool returns whether to use system tool
func (p *ToolConfigProvider) ShouldUseSystemTool() bool {
	envVar := getSystemToolEnvVar(p.toolName)
	return p.configProvider.GetBool(envVar, false)
}

// GetVersionOverride returns version override if set
func (p *ToolConfigProvider) GetVersionOverride() string {
	envVar := getToolVersionOverrideEnvVar(p.toolName)
	return p.configProvider.GetString(envVar, "")
}

// HasVersionOverride returns whether version override is set
func (p *ToolConfigProvider) HasVersionOverride() bool {
	return p.GetVersionOverride() != ""
}
