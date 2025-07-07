package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/gnodet/mvx/pkg/config"
	"github.com/gnodet/mvx/pkg/executor"
	"github.com/gnodet/mvx/pkg/tools"
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info [command]",
	Short: "Show detailed information about a command",
	Long: `Show detailed information about a command including its description,
script, requirements, and environment variables.

Examples:
  mvx info build             # Show info about build command
  mvx info test              # Show info about test command
  mvx info                   # Show project information`,
	
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// No command specified, show project info
			if err := showProjectInfo(); err != nil {
				printError("%v", err)
				os.Exit(1)
			}
			return
		}
		
		commandName := args[0]
		if err := showCommandInfo(commandName); err != nil {
			printError("%v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

// showProjectInfo displays general project information
func showProjectInfo() error {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}
	
	// Load configuration
	cfg, err := config.LoadConfig(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	
	printInfo("📋 Project Information")
	printInfo("")
	printInfo("Name:        %s", cfg.Project.Name)
	printInfo("Description: %s", cfg.Project.Description)
	printInfo("Root:        %s", projectRoot)
	printInfo("")
	
	// Show configured tools
	if len(cfg.Tools) > 0 {
		printInfo("🛠️  Configured Tools:")
		for toolName, toolConfig := range cfg.Tools {
			distribution := ""
			if toolConfig.Distribution != "" {
				distribution = fmt.Sprintf(" (%s)", toolConfig.Distribution)
			}
			printInfo("  %s: %s%s", toolName, toolConfig.Version, distribution)
		}
		printInfo("")
	}
	
	// Show available commands
	if len(cfg.Commands) > 0 {
		printInfo("⚡ Available Commands: %d", len(cfg.Commands))
		printInfo("  Run 'mvx run' to see all commands")
		printInfo("  Run 'mvx info <command>' for command details")
	}
	
	return nil
}

// showCommandInfo displays detailed information about a specific command
func showCommandInfo(commandName string) error {
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
	
	// Get command info
	cmdInfo, err := exec.GetCommandInfo(commandName)
	if err != nil {
		return err
	}
	
	printInfo("⚡ Command: %s", commandName)
	printInfo("")
	
	if cmdInfo.Description != "" {
		printInfo("Description: %s", cmdInfo.Description)
		printInfo("")
	}
	
	// Show script
	printInfo("Script:")
	scriptLines := strings.Split(strings.TrimSpace(cmdInfo.Script), "\n")
	for _, line := range scriptLines {
		printInfo("  %s", line)
	}
	printInfo("")
	
	// Show working directory if specified
	if cmdInfo.WorkingDir != "" {
		printInfo("Working Directory: %s", cmdInfo.WorkingDir)
		printInfo("")
	}
	
	// Show required tools
	if len(cmdInfo.Requires) > 0 {
		printInfo("Required Tools:")
		for _, toolName := range cmdInfo.Requires {
			if toolConfig, exists := cfg.Tools[toolName]; exists {
				status := "❌ Not installed"
				if tool, err := manager.GetTool(toolName); err == nil {
					if tool.IsInstalled(toolConfig.Version, toolConfig) {
						status = "✅ Installed"
					}
				}
				printInfo("  %s %s - %s", toolName, toolConfig.Version, status)
			} else {
				printInfo("  %s - ❌ Not configured", toolName)
			}
		}
		printInfo("")
	}
	
	// Show environment variables
	if len(cmdInfo.Environment) > 0 {
		printInfo("Environment Variables:")
		for key, value := range cmdInfo.Environment {
			printInfo("  %s=%s", key, value)
		}
		printInfo("")
	}
	
	// Show arguments if defined
	if len(cmdInfo.Args) > 0 {
		printInfo("Arguments:")
		for _, arg := range cmdInfo.Args {
			required := ""
			if arg.Required {
				required = " (required)"
			}
			defaultVal := ""
			if arg.Default != "" {
				defaultVal = fmt.Sprintf(" [default: %s]", arg.Default)
			}
			printInfo("  %s%s - %s%s", arg.Name, required, arg.Description, defaultVal)
		}
		printInfo("")
	}
	
	// Show usage
	printInfo("Usage:")
	printInfo("  mvx run %s", commandName)
	if commandName == "build" || commandName == "test" {
		printInfo("  mvx %s", commandName)
	}
	
	return nil
}
