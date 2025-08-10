package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

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
