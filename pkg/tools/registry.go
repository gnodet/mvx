package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gnodet/mvx/pkg/version"
)

// ToolRegistry provides version discovery for tools
type ToolRegistry struct {
	httpClient *http.Client
}

// NewToolRegistry creates a new tool registry
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// JavaDistribution represents a Java distribution
type JavaDistribution struct {
	Name        string
	DisplayName string
	APIBase     string
}

// DiscoDistribution represents a Java distribution from Disco API
type DiscoDistribution struct {
	APIParameter string `json:"api_parameter"`
	Name         string `json:"name"`
	Maintained   bool   `json:"maintained"`
	Available    bool   `json:"available"`
}

// GetJavaDistributions returns available Java distributions from Disco API
func (r *ToolRegistry) GetJavaDistributions() []JavaDistribution {
	// Try to get distributions from Disco API
	if distributions, err := r.getDiscoDistributions(); err == nil {
		return distributions
	}

	// Fallback to known distributions
	return []JavaDistribution{
		{
			Name:        "temurin",
			DisplayName: "Eclipse Temurin (OpenJDK)",
			APIBase:     "https://api.foojay.io/disco/v3.0",
		},
		{
			Name:        "graalvm_ce",
			DisplayName: "GraalVM Community Edition",
			APIBase:     "https://api.foojay.io/disco/v3.0",
		},
		{
			Name:        "oracle",
			DisplayName: "Oracle JDK",
			APIBase:     "https://api.foojay.io/disco/v3.0",
		},
		{
			Name:        "corretto",
			DisplayName: "Amazon Corretto",
			APIBase:     "https://api.foojay.io/disco/v3.0",
		},
		{
			Name:        "liberica",
			DisplayName: "BellSoft Liberica",
			APIBase:     "https://api.foojay.io/disco/v3.0",
		},
		{
			Name:        "zulu",
			DisplayName: "Azul Zulu",
			APIBase:     "https://api.foojay.io/disco/v3.0",
		},
		{
			Name:        "microsoft",
			DisplayName: "Microsoft Build of OpenJDK",
			APIBase:     "https://api.foojay.io/disco/v3.0",
		},
	}
}

// GetJavaVersions returns available Java versions for a distribution using Disco API
func (r *ToolRegistry) GetJavaVersions(distribution string) ([]string, error) {
	return r.getDiscoVersions(distribution)
}

// getDiscoDistributions fetches available distributions from Disco API
func (r *ToolRegistry) getDiscoDistributions() ([]JavaDistribution, error) {
	resp, err := r.httpClient.Get("https://api.foojay.io/disco/v3.0/distributions")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var discoDistributions []DiscoDistribution
	if err := json.NewDecoder(resp.Body).Decode(&discoDistributions); err != nil {
		return nil, err
	}

	var distributions []JavaDistribution
	for _, dist := range discoDistributions {
		if dist.Available && dist.Maintained {
			distributions = append(distributions, JavaDistribution{
				Name:        dist.APIParameter,
				DisplayName: dist.Name,
				APIBase:     "https://api.foojay.io/disco/v3.0",
			})
		}
	}

	return distributions, nil
}

// getDiscoVersions fetches available versions for a distribution from Disco API
func (r *ToolRegistry) getDiscoVersions(distribution string) ([]string, error) {
	if distribution == "" {
		distribution = "temurin" // Default to Temurin
	}

	// Get major versions available for this distribution
	url := fmt.Sprintf("https://api.foojay.io/disco/v3.0/major_versions?distribution=%s&maintained=true", distribution)
	resp, err := r.httpClient.Get(url)
	if err != nil {
		// Fallback to known versions if API is unavailable
		return []string{"8", "11", "17", "21", "22", "23", "24", "25"}, nil
	}
	defer resp.Body.Close()

	var majorVersions []struct {
		MajorVersion int  `json:"major_version"`
		Maintained   bool `json:"maintained"`
		EarlyAccess  bool `json:"early_access"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&majorVersions); err != nil {
		return []string{"8", "11", "17", "21", "22", "23", "24", "25"}, nil
	}

	var versions []string
	for _, mv := range majorVersions {
		if mv.EarlyAccess {
			versions = append(versions, fmt.Sprintf("%d-ea", mv.MajorVersion))
		} else {
			versions = append(versions, fmt.Sprintf("%d", mv.MajorVersion))
		}
	}

	// Sort in descending order (newest first)
	sort.Slice(versions, func(i, j int) bool {
		return versions[i] > versions[j]
	})

	return versions, nil
}

// GetMavenVersions returns available Maven versions
func (r *ToolRegistry) GetMavenVersions() ([]string, error) {
	// Try to fetch versions from Apache distribution repos
	versions, err := r.fetchMavenVersionsFromApache()
	if err != nil {
		// Fallback to known versions if API is unavailable
		return r.getFallbackMavenVersions(), nil
	}
	return version.SortVersions(versions), nil
}

// fetchMavenVersionsFromApache fetches Maven versions from Apache archive repositories
func (r *ToolRegistry) fetchMavenVersionsFromApache() ([]string, error) {
	var allVersions []string

	// Fetch Maven 3.x versions from archive
	maven3Versions, err := r.fetchVersionsFromApacheRepo("https://archive.apache.org/dist/maven/maven-3/")
	if err == nil {
		allVersions = append(allVersions, maven3Versions...)
	}

	// Fetch Maven 4.x versions from archive
	maven4Versions, err := r.fetchVersionsFromApacheRepo("https://archive.apache.org/dist/maven/maven-4/")
	if err == nil {
		allVersions = append(allVersions, maven4Versions...)
	}

	if len(allVersions) == 0 {
		return nil, fmt.Errorf("no versions found from Apache repositories")
	}

	return allVersions, nil
}

// getFallbackMavenVersions returns known Maven versions as fallback
func (r *ToolRegistry) getFallbackMavenVersions() []string {
	return []string{
		// Maven 4.x (pre-release versions)
		"4.0.0", "4.0.0-rc-4", "4.0.0-rc-3", "4.0.0-rc-2", "4.0.0-rc-1",
		"4.0.0-beta-5", "4.0.0-beta-4", "4.0.0-beta-3", "4.0.0-beta-2", "4.0.0-beta-1",
		"4.0.0-alpha-13", "4.0.0-alpha-12", "4.0.0-alpha-11", "4.0.0-alpha-10",

		// Maven 3.9.x (latest stable)
		"3.9.11", "3.9.10", "3.9.9", "3.9.8", "3.9.7", "3.9.6", "3.9.5", "3.9.4", "3.9.3", "3.9.2", "3.9.1", "3.9.0",

		// Maven 3.8.x (previous stable)
		"3.8.8", "3.8.7", "3.8.6", "3.8.5", "3.8.4", "3.8.3", "3.8.2", "3.8.1",

		// Maven 3.6.x (older stable)
		"3.6.3", "3.6.2", "3.6.1", "3.6.0",
	}
}

// fetchVersionsFromApacheRepo fetches version directories from Apache repository
func (r *ToolRegistry) fetchVersionsFromApacheRepo(repoURL string) ([]string, error) {
	resp, err := r.httpClient.Get(repoURL)
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

// GetMvndVersions returns available Maven Daemon versions
func (r *ToolRegistry) GetMvndVersions() ([]string, error) {
	// Try to fetch versions from Apache archive
	versions, err := r.fetchMvndVersionsFromApache()
	if err != nil {
		// Fallback to known versions if API is unavailable
		return r.getFallbackMvndVersions(), nil
	}
	return version.SortVersions(versions), nil
}

// fetchMvndVersionsFromApache fetches mvnd versions from Apache archive
func (r *ToolRegistry) fetchMvndVersionsFromApache() ([]string, error) {
	// Fetch mvnd versions from archive
	mvndVersions, err := r.fetchVersionsFromApacheRepo("https://archive.apache.org/dist/maven/mvnd/")
	if err != nil {
		return nil, fmt.Errorf("no mvnd versions found from Apache archive: %w", err)
	}

	return mvndVersions, nil
}

// getFallbackMvndVersions returns known mvnd versions as fallback
func (r *ToolRegistry) getFallbackMvndVersions() []string {
	return []string{
		// Maven Daemon 2.x
		"2.0.0", "2.0.0-beta-1", "2.0.0-alpha-1",

		// Maven Daemon 1.x
		"1.0.2", "1.0.1", "1.0.0", "1.0.0-m8", "1.0.0-m7", "1.0.0-m6", "1.0.0-m5",
		"1.0.0-m4", "1.0.0-m3", "1.0.0-m2", "1.0.0-m1",

		// Maven Daemon 0.x
		"0.9.0", "0.8.2", "0.8.1", "0.8.0", "0.7.1", "0.7.0", "0.6.0", "0.5.2", "0.5.1", "0.5.0",
	}
}

// ResolveMvndVersion resolves a mvnd version specification to a concrete version
func (r *ToolRegistry) ResolveMvndVersion(versionSpec string) (string, error) {
	availableVersions, err := r.GetMvndVersions()
	if err != nil {
		return "", err
	}

	spec, err := version.ParseSpec(versionSpec)
	if err != nil {
		return "", fmt.Errorf("invalid version specification %s: %w", versionSpec, err)
	}

	resolved, err := spec.Resolve(availableVersions)
	if err != nil {
		return "", fmt.Errorf("failed to resolve mvnd version %s: %w", versionSpec, err)
	}

	return resolved, nil
}

// GetNodeVersions returns available Node.js versions
func (r *ToolRegistry) GetNodeVersions() ([]string, error) {
	versions, err := r.fetchNodeVersions()
	if err != nil {
		// minimal fallback
		return []string{"22.5.1", "22.4.1", "20.15.0", "18.20.4"}, nil
	}
	return version.SortVersions(versions), nil
}

// ResolveNodeVersion resolves a Node version specification
func (r *ToolRegistry) ResolveNodeVersion(versionSpec string) (string, error) {
	if versionSpec == "lts" {
		lts, err := r.fetchNodeLTSVersions()
		if err != nil || len(lts) == 0 {
			return "", fmt.Errorf("failed to resolve Node LTS version")
		}
		// Return highest LTS
		sorted := version.SortVersions(lts)
		return sorted[len(sorted)-1], nil
	}
	available, err := r.GetNodeVersions()
	if err != nil { return "", err }
	spec, err := version.ParseSpec(versionSpec)
	if err != nil { return "", fmt.Errorf("invalid version specification %s: %w", versionSpec, err) }
	resolved, err := spec.Resolve(available)
	if err != nil { return "", fmt.Errorf("failed to resolve Node version %s: %w", versionSpec, err) }
	return resolved, nil
}

func (r *ToolRegistry) fetchNodeVersions() ([]string, error) {
	entries, err := r.fetchNodeIndex()
	if err != nil { return nil, err }
	var versions []string
	for _, e := range entries {
		v := strings.TrimPrefix(e.Version, "v")
		versions = append(versions, v)
	}
	return versions, nil
}

func (r *ToolRegistry) fetchNodeLTSVersions() ([]string, error) {
	entries, err := r.fetchNodeIndex()
	if err != nil { return nil, err }
	var versions []string
	for _, e := range entries {
		if e.LTS != nil && e.LTS != false {
			versions = append(versions, strings.TrimPrefix(e.Version, "v"))
		}
	}
	return versions, nil
}

type nodeIndexEntry struct {
	Version string      `json:"version"`
	LTS     interface{} `json:"lts"`
}

func (r *ToolRegistry) fetchNodeIndex() ([]nodeIndexEntry, error) {
	resp, err := r.httpClient.Get("https://nodejs.org/dist/index.json")
	if err != nil { return nil, err }
	defer resp.Body.Close()
	if resp.StatusCode != 200 { return nil, fmt.Errorf("node index fetch failed: %d", resp.StatusCode) }
	var entries []nodeIndexEntry
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&entries); err != nil { return nil, err }
	return entries, nil
}


// ResolveJavaVersion resolves a Java version specification to a concrete version
func (r *ToolRegistry) ResolveJavaVersion(versionSpec, distribution string) (string, error) {
	availableVersions, err := r.GetJavaVersions(distribution)
	if err != nil {
		return "", err
	}

	spec, err := version.ParseSpec(versionSpec)
	if err != nil {
		return "", fmt.Errorf("invalid version specification %s: %w", versionSpec, err)
	}

	resolved, err := spec.Resolve(availableVersions)
	if err != nil {
		return "", fmt.Errorf("failed to resolve Java %s version %s: %w", distribution, versionSpec, err)
	}

	return resolved, nil
}

// ResolveMavenVersion resolves a Maven version specification to a concrete version
func (r *ToolRegistry) ResolveMavenVersion(versionSpec string) (string, error) {
	availableVersions, err := r.GetMavenVersions()
	if err != nil {
		return "", err
	}

	spec, err := version.ParseSpec(versionSpec)
	if err != nil {
		return "", fmt.Errorf("invalid version specification %s: %w", versionSpec, err)
	}

	resolved, err := spec.Resolve(availableVersions)
	if err != nil {
		return "", fmt.Errorf("failed to resolve Maven version %s: %w", versionSpec, err)
	}

	return resolved, nil
}

// GetToolInfo returns information about a tool and its available versions
func (r *ToolRegistry) GetToolInfo(toolName string) (map[string]interface{}, error) {
	switch toolName {
	case "java":
		distributions := r.GetJavaDistributions()
		info := map[string]interface{}{
			"name":          "Java Development Kit",
			"distributions": distributions,
		}

		// Add version info for each distribution
		for _, dist := range distributions {
			if versions, err := r.GetJavaVersions(dist.Name); err == nil {
				info[dist.Name+"_versions"] = versions
			}
		}

		return info, nil

	case "maven":
		versions, err := r.GetMavenVersions()
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"name":     "Apache Maven",
			"versions": versions,
		}, nil

	case "mvnd":
		versions, err := r.GetMvndVersions()
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"name":     "Maven Daemon",
			"versions": versions,
		}, nil

	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}
