package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gnodet/mvx/pkg/config"
	"github.com/gnodet/mvx/pkg/tools"
	"github.com/spf13/cobra"
)

var (
	envShell string
)

// envCmd represents the env command
var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Output environment variables for shell integration",
	Long: `Output environment variables in shell-specific format for integration
with shell activation hooks.

This command is primarily used by 'mvx activate' to set up the environment
when entering directories with .mvx configuration.

The command outputs shell-specific export statements that can be evaluated
to set up PATH and other environment variables.

Examples:
  # Bash/Zsh
  eval "$(mvx env --shell bash)"
  
  # Fish
  mvx env --shell fish | source
  
  # PowerShell
  Invoke-Expression (mvx env --shell powershell | Out-String)`,

	Run: func(cmd *cobra.Command, args []string) {
		if err := outputEnvironment(); err != nil {
			printError("%v", err)
			os.Exit(1)
		}
	},
}

func init() {
	envCmd.Flags().StringVar(&envShell, "shell", detectShell(), "shell type (bash, zsh, fish, powershell)")
}

// detectShell attempts to detect the current shell
func detectShell() string {
	// Try SHELL environment variable first
	shell := os.Getenv("SHELL")
	if shell != "" {
		if strings.Contains(shell, "bash") {
			return "bash"
		}
		if strings.Contains(shell, "zsh") {
			return "zsh"
		}
		if strings.Contains(shell, "fish") {
			return "fish"
		}
	}

	// Check for PowerShell on Windows
	if runtime.GOOS == "windows" {
		return "powershell"
	}

	// Default to bash
	return "bash"
}

// outputEnvironment outputs environment variables in shell-specific format
func outputEnvironment() error {
	// Find project root
	projectRoot, err := findProjectRoot()
	if err != nil {
		// No .mvx directory found - output nothing (silent)
		return nil
	}

	// Check if .mvx directory exists
	mvxDir := filepath.Join(projectRoot, ".mvx")
	if _, err := os.Stat(mvxDir); os.IsNotExist(err) {
		// No .mvx directory - output nothing (silent)
		return nil
	}

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

	// Get environment variables
	env, err := manager.SetupEnvironment(cfg)
	if err != nil {
		return fmt.Errorf("failed to setup environment: %w", err)
	}

	// Build PATH from tool bin directories
	pathDirs := []string{}

	for toolName, toolConfig := range cfg.Tools {
		// Check if user wants to use system tool instead
		systemEnvVar := fmt.Sprintf("MVX_USE_SYSTEM_%s", strings.ToUpper(toolName))
		if os.Getenv(systemEnvVar) == "true" {
			continue
		}

		tool, err := manager.GetTool(toolName)
		if err != nil {
			continue
		}

		// Resolve version
		resolvedVersion, err := manager.ResolveVersion(toolName, toolConfig)
		if err != nil {
			continue
		}

		// Create resolved config
		resolvedConfig := toolConfig
		resolvedConfig.Version = resolvedVersion

		// Try to get tool path
		binPath, err := tool.GetPath(resolvedVersion, resolvedConfig)
		if err == nil {
			pathDirs = append(pathDirs, binPath)
		}
	}

	// Output environment in shell-specific format
	switch envShell {
	case "bash", "zsh":
		return outputBashEnv(pathDirs, env)
	case "fish":
		return outputFishEnv(pathDirs, env)
	case "powershell":
		return outputPowerShellEnv(pathDirs, env)
	default:
		return fmt.Errorf("unsupported shell: %s", envShell)
	}
}

// outputBashEnv outputs environment in bash/zsh format
func outputBashEnv(pathDirs []string, env map[string]string) error {
	// Export PATH
	if len(pathDirs) > 0 {
		pathStr := strings.Join(pathDirs, string(os.PathListSeparator))
		fmt.Printf("export PATH=\"%s:$PATH\"\n", pathStr)
	}

	// Export other environment variables
	for key, value := range env {
		if key != "PATH" {
			// Escape quotes in value
			escapedValue := strings.ReplaceAll(value, `"`, `\"`)
			fmt.Printf("export %s=\"%s\"\n", key, escapedValue)
		}
	}

	return nil
}

// outputFishEnv outputs environment in fish format
func outputFishEnv(pathDirs []string, env map[string]string) error {
	// Set PATH
	if len(pathDirs) > 0 {
		for _, dir := range pathDirs {
			fmt.Printf("set -gx PATH \"%s\" $PATH\n", dir)
		}
	}

	// Set other environment variables
	for key, value := range env {
		if key != "PATH" {
			// Escape quotes in value
			escapedValue := strings.ReplaceAll(value, `"`, `\"`)
			fmt.Printf("set -gx %s \"%s\"\n", key, escapedValue)
		}
	}

	return nil
}

// outputPowerShellEnv outputs environment in PowerShell format
func outputPowerShellEnv(pathDirs []string, env map[string]string) error {
	// Set PATH
	if len(pathDirs) > 0 {
		pathStr := strings.Join(pathDirs, string(os.PathListSeparator))
		fmt.Printf("$env:PATH = \"%s;$env:PATH\"\n", pathStr)
	}

	// Set other environment variables
	for key, value := range env {
		if key != "PATH" {
			// Escape quotes in value
			escapedValue := strings.ReplaceAll(value, `"`, `\"`+"`"+`"`)
			fmt.Printf("$env:%s = \"%s\"\n", key, escapedValue)
		}
	}

	return nil
}
