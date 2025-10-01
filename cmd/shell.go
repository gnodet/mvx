package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/gnodet/mvx/pkg/config"
	"github.com/gnodet/mvx/pkg/shell"
	"github.com/gnodet/mvx/pkg/tools"
	"github.com/spf13/cobra"
)

// shellCmd represents the shell command for executing shell commands in mvx environment
var shellCmd = &cobra.Command{
	Use:   "shell [command...]",
	Short: "Execute shell commands in the mvx environment",
	Long: `Execute shell commands with the mvx-managed tools and environment setup.

This command runs shell commands using the mvx-shell interpreter with access to
all mvx-managed tools and their environment variables.

Examples:
  mvx shell echo '$JAVA_HOME'           # Show Java home directory
  mvx shell env                         # Show all environment variables
  mvx shell "mvn --version"             # Run Maven with mvx environment
  mvx shell "echo '$PATH' | grep mvx"   # Show mvx paths in PATH
  mvx shell cd /tmp && pwd              # Change directory and show current path`,

	Run: func(cmd *cobra.Command, args []string) {
		if err := runShellCommand(args); err != nil {
			printError("%v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(shellCmd)
}

// runShellCommand executes shell commands in the mvx environment
func runShellCommand(args []string) error {
	// Get current working directory
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Load configuration
	cfg, err := config.LoadConfig(workDir)
	if err != nil {
		printVerbose("No configuration found, using defaults: %v", err)
		cfg = &config.Config{}
	}

	// Create tool manager
	manager, err := tools.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create tool manager: %w", err)
	}

	// Setup environment with mvx-managed tools
	env, err := setupShellEnvironment(cfg, manager, workDir)
	if err != nil {
		return fmt.Errorf("failed to setup environment: %w", err)
	}

	// Join all arguments into a single command string
	var command string
	if len(args) == 0 {
		return fmt.Errorf("no command specified. Use 'mvx shell <command>' to execute shell commands")
	} else if len(args) == 1 {
		command = args[0]
	} else {
		command = strings.Join(args, " ")
	}

	printVerbose("Executing shell command: %s", command)
	printVerbose("Working directory: %s", workDir)
	printVerbose("Environment variables: %d", len(env))

	// Create mvx-shell instance and execute command
	mvxShell := shell.NewMVXShell(workDir, env)
	return mvxShell.Execute(command)
}

// setupShellEnvironment sets up the environment for shell execution
func setupShellEnvironment(cfg *config.Config, manager *tools.Manager, workDir string) ([]string, error) {
	// Start with current environment
	env := os.Environ()

	// Get environment variables from tool manager (this handles all configured tools)
	toolEnv, err := manager.SetupEnvironment(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to setup tool environment: %w", err)
	}

	// Merge tool environment into shell environment
	env = mergeEnvironment(env, toolEnv)

	return env, nil
}

// mergeEnvironment merges tool environment variables into the shell environment
func mergeEnvironment(baseEnv []string, toolEnv map[string]string) []string {
	// Create a map of existing environment variables
	envMap := make(map[string]string)
	for _, envVar := range baseEnv {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	// Override with tool environment variables
	for key, value := range toolEnv {
		envMap[key] = value
	}

	// Convert back to slice
	result := make([]string, 0, len(envMap))
	for key, value := range envMap {
		result = append(result, fmt.Sprintf("%s=%s", key, value))
	}

	return result
}
