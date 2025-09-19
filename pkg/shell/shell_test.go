package shell

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseCommands(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		expected []Command
	}{
		{
			name:   "single command",
			script: "echo hello",
			expected: []Command{
				{Name: "echo", Args: []string{"hello"}},
			},
		},
		{
			name:   "multiple commands with &&",
			script: "cd test && echo hello && mkdir temp",
			expected: []Command{
				{Name: "cd", Args: []string{"test"}},
				{Name: "echo", Args: []string{"hello"}},
				{Name: "mkdir", Args: []string{"temp"}},
			},
		},
		{
			name:   "command with multiple arguments",
			script: "mvn clean install -DskipTests",
			expected: []Command{
				{Name: "mvn", Args: []string{"clean", "install", "-DskipTests"}},
			},
		},
		{
			name:     "empty script",
			script:   "",
			expected: []Command{},
		},
		{
			name:   "script with extra spaces",
			script: "  echo   hello   &&   cd   test  ",
			expected: []Command{
				{Name: "echo", Args: []string{"hello"}},
				{Name: "cd", Args: []string{"test"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseCommands(tt.script)
			if err != nil {
				t.Errorf("parseCommands() error = %v", err)
				return
			}
			if len(result) != len(tt.expected) {
				t.Errorf("parseCommands() returned %d commands, expected %d", len(result), len(tt.expected))
				return
			}
			for i, cmd := range result {
				if cmd.Name != tt.expected[i].Name {
					t.Errorf("parseCommands() command %d name = %v, expected %v", i, cmd.Name, tt.expected[i].Name)
				}
				if len(cmd.Args) != len(tt.expected[i].Args) {
					t.Errorf("parseCommands() command %d args length = %d, expected %d", i, len(cmd.Args), len(tt.expected[i].Args))
					continue
				}
				for j, arg := range cmd.Args {
					if arg != tt.expected[i].Args[j] {
						t.Errorf("parseCommands() command %d arg %d = %v, expected %v", i, j, arg, tt.expected[i].Args[j])
					}
				}
			}
		})
	}
}

func TestMVXShell_Echo(t *testing.T) {
	tempDir := t.TempDir()
	shell := NewMVXShell(tempDir, os.Environ())

	err := shell.echo([]string{"hello", "world"})
	if err != nil {
		t.Errorf("echo() error = %v", err)
	}
}

func TestMVXShell_MakeDirectory(t *testing.T) {
	tempDir := t.TempDir()
	shell := NewMVXShell(tempDir, os.Environ())

	testDir := "test-dir"
	err := shell.makeDirectory([]string{testDir})
	if err != nil {
		t.Errorf("makeDirectory() error = %v", err)
	}

	// Check if directory was created
	fullPath := filepath.Join(tempDir, testDir)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Errorf("makeDirectory() did not create directory %s", fullPath)
	}
}

func TestMVXShell_ChangeDirectory(t *testing.T) {
	tempDir := t.TempDir()
	shell := NewMVXShell(tempDir, os.Environ())

	// Create a subdirectory
	subDir := "subdir"
	err := os.Mkdir(filepath.Join(tempDir, subDir), 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Test changing to subdirectory
	err = shell.changeDirectory([]string{subDir})
	if err != nil {
		t.Errorf("changeDirectory() error = %v", err)
	}

	expectedPath := filepath.Join(tempDir, subDir)
	if shell.workDir != expectedPath {
		t.Errorf("changeDirectory() workDir = %v, expected %v", shell.workDir, expectedPath)
	}

	// Test changing to non-existent directory
	err = shell.changeDirectory([]string{"nonexistent"})
	if err == nil {
		t.Errorf("changeDirectory() should have failed for non-existent directory")
	}
}

func TestMVXShell_Remove(t *testing.T) {
	tempDir := t.TempDir()
	shell := NewMVXShell(tempDir, os.Environ())

	// Create a test file
	testFile := "test-file.txt"
	fullPath := filepath.Join(tempDir, testFile)
	err := os.WriteFile(fullPath, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Remove the file
	err = shell.remove([]string{testFile})
	if err != nil {
		t.Errorf("remove() error = %v", err)
	}

	// Check if file was removed
	if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
		t.Errorf("remove() did not remove file %s", fullPath)
	}
}

func TestMVXShell_Copy(t *testing.T) {
	tempDir := t.TempDir()
	shell := NewMVXShell(tempDir, os.Environ())

	// Create a test file
	srcFile := "source.txt"
	dstFile := "destination.txt"
	srcPath := filepath.Join(tempDir, srcFile)
	dstPath := filepath.Join(tempDir, dstFile)

	testContent := "test content"
	err := os.WriteFile(srcPath, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Copy the file
	err = shell.copy([]string{srcFile, dstFile})
	if err != nil {
		t.Errorf("copy() error = %v", err)
	}

	// Check if destination file exists and has correct content
	content, err := os.ReadFile(dstPath)
	if err != nil {
		t.Errorf("copy() destination file not readable: %v", err)
	}
	if string(content) != testContent {
		t.Errorf("copy() destination content = %v, expected %v", string(content), testContent)
	}
}

func TestMVXShell_Execute(t *testing.T) {
	tempDir := t.TempDir()
	shell := NewMVXShell(tempDir, os.Environ())

	// Test simple command execution
	err := shell.Execute("echo hello")
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	// Test command chaining
	err = shell.Execute("mkdir testdir && cd testdir")
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	// Check if directory was created and we changed into it
	expectedPath := filepath.Join(tempDir, "testdir")
	if shell.workDir != expectedPath {
		t.Errorf("Execute() workDir = %v, expected %v", shell.workDir, expectedPath)
	}
}
