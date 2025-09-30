package tools

import (
	"github.com/gnodet/mvx/pkg/config"
)

// ChecksumRegistry provides checksum configuration parsing for tools.
// This is a simplified version that only handles configuration parsing.
// Tools implement their own GetChecksum() methods for dynamic checksum fetching.
type ChecksumRegistry struct {
	// Empty struct - all logic is stateless
}

// NewChecksumRegistry creates a new checksum registry
func NewChecksumRegistry() *ChecksumRegistry {
	return &ChecksumRegistry{}
}

// GetChecksumFromConfig returns checksum information from tool configuration
func (cr *ChecksumRegistry) GetChecksumFromConfig(toolName, version string, cfg config.ToolConfig) (ChecksumInfo, bool) {
	// Check if checksum is provided in configuration
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

	// No checksum in config
	return ChecksumInfo{}, false
}

// IsChecksumRequired returns whether checksum verification is required for a tool
func (cr *ChecksumRegistry) IsChecksumRequired(cfg config.ToolConfig) bool {
	if cfg.Checksum != nil {
		return cfg.Checksum.Required
	}
	return false // Default to optional
}

// SupportsChecksumVerification returns whether a tool supports checksum verification
func (cr *ChecksumRegistry) SupportsChecksumVerification(toolName string) bool {
	// All tools support checksum verification via their GetChecksum() method
	supportedTools := []string{"maven", "mvnd", "go", "java", "node"}
	for _, tool := range supportedTools {
		if tool == toolName {
			return true
		}
	}
	return false
}
