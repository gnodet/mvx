package tools

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
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

// GetJavaDistributions returns available Java distributions
func (r *ToolRegistry) GetJavaDistributions() []JavaDistribution {
	return []JavaDistribution{
		{
			Name:        "temurin",
			DisplayName: "Eclipse Temurin (OpenJDK)",
			APIBase:     "https://api.adoptium.net/v3",
		},
		{
			Name:        "graalvm",
			DisplayName: "GraalVM",
			APIBase:     "https://api.github.com/repos/graalvm/graalvm-ce-builds",
		},
		{
			Name:        "oracle",
			DisplayName: "Oracle JDK",
			APIBase:     "", // Oracle doesn't have a public API
		},
		{
			Name:        "corretto",
			DisplayName: "Amazon Corretto",
			APIBase:     "https://api.github.com/repos/corretto/corretto-8", // Different repos per version
		},
		{
			Name:        "liberica",
			DisplayName: "BellSoft Liberica",
			APIBase:     "", // Would need custom implementation
		},
	}
}

// GetJavaVersions returns available Java versions for a distribution
func (r *ToolRegistry) GetJavaVersions(distribution string) ([]string, error) {
	switch distribution {
	case "temurin", "":
		return r.getTemurinVersions()
	case "graalvm":
		return r.getGraalVMVersions()
	case "oracle":
		return r.getOracleVersions()
	case "corretto":
		return r.getCorrettoVersions()
	default:
		return nil, fmt.Errorf("unsupported Java distribution: %s", distribution)
	}
}

// getTemurinVersions fetches available Temurin versions from Adoptium API
func (r *ToolRegistry) getTemurinVersions() ([]string, error) {
	// Get available feature versions (major versions)
	resp, err := r.httpClient.Get("https://api.adoptium.net/v3/info/available_releases")
	if err != nil {
		// Fallback to known versions if API is unavailable
		return []string{"8", "11", "17", "21", "22", "23"}, nil
	}
	defer resp.Body.Close()
	
	var releases struct {
		AvailableReleases []int `json:"available_releases"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return []string{"8", "11", "17", "21", "22", "23"}, nil
	}
	
	var versions []string
	for _, release := range releases.AvailableReleases {
		versions = append(versions, fmt.Sprintf("%d", release))
	}
	
	// Sort in descending order (newest first)
	sort.Slice(versions, func(i, j int) bool {
		return versions[i] > versions[j]
	})
	
	return versions, nil
}

// getGraalVMVersions returns available GraalVM versions
func (r *ToolRegistry) getGraalVMVersions() ([]string, error) {
	// For now, return known GraalVM versions
	// TODO: Implement GitHub API integration
	return []string{"23.0.0", "22.3.0", "22.2.0", "22.1.0", "21.3.0"}, nil
}

// getOracleVersions returns available Oracle JDK versions
func (r *ToolRegistry) getOracleVersions() ([]string, error) {
	// Oracle doesn't provide a public API, return known versions
	return []string{"8", "11", "17", "21", "22", "23"}, nil
}

// getCorrettoVersions returns available Amazon Corretto versions
func (r *ToolRegistry) getCorrettoVersions() ([]string, error) {
	// For now, return known Corretto versions
	// TODO: Implement GitHub API integration for multiple Corretto repos
	return []string{"8", "11", "17", "21", "22"}, nil
}

// GetMavenVersions returns available Maven versions
func (r *ToolRegistry) GetMavenVersions() ([]string, error) {
	// For now, return known Maven versions
	// TODO: Implement Maven Central API integration
	versions := []string{
		"4.0.0", "4.0.0-rc-3", "4.0.0-beta-5", "4.0.0-beta-4", "4.0.0-beta-3",
		"3.9.6", "3.9.5", "3.9.4", "3.9.3", "3.9.2", "3.9.1", "3.9.0",
		"3.8.8", "3.8.7", "3.8.6", "3.8.5", "3.8.4", "3.8.3", "3.8.2", "3.8.1",
		"3.6.3", "3.6.2", "3.6.1", "3.6.0",
	}
	
	return version.SortVersions(versions), nil
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
		
	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}
