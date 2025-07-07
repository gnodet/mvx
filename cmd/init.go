package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	// Init command flags
	initFormat string
	initForce  bool
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize mvx in the current project",
	Long: `Initialize mvx configuration in the current project directory.

This creates a .mvx directory with a default configuration file that you can
customize for your project's needs.

Examples:
  mvx init                    # Create config.json5 (default)
  mvx init --format=yaml      # Create config.yml instead
  mvx init --force            # Overwrite existing configuration`,
	
	Run: func(cmd *cobra.Command, args []string) {
		if err := initProject(); err != nil {
			printError("%v", err)
			os.Exit(1)
		}
	},
}

func init() {
	initCmd.Flags().StringVar(&initFormat, "format", "json5", "configuration format (json5, yaml)")
	initCmd.Flags().BoolVar(&initForce, "force", false, "overwrite existing configuration")
}

func initProject() error {
	projectRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	
	mvxDir := filepath.Join(projectRoot, ".mvx")
	
	// Create .mvx directory
	if err := os.MkdirAll(mvxDir, 0755); err != nil {
		return fmt.Errorf("failed to create .mvx directory: %w", err)
	}
	
	// Determine config file name and content based on format
	var configFile string
	var configContent string
	
	switch initFormat {
	case "json5":
		configFile = "config.json5"
		configContent = getDefaultJSON5Config()
	case "yaml", "yml":
		configFile = "config.yml"
		configContent = getDefaultYAMLConfig()
	default:
		return fmt.Errorf("unsupported format: %s (supported: json5, yaml)", initFormat)
	}
	
	configPath := filepath.Join(mvxDir, configFile)
	
	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil && !initForce {
		return fmt.Errorf("configuration file already exists: %s (use --force to overwrite)", configPath)
	}
	
	// Write config file
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}
	
	printInfo("✅ Initialized mvx configuration in %s", configPath)
	printInfo("")
	printInfo("Next steps:")
	printInfo("  1. Edit %s to configure your project", configPath)
	printInfo("  2. Run 'mvx setup' to install required tools")
	printInfo("  3. Run 'mvx build' to build your project")
	
	return nil
}

func getDefaultJSON5Config() string {
	return `{
  // mvx configuration
  // See: https://github.com/gnodet/mvx for documentation
  
  project: {
    name: "my-project",
    description: "A sample project",
  },
  
  tools: {
    // Java configuration
    java: {
      version: "21",
      distribution: "temurin",
    },
    
    // Maven configuration
    maven: {
      version: "3.9.6",
    },
  },
  
  environment: {
    // JVM options for Maven builds
    MAVEN_OPTS: "-Xmx2g",
  },
  
  commands: {
    build: {
      description: "Build the project",
      script: "./mvnw clean install",
    },
    
    test: {
      description: "Run tests",
      script: "./mvnw test",
    },
    
    clean: {
      description: "Clean build artifacts",
      script: "./mvnw clean",
    },
  },
}
`
}

func getDefaultYAMLConfig() string {
	return `# mvx configuration
# See: https://github.com/gnodet/mvx for documentation

project:
  name: my-project
  description: A sample project

tools:
  # Java configuration
  java:
    version: "21"
    distribution: temurin
  
  # Maven configuration
  maven:
    version: "3.9.6"

environment:
  # JVM options for Maven builds
  MAVEN_OPTS: "-Xmx2g"

commands:
  build:
    description: Build the project
    script: ./mvnw clean install
  
  test:
    description: Run tests
    script: ./mvnw test
  
  clean:
    description: Clean build artifacts
    script: ./mvnw clean
`
}
