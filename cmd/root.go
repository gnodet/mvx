package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/gnodet/mvx/pkg/config"
	"github.com/gnodet/mvx/pkg/executor"
	"github.com/gnodet/mvx/pkg/tools"
	"github.com/spf13/cobra"
)

var (
	// Version information set from main
	version = "dev"
	commit  = "unknown"
	date    = "unknown"

	// Global flags
	verbose bool
	quiet   bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mvx",
	Short: "Maven eXtended - Universal build environment bootstrap",
	Long: `mvx is a universal build environment bootstrap tool that goes beyond Maven.

It provides zero-dependency bootstrapping, universal tool management, and simple
command interfaces for any project. Think of it as "Maven Wrapper for the modern era."

Examples:
  mvx setup          # Install all required tools automatically
  mvx build          # Build the project with the right environment
  mvx test           # Run tests with proper configuration
  mvx demo           # Launch project-specific demos

For more information, visit: https://github.com/gnodet/mvx`,

	// Show help if no command is provided
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	// Add dynamic custom commands before execution
	if err := addCustomCommands(); err != nil {
		// If we can't load custom commands, continue with built-in commands only
		printVerbose("Failed to load custom commands: %v", err)
	}

	return rootCmd.Execute()
}

// SetVersionInfo sets the version information from main
func SetVersionInfo(v, c, d string) {
	version = v
	commit = c
	date = d
}

func isWindows() bool { return runtime.GOOS == "windows" }

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "quiet output (errors only)")

	// Add subcommands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(updateBootstrapCmd)
}

// Helper functions for output
func printVerbose(format string, args ...interface{}) {
	if verbose && !quiet {
		fmt.Fprintf(os.Stderr, "[VERBOSE] "+format+"\n", args...)
	}
}

func printInfo(format string, args ...interface{}) {
	if !quiet {
		fmt.Printf(format+"\n", args...)
	}
}

func printError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
}

// Helper to find project root (directory containing .mvx/)
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		mvxDir := filepath.Join(dir, ".mvx")
		if info, err := os.Stat(mvxDir); err == nil && info.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			break
		}
		dir = parent
	}

	// If no .mvx directory found, use current directory
	return os.Getwd()
}

// addCustomCommands dynamically adds custom commands from configuration as top-level commands
func addCustomCommands() error {
	// Try to find project root and load configuration
	projectRoot, err := findProjectRoot()
	if err != nil {
		return err // No project root found, skip custom commands
	}

	// Load configuration
	cfg, err := config.LoadConfig(projectRoot)
	if err != nil {
		return err // No configuration found, skip custom commands
	}

	// Create tool manager and executor
	manager, err := tools.NewManager()
	if err != nil {
		return err
	}

	exec := executor.NewExecutor(cfg, manager, projectRoot)

	// Add each custom command as a top-level command
	for cmdName, cmdConfig := range cfg.Commands {
		// Skip built-in commands unless they have override flag
		if isBuiltinCommand(cmdName) && !cmdConfig.Override {
			continue
		}

		// Create a new cobra command for this custom command
		customCmd := createCustomCommand(cmdName, cmdConfig, exec)
		rootCmd.AddCommand(customCmd)
	}

	return nil
}

// createCustomCommand creates a cobra command for a custom command
func createCustomCommand(cmdName string, cmdConfig config.CommandConfig, exec *executor.Executor) *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdName + " [args...]",
		Short: cmdConfig.Description,
		Long:  fmt.Sprintf("%s\n\nThis is a custom command defined in your .mvx/config file.", cmdConfig.Description),
		Run: func(cmd *cobra.Command, args []string) {
			if err := exec.ExecuteCommand(cmdName, args); err != nil {
				printError("%v", err)
				os.Exit(1)
			}
		},
	}

	// Add arguments if defined
	for _, arg := range cmdConfig.Args {
		if arg.Required {
			cmd.MarkFlagRequired(arg.Name)
		}
	}

	return cmd
}

// isBuiltinCommand checks if a command name is a built-in mvx command
func isBuiltinCommand(commandName string) bool {
	builtinCommands := map[string]bool{
		"build":            true,
		"test":             true,
		"setup":            true,
		"init":             true,
		"tools":            true,
		"version":          true,
		"help":             true,
		"completion":       true,
		"info":             true,
		"update-bootstrap": true,
		"run":              true, // Don't override the run command itself
	}
	return builtinCommands[commandName]
}
