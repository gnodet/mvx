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
	// For now, return known Maven versions including latest RCs
	// TODO: Implement Maven Central API integration
	versions := []string{
		// Maven 4.x (pre-release versions)
		"4.0.0", "4.0.0-rc-4", "4.0.0-rc-3", "4.0.0-rc-2", "4.0.0-rc-1",
		"4.0.0-beta-5", "4.0.0-beta-4", "4.0.0-beta-3", "4.0.0-beta-2", "4.0.0-beta-1",
		"4.0.0-alpha-13", "4.0.0-alpha-12", "4.0.0-alpha-11", "4.0.0-alpha-10",

		// Maven 3.9.x (latest stable)
		"3.9.6", "3.9.5", "3.9.4", "3.9.3", "3.9.2", "3.9.1", "3.9.0",

		// Maven 3.8.x (previous stable)
		"3.8.8", "3.8.7", "3.8.6", "3.8.5", "3.8.4", "3.8.3", "3.8.2", "3.8.1",

		// Maven 3.6.x (older stable)
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
