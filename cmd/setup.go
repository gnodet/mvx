package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gnodet/mvx/pkg/config"
	"github.com/gnodet/mvx/pkg/tools"
	"github.com/spf13/cobra"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup the build environment",
	Long: `Setup the build environment by installing all required tools and 
configuring the environment as specified in the mvx configuration.

This command will:
  - Read the project configuration (.mvx/config.json5 or .mvx/config.yml)
  - Download and install required tools (Java, Maven, Node.js, etc.)
  - Set up environment variables
  - Verify the installation

Examples:
  mvx setup                   # Setup everything
  mvx setup --tools-only      # Only install tools, skip environment setup`,

	Run: func(cmd *cobra.Command, args []string) {
		if err := setupEnvironment(); err != nil {
			printError("%v", err)
			os.Exit(1)
		}
	},
}

var (
	toolsOnly bool
)

func init() {
	setupCmd.Flags().BoolVar(&toolsOnly, "tools-only", false, "only install tools, skip environment setup")
}

func setupEnvironment() error {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	printVerbose("Project root: %s", projectRoot)

	// Check if .mvx directory exists
	mvxDir := filepath.Join(projectRoot, ".mvx")
	if _, err := os.Stat(mvxDir); os.IsNotExist(err) {
		return fmt.Errorf("no mvx configuration found. Run 'mvx init' first")
	}

	// Load configuration
	printInfo("🔍 Loading configuration...")
	cfg, err := config.LoadConfig(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	printVerbose("Loaded configuration for project: %s", cfg.Project.Name)

	// Create tool manager
	manager, err := tools.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create tool manager: %w", err)
	}

	// Install tools
	printInfo("📦 Installing tools...")
	if err := manager.InstallTools(cfg); err != nil {
		return fmt.Errorf("failed to install tools: %w", err)
	}

	if !toolsOnly {
		printInfo("🔧 Setting up environment...")
		env, err := manager.SetupEnvironment(cfg)
		if err != nil {
			return fmt.Errorf("failed to setup environment: %w", err)
		}

		// Show environment variables that would be set
		if verbose {
			printVerbose("Environment variables:")
			for key, value := range env {
				printVerbose("  %s=%s", key, value)
			}
		}

		printInfo("  ✅ Environment variables configured")
	}

	printInfo("")
	printInfo("✅ Setup complete! Your build environment is ready.")
	printInfo("")
	printInfo("Try running:")
	printInfo("  mvx build    # Build your project")
	printInfo("  mvx test     # Run tests")

	return nil
}
