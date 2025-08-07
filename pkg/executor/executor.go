package executor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gnodet/mvx/pkg/config"
	"github.com/gnodet/mvx/pkg/tools"
)

// Executor handles command execution with proper environment setup
type Executor struct {
	config      *config.Config
	toolManager *tools.Manager
	projectRoot string
}

// NewExecutor creates a new command executor
func NewExecutor(cfg *config.Config, toolManager *tools.Manager, projectRoot string) *Executor {
	return &Executor{
		config:      cfg,
		toolManager: toolManager,
		projectRoot: projectRoot,
	}
}

// ExecuteCommand executes a configured command with arguments
func (e *Executor) ExecuteCommand(commandName string, args []string) error {
	// Get command configuration
	cmdConfig, exists := e.config.Commands[commandName]
	if !exists {
		return fmt.Errorf("unknown command: %s", commandName)
	}

	// Setup environment
	env, err := e.setupEnvironment(cmdConfig)
	if err != nil {
		return fmt.Errorf("failed to setup environment: %w", err)
	}

	// Determine working directory
	workDir := e.projectRoot
	if cmdConfig.WorkingDir != "" {
		workDir = filepath.Join(e.projectRoot, cmdConfig.WorkingDir)
	}

	// Process script (handle multiline scripts)
	script := e.processScript(cmdConfig.Script, args)

	// Execute command
	fmt.Printf("🔨 Running command: %s\n", commandName)
	if cmdConfig.Description != "" {
		fmt.Printf("   %s\n", cmdConfig.Description)
	}

	return e.executeScript(script, workDir, env)
}

// ListCommands returns available commands from configuration
func (e *Executor) ListCommands() map[string]string {
	commands := make(map[string]string)
	for name, cmd := range e.config.Commands {
		commands[name] = cmd.Description
	}
	return commands
}

// GetCommandInfo returns detailed information about a command
func (e *Executor) GetCommandInfo(commandName string) (*config.CommandConfig, error) {
	cmdConfig, exists := e.config.Commands[commandName]
	if !exists {
		return nil, fmt.Errorf("unknown command: %s", commandName)
	}
	return &cmdConfig, nil
}

// setupEnvironment prepares the environment for command execution
func (e *Executor) setupEnvironment(cmdConfig config.CommandConfig) ([]string, error) {
	// Start with current environment
	env := os.Environ()

	// Add global environment variables from config
	globalEnv, err := e.toolManager.SetupEnvironment(e.config)
	if err != nil {
		return nil, err
	}

	// Add global environment variables
	for key, value := range globalEnv {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	// Add command-specific environment variables
	for key, value := range cmdConfig.Environment {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	// Add tool paths to PATH
	pathDirs := []string{}

	// Get required tools for this command
	requiredTools := cmdConfig.Requires
	if len(requiredTools) == 0 {
		// If no specific requirements, use all configured tools
		for toolName := range e.config.Tools {
			requiredTools = append(requiredTools, toolName)
		}
	}

	// Add tool bin directories to PATH
	for _, toolName := range requiredTools {
		if toolConfig, exists := e.config.Tools[toolName]; exists {
			tool, err := e.toolManager.GetTool(toolName)
			if err != nil {
				continue // Skip unknown tools
			}

			if tool.IsInstalled(toolConfig.Version, toolConfig) {
				binPath, err := tool.GetBinPath(toolConfig.Version, toolConfig)
				if err != nil {
					continue
				}
				pathDirs = append(pathDirs, binPath)
			}
		}
	}

	// Prepend tool paths to existing PATH
	if len(pathDirs) > 0 {
		currentPath := os.Getenv("PATH")
		newPath := strings.Join(pathDirs, string(os.PathListSeparator))
		if currentPath != "" {
			newPath = newPath + string(os.PathListSeparator) + currentPath
		}

		// Update PATH in environment
		for i, envVar := range env {
			if strings.HasPrefix(envVar, "PATH=") {
				env[i] = "PATH=" + newPath
				break
			}
		}
	}

	return env, nil
}

// processScript processes the script string, handling multiline scripts and arguments
func (e *Executor) processScript(script string, args []string) string {
	// If there are arguments, append them to the script
	if len(args) > 0 {
		// Join arguments with spaces and append to script
		argsStr := strings.Join(args, " ")
		script = script + " " + argsStr
	}

	return script
}

// executeScript executes a script in the specified working directory with environment
func (e *Executor) executeScript(script, workDir string, env []string) error {
	// Determine shell
	shell := "/bin/bash"
	shellArgs := []string{"-c"}

	if runtime.GOOS == "windows" {
		shell = "cmd"
		shellArgs = []string{"/c"}
	}

	// Create command
	cmd := exec.Command(shell, append(shellArgs, script)...)
	cmd.Dir = workDir
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Execute command
	return cmd.Run()
}

// ValidateCommand checks if a command can be executed
func (e *Executor) ValidateCommand(commandName string) error {
	cmdConfig, exists := e.config.Commands[commandName]
	if !exists {
		return fmt.Errorf("unknown command: %s", commandName)
	}

	// Check if required tools are installed
	for _, toolName := range cmdConfig.Requires {
		toolConfig, exists := e.config.Tools[toolName]
		if !exists {
			return fmt.Errorf("command %s requires tool %s, but it's not configured", commandName, toolName)
		}

		tool, err := e.toolManager.GetTool(toolName)
		if err != nil {
			return fmt.Errorf("unknown tool %s required by command %s", toolName, commandName)
		}

		if !tool.IsInstalled(toolConfig.Version, toolConfig) {
			return fmt.Errorf("tool %s %s is required by command %s but not installed. Run 'mvx setup' first",
				toolName, toolConfig.Version, commandName)
		}
	}

	return nil
}
