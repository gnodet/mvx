package shell

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// MVXShell provides cross-platform command execution
type MVXShell struct {
	workDir string
	env     []string
}

// NewMVXShell creates a new cross-platform shell instance
func NewMVXShell(workDir string, env []string) *MVXShell {
	return &MVXShell{
		workDir: workDir,
		env:     env,
	}
}

// Execute executes a script using the cross-platform interpreter
func (s *MVXShell) Execute(script string) error {
	chains, err := parseCommands(script)
	if err != nil {
		return fmt.Errorf("failed to parse script: %w", err)
	}

	var lastError error
	for _, chain := range chains {
		if err := s.executeCommandChain(chain); err != nil {
			lastError = err
			// Continue executing other chains (semicolon behavior)
		}
	}
	return lastError
}

// Command represents a parsed command
type Command struct {
	Name string
	Args []string
	Env  map[string]string // Environment variables for this command
}

// CommandChain represents a chain of commands with operators
type CommandChain struct {
	Commands  []Command
	Operators []string // "&&", "||", ";", "|"
}

// parseCommands parses a script into command chains
// Supports &&, ||, ;, pipes (|), and parentheses
func parseCommands(script string) ([]CommandChain, error) {
	tokens, err := tokenize(script)
	if err != nil {
		return nil, err
	}

	return parseTokens(tokens)
}

// Token represents a parsed token
type Token struct {
	Type  TokenType
	Value string
}

type TokenType int

const (
	TokenCommand TokenType = iota
	TokenOperator
	TokenPipe
	TokenLeftParen
	TokenRightParen
	TokenSemicolon
)

// tokenize breaks a script into tokens
func tokenize(script string) ([]Token, error) {
	var tokens []Token
	var current strings.Builder
	inQuotes := false
	quoteChar := byte(0)

	for i := 0; i < len(script); i++ {
		char := script[i]

		// Handle quotes
		if (char == '"' || char == '\'') && !inQuotes {
			inQuotes = true
			quoteChar = char
			current.WriteByte(char)
			continue
		} else if char == quoteChar && inQuotes {
			inQuotes = false
			quoteChar = 0
			current.WriteByte(char)
			continue
		}

		if inQuotes {
			current.WriteByte(char)
			continue
		}

		// Handle operators and special characters
		switch char {
		case '&':
			if i+1 < len(script) && script[i+1] == '&' {
				if current.Len() > 0 {
					tokens = append(tokens, Token{TokenCommand, strings.TrimSpace(current.String())})
					current.Reset()
				}
				tokens = append(tokens, Token{TokenOperator, "&&"})
				i++ // Skip next &
			} else {
				current.WriteByte(char)
			}
		case '|':
			if i+1 < len(script) && script[i+1] == '|' {
				if current.Len() > 0 {
					tokens = append(tokens, Token{TokenCommand, strings.TrimSpace(current.String())})
					current.Reset()
				}
				tokens = append(tokens, Token{TokenOperator, "||"})
				i++ // Skip next |
			} else {
				if current.Len() > 0 {
					tokens = append(tokens, Token{TokenCommand, strings.TrimSpace(current.String())})
					current.Reset()
				}
				tokens = append(tokens, Token{TokenPipe, "|"})
			}
		case ';':
			if current.Len() > 0 {
				tokens = append(tokens, Token{TokenCommand, strings.TrimSpace(current.String())})
				current.Reset()
			}
			tokens = append(tokens, Token{TokenSemicolon, ";"})
		case '(':
			if current.Len() > 0 {
				tokens = append(tokens, Token{TokenCommand, strings.TrimSpace(current.String())})
				current.Reset()
			}
			tokens = append(tokens, Token{TokenLeftParen, "("})
		case ')':
			if current.Len() > 0 {
				tokens = append(tokens, Token{TokenCommand, strings.TrimSpace(current.String())})
				current.Reset()
			}
			tokens = append(tokens, Token{TokenRightParen, ")"})
		case ' ', '\t', '\n', '\r':
			if current.Len() > 0 {
				// Don't break on whitespace, just add it to current command
				current.WriteByte(char)
			}
		default:
			current.WriteByte(char)
		}
	}

	if inQuotes {
		return nil, fmt.Errorf("unterminated quote in script")
	}

	if current.Len() > 0 {
		tokens = append(tokens, Token{TokenCommand, strings.TrimSpace(current.String())})
	}

	return tokens, nil
}

// parseTokens converts tokens into command chains
func parseTokens(tokens []Token) ([]CommandChain, error) {
	var chains []CommandChain
	var currentChain CommandChain
	lastWasOperator := false

	for _, token := range tokens {
		switch token.Type {
		case TokenCommand:
			if token.Value == "" {
				continue
			}
			cmd, err := parseCommand(token.Value)
			if err != nil {
				return nil, err
			}
			currentChain.Commands = append(currentChain.Commands, cmd)
			lastWasOperator = false

		case TokenOperator, TokenPipe:
			if len(currentChain.Commands) == 0 {
				return nil, fmt.Errorf("operator %s without preceding command", token.Value)
			}
			if lastWasOperator {
				return nil, fmt.Errorf("consecutive operators: %s", token.Value)
			}
			// Add operator to current chain
			currentChain.Operators = append(currentChain.Operators, token.Value)
			lastWasOperator = true

		case TokenSemicolon:
			// Semicolon ends the current chain (even if empty)
			if len(currentChain.Commands) > 0 {
				chains = append(chains, currentChain)
			}
			currentChain = CommandChain{}
			lastWasOperator = false

		case TokenLeftParen, TokenRightParen:
			// For now, ignore parentheses - treat them as whitespace
			// Full subshell support can be added later
			continue

		default:
			return nil, fmt.Errorf("unexpected token: %s", token.Value)
		}
	}

	if len(currentChain.Commands) > 0 {
		chains = append(chains, currentChain)
	}

	return chains, nil
}

// parseCommand parses a command string into Command struct
// Supports environment variable assignments like: VAR=value command args
func parseCommand(cmdStr string) (Command, error) {
	fields := strings.Fields(cmdStr)
	if len(fields) == 0 {
		return Command{}, fmt.Errorf("empty command")
	}

	env := make(map[string]string)
	var cmdStart int

	// Parse environment variable assignments at the beginning
	for i, field := range fields {
		if strings.Contains(field, "=") && !strings.HasPrefix(field, "-") {
			// This looks like an environment variable assignment
			parts := strings.SplitN(field, "=", 2)
			if len(parts) == 2 && isValidEnvVarName(parts[0]) {
				env[parts[0]] = parts[1]
				cmdStart = i + 1
				continue
			}
		}
		// First non-env-var field is the command
		cmdStart = i
		break
	}

	if cmdStart >= len(fields) {
		return Command{}, fmt.Errorf("no command found after environment variables")
	}

	return Command{
		Name: fields[cmdStart],
		Args: fields[cmdStart+1:],
		Env:  env,
	}, nil
}

// isValidEnvVarName checks if a string is a valid environment variable name
func isValidEnvVarName(name string) bool {
	if len(name) == 0 {
		return false
	}

	// First character must be letter or underscore
	first := name[0]
	if !((first >= 'A' && first <= 'Z') || (first >= 'a' && first <= 'z') || first == '_') {
		return false
	}

	// Remaining characters must be letters, digits, or underscores
	for i := 1; i < len(name); i++ {
		c := name[i]
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}

	return true
}

// executeCommandChain executes a chain of commands with operators
func (s *MVXShell) executeCommandChain(chain CommandChain) error {
	if len(chain.Commands) == 0 {
		return nil
	}

	// Handle single command
	if len(chain.Commands) == 1 {
		return s.executeCommand(chain.Commands[0])
	}

	// Handle command chain with operators
	for i, cmd := range chain.Commands {
		err := s.executeCommand(cmd)

		// If this is the last command, we're done
		if i >= len(chain.Operators) {
			return err
		}

		operator := chain.Operators[i]

		switch operator {
		case "&&":
			// Continue only if current command succeeded
			if err != nil {
				return err
			}
		case "||":
			// Continue only if current command failed
			if err == nil {
				// Command succeeded, skip the rest of the chain
				return nil
			}
			// Command failed, continue to next command
		case "|":
			// For pipes, we need to handle input/output redirection
			// For now, treat as sequential execution
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported operator: %s", operator)
		}
	}

	return nil
}

// executeCommand executes a single command
func (s *MVXShell) executeCommand(cmd Command) error {
	// Create environment map for variable expansion
	envMap := make(map[string]string)

	// Start with shell environment
	for _, envVar := range s.env {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	// Override with command-specific environment variables
	for key, value := range cmd.Env {
		envMap[key] = value
	}

	// Expand variables in command name and arguments
	expandedName := s.ExpandVariables(cmd.Name, envMap)
	expandedArgs := make([]string, len(cmd.Args))
	for i, arg := range cmd.Args {
		expandedArgs[i] = s.ExpandVariables(arg, envMap)
	}

	// Create new command with expanded values
	expandedCmd := Command{
		Name: expandedName,
		Args: expandedArgs,
		Env:  cmd.Env,
	}

	switch expandedCmd.Name {
	case "cd":
		return s.changeDirectory(expandedCmd.Args)
	case "echo":
		return s.echo(expandedCmd.Args, expandedCmd.Env)
	case "mkdir":
		return s.makeDirectory(expandedCmd.Args)
	case "rm":
		return s.remove(expandedCmd.Args)
	case "copy", "cp":
		return s.copy(expandedCmd.Args)
	case "open":
		return s.open(expandedCmd.Args)
	default:
		// Execute as external command
		return s.executeExternal(expandedCmd)
	}
}

// changeDirectory changes the current working directory
func (s *MVXShell) changeDirectory(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("cd: expected 1 argument, got %d", len(args))
	}

	newDir := args[0]
	if !filepath.IsAbs(newDir) {
		newDir = filepath.Join(s.workDir, newDir)
	}

	// Check if directory exists
	if _, err := os.Stat(newDir); os.IsNotExist(err) {
		return fmt.Errorf("cd: directory does not exist: %s", newDir)
	}

	s.workDir = newDir
	return nil
}

// echo prints text to stdout (variable expansion handled at command level)
func (s *MVXShell) echo(args []string, cmdEnv map[string]string) error {
	fmt.Println(strings.Join(args, " "))
	return nil
}

// ExpandVariables expands $VAR and ${VAR} syntax in a string
func (s *MVXShell) ExpandVariables(text string, envMap map[string]string) string {
	result := text

	// Handle ${VAR} syntax
	for {
		start := strings.Index(result, "${")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], "}")
		if end == -1 {
			break
		}
		end += start

		varName := result[start+2 : end]
		varValue := envMap[varName]
		result = result[:start] + varValue + result[end+1:]
	}

	// Handle $VAR syntax (simple variable names)
	for {
		start := strings.Index(result, "$")
		if start == -1 {
			break
		}

		// Find the end of the variable name
		end := start + 1
		for end < len(result) {
			c := result[end]
			if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_') {
				break
			}
			end++
		}

		if end == start+1 {
			// Just a $ with no variable name, skip it
			break
		}

		varName := result[start+1 : end]
		varValue := envMap[varName]
		result = result[:start] + varValue + result[end:]
	}

	return result
}

// makeDirectory creates directories
func (s *MVXShell) makeDirectory(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("mkdir: missing directory argument")
	}

	// Filter out flags (like -p) and collect directory names
	var dirs []string
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			dirs = append(dirs, arg)
		}
		// Note: We always use MkdirAll behavior (equivalent to -p flag)
	}

	if len(dirs) == 0 {
		return fmt.Errorf("mkdir: missing directory argument")
	}

	for _, dir := range dirs {
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(s.workDir, dir)
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("mkdir: failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

// remove removes files and directories
func (s *MVXShell) remove(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("rm: missing file argument")
	}

	for _, path := range args {
		if !filepath.IsAbs(path) {
			path = filepath.Join(s.workDir, path)
		}
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("rm: failed to remove %s: %w", path, err)
		}
	}
	return nil
}

// copy copies files
func (s *MVXShell) copy(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("copy: expected 2 arguments (source, destination), got %d", len(args))
	}

	src := args[0]
	dst := args[1]

	if !filepath.IsAbs(src) {
		src = filepath.Join(s.workDir, src)
	}
	if !filepath.IsAbs(dst) {
		dst = filepath.Join(s.workDir, dst)
	}

	return copyFile(src, dst)
}

// open opens a file or directory using the platform's default application
func (s *MVXShell) open(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("open: expected 1 argument, got %d", len(args))
	}

	path := args[0]
	if !filepath.IsAbs(path) {
		path = filepath.Join(s.workDir, path)
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}

	cmd.Dir = s.workDir
	cmd.Env = s.env
	return cmd.Run()
}

// logVerbose prints verbose log messages for mvx-shell
func logVerbose(format string, args ...interface{}) {
	if os.Getenv("MVX_VERBOSE") == "true" {
		fmt.Printf("[VERBOSE] "+format+"\n", args...)
	}
}

// executeExternal executes an external command
func (s *MVXShell) executeExternal(cmd Command) error {
	logVerbose("mvx-shell executing external command: %s %v", cmd.Name, cmd.Args)
	logVerbose("mvx-shell working directory: %s", s.workDir)
	logVerbose("mvx-shell environment variables count: %d", len(s.env))

	execCmd := exec.Command(cmd.Name, cmd.Args...)
	execCmd.Dir = s.workDir

	// Start with the shell's environment
	env := make([]string, len(s.env))
	copy(env, s.env)

	// Apply command-specific environment variables
	if len(cmd.Env) > 0 {
		// Create a map of existing environment variables for easy lookup
		envMap := make(map[string]string)
		for _, envVar := range s.env {
			parts := strings.SplitN(envVar, "=", 2)
			if len(parts) == 2 {
				envMap[parts[0]] = parts[1]
			}
		}

		// Override with command-specific environment variables
		for key, value := range cmd.Env {
			envMap[key] = value
		}

		// Convert back to slice format
		env = make([]string, 0, len(envMap))
		for key, value := range envMap {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Log PATH specifically
	for _, envVar := range env {
		if strings.HasPrefix(envVar, "PATH=") {
			logVerbose("mvx-shell PATH in environment: %s", envVar)
			break
		}
	}

	execCmd.Env = env
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	execCmd.Stdin = os.Stdin

	return execCmd.Run()
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(sourceFile)
	return err
}
