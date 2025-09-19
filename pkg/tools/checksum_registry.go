package tools

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/gnodet/mvx/pkg/config"
)

// ChecksumRegistry provides checksum information for known tools and versions
type ChecksumRegistry struct {
	// Known checksums for tools - this could be loaded from external sources in the future
	knownChecksums map[string]map[string]ChecksumInfo
	// HTTP client for fetching checksums from APIs
	client *http.Client
}

// NewChecksumRegistry creates a new checksum registry
func NewChecksumRegistry() *ChecksumRegistry {
	registry := &ChecksumRegistry{
		knownChecksums: make(map[string]map[string]ChecksumInfo),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Initialize with known checksums
	registry.initializeKnownChecksums()

	return registry
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
func (cr *ChecksumRegistry) getPlatformString(toolName string) string {
	osName := runtime.GOOS
	arch := runtime.GOARCH

	switch toolName {
	case "maven", "mvnd":
		// Maven and mvnd use platform-independent archives
		return "bin"
	case "java":
		// Java uses specific platform naming
		return fmt.Sprintf("%s-%s", osName, arch)
	case "go":
		// Go uses its own platform naming
		return fmt.Sprintf("%s-%s", osName, arch)
	case "node":
		// Node.js platform naming
		if osName == "windows" {
			return fmt.Sprintf("win-%s", arch)
		}
		return fmt.Sprintf("%s-%s", osName, arch)
	default:
		return fmt.Sprintf("%s-%s", osName, arch)
	}
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
func (cr *ChecksumRegistry) GetChecksumURL(toolName, version, filename string) string {
	switch toolName {
	case "maven":
		if strings.HasPrefix(version, "4.") {
			return fmt.Sprintf("https://archive.apache.org/dist/maven/maven-4/%s/binaries/%s.sha512", version, filename)
		}
		return fmt.Sprintf("https://archive.apache.org/dist/maven/maven-3/%s/binaries/%s.sha512", version, filename)
	case "mvnd":
		return fmt.Sprintf("https://archive.apache.org/dist/maven/mvnd/%s/%s.sha512", version, filename)
	case "go":
		// Go provides checksums at https://go.dev/dl/
		return fmt.Sprintf("https://go.dev/dl/?mode=json&include=all")
	case "java":
		// Java checksums are provided via Adoptium API
		// We'll use the API to get the checksum URL dynamically
		return ""
	case "node":
		// Node.js provides SHASUMS256.txt files for each version
		return fmt.Sprintf("https://nodejs.org/dist/v%s/SHASUMS256.txt", version)
	default:
		return ""
	}
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

// AdoptiumAsset represents a binary asset from the Adoptium API
type AdoptiumAsset struct {
	Binary struct {
		Architecture string `json:"architecture"`
		ImageType    string `json:"image_type"`
		OS           string `json:"os"`
		Package      struct {
			Checksum     string `json:"checksum"`
			ChecksumLink string `json:"checksum_link"`
			Name         string `json:"name"`
			Link         string `json:"link"`
		} `json:"package"`
	} `json:"binary"`
}

// GetJavaChecksumFromAPI fetches Java checksum from Adoptium API
func (cr *ChecksumRegistry) GetJavaChecksumFromAPI(version, arch, osName string) (ChecksumInfo, error) {
	// Convert mvx architecture names to Adoptium API names
	adoptiumArch := arch
	if arch == "amd64" {
		adoptiumArch = "x64"
	}

	// Convert mvx OS names to Adoptium API names
	adoptiumOS := osName
	if osName == "darwin" {
		adoptiumOS = "mac"
	}

	// Construct API URL
	url := fmt.Sprintf("https://api.adoptium.net/v3/assets/latest/%s/hotspot?architecture=%s&os=%s&image_type=jdk",
		version, adoptiumArch, adoptiumOS)

	resp, err := cr.client.Get(url)
	if err != nil {
		return ChecksumInfo{}, fmt.Errorf("failed to fetch from Adoptium API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ChecksumInfo{}, fmt.Errorf("Adoptium API returned status %d", resp.StatusCode)
	}

	var assets []AdoptiumAsset
	if err := json.NewDecoder(resp.Body).Decode(&assets); err != nil {
		return ChecksumInfo{}, fmt.Errorf("failed to decode Adoptium API response: %w", err)
	}

	// Find the JDK package
	for _, asset := range assets {
		if asset.Binary.ImageType == "jdk" &&
			asset.Binary.Architecture == adoptiumArch &&
			asset.Binary.OS == adoptiumOS {
			return ChecksumInfo{
				Type:  SHA256,
				Value: asset.Binary.Package.Checksum,
				URL:   asset.Binary.Package.ChecksumLink,
			}, nil
		}
	}

	return ChecksumInfo{}, fmt.Errorf("no matching JDK found for %s %s %s", version, arch, osName)
}

// GetNodeChecksumFromSHASUMS fetches Node.js checksum from SHASUMS256.txt
func (cr *ChecksumRegistry) GetNodeChecksumFromSHASUMS(version, filename string) (ChecksumInfo, error) {
	url := fmt.Sprintf("https://nodejs.org/dist/v%s/SHASUMS256.txt", version)

	resp, err := cr.client.Get(url)
	if err != nil {
		return ChecksumInfo{}, fmt.Errorf("failed to fetch Node.js checksums: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ChecksumInfo{}, fmt.Errorf("Node.js checksums returned status %d", resp.StatusCode)
	}

	return ChecksumInfo{
		Type:     SHA256,
		URL:      url,
		Filename: filename,
	}, nil
}
