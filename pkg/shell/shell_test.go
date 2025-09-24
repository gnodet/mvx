package shell

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		expected []Token
	}{
		{
			name:   "simple command",
			script: "echo hello",
			expected: []Token{
				{TokenCommand, "echo hello"},
			},
		},
		{
			name:   "command with &&",
			script: "echo hello && echo world",
			expected: []Token{
				{TokenCommand, "echo hello"},
				{TokenOperator, "&&"},
				{TokenCommand, "echo world"},
			},
		},
		{
			name:   "command with ||",
			script: "false || echo backup",
			expected: []Token{
				{TokenCommand, "false"},
				{TokenOperator, "||"},
				{TokenCommand, "echo backup"},
			},
		},
		{
			name:   "command with pipe",
			script: "echo hello | grep hello",
			expected: []Token{
				{TokenCommand, "echo hello"},
				{TokenPipe, "|"},
				{TokenCommand, "grep hello"},
			},
		},
		{
			name:   "command with semicolon",
			script: "echo hello; echo world",
			expected: []Token{
				{TokenCommand, "echo hello"},
				{TokenSemicolon, ";"},
				{TokenCommand, "echo world"},
			},
		},
		{
			name:   "command with parentheses",
			script: "(echo hello) && echo world",
			expected: []Token{
				{TokenLeftParen, "("},
				{TokenCommand, "echo hello"},
				{TokenRightParen, ")"},
				{TokenOperator, "&&"},
				{TokenCommand, "echo world"},
			},
		},
		{
			name:   "quoted arguments",
			script: "echo 'hello world' && echo \"test\"",
			expected: []Token{
				{TokenCommand, "echo 'hello world'"},
				{TokenOperator, "&&"},
				{TokenCommand, "echo \"test\""},
			},
		},
		{
			name:   "complex chain",
			script: "cd test && mvn clean install || echo failed; echo done",
			expected: []Token{
				{TokenCommand, "cd test"},
				{TokenOperator, "&&"},
				{TokenCommand, "mvn clean install"},
				{TokenOperator, "||"},
				{TokenCommand, "echo failed"},
				{TokenSemicolon, ";"},
				{TokenCommand, "echo done"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tokenize(tt.script)
			if err != nil {
				t.Errorf("tokenize() error = %v", err)
				return
			}
			if len(result) != len(tt.expected) {
				t.Errorf("tokenize() returned %d tokens, expected %d", len(result), len(tt.expected))
				t.Errorf("Got: %+v", result)
				t.Errorf("Expected: %+v", tt.expected)
				return
			}
			for i, token := range result {
				if token.Type != tt.expected[i].Type || token.Value != tt.expected[i].Value {
					t.Errorf("tokenize() token %d = %+v, expected %+v", i, token, tt.expected[i])
				}
			}
		})
	}
}

func TestParseCommands(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		expected []CommandChain
	}{
		{
			name:   "single command",
			script: "echo hello",
			expected: []CommandChain{
				{
					Commands: []Command{{Name: "echo", Args: []string{"hello"}}},
				},
			},
		},
		{
			name:   "commands with &&",
			script: "cd test && echo hello",
			expected: []CommandChain{
				{
					Commands:  []Command{{Name: "cd", Args: []string{"test"}}, {Name: "echo", Args: []string{"hello"}}},
					Operators: []string{"&&"},
				},
			},
		},
		{
			name:   "commands with ||",
			script: "false || echo backup",
			expected: []CommandChain{
				{
					Commands:  []Command{{Name: "false", Args: []string{}}, {Name: "echo", Args: []string{"backup"}}},
					Operators: []string{"||"},
				},
			},
		},
		{
			name:   "commands with pipe",
			script: "echo hello | grep hello",
			expected: []CommandChain{
				{
					Commands:  []Command{{Name: "echo", Args: []string{"hello"}}, {Name: "grep", Args: []string{"hello"}}},
					Operators: []string{"|"},
				},
			},
		},
		{
			name:   "commands with semicolon",
			script: "echo hello; echo world",
			expected: []CommandChain{
				{Commands: []Command{{Name: "echo", Args: []string{"hello"}}}},
				{Commands: []Command{{Name: "echo", Args: []string{"world"}}}},
			},
		},
		{
			name:   "complex chain",
			script: "cd test && mvn clean install || echo failed",
			expected: []CommandChain{
				{
					Commands: []Command{
						{Name: "cd", Args: []string{"test"}},
						{Name: "mvn", Args: []string{"clean", "install"}},
						{Name: "echo", Args: []string{"failed"}},
					},
					Operators: []string{"&&", "||"},
				},
			},
		},
		{
			name:   "multiple chains with semicolon",
			script: "echo first; echo second && echo third",
			expected: []CommandChain{
				{Commands: []Command{{Name: "echo", Args: []string{"first"}}}},
				{
					Commands:  []Command{{Name: "echo", Args: []string{"second"}}, {Name: "echo", Args: []string{"third"}}},
					Operators: []string{"&&"},
				},
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
				t.Errorf("parseCommands() returned %d chains, expected %d", len(result), len(tt.expected))
				t.Errorf("Got: %+v", result)
				t.Errorf("Expected: %+v", tt.expected)
				return
			}
			for i, chain := range result {
				expectedChain := tt.expected[i]
				if len(chain.Commands) != len(expectedChain.Commands) {
					t.Errorf("parseCommands() chain %d has %d commands, expected %d", i, len(chain.Commands), len(expectedChain.Commands))
					continue
				}
				if len(chain.Operators) != len(expectedChain.Operators) {
					t.Errorf("parseCommands() chain %d has %d operators, expected %d", i, len(chain.Operators), len(expectedChain.Operators))
					continue
				}
				for j, cmd := range chain.Commands {
					expectedCmd := expectedChain.Commands[j]
					if cmd.Name != expectedCmd.Name {
						t.Errorf("parseCommands() chain %d command %d name = %v, expected %v", i, j, cmd.Name, expectedCmd.Name)
					}
					if len(cmd.Args) != len(expectedCmd.Args) {
						t.Errorf("parseCommands() chain %d command %d args length = %d, expected %d", i, j, len(cmd.Args), len(expectedCmd.Args))
						continue
					}
					for k, arg := range cmd.Args {
						if arg != expectedCmd.Args[k] {
							t.Errorf("parseCommands() chain %d command %d arg %d = %v, expected %v", i, j, k, arg, expectedCmd.Args[k])
						}
					}
				}
				for j, op := range chain.Operators {
					if op != expectedChain.Operators[j] {
						t.Errorf("parseCommands() chain %d operator %d = %v, expected %v", i, j, op, expectedChain.Operators[j])
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

func TestMVXShell_ExecuteCommandChain(t *testing.T) {
	tempDir := t.TempDir()
	shell := NewMVXShell(tempDir, os.Environ())

	tests := []struct {
		name        string
		chain       CommandChain
		expectError bool
		description string
	}{
		{
			name: "single command success",
			chain: CommandChain{
				Commands: []Command{{Name: "echo", Args: []string{"hello"}}},
			},
			expectError: false,
			description: "Single echo command should succeed",
		},
		{
			name: "AND chain success",
			chain: CommandChain{
				Commands:  []Command{{Name: "echo", Args: []string{"first"}}, {Name: "echo", Args: []string{"second"}}},
				Operators: []string{"&&"},
			},
			expectError: false,
			description: "Both commands should execute with &&",
		},
		{
			name: "AND chain with failure",
			chain: CommandChain{
				Commands:  []Command{{Name: "nonexistent-command"}, {Name: "echo", Args: []string{"should not run"}}},
				Operators: []string{"&&"},
			},
			expectError: true,
			description: "Second command should not run if first fails with &&",
		},
		{
			name: "OR chain with success",
			chain: CommandChain{
				Commands:  []Command{{Name: "echo", Args: []string{"first"}}, {Name: "echo", Args: []string{"should not run"}}},
				Operators: []string{"||"},
			},
			expectError: false,
			description: "Second command should not run if first succeeds with ||",
		},
		{
			name: "OR chain with failure",
			chain: CommandChain{
				Commands:  []Command{{Name: "nonexistent-command"}, {Name: "echo", Args: []string{"backup"}}},
				Operators: []string{"||"},
			},
			expectError: false,
			description: "Second command should run if first fails with ||",
		},
		{
			name: "pipe chain",
			chain: CommandChain{
				Commands:  []Command{{Name: "echo", Args: []string{"hello"}}, {Name: "echo", Args: []string{"world"}}},
				Operators: []string{"|"},
			},
			expectError: false,
			description: "Pipe should work as sequential execution for now",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := shell.executeCommandChain(tt.chain)
			if tt.expectError && err == nil {
				t.Errorf("executeCommandChain() expected error but got none: %s", tt.description)
			}
			if !tt.expectError && err != nil {
				t.Errorf("executeCommandChain() unexpected error = %v: %s", err, tt.description)
			}
		})
	}
}

func TestMVXShell_Execute(t *testing.T) {
	tempDir := t.TempDir()
	shell := NewMVXShell(tempDir, os.Environ())

	tests := []struct {
		name        string
		script      string
		expectError bool
		description string
	}{
		{
			name:        "simple command",
			script:      "echo hello",
			expectError: false,
			description: "Simple echo should work",
		},
		{
			name:        "AND chain success",
			script:      "mkdir testdir && cd testdir",
			expectError: false,
			description: "Create directory and change into it",
		},
		{
			name:        "AND chain with failure",
			script:      "nonexistent-command && echo should not run",
			expectError: true,
			description: "Second command should not run if first fails",
		},
		{
			name:        "OR chain with success",
			script:      "echo success || echo should not run",
			expectError: false,
			description: "Second command should not run if first succeeds",
		},
		{
			name:        "OR chain with failure",
			script:      "nonexistent-command || echo backup executed",
			expectError: false,
			description: "Second command should run if first fails",
		},
		{
			name:        "semicolon separation",
			script:      "echo first; echo second",
			expectError: false,
			description: "Both commands should run regardless of first result",
		},
		{
			name:        "complex chain",
			script:      "mkdir complex && cd complex || echo failed to setup",
			expectError: false,
			description: "Complex chain should work",
		},
		{
			name:        "pipe chain",
			script:      "echo hello | echo world",
			expectError: false,
			description: "Pipe should work as sequential execution",
		},
		{
			name:        "parentheses",
			script:      "echo grouped && echo after",
			expectError: false,
			description: "Commands should work without parentheses for now",
		},
		{
			name:        "quoted arguments",
			script:      "echo 'hello world' && echo \"test string\"",
			expectError: false,
			description: "Quoted arguments should be preserved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset to temp directory for each test
			shell.workDir = tempDir

			err := shell.Execute(tt.script)
			if tt.expectError && err == nil {
				t.Errorf("Execute() expected error but got none: %s", tt.description)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Execute() unexpected error = %v: %s", err, tt.description)
			}
		})
	}

	// Specific test for directory change
	shell.workDir = tempDir
	err := shell.Execute("mkdir testdir && cd testdir")
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	// Check if directory was created and we changed into it
	expectedPath := filepath.Join(tempDir, "testdir")
	if shell.workDir != expectedPath {
		t.Errorf("Execute() workDir = %v, expected %v", shell.workDir, expectedPath)
	}
}

func TestTokenizeErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		script      string
		expectError bool
		description string
	}{
		{
			name:        "unterminated single quote",
			script:      "echo 'hello",
			expectError: true,
			description: "Should fail with unterminated quote",
		},
		{
			name:        "unterminated double quote",
			script:      "echo \"hello",
			expectError: true,
			description: "Should fail with unterminated quote",
		},
		{
			name:        "empty script",
			script:      "",
			expectError: false,
			description: "Empty script should be valid",
		},
		{
			name:        "only whitespace",
			script:      "   \t\n  ",
			expectError: false,
			description: "Whitespace-only script should be valid",
		},
		{
			name:        "only operators",
			script:      "&&",
			expectError: false,
			description: "Tokenize should succeed, parsing will catch the error",
		},
		{
			name:        "trailing operator",
			script:      "echo hello &&",
			expectError: false,
			description: "Trailing operator should be handled gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tokenize(tt.script)
			if tt.expectError && err == nil {
				t.Errorf("tokenize() expected error but got none: %s", tt.description)
			}
			if !tt.expectError && err != nil {
				t.Errorf("tokenize() unexpected error = %v: %s", err, tt.description)
			}
		})
	}
}

func TestParseCommandsErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		script      string
		expectError bool
		description string
	}{
		{
			name:        "operator without command",
			script:      "&&",
			expectError: true,
			description: "Should fail when operator has no preceding command",
		},
		{
			name:        "pipe without command",
			script:      "|",
			expectError: true,
			description: "Should fail when pipe has no preceding command",
		},
		{
			name:        "empty command between operators",
			script:      "echo hello && && echo world",
			expectError: true,
			description: "Should fail with empty command between operators",
		},
		{
			name:        "valid complex script",
			script:      "cd test && mvn clean install || echo failed; echo done",
			expectError: false,
			description: "Complex valid script should parse successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseCommands(tt.script)
			if tt.expectError && err == nil {
				t.Errorf("parseCommands() expected error but got none: %s", tt.description)
			}
			if !tt.expectError && err != nil {
				t.Errorf("parseCommands() unexpected error = %v: %s", err, tt.description)
			}
		})
	}
}

func TestCommandChainOperatorLogic(t *testing.T) {
	tempDir := t.TempDir()
	shell := NewMVXShell(tempDir, os.Environ())

	// Test && operator - second command should not run if first fails
	t.Run("AND operator with failure", func(t *testing.T) {
		// This should fail and not run the second command
		err := shell.Execute("nonexistent-command && mkdir should-not-exist")
		if err == nil {
			t.Error("Expected error from nonexistent command")
		}

		// Directory should not exist because second command shouldn't run
		shouldNotExist := filepath.Join(tempDir, "should-not-exist")
		if _, err := os.Stat(shouldNotExist); !os.IsNotExist(err) {
			t.Error("Directory should not exist because second command should not have run")
		}
	})

	// Test || operator - second command should not run if first succeeds
	t.Run("OR operator with success", func(t *testing.T) {
		// This should succeed and not run the second command
		err := shell.Execute("echo success || echo this should not run")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		// We can't easily test that the second command didn't run without output capture
		// but the test verifies the logic doesn't error
	})

	// Test || operator - second command should run if first fails
	t.Run("OR operator with failure", func(t *testing.T) {
		testDir := filepath.Join(tempDir, "or_test_dir")

		// First command fails, second should run and create directory
		err := shell.Execute("nonexistent-command || mkdir " + testDir)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Directory should exist because second command should have run
		if _, err := os.Stat(testDir); os.IsNotExist(err) {
			t.Error("Directory should exist because second command should have run")
		}

		// Clean up
		os.Remove(testDir)
	})

	// Test semicolon - both commands should run regardless
	t.Run("semicolon separator", func(t *testing.T) {
		testDir := filepath.Join(tempDir, "semi_test_dir")

		// First command fails, but second should still run
		// We expect an error from the first command, but the second should still execute
		err := shell.Execute("nonexistent-command; mkdir " + testDir)
		if err == nil {
			t.Error("Expected error from nonexistent command")
		}

		// Directory should exist because second command should have run
		if _, err := os.Stat(testDir); os.IsNotExist(err) {
			t.Error("Directory should exist because second command should have run after semicolon")
		}

		// Clean up
		os.Remove(testDir)
	})
}
