package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

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
  info       Show detailed information about a tool`,

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
	printInfo("")

	printInfo("")

	printInfo("Usage:")
	printInfo("  mvx tools search java        # Search Java versions")
	printInfo("  mvx tools search maven       # Search Maven versions")
	printInfo("  mvx tools search mvnd        # Search Maven Daemon versions")
	printInfo("  mvx tools search node        # Search Node.js versions")
	printInfo("  mvx tools info java          # Show Java details")

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
	}

	return nil
}
