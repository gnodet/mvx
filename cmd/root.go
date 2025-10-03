package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

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

	// Auto-setup cache to avoid repeated setup
	autoSetupDone bool
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
	// Auto-setup tools and environment before executing any command
	if err := autoSetupEnvironment(); err != nil {
		// If auto-setup fails, we should fail the command execution
		// This prevents commands from running with missing tools
		return fmt.Errorf("auto-setup failed: %w", err)
	}

	// Add dynamic custom commands and tool commands before execution
	if err := addCustomCommands(); err != nil {
		// If we can't load custom commands, continue with built-in commands only
		printVerbose("Failed to load custom commands: %v", err)
	}

	// Add automatic tool commands
	if err := addToolCommands(); err != nil {
		// If we can't load tool commands, continue without them
		printVerbose("Failed to load tool commands: %v", err)
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
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(updateBootstrapCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(activateCmd)
	rootCmd.AddCommand(deactivateCmd)
	rootCmd.AddCommand(envCmd)
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

// autoSetupEnvironment automatically installs tools and sets up environment
func autoSetupEnvironment() error {
	// Skip auto-setup if already done in this process
	if autoSetupDone {
		printVerbose("Auto-setup already completed in this process")
		return nil
	}

	// Skip auto-setup if explicitly disabled
	if os.Getenv("MVX_NO_AUTO_SETUP") == "true" {
		printVerbose("Auto-setup disabled by MVX_NO_AUTO_SETUP")
		autoSetupDone = true
		return nil
	}

	// Try to find project root
	projectRoot, err := findProjectRoot()
	if err != nil {
		// No project found, skip auto-setup
		printVerbose("No mvx project found, skipping auto-setup")
		return nil
	}

	// Try to load configuration
	cfg, err := config.LoadConfig(projectRoot)
	if err != nil {
		// No valid config found, skip auto-setup
		printVerbose("No valid mvx config found, skipping auto-setup")
		return nil
	}

	// Skip if no tools configured
	if len(cfg.Tools) == 0 {
		printVerbose("No tools configured, skipping auto-setup")
		return nil
	}

	// Create tool manager
	manager, err := tools.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create tool manager: %w", err)
	}

	// Check if tools need installation (excluding system tools)
	toolsToInstall, err := manager.GetToolsNeedingInstallation(cfg)
	if err != nil {
		return fmt.Errorf("failed to check tool installation status: %w", err)
	}

	// Filter out tools that should use system versions
	filteredToolsToInstall := make(map[string]config.ToolConfig)
	for toolName, toolConfig := range toolsToInstall {
		systemEnvVar := fmt.Sprintf("MVX_USE_SYSTEM_%s", strings.ToUpper(toolName))
		if os.Getenv(systemEnvVar) == "true" {
			printVerbose("Skipping %s installation: %s=true (using system tool)", toolName, systemEnvVar)
			continue
		}
		filteredToolsToInstall[toolName] = toolConfig
	}

	// Install missing tools if any
	if len(filteredToolsToInstall) > 0 {
		printInfo("🔧 Auto-installing %d missing tool(s)...", len(filteredToolsToInstall))

		// Create a temporary config with only tools that need installation
		tempCfg := *cfg
		tempCfg.Tools = filteredToolsToInstall

		opts := &tools.InstallOptions{
			MaxConcurrent: 3,
			Parallel:      true,
			Verbose:       verbose,
		}

		if err := manager.InstallToolsWithOptions(&tempCfg, opts); err != nil {
			return fmt.Errorf("failed to auto-install tools: %w", err)
		}

		printVerbose("Auto-installation complete")
	} else {
		printVerbose("All tools already installed or using system versions")
	}

	// Set up environment variables globally
	if err := setupGlobalEnvironment(cfg, manager); err != nil {
		return fmt.Errorf("failed to setup global environment: %w", err)
	}

	// Mark auto-setup as completed
	autoSetupDone = true
	printVerbose("Auto-setup completed successfully")

	return nil
}

// setupGlobalEnvironment sets up PATH and environment variables globally
func setupGlobalEnvironment(cfg *config.Config, manager *tools.Manager) error {
	// Get environment variables from tool manager (includes PATH with tool directories)
	env, err := manager.SetupEnvironment(cfg)
	if err != nil {
		return fmt.Errorf("failed to get tool environment: %w", err)
	}

	// Set all environment variables globally
	for key, value := range env {
		if err := os.Setenv(key, value); err != nil {
			printVerbose("Failed to set environment variable %s: %v", key, err)
		} else {
			if key == "PATH" {
				printVerbose("Set global PATH with tool directories")
				printVerbose("New PATH: %s", value)
			} else {
				printVerbose("Set global environment variable: %s=%s", key, value)
			}
		}
	}

	return nil
}

func printWarning(format string, args ...interface{}) {
	if !quiet {
		fmt.Fprintf(os.Stderr, "Warning: "+format+"\n", args...)
	}
}

func printSuccess(format string, args ...interface{}) {
	if !quiet {
		fmt.Printf(format+"\n", args...)
	}
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
		// Skip commands with spaces (they're subcommands, handled by their parent command)
		if strings.Contains(cmdName, " ") {
			printVerbose("Skipping custom command with space: %s (handled by parent command)", cmdName)
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

// addToolCommands dynamically adds tool commands (mvn, go, node, etc.) as top-level commands
func addToolCommands() error {
	// Try to find project root and load configuration
	projectRoot, err := findProjectRoot()
	if err != nil {
		return err // No project root found, skip tool commands
	}

	// Load configuration
	cfg, err := config.LoadConfig(projectRoot)
	if err != nil {
		return err // No configuration found, skip tool commands
	}

	// Create tool manager
	manager, err := tools.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create tool manager: %w", err)
	}

	// Create executor
	exec := executor.NewExecutor(cfg, manager, projectRoot)

	// Get all registered tool names from the manager
	registeredToolNames := manager.GetToolNames()

	for _, toolName := range registeredToolNames {
		// Check if this tool is configured in the project
		if _, exists := cfg.Tools[toolName]; !exists {
			continue // Skip tools not configured in this project
		}

		// Check if a command with this name already exists (avoid conflicts)
		if hasCommand(rootCmd, toolName) {
			continue // Skip if command already exists
		}

		// Create the tool command
		toolCmd := createToolCommand(toolName, exec)
		rootCmd.AddCommand(toolCmd)
		printVerbose("Added automatic tool command: %s", toolName)
	}

	return nil
}

// hasCommand checks if a command with the given name already exists
func hasCommand(cmd *cobra.Command, name string) bool {
	for _, subCmd := range cmd.Commands() {
		if subCmd.Name() == name {
			return true
		}
	}
	return false
}

// createToolCommand creates a cobra command for a specific tool
func createToolCommand(toolName string, exec *executor.Executor) *cobra.Command {
	return &cobra.Command{
		Use:   toolName + " [tool-args...]",
		Short: fmt.Sprintf("Run %s with mvx-managed environment", toolName),
		Long: fmt.Sprintf(`Run %s with the mvx-managed %s installation and proper environment setup.

This command automatically uses the %s version specified in your mvx configuration
and sets up the appropriate environment variables.

Examples:
  mvx %s --version           # Show %s version
  mvx %s [args...]           # Run %s with arguments`, toolName, toolName, toolName, toolName, toolName, toolName, toolName),

		DisableFlagParsing: true, // Allow all flags to be passed through to the tool
		Run: func(cmd *cobra.Command, args []string) {
			if err := exec.ExecuteTool(toolName, args); err != nil {
				printError("%v", err)
				os.Exit(1)
			}
		},
	}
}
