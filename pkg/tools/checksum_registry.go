package tools

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gnodet/mvx/pkg/config"
)

// ChecksumProvider interface is deprecated - tools should implement GetChecksum in Tool interface
// This is kept for backward compatibility only
type ChecksumProvider interface {
	GetChecksum(version, filename string) (ChecksumInfo, error)
}

// ChecksumRegistry provides checksum information for known tools and versions
type ChecksumRegistry struct {
	// Known checksums for tools - this could be loaded from external sources in the future
	knownChecksums map[string]map[string]ChecksumInfo
	// HTTP client for fetching checksums from APIs
	client *http.Client
	// Deprecated: providers map is no longer used - tools implement GetChecksum directly
	providers map[string]ChecksumProvider
}

// NewChecksumRegistry creates a new checksum registry
func NewChecksumRegistry() *ChecksumRegistry {
	registry := &ChecksumRegistry{
		knownChecksums: make(map[string]map[string]ChecksumInfo),
		client: &http.Client{
			Timeout: getTimeoutFromEnv("MVX_REGISTRY_TIMEOUT", 120*time.Second), // Default: 2 minutes for slow Apache servers
		},
		providers: make(map[string]ChecksumProvider),
	}

	// Initialize with known checksums
	registry.initializeKnownChecksums()

	return registry
}

// RegisterChecksumProvider is deprecated - tools implement GetChecksum directly
func (cr *ChecksumRegistry) RegisterChecksumProvider(toolName string, provider ChecksumProvider) {
	// This method is kept for backward compatibility but is no longer used
	cr.providers[toolName] = provider
}

// GetChecksum returns checksum information for a tool, version, and platform
func (cr *ChecksumRegistry) GetChecksum(toolName, version, platform string) (ChecksumInfo, bool) {
	toolChecksums, exists := cr.knownChecksums[toolName]
	if !exists {
		return ChecksumInfo{}, false
	}

	key := fmt.Sprintf("%s-%s", version, platform)
	checksum, exists := toolChecksums[key]
	return checksum, exists
}

// GetChecksumFromConfig returns checksum information from tool configuration
func (cr *ChecksumRegistry) GetChecksumFromConfig(toolName, version string, cfg config.ToolConfig) (ChecksumInfo, bool) {
	// First check if checksum is provided in configuration
	if cfg.Checksum != nil {
		checksumType := cfg.Checksum.Type
		if checksumType == "" {
			checksumType = "sha256" // default
		}

		return ChecksumInfo{
			Type:     ChecksumType(checksumType),
			Value:    cfg.Checksum.Value,
			URL:      cfg.Checksum.URL,
			Filename: cfg.Checksum.Filename,
		}, true
	}

	// Fall back to known checksums
	platform := cr.getPlatformString(toolName)
	return cr.GetChecksum(toolName, version, platform)
}

// IsChecksumRequired returns whether checksum verification is required for a tool
func (cr *ChecksumRegistry) IsChecksumRequired(cfg config.ToolConfig) bool {
	if cfg.Checksum != nil {
		return cfg.Checksum.Required
	}
	return false // Default to optional
}

// getPlatformString returns the platform string for checksum lookup
// This method should be deprecated in favor of tool-specific platform handling
func (cr *ChecksumRegistry) getPlatformString(toolName string) string {
	platformMapper := NewPlatformMapper()

	// Default to generic platform for all tools
	// Individual tools should handle their own platform-specific logic
	return platformMapper.GetGenericPlatform()
}

// initializeKnownChecksums initializes the registry with known checksums
func (cr *ChecksumRegistry) initializeKnownChecksums() {
	// Maven checksums (these would typically be fetched from Apache's checksum files)
	cr.addMavenChecksums()

	// Go checksums (these would typically be fetched from Go's checksum database)
	cr.addGoChecksums()

	// Add checksum URL patterns for tools that provide them
	cr.addChecksumURLPatterns()

	// Add Java checksums
	cr.addJavaChecksums()

	// Add Node.js checksums
	cr.addNodeChecksums()
}

// addMavenChecksums adds known Maven checksums
func (cr *ChecksumRegistry) addMavenChecksums() {
	cr.knownChecksums["maven"] = map[string]ChecksumInfo{
		// Maven 4.0.0-rc-4
		"4.0.0-rc-4-bin": {
			Type: SHA512,
			URL:  "https://archive.apache.org/dist/maven/maven-4/4.0.0-rc-4/binaries/apache-maven-4.0.0-rc-4-bin.zip.sha512",
		},
		// Maven 3.9.6
		"3.9.6-bin": {
			Type: SHA512,
			URL:  "https://archive.apache.org/dist/maven/maven-3/3.9.6/binaries/apache-maven-3.9.6-bin.zip.sha512",
		},
		// Maven 3.9.5
		"3.9.5-bin": {
			Type: SHA512,
			URL:  "https://archive.apache.org/dist/maven/maven-3/3.9.5/binaries/apache-maven-3.9.5-bin.zip.sha512",
		},
		// Maven 3.9.4
		"3.9.4-bin": {
			Type: SHA512,
			URL:  "https://archive.apache.org/dist/maven/maven-3/3.9.4/binaries/apache-maven-3.9.4-bin.zip.sha512",
		},
	}
}

// addGoChecksums adds known Go checksums
func (cr *ChecksumRegistry) addGoChecksums() {
	cr.knownChecksums["go"] = map[string]ChecksumInfo{
		// Go checksums can be fetched from https://go.dev/dl/
		// For now, we'll use URL patterns
	}
}

// addJavaChecksums adds known Java checksums
func (cr *ChecksumRegistry) addJavaChecksums() {
	cr.knownChecksums["java"] = map[string]ChecksumInfo{
		// Java checksums are provided via Adoptium API
		// We'll use the API to fetch checksums dynamically
	}
}

// addNodeChecksums adds known Node.js checksums
func (cr *ChecksumRegistry) addNodeChecksums() {
	cr.knownChecksums["node"] = map[string]ChecksumInfo{
		// Node.js checksums are available in SHASUMS256.txt files
		// Pattern: https://nodejs.org/dist/v{version}/SHASUMS256.txt
	}
}

// addChecksumURLPatterns adds URL patterns for fetching checksums
func (cr *ChecksumRegistry) addChecksumURLPatterns() {
	// Maven Daemon checksums
	cr.knownChecksums["mvnd"] = map[string]ChecksumInfo{
		// Pattern for mvnd checksums
		"pattern": {
			Type: SHA256,
			URL:  "https://archive.apache.org/dist/maven/mvnd/{version}/maven-mvnd-{version}-{platform}.zip.sha256",
		},
	}
}

// GetChecksumURL generates a checksum URL for a tool based on patterns
// This method is deprecated - tools should use their own GetChecksumURL method
func (cr *ChecksumRegistry) GetChecksumURL(toolName, version, filename string) string {
	// This method is kept for backward compatibility but should not be used
	// Tools should call their own GetChecksumURL method directly
	return ""
}

// SupportsChecksumVerification returns whether a tool supports checksum verification
func (cr *ChecksumRegistry) SupportsChecksumVerification(toolName string) bool {
	supportedTools := []string{"maven", "mvnd", "go", "java", "node"}
	for _, tool := range supportedTools {
		if tool == toolName {
			return true
		}
	}
	return false
}

// GetDynamicChecksum attempts to get checksum using the tool's GetChecksum method
func (cr *ChecksumRegistry) GetDynamicChecksum(tool Tool, version, filename string) (ChecksumInfo, error) {
	return tool.GetChecksum(version, filename)
}

// GetDynamicChecksumByName is deprecated - use GetDynamicChecksum with Tool parameter
func (cr *ChecksumRegistry) GetDynamicChecksumByName(toolName, version, filename string) (ChecksumInfo, error) {
	if provider, exists := cr.providers[toolName]; exists {
		return provider.GetChecksum(version, filename)
	}
	return ChecksumInfo{}, fmt.Errorf("no checksum provider registered for tool: %s", toolName)
}
