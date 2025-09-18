package cmd

import (
	"fmt"
	"os"

	"github.com/gnodet/mvx/pkg/config"
	"github.com/gnodet/mvx/pkg/executor"
	"github.com/gnodet/mvx/pkg/tools"
	"github.com/spf13/cobra"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test [args...]",
	Short: "Run tests",
	Long: `Run tests using the configured test command.

This command will:
  - Load the project configuration
  - Set up the environment
  - Execute the test script defined in the configuration
  - Pass any additional arguments to the test script

Examples:
  mvx test                    # Run all tests
  mvx test --unit             # Run only unit tests
  mvx test --integration      # Run only integration tests
  mvx test -Dtest=MyTest      # Pass Maven test arguments through`,

	Run: func(cmd *cobra.Command, args []string) {
		if err := runTests(args); err != nil {
			printError("%v", err)
			os.Exit(1)
		}
	},
}

var (
	unitTests        bool
	integrationTests bool
)

func init() {
	testCmd.Flags().BoolVar(&unitTests, "unit", false, "run only unit tests")
	testCmd.Flags().BoolVar(&integrationTests, "integration", false, "run only integration tests")
}

func runTests(args []string) error {
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

	// Determine which test command to run
	commandName := "test"
	if unitTests {
		commandName = "test-unit"
	} else if integrationTests {
		commandName = "test-integration"
	}

	// Check if the specific test command exists, fall back to "test"
	if _, exists := cfg.Commands[commandName]; !exists {
		if commandName != "test" {
			printInfo("⚠️  Command '%s' not found, falling back to 'test'", commandName)
			commandName = "test"
		}
	}

	// Execute test command with hooks/overrides support
	return exec.ExecuteBuiltinCommand(commandName, args, func(args []string) error {
		// Default built-in test behavior
		if err := exec.ValidateCommand(commandName); err != nil {
			return err
		}
		return exec.ExecuteCommand(commandName, args)
	})
}
