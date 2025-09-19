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
	commands, err := parseCommands(script)
	if err != nil {
		return fmt.Errorf("failed to parse script: %w", err)
	}

	for _, cmd := range commands {
		if err := s.executeCommand(cmd); err != nil {
			return err
		}
	}
	return nil
}

// Command represents a parsed command
type Command struct {
	Name string
	Args []string
}

// parseCommands parses a script into individual commands
// Supports basic command chaining with && and ||
func parseCommands(script string) ([]Command, error) {
	// Simple implementation - split by && for now
	// TODO: Add proper parsing for ||, ;, quotes, etc.
	parts := strings.Split(script, "&&")
	var commands []Command

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Split command and arguments
		fields := strings.Fields(part)
		if len(fields) == 0 {
			continue
		}

		commands = append(commands, Command{
			Name: fields[0],
			Args: fields[1:],
		})
	}

	return commands, nil
}

// executeCommand executes a single command
func (s *MVXShell) executeCommand(cmd Command) error {
	switch cmd.Name {
	case "cd":
		return s.changeDirectory(cmd.Args)
	case "echo":
		return s.echo(cmd.Args)
	case "mkdir":
		return s.makeDirectory(cmd.Args)
	case "rm":
		return s.remove(cmd.Args)
	case "copy", "cp":
		return s.copy(cmd.Args)
	case "open":
		return s.open(cmd.Args)
	default:
		// Execute as external command
		return s.executeExternal(cmd)
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

// echo prints text to stdout
func (s *MVXShell) echo(args []string) error {
	fmt.Println(strings.Join(args, " "))
	return nil
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

// executeExternal executes an external command
func (s *MVXShell) executeExternal(cmd Command) error {
	execCmd := exec.Command(cmd.Name, cmd.Args...)
	execCmd.Dir = s.workDir
	execCmd.Env = s.env
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
