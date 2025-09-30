package tools

import (
	"fmt"
	"io"
	"strings"
)

// ToolRegistry provides version discovery for tools
type ToolRegistry struct {
	manager *Manager
}

// NewToolRegistry creates a new tool registry
func NewToolRegistry(manager *Manager) *ToolRegistry {
	return &ToolRegistry{
		manager: manager,
	}
}

// FetchVersionsFromApacheRepo fetches version directories from Apache repository
func (r *ToolRegistry) FetchVersionsFromApacheRepo(repoURL string) ([]string, error) {
	resp, err := r.manager.Get(repoURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch from %s: status %d", repoURL, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse HTML directory listing to extract version directories
	// Apache directory listings contain links like: <a href="3.9.11/">3.9.11/</a>
	content := string(body)
	var versions []string

	// Look for directory links that look like version numbers
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		// Look for href patterns like href="3.9.11/" or href="4.0.0-rc-1/"
		if strings.Contains(line, `href="`) && strings.Contains(line, `/">`) {
			start := strings.Index(line, `href="`) + 6
			end := strings.Index(line[start:], `/"`)
			if end > 0 {
				candidate := line[start : start+end]
				// Validate that this looks like a version number
				if r.isValidVersionString(candidate) {
					versions = append(versions, candidate)
				}
			}
		}
	}

	return versions, nil
}

// isValidVersionString checks if a string looks like a valid version number
func (r *ToolRegistry) isValidVersionString(s string) bool {
	// Must start with a digit
	if len(s) == 0 || !strings.ContainsAny(s[:1], "0123456789") {
		return false
	}

	// Should contain dots or hyphens (for pre-release versions)
	if !strings.ContainsAny(s, ".-") {
		return false
	}

	// Exclude common non-version directories
	excludes := []string{"KEYS", "archetype-catalog.xml", "maven-metadata.xml"}
	for _, exclude := range excludes {
		if s == exclude {
			return false
		}
	}

	return true
}
