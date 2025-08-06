package cmd

import (
	"fmt"
	"os"

	"github.com/gnodet/mvx/pkg/config"
	"github.com/gnodet/mvx/pkg/executor"
	"github.com/gnodet/mvx/pkg/tools"
	"github.com/spf13/cobra"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the project",
	Long: `Build the project using the configured build command.

This command will:
  - Load the project configuration
  - Set up the environment
  - Execute the build script defined in the configuration

Examples:
  mvx build                   # Run the default build command
  mvx build --clean           # Clean before building`,

	Run: func(cmd *cobra.Command, args []string) {
		if err := buildProject(); err != nil {
			printError("%v", err)
			os.Exit(1)
		}
	},
}

var (
	cleanBuild bool
)

func init() {
	buildCmd.Flags().BoolVar(&cleanBuild, "clean", false, "clean before building")
}

func buildProject() error {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	printVerbose("Project root: %s", projectRoot)

	// Load configuration
	cfg, err := config.LoadConfig(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create tool manager
	manager, err := tools.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create tool manager: %w", err)
	}

	// Create executor
	exec := executor.NewExecutor(cfg, manager, projectRoot)

	// If clean flag is set, run clean command first
	if cleanBuild {
		if _, exists := cfg.Commands["clean"]; exists {
			printInfo("🧹 Cleaning project...")
			if err := exec.ExecuteCommand("clean", []string{}); err != nil {
				return fmt.Errorf("clean command failed: %w", err)
			}
		} else {
			printInfo("🧹 Clean requested but no clean command configured")
		}
	}

	// Execute build command
	if err := exec.ValidateCommand("build"); err != nil {
		return err
	}

	return exec.ExecuteCommand("build", []string{})
}
