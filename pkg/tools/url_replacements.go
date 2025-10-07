package tools

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/gnodet/mvx/pkg/config"
	"github.com/gnodet/mvx/pkg/util"
)

// URLReplacer handles URL replacements for enterprise networks and mirrors
type URLReplacer struct {
	replacements map[string]string
}

// NewURLReplacer creates a new URL replacer with the given replacements
func NewURLReplacer(replacements map[string]string) *URLReplacer {
	return &URLReplacer{
		replacements: replacements,
	}
}

// LoadURLReplacer loads URL replacements from global configuration only
func LoadURLReplacer() (*URLReplacer, error) {
	// Load global configuration
	globalConfig, err := config.LoadGlobalConfig()
	if err != nil {
		// Log warning but continue - global config is optional
		util.LogVerbose("Warning: failed to load global config: %v", err)
		globalConfig = &config.GlobalConfig{}
	}

	// Use only global replacements for simplicity and security
	replacements := make(map[string]string)

	// Add global replacements
	if globalConfig.URLReplacements != nil {
		for pattern, replacement := range globalConfig.URLReplacements {
			replacements[pattern] = replacement
		}
	}

	return NewURLReplacer(replacements), nil
}

// ApplyReplacements applies URL replacements to the given URL
func (r *URLReplacer) ApplyReplacements(originalURL string) string {
	if len(r.replacements) == 0 {
		return originalURL
	}

	currentURL := originalURL

	// Process replacements in deterministic order by sorting keys
	// This ensures consistent behavior across runs
	patterns := make([]string, 0, len(r.replacements))
	for pattern := range r.replacements {
		patterns = append(patterns, pattern)
	}

	// Sort patterns to ensure deterministic order
	// Simple string patterns come before regex patterns for predictable behavior
	sort.Slice(patterns, func(i, j int) bool {
		// Simple patterns (no "regex:" prefix) come first
		iIsRegex := strings.HasPrefix(patterns[i], "regex:")
		jIsRegex := strings.HasPrefix(patterns[j], "regex:")

		if iIsRegex != jIsRegex {
			return !iIsRegex // Simple patterns first
		}

		// Within the same type, sort alphabetically
		return patterns[i] < patterns[j]
	})

	for _, pattern := range patterns {
		replacement := r.replacements[pattern]
		newURL := r.applyReplacement(currentURL, pattern, replacement)
		if newURL != currentURL {
			util.LogVerbose("URL replacement applied: %s -> %s (pattern: %s)", currentURL, newURL, pattern)
			return newURL // Return after first match (like mise)
		}
	}

	return currentURL
}

// applyReplacement applies a single replacement pattern to a URL
func (r *URLReplacer) applyReplacement(url, pattern, replacement string) string {
	// Check if pattern is a regex (starts with "regex:")
	if strings.HasPrefix(pattern, "regex:") {
		regexPattern := strings.TrimPrefix(pattern, "regex:")
		return r.applyRegexReplacement(url, regexPattern, replacement)
	}

	// Simple string replacement
	return strings.ReplaceAll(url, pattern, replacement)
}

// applyRegexReplacement applies a regex replacement to a URL
func (r *URLReplacer) applyRegexReplacement(url, regexPattern, replacement string) string {
	// Compile the regex
	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		util.LogVerbose("Warning: invalid regex pattern '%s': %v", regexPattern, err)
		return url
	}

	// Apply the replacement
	return regex.ReplaceAllString(url, replacement)
}

// GetReplacementCount returns the number of configured replacements
func (r *URLReplacer) GetReplacementCount() int {
	return len(r.replacements)
}

// GetReplacements returns a copy of the replacements map
func (r *URLReplacer) GetReplacements() map[string]string {
	result := make(map[string]string)
	for k, v := range r.replacements {
		result[k] = v
	}
	return result
}

// ValidateReplacements validates all replacement patterns
func (r *URLReplacer) ValidateReplacements() []error {
	var errors []error

	for pattern := range r.replacements {
		if strings.HasPrefix(pattern, "regex:") {
			regexPattern := strings.TrimPrefix(pattern, "regex:")
			if _, err := regexp.Compile(regexPattern); err != nil {
				errors = append(errors, fmt.Errorf("invalid regex pattern '%s': %w", regexPattern, err))
			}
		}
	}

	return errors
}

// Example usage and common patterns
const (
	// Common replacement patterns (for documentation)
	ExampleGitHubMirror    = `"github.com": "nexus.mycompany.net"`
	ExampleHTTPSUpgrade    = `"regex:^http://(.+)": "https://$1"`
	ExampleGitHubReleases  = `"regex:https://github\\.com/([^/]+)/([^/]+)/releases/download/(.+)": "https://hub.corp.com/artifactory/github/$1/$2/$3"`
	ExampleHashiCorpMirror = `"releases.hashicorp.com": "hashicorp-mirror.internal.com"`
	ExampleApacheMirror    = `"archive.apache.org": "apache-mirror.internal.com"`
)

// GetExampleReplacements returns example URL replacement configurations
func GetExampleReplacements() map[string]string {
	return map[string]string{
		"github.com":         "nexus.mycompany.net",
		"regex:^http://(.+)": "https://$1",
		"regex:https://github\\.com/([^/]+)/([^/]+)/releases/download/(.+)": "https://hub.corp.com/artifactory/github/$1/$2/$3",
		"releases.hashicorp.com":                           "hashicorp-mirror.internal.com",
		"archive.apache.org":                               "apache-mirror.internal.com",
		"dist.apache.org":                                  "apache-dist-mirror.internal.com",
		"regex:https://([^.]+)\\.cdn\\.example\\.com/(.+)": "https://unified-cdn.com/$1/$2",
	}
}
