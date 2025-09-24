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

// ExecuteBuiltinCommand executes a built-in command with optional hooks and overrides
func (e *Executor) ExecuteBuiltinCommand(commandName string, args []string, builtinFunc func([]string) error) error {
	// Check if command is overridden
	if e.config.HasCommandOverride(commandName) {
		cmdConfig, _ := e.config.GetCommandConfig(commandName)
		fmt.Printf("🔨 Running overridden command: %s\n", commandName)
		if cmdConfig.Description != "" {
			fmt.Printf("   %s\n", cmdConfig.Description)
		}
		return e.ExecuteCommand(commandName, args)
	}

	// Check if command has hooks
	if e.config.HasCommandHooks(commandName) {
		return e.executeWithHooks(commandName, args, builtinFunc)
	}

	// Execute built-in command normally
	return builtinFunc(args)
}

// executeWithHooks executes a built-in command with pre/post hooks
func (e *Executor) executeWithHooks(commandName string, args []string, builtinFunc func([]string) error) error {
	cmdConfig, _ := e.config.GetCommandConfig(commandName)

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

	fmt.Printf("🔨 Running command with hooks: %s\n", commandName)
	if cmdConfig.Description != "" {
		fmt.Printf("   %s\n", cmdConfig.Description)
	}

	// Execute pre-hook
	if config.HasValidScript(cmdConfig.Pre) {
		fmt.Printf("   ⚡ Running pre-hook...\n")
		preScript, preInterpreter, err := config.ResolvePlatformScriptWithInterpreter(cmdConfig.Pre, cmdConfig.Interpreter)
		if err != nil {
			return fmt.Errorf("pre-hook script resolution failed: %w", err)
		}
		processedPreScript := e.processScriptString(preScript, args)
		if err := e.executeScriptWithInterpreter(processedPreScript, workDir, env, preInterpreter); err != nil {
			return fmt.Errorf("pre-hook failed: %w", err)
		}
	}

	// Execute built-in command or custom script
	if config.HasValidScript(cmdConfig.Script) {
		// Custom script instead of built-in
		fmt.Printf("   🔧 Running custom script...\n")
		script, scriptInterpreter, err := config.ResolvePlatformScriptWithInterpreter(cmdConfig.Script, cmdConfig.Interpreter)
		if err != nil {
			return fmt.Errorf("custom script resolution failed: %w", err)
		}
		processedScript := e.processScriptString(script, args)
		if err := e.executeScriptWithInterpreter(processedScript, workDir, env, scriptInterpreter); err != nil {
			return fmt.Errorf("custom script failed: %w", err)
		}
	} else {
		// Built-in command
		fmt.Printf("   🏗️  Running built-in command...\n")
		if err := builtinFunc(args); err != nil {
			return fmt.Errorf("built-in command failed: %w", err)
		}
	}

	// Execute post-hook
	if config.HasValidScript(cmdConfig.Post) {
		fmt.Printf("   ⚡ Running post-hook...\n")
		postScript, postInterpreter, err := config.ResolvePlatformScriptWithInterpreter(cmdConfig.Post, cmdConfig.Interpreter)
		if err != nil {
			return fmt.Errorf("post-hook script resolution failed: %w", err)
		}
		processedPostScript := e.processScriptString(postScript, args)
		if err := e.executeScriptWithInterpreter(processedPostScript, workDir, env, postInterpreter); err != nil {
			return fmt.Errorf("post-hook failed: %w", err)
		}
	}

	return nil
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
			tool, err := e.toolManager.GetTool(toolName)
			if err != nil {
				logVerbose("Skipping tool %s: %v", toolName, err)
				continue // Skip unknown tools
			}

			// Resolve version to handle any overrides
			resolvedVersion, err := e.toolManager.ResolveVersion(toolName, toolConfig)
			if err != nil {
				logVerbose("Skipping tool %s: version resolution failed: %v", toolName, err)
				continue // Skip tools with resolution errors
			}

			// Create resolved config
			resolvedConfig := toolConfig
			resolvedConfig.Version = resolvedVersion

			if tool.IsInstalled(resolvedVersion, resolvedConfig) {
				binPath, err := tool.GetBinPath(resolvedVersion, resolvedConfig)
				if err != nil {
					logVerbose("Skipping tool %s: failed to get bin path: %v", toolName, err)
					continue
				}
				logVerbose("Adding %s bin path to PATH: %s", toolName, binPath)
				pathDirs = append(pathDirs, binPath)
			} else {
				logVerbose("Tool %s version %s is not installed", toolName, resolvedVersion)
			}
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

// processScript processes the script, handling platform-specific scripts and arguments
func (e *Executor) processScript(script interface{}, args []string) (string, error) {
	// Resolve platform-specific script
	resolvedScript, err := config.ResolvePlatformScript(script)
	if err != nil {
		return "", err
	}

	return e.processScriptString(resolvedScript, args), nil
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

// executeScript executes a script in the specified working directory with environment
// This method is kept for backward compatibility
func (e *Executor) executeScript(script, workDir string, env []string) error {
	return e.executeScriptWithInterpreter(script, workDir, env, "")
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
