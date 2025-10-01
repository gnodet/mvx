package executor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gnodet/mvx/pkg/config"
	"github.com/gnodet/mvx/pkg/shell"
	"github.com/gnodet/mvx/pkg/tools"
)

// logVerbose prints verbose log messages
func logVerbose(format string, args ...interface{}) {
	if os.Getenv("MVX_VERBOSE") == "true" {
		fmt.Printf("[VERBOSE] "+format+"\n", args...)
	}
}

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

	// Process script and resolve interpreter (handle platform-specific scripts)
	script, interpreter, err := config.ResolvePlatformScriptWithInterpreter(cmdConfig.Script, cmdConfig.Interpreter)
	if err != nil {
		return fmt.Errorf("failed to resolve script: %w", err)
	}

	// Process script arguments
	processedScript := e.processScriptString(script, args)

	// Execute command
	fmt.Printf("🔨 Running command: %s\n", commandName)
	if cmdConfig.Description != "" {
		fmt.Printf("   %s\n", cmdConfig.Description)
	}

	return e.executeScriptWithInterpreter(processedScript, workDir, env, interpreter)
}

// ExecuteTool executes a tool command with mvx-managed environment
func (e *Executor) ExecuteTool(toolName string, args []string) error {
	// Check if the tool is configured
	toolConfig, exists := e.config.Tools[toolName]
	if !exists {
		return fmt.Errorf("tool %s is not configured in this project", toolName)
	}

	// EnsureTool handles everything: resolve, check, install, get path
	toolBinPath, err := e.toolManager.EnsureTool(toolName, toolConfig)
	if err != nil {
		return fmt.Errorf("failed to ensure %s is installed: %w", toolName, err)
	}

	// Setup environment with tool paths
	env, err := e.setupToolEnvironment(toolName, toolBinPath)
	if err != nil {
		return fmt.Errorf("failed to setup environment for %s: %w", toolName, err)
	}

	// Execute the tool
	toolExecutable := toolName
	if len(args) == 0 {
		args = []string{"--version"} // Default to showing version if no args
	}

	// Create and execute command
	cmd := exec.Command(toolExecutable, args...)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir = e.projectRoot

	return cmd.Run()
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
	// Create environment map starting with current environment
	envVars := make(map[string]string)
	for _, envVar := range os.Environ() {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) == 2 {
			envVars[parts[0]] = parts[1]
		}
	}

	// Add global environment variables from config (these override system ones)
	globalEnv, err := e.toolManager.SetupEnvironment(e.config)
	if err != nil {
		return nil, err
	}

	for key, value := range globalEnv {
		envVars[key] = value
	}

	// Add command-specific environment variables (these override global ones)
	for key, value := range cmdConfig.Environment {
		envVars[key] = value
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
	logVerbose("Required tools for command: %v", requiredTools)

	// Add tool bin directories to PATH
	for _, toolName := range requiredTools {
		if toolConfig, exists := e.config.Tools[toolName]; exists {
			// EnsureTool handles version resolution, installation check, auto-install, and path retrieval
			binPath, err := e.toolManager.EnsureTool(toolName, toolConfig)
			if err != nil {
				logVerbose("Skipping tool %s: %v", toolName, err)
				continue
			}

			logVerbose("Adding %s bin path to PATH: %s", toolName, binPath)
			pathDirs = append(pathDirs, binPath)
		}
	}

	// Prepend tool paths to existing PATH
	if len(pathDirs) > 0 {
		currentPath := envVars["PATH"]
		newPath := strings.Join(pathDirs, string(os.PathListSeparator))
		if currentPath != "" {
			newPath = newPath + string(os.PathListSeparator) + currentPath
		}
		envVars["PATH"] = newPath
		logVerbose("Updated PATH with %d tool directories: %s", len(pathDirs), newPath)
	} else {
		logVerbose("No tool directories added to PATH")
	}

	// Convert environment map back to slice format
	var env []string
	for key, value := range envVars {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	return env, nil
}

// processScriptString processes a script string with arguments
func (e *Executor) processScriptString(script string, args []string) string {
	// If there are arguments, append them to the script
	if len(args) > 0 {
		// Join arguments with spaces and append to script
		argsStr := strings.Join(args, " ")
		script = script + " " + argsStr
	}

	return script
}

// executeScriptWithInterpreter executes a script using the specified interpreter
func (e *Executor) executeScriptWithInterpreter(script, workDir string, env []string, interpreter string) error {
	logVerbose("executeScriptWithInterpreter called with interpreter: '%s', script: '%s'", interpreter, script)

	// Default to native interpreter if not specified
	if interpreter == "" || interpreter == "native" {
		logVerbose("Using native interpreter")
		return e.executeNativeScript(script, workDir, env)
	}

	// Use mvx-shell interpreter
	if interpreter == "mvx-shell" {
		mvxShell := shell.NewMVXShell(workDir, env)
		return mvxShell.Execute(script)
	}

	return fmt.Errorf("unknown interpreter: %s", interpreter)
}

// executeNativeScript executes a script using the native system shell
func (e *Executor) executeNativeScript(script, workDir string, env []string) error {
	// Determine shell
	shell := "/bin/bash"
	shellArgs := []string{"-c"}

	if runtime.GOOS == "windows" {
		shell = "cmd"
		shellArgs = []string{"/c"}
	}

	logVerbose("Executing native script: %s", script)
	logVerbose("Working directory: %s", workDir)
	logVerbose("Environment variables count: %d", len(env))

	// Log PATH specifically
	for _, envVar := range env {
		if strings.HasPrefix(envVar, "PATH=") {
			logVerbose("PATH in environment: %s", envVar)
			break
		}
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

// ValidateCommand is deprecated - tools are now auto-installed via EnsureTool
// This method is kept for backward compatibility but does nothing
func (e *Executor) ValidateCommand(commandName string) error {
	// Just check if command exists
	_, exists := e.config.Commands[commandName]
	if !exists {
		return fmt.Errorf("unknown command: %s", commandName)
	}
	// Note: Tool installation checks removed - EnsureTool handles automatic installation
	return nil
}

// setupToolEnvironment prepares the environment for tool execution
func (e *Executor) setupToolEnvironment(toolName, toolBinPath string) ([]string, error) {
	// Create environment map starting with current environment
	envVars := make(map[string]string)
	for _, envVar := range os.Environ() {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) == 2 {
			envVars[parts[0]] = parts[1]
		}
	}

	// Add global environment variables from config
	globalEnv, err := e.toolManager.SetupEnvironment(e.config)
	if err != nil {
		return nil, err
	}

	for key, value := range globalEnv {
		envVars[key] = value
	}

	// Add tool binary directory to PATH
	pathDirs := []string{toolBinPath}

	// Add existing PATH
	if existingPath, exists := envVars["PATH"]; exists {
		pathDirs = append(pathDirs, existingPath)
	}

	// Set PATH with tool directory first
	envVars["PATH"] = strings.Join(pathDirs, string(os.PathListSeparator))

	// Tool-specific environment variables are already set by SetupEnvironment above
	// which calls each tool's EnvironmentProvider.SetupEnvironment() method

	// Convert map back to slice
	env := make([]string, 0, len(envVars))
	for key, value := range envVars {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	return env, nil
}
