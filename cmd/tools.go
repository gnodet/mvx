package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/gnodet/mvx/pkg/config"
	"github.com/gnodet/mvx/pkg/tools"
	"github.com/spf13/cobra"
)

// toolsCmd represents the tools command
var toolsCmd = &cobra.Command{
	Use:   "tools [subcommand]",
	Short: "Manage and discover tools",
	Long: `Manage and discover available tools and versions.

Subcommands:
  list       List available tools and their versions
  search     Search for specific tool versions
  info       Show detailed information about a tool
  add        Add a tool to the project configuration`,

	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// Default to list
			if err := listTools(); err != nil {
				printError("%v", err)
				os.Exit(1)
			}
			return
		}

		subcommand := args[0]
		switch subcommand {
		case "list":
			if err := listTools(); err != nil {
				printError("%v", err)
				os.Exit(1)
			}
		case "search":
			if len(args) < 2 {
				printError("search requires a tool name")
				os.Exit(1)
			}
			if err := searchTool(args[1], args[2:]); err != nil {
				printError("%v", err)
				os.Exit(1)
			}
		case "info":
			if len(args) < 2 {
				printError("info requires a tool name")
				os.Exit(1)
			}
			if err := showToolInfo(args[1]); err != nil {
				printError("%v", err)
				os.Exit(1)
			}
		case "add":
			if len(args) < 3 {
				printError("add requires a tool name and version")
				printError("Usage: mvx tools add <tool> <version> [distribution]")
				os.Exit(1)
			}
			distribution := ""
			if len(args) >= 4 {
				distribution = args[3]
			}
			if err := addTool(args[1], args[2], distribution); err != nil {
				printError("%v", err)
				os.Exit(1)
			}
		default:
			printError("unknown subcommand: %s", subcommand)
			cmd.Help()
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(toolsCmd)
}

// listTools shows all available tools
func listTools() error {
	manager, err := tools.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create tool manager: %w", err)
	}

	registry := manager.GetRegistry()

	printInfo("🛠️  Available Tools")
	printInfo("")

	// Java
	printInfo("📦 Java Development Kit")
	distributions := registry.GetJavaDistributions()
	for _, dist := range distributions {
		printInfo("  %s - %s", dist.Name, dist.DisplayName)
		if versions, err := registry.GetJavaVersions(dist.Name); err == nil && len(versions) > 0 {
			// Show first few versions
			shown := versions
			if len(versions) > 5 {
				shown = versions[:5]
			}
			printInfo("    Versions: %s", strings.Join(shown, ", "))
			if len(versions) > 5 {
				printInfo("    ... and %d more", len(versions)-5)
			}
		}
	}
	printInfo("")

	// Maven
	printInfo("📦 Apache Maven")
	if versions, err := registry.GetMavenVersions(); err == nil && len(versions) > 0 {
		// Show first few versions
		shown := versions
		if len(versions) > 8 {
			shown = versions[:8]
		}
		printInfo("  Versions: %s", strings.Join(shown, ", "))
		if len(versions) > 8 {
			printInfo("  ... and %d more", len(versions)-8)
		}
	}
	printInfo("")

	// Maven Daemon
	printInfo("🚀 Maven Daemon (mvnd)")
	if versions, err := registry.GetMvndVersions(); err == nil && len(versions) > 0 {
		shown := versions
		if len(versions) > 8 {
			shown = versions[:8]
		}
		printInfo("  Versions: %s", strings.Join(shown, ", "))
		if len(versions) > 8 {
			printInfo("  ... and %d more", len(versions)-8)
		}
	}

	// Node.js
	printInfo("📦 Node.js")
	if versions, err := registry.GetNodeVersions(); err == nil && len(versions) > 0 {
		shown := versions
		if len(versions) > 8 {
			shown = versions[:8]
		}
		printInfo("  Versions: %s", strings.Join(shown, ", "))
		if len(versions) > 8 {
			printInfo("  ... and %d more", len(versions)-8)
		}
	}

	// Go
	printInfo("🐹 Go Programming Language")
	if versions, err := registry.GetGoVersions(); err == nil && len(versions) > 0 {
		shown := versions
		if len(versions) > 8 {
			shown = versions[:8]
		}
		printInfo("  Versions: %s", strings.Join(shown, ", "))
		if len(versions) > 8 {
			printInfo("  ... and %d more", len(versions)-8)
		}
	}

	printInfo("")

	printInfo("Usage:")
	printInfo("  mvx tools search java        # Search Java versions")
	printInfo("  mvx tools search maven       # Search Maven versions")
	printInfo("  mvx tools search mvnd        # Search Maven Daemon versions")
	printInfo("  mvx tools search node        # Search Node.js versions")
	printInfo("  mvx tools search go          # Search Go versions")
	printInfo("  mvx tools info java          # Show Java details")
	printInfo("")
	printInfo("Add tools to your project:")
	printInfo("  mvx tools add java 21        # Add Java 21 (Temurin)")
	printInfo("  mvx tools add java 17 zulu   # Add Java 17 (Azul Zulu)")
	printInfo("  mvx tools add maven 4.0.0-rc-4  # Add Maven 4.0.0-rc-4")
	printInfo("  mvx tools add node lts       # Add Node.js LTS")
	printInfo("  mvx tools add go 1.23.1      # Add Go 1.23.1")

	return nil
}

// searchTool searches for versions of a specific tool
func searchTool(toolName string, filters []string) error {
	manager, err := tools.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create tool manager: %w", err)
	}

	registry := manager.GetRegistry()

	switch toolName {
	case "java":
		return searchJavaVersions(registry, filters)
	case "maven":
		return searchMavenVersions(registry, filters)
	case "mvnd":
		return searchMvndVersions(registry, filters)
	case "node":
		return searchNodeVersions(registry, filters)
	case "go":
		return searchGoVersions(registry, filters)
	default:
		return fmt.Errorf("unknown tool: %s", toolName)
	}
}

// searchJavaVersions searches for Java versions
func searchJavaVersions(registry *tools.ToolRegistry, filters []string) error {
	distributions := registry.GetJavaDistributions()

	// If distribution filter is provided, use it
	var targetDistributions []tools.JavaDistribution
	if len(filters) > 0 {
		distFilter := filters[0]
		for _, dist := range distributions {
			if dist.Name == distFilter {
				targetDistributions = []tools.JavaDistribution{dist}
				break
			}
		}
		if len(targetDistributions) == 0 {
			return fmt.Errorf("unknown Java distribution: %s", distFilter)
		}
	} else {
		targetDistributions = distributions
	}

	printInfo("☕ Java Versions")
	printInfo("")

	for _, dist := range targetDistributions {
		printInfo("📦 %s (%s)", dist.DisplayName, dist.Name)

		versions, err := registry.GetJavaVersions(dist.Name)
		if err != nil {
			printInfo("  ❌ Error: %v", err)
			continue
		}

		if len(versions) == 0 {
			printInfo("  No versions available")
			continue
		}

		// Group by major version
		majorVersions := make(map[string][]string)
		for _, v := range versions {
			major := strings.Split(v, ".")[0]
			majorVersions[major] = append(majorVersions[major], v)
		}

		// Sort major versions
		var majors []string
		for major := range majorVersions {
			majors = append(majors, major)
		}
		sort.Slice(majors, func(i, j int) bool {
			return majors[i] > majors[j] // Newest first
		})

		for _, major := range majors {
			versions := majorVersions[major]
			if len(versions) == 1 {
				printInfo("  %s", versions[0])
			} else {
				printInfo("  %s (%d versions available)", major, len(versions))
			}
		}
		printInfo("")
	}

	printInfo("Usage examples:")
	printInfo("  version: \"21\"           # Latest Java 21")
	printInfo("  version: \"17\"           # Latest Java 17")
	printInfo("  version: \"11\"           # Latest Java 11")
	printInfo("  distribution: \"graalvm\"  # Use GraalVM instead of Temurin")

	return nil
}

// searchMavenVersions searches for Maven versions
func searchMavenVersions(registry *tools.ToolRegistry, filters []string) error {
	versions, err := registry.GetMavenVersions()
	if err != nil {
		return fmt.Errorf("failed to get Maven versions: %w", err)
	}

	printInfo("📦 Apache Maven Versions")
	printInfo("")

	// Group by major version
	majorVersions := make(map[string][]string)
	for _, v := range versions {
		major := strings.Split(v, ".")[0]
		majorVersions[major] = append(majorVersions[major], v)
	}

	// Sort major versions
	var majors []string
	for major := range majorVersions {
		majors = append(majors, major)
	}
	sort.Slice(majors, func(i, j int) bool {
		return majors[i] > majors[j] // Newest first
	})

	for _, major := range majors {
		versions := majorVersions[major]
		printInfo("Maven %s.x:", major)

		// Show first few versions in each major
		shown := versions
		if len(versions) > 6 {
			shown = versions[:6]
		}

		for _, v := range shown {
			status := ""
			if strings.Contains(v, "rc") || strings.Contains(v, "beta") || strings.Contains(v, "alpha") {
				status = " (pre-release)"
			}
			printInfo("  %s%s", v, status)
		}

		if len(versions) > 6 {
			printInfo("  ... and %d more", len(versions)-6)
		}
		printInfo("")
	}

	printInfo("Usage examples:")
	printInfo("  version: \"3\"             # Latest Maven 3.x")
	printInfo("  version: \"3.9\"           # Latest Maven 3.9.x")
	printInfo("  version: \"3.9.6\"         # Exact version")
	printInfo("  version: \"4\"             # Latest Maven 4.x (pre-release)")

	return nil
}

// searchMvndVersions searches for Maven Daemon versions
func searchMvndVersions(registry *tools.ToolRegistry, filters []string) error {
	versions, err := registry.GetMvndVersions()
	if err != nil {
		return fmt.Errorf("failed to get mvnd versions: %w", err)
	}

	printInfo("🚀 Maven Daemon (mvnd) Versions")
	printInfo("")

	// Group by major version
	majorVersions := make(map[string][]string)
	for _, v := range versions {
		major := strings.Split(v, ".")[0]
		majorVersions[major] = append(majorVersions[major], v)
	}

	// Sort major versions
	var majors []string
	for major := range majorVersions {
		majors = append(majors, major)
	}
	sort.Slice(majors, func(i, j int) bool {
		return majors[i] > majors[j] // Newest first
	})

	// Display versions by major version
	for _, major := range majors {
		versions := majorVersions[major]
		printInfo("📦 mvnd %s.x:", major)

		// Show versions in rows of 6
		for i := 0; i < len(versions); i += 6 {
			end := i + 6
			if end > len(versions) {
				end = len(versions)
			}
			row := versions[i:end]
			printInfo("  %s", strings.Join(row, ", "))
		}
		printInfo("")
	}

	printInfo("Usage examples:")
	printInfo("  version: \"2\"             # Latest mvnd 2.x")
	printInfo("  version: \"1.0\"           # Latest mvnd 1.0.x")
	printInfo("  version: \"1.0.2\"         # Exact version")

	return nil
}

// searchNodeVersions searches for Node.js versions
func searchNodeVersions(registry *tools.ToolRegistry, filters []string) error {
	versions, err := registry.GetNodeVersions()
	if err != nil {
		return fmt.Errorf("failed to get Node.js versions: %w", err)
	}

	printInfo("📦 Node.js Versions")
	printInfo("")

	// Group by major version
	majorVersions := make(map[string][]string)
	for _, v := range versions {
		major := strings.Split(v, ".")[0]
		majorVersions[major] = append(majorVersions[major], v)
	}

	// Sort major versions
	var majors []string
	for major := range majorVersions {
		majors = append(majors, major)
	}
	sort.Slice(majors, func(i, j int) bool { return majors[i] > majors[j] })

	for _, major := range majors {
		versions := majorVersions[major]
		printInfo("Node %s.x:", major)
		shown := versions
		if len(versions) > 6 {
			shown = versions[:6]
		}
		for _, v := range shown {
			printInfo("  %s", v)
		}
		if len(versions) > 6 {
			printInfo("  ... and %d more", len(versions)-6)
		}
		printInfo("")
	}

	printInfo("Usage examples:")
	printInfo("  version: \"lts\"         # Latest LTS")
	printInfo("  version: \"20\"          # Latest Node 20.x")
	printInfo("  version: \"22.5.1\"      # Exact version")

	return nil
}

// searchGoVersions searches for Go versions
func searchGoVersions(registry *tools.ToolRegistry, filters []string) error {
	versions, err := registry.GetGoVersions()
	if err != nil {
		return fmt.Errorf("failed to get Go versions: %w", err)
	}

	printInfo("🐹 Go Versions")
	printInfo("")

	// Group by major.minor version
	majorVersions := make(map[string][]string)
	for _, v := range versions {
		parts := strings.Split(v, ".")
		if len(parts) >= 2 {
			major := parts[0] + "." + parts[1]
			majorVersions[major] = append(majorVersions[major], v)
		}
	}

	// Sort major versions
	var majors []string
	for major := range majorVersions {
		majors = append(majors, major)
	}
	sort.Slice(majors, func(i, j int) bool { return majors[i] > majors[j] })

	for _, major := range majors {
		versions := majorVersions[major]
		printInfo("Go %s.x:", major)
		shown := versions
		if len(versions) > 5 {
			shown = versions[:5]
		}
		for _, v := range shown {
			printInfo("  %s", v)
		}
		if len(versions) > 5 {
			printInfo("  ... and %d more", len(versions)-5)
		}
		printInfo("")
	}

	printInfo("Usage examples:")
	printInfo("  version: \"1.24\"         # Latest Go 1.24.x")
	printInfo("  version: \"1.23\"         # Latest Go 1.23.x")
	printInfo("  version: \"1.24.2\"       # Exact version")

	return nil
}

// showToolInfo shows detailed information about a tool
func showToolInfo(toolName string) error {
	manager, err := tools.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create tool manager: %w", err)
	}

	registry := manager.GetRegistry()
	info, err := registry.GetToolInfo(toolName)
	if err != nil {
		return err
	}

	printInfo("🔍 Tool Information: %s", toolName)
	printInfo("")
	printInfo("Name: %s", info["name"])

	// Tool-specific information
	switch toolName {
	case "java":
		if distributions, ok := info["distributions"].([]tools.JavaDistribution); ok {
			printInfo("")
			printInfo("Available Distributions:")
			for _, dist := range distributions {
				printInfo("  %s - %s", dist.Name, dist.DisplayName)
			}
		}
	case "maven":
		if versions, ok := info["versions"].([]string); ok {
			printInfo("")
			printInfo("Available Versions: %d", len(versions))
			printInfo("Latest: %s", versions[0])
		}
	case "mvnd":
		if versions, ok := info["versions"].([]string); ok {
			printInfo("")
			printInfo("Available Versions: %d", len(versions))
			printInfo("Latest: %s", versions[0])
		}
	case "go":
		if versions, ok := info["versions"].([]string); ok {
			printInfo("")
			printInfo("Available Versions: %d", len(versions))
			printInfo("Latest: %s", versions[0])
		}
	}

	return nil
}

// addTool adds a tool to the project configuration
func addTool(toolName, version, distribution string) error {
	// Find project root
	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	// Load existing configuration
	cfg, err := config.LoadConfig(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate tool name
	manager, err := tools.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create tool manager: %w", err)
	}

	registry := manager.GetRegistry()

	// Validate that the tool exists
	switch toolName {
	case "java", "maven", "mvnd", "node", "go":
		// Valid tools
	default:
		return fmt.Errorf("unknown tool: %s (supported: java, maven, mvnd, node, go)", toolName)
	}

	// Validate version exists for the tool
	if err := validateToolVersion(registry, toolName, version, distribution); err != nil {
		return err
	}

	// Initialize tools map if it doesn't exist
	if cfg.Tools == nil {
		cfg.Tools = make(map[string]config.ToolConfig)
	}

	// Create tool configuration
	toolConfig := config.ToolConfig{
		Version: version,
	}

	// Add distribution if specified and applicable
	if distribution != "" {
		if toolName == "java" {
			toolConfig.Distribution = distribution
		} else {
			printWarning("Distribution '%s' ignored for tool '%s' (only applicable to Java)", distribution, toolName)
		}
	}

	// Check if tool already exists
	if existingConfig, exists := cfg.Tools[toolName]; exists {
		printInfo("Tool '%s' already configured with version '%s'", toolName, existingConfig.Version)
		printInfo("Updating to version '%s'", version)
	}

	// Add/update the tool
	cfg.Tools[toolName] = toolConfig

	// Save the configuration
	if err := config.SaveConfig(cfg, projectRoot); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	printSuccess("✅ Added %s %s to project configuration", toolName, version)
	if distribution != "" && toolName == "java" {
		printSuccess("   Distribution: %s", distribution)
	}

	printInfo("")
	printInfo("To install the tool, run: mvx install")

	return nil
}

// validateToolVersion validates that a version exists for the given tool
func validateToolVersion(registry *tools.ToolRegistry, toolName, version, distribution string) error {
	switch toolName {
	case "java":
		// For Java, we need to validate the distribution if specified
		dist := distribution
		if dist == "" {
			dist = "temurin" // Default distribution
		}

		// Check if distribution exists
		distributions := registry.GetJavaDistributions()
		validDist := false
		for _, d := range distributions {
			if d.Name == dist {
				validDist = true
				break
			}
		}
		if !validDist {
			return fmt.Errorf("unknown Java distribution: %s", dist)
		}

		// For Java, we allow version patterns like "21", "17", "11" which resolve to latest
		// So we don't need strict validation here - the tool manager will resolve it
		printInfo("🔍 Java %s (%s) - version will be resolved during installation", version, dist)
		return nil

	case "maven":
		versions, err := registry.GetMavenVersions()
		if err != nil {
			return fmt.Errorf("failed to get Maven versions: %w", err)
		}

		// Check if exact version exists
		for _, v := range versions {
			if v == version {
				return nil
			}
		}

		printInfo("🔍 Maven %s - version will be resolved during installation", version)
		return nil

	case "mvnd":
		versions, err := registry.GetMvndVersions()
		if err != nil {
			return fmt.Errorf("failed to get mvnd versions: %w", err)
		}

		// Check if exact version exists
		for _, v := range versions {
			if v == version {
				return nil
			}
		}

		printInfo("🔍 Maven Daemon %s - version will be resolved during installation", version)
		return nil

	case "node":
		versions, err := registry.GetNodeVersions()
		if err != nil {
			return fmt.Errorf("failed to get Node.js versions: %w", err)
		}

		// For Node.js, we allow patterns like "lts", "20", "22.5.1"
		// Check for exact match or pattern
		if version == "lts" || version == "latest" {
			return nil
		}

		for _, v := range versions {
			if v == version {
				return nil
			}
		}

		printInfo("🔍 Node.js %s - version will be resolved during installation", version)
		return nil

	case "go":
		versions, err := registry.GetGoVersions()
		if err != nil {
			return fmt.Errorf("failed to get Go versions: %w", err)
		}

		// Check if exact version exists
		for _, v := range versions {
			if v == version {
				return nil
			}
		}

		printInfo("🔍 Go %s - version will be resolved during installation", version)
		return nil

	default:
		return fmt.Errorf("unknown tool: %s", toolName)
	}
}
