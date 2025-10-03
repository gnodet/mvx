package cmd

import (
	"fmt"
	"os"
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

	printInfo("🛠️  Available Tools")
	printInfo("")

	// Get all tools from manager
	allTools := manager.GetAllTools()

	// Define tool order for consistent display
	toolOrder := []string{tools.ToolJava, tools.ToolMaven, tools.ToolMvnd, tools.ToolNode, tools.ToolGo}

	for _, toolName := range toolOrder {
		tool, exists := allTools[toolName]
		if !exists {
			continue
		}

		// Get display metadata
		emoji := "📦" // default emoji for all tools
		displayName := tool.GetDisplayName()

		printInfo("%s %s", emoji, displayName)

		// Check if tool supports distributions
		if distProvider, ok := tool.(tools.DistributionProvider); ok {
			distributions := distProvider.GetDistributions()
			for _, dist := range distributions {
				printInfo("  %s - %s", dist.Name, dist.DisplayName)

				// Get versions for this distribution if supported
				if distVersionProvider, ok := tool.(tools.DistributionVersionProvider); ok {
					if versions, err := distVersionProvider.ListVersionsForDistribution(dist.Name); err == nil && len(versions) > 0 {
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
			}
		} else {
			// Tool doesn't support distributions, just list versions
			if versions, err := tool.ListVersions(); err == nil && len(versions) > 0 {
				shown := versions
				if len(versions) > 8 {
					shown = versions[:8]
				}
				printInfo("  Versions: %s", strings.Join(shown, ", "))
				if len(versions) > 8 {
					printInfo("  ... and %d more", len(versions)-8)
				}
			}
		}
		printInfo("")
	}

	printInfo("Usage:")
	printInfo("  mvx tools search java           # Search Java versions")
	printInfo("  mvx tools search maven          # Search Maven versions")
	printInfo("  mvx tools search mvnd           # Search Maven Daemon versions")
	printInfo("  mvx tools search node           # Search Node.js versions")
	printInfo("  mvx tools search go             # Search Go versions")

	printInfo("  mvx tools info java             # Show Java details")
	printInfo("")
	printInfo("Add tools to your project:")
	printInfo("  mvx tools add java 21           # Add Java 21 (Temurin)")
	printInfo("  mvx tools add java 17 zulu      # Add Java 17 (Azul Zulu)")
	printInfo("  mvx tools add maven 4.0.0-rc-4  # Add Maven 4.0.0-rc-4")
	printInfo("  mvx tools add node lts          # Add Node.js LTS")

	printInfo("  mvx tools add go 1.23.1         # Add Go 1.23.1")

	return nil
}

// searchTool searches for versions of a specific tool
func searchTool(toolName string, filters []string) error {
	manager, err := tools.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create tool manager: %w", err)
	}

	// Use manager's search functionality instead of switch statement
	versions, err := manager.SearchToolVersions(toolName, filters)
	if err != nil {
		return err
	}

	// Print the search results
	printInfo("🔍 %s Versions", strings.Title(toolName))
	printInfo("")

	if len(versions) == 0 {
		printInfo("No versions found")
		return nil
	}

	// Display versions (limit to first 20 for readability)
	displayed := versions
	if len(versions) > 20 {
		displayed = versions[:20]
	}

	for _, version := range displayed {
		printInfo("  %s", version)
	}

	if len(versions) > 20 {
		printInfo("  ... and %d more", len(versions)-20)
	}

	printInfo("")
	printInfo("Usage: mvx tools add %s <version>", toolName)

	return nil
}

// showToolInfo shows detailed information about a tool
func showToolInfo(toolName string) error {
	manager, err := tools.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create tool manager: %w", err)
	}

	// Use manager's tool info functionality instead of switch statement
	info, err := manager.GetToolInfo(toolName)
	if err != nil {
		return err
	}

	printInfo("🔍 Tool Information: %s", toolName)
	printInfo("")
	printInfo("Name: %s", info["name"])

	// Display distributions if available (for Java)
	if distributions, ok := info["distributions"].([]tools.Distribution); ok {
		printInfo("")
		printInfo("Available Distributions:")
		for _, dist := range distributions {
			printInfo("  %s - %s", dist.Name, dist.DisplayName)
		}
	}

	// Display version information if available
	if versions, ok := info["versions"].([]string); ok {
		printInfo("")
		printInfo("Available Versions: %d", len(versions))
		if len(versions) > 0 {
			printInfo("Latest: %s", versions[0])
		}
	}

	// Display any additional tool-specific information
	for key, value := range info {
		if key != "name" && key != "distributions" && key != "versions" && !strings.HasSuffix(key, "_versions") {
			printInfo("%s: %v", key, value)
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

	manager, err := tools.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create tool manager: %w", err)
	}

	// Validate that the tool exists and version is valid
	if err := manager.ValidateToolVersion(toolName, version, distribution); err != nil {
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
	printInfo("To install the tool, run: mvx setup")

	return nil
}
