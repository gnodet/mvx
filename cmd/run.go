package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/gnodet/mvx/pkg/config"
	"github.com/gnodet/mvx/pkg/executor"
	"github.com/gnodet/mvx/pkg/tools"
	"github.com/spf13/cobra"
)

// runCmd represents the run command for executing custom commands
var runCmd = &cobra.Command{
	Use:   "run [command] [args...]",
	Short: "Run a custom command defined in configuration",
	Long: `Run a custom command defined in the mvx configuration file.

This command executes scripts defined in the .mvx/config file with the proper
environment setup including tool paths and environment variables.

Examples:
  mvx run build              # Run the build command
  mvx run test               # Run the test command  
  mvx run demo gogo          # Run demo command with arguments
  mvx run                    # List all available commands`,

	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// No command specified, list available commands
			if err := listCommands(); err != nil {
				printError("%v", err)
				os.Exit(1)
			}
			return
		}

		commandName := args[0]
		commandArgs := args[1:]

		if err := runCustomCommand(commandName, commandArgs); err != nil {
			printError("%v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}

// listCommands shows all available commands from configuration
func listCommands() error {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	// Load configuration
	cfg, err := config.LoadConfig(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create tool manager and executor
	manager, err := tools.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create tool manager: %w", err)
	}

	exec := executor.NewExecutor(cfg, manager, projectRoot)
	commands := exec.ListCommands()

	if len(commands) == 0 {
		printInfo("No custom commands defined in configuration.")
		printInfo("Add commands to your .mvx/config file to get started.")
		return nil
	}

	printInfo("Available commands:")
	printInfo("")

	// Sort commands for consistent output
	var names []string
	for name := range commands {
		names = append(names, name)
	}
	sort.Strings(names)

	// Find the longest command name for alignment
	maxLen := 0
	for _, name := range names {
		if len(name) > maxLen {
			maxLen = len(name)
		}
	}

	// Print commands with descriptions
	for _, name := range names {
		description := commands[name]
		if description == "" {
			description = "No description"
		}

		// Pad command name for alignment
		padding := strings.Repeat(" ", maxLen-len(name)+2)
		printInfo("  %s%s%s", name, padding, description)
	}

	printInfo("")
	printInfo("Usage: mvx run <command> [args...]")
	printInfo("   or: mvx <command> [args...]  (for built-in commands)")

	return nil
}

// runCustomCommand executes a custom command with arguments
func runCustomCommand(commandName string, args []string) error {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	printVerbose("Project root: %s", projectRoot)
	printVerbose("Command: %s", commandName)
	printVerbose("Arguments: %v", args)

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

	// Execute command (tools are auto-installed via EnsureTool)
	return exec.ExecuteCommand(commandName, args)
}
