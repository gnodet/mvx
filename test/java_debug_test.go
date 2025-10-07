package test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestJavaInstallationDebug is a focused test specifically for debugging Java 17 installation failures in CI
func TestJavaInstallationDebug(t *testing.T) {
	t.Logf("========== JAVA 17 INSTALLATION DEBUG TEST ==========")

	// Immediate panic recovery
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("PANIC in Java debug test: %v", r)
		}
	}()

	// Step 1: Environment Analysis
	t.Logf("STEP 1: ENVIRONMENT ANALYSIS")
	t.Logf("GOOS: %s, GOARCH: %s", runtime.GOOS, runtime.GOARCH)
	t.Logf("NumCPU: %d, NumGoroutine: %d", runtime.NumCPU(), runtime.NumGoroutine())

	// Check if running in CI
	ciEnvVars := []string{"CI", "GITHUB_ACTIONS", "GITLAB_CI", "TRAVIS", "CIRCLECI", "JENKINS_URL"}
	isCI := false
	for _, envVar := range ciEnvVars {
		if value := os.Getenv(envVar); value != "" {
			t.Logf("CI DETECTED - %s=%s", envVar, value)
			isCI = true
		}
	}
	if !isCI {
		t.Logf("NOT RUNNING IN CI")
	}

	// Step 2: Build MVX Binary
	t.Logf("STEP 2: BUILDING MVX BINARY")
	mvxBinary := buildMVXBinary(t)
	t.Logf("MVX binary built: %s", mvxBinary)

	// Step 3: Create Test Environment
	t.Logf("STEP 3: CREATING TEST ENVIRONMENT")
	tempDir, oldDir := createTestEnvironment(t)
	defer cleanupTestEnvironment(t, tempDir, oldDir)

	// Step 4: Initialize Project
	t.Logf("STEP 4: INITIALIZING PROJECT")
	initializeProject(t, mvxBinary, tempDir)

	// Step 5: Add Java Tool
	t.Logf("STEP 5: ADDING JAVA TOOL")
	addJavaTool(t, mvxBinary)

	// Step 6: Install Java (The Critical Step)
	t.Logf("STEP 6: INSTALLING JAVA (CRITICAL STEP)")
	installJava(t, mvxBinary)

	// Step 7: Verify Installation
	t.Logf("STEP 7: VERIFYING INSTALLATION")
	verifyJavaInstallation(t)

	t.Logf("========== JAVA 17 DEBUG TEST COMPLETED SUCCESSFULLY ==========")
}

func buildMVXBinary(t *testing.T) string {
	t.Logf("Building MVX binary...")

	// Get workspace root
	workspaceRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Go up one level if we're in test directory
	if filepath.Base(workspaceRoot) == "test" {
		workspaceRoot = filepath.Dir(workspaceRoot)
	}

	t.Logf("Workspace root: %s", workspaceRoot)

	// Build the binary
	mvxBinary := filepath.Join(workspaceRoot, "mvx-debug")
	cmd := exec.Command("go", "build", "-o", mvxBinary, ".")
	cmd.Dir = workspaceRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build MVX binary: %v\nOutput: %s", err, output)
	}

	t.Logf("MVX binary built successfully")
	return mvxBinary
}

func createTestEnvironment(t *testing.T) (string, string) {
	t.Logf("Creating test environment...")

	tempDir, err := os.MkdirTemp("", "mvx-java-debug-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	t.Logf("Created temp directory: %s", tempDir)

	// Change to temp directory
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	t.Logf("Changed to temp directory, old dir: %s", oldDir)
	return tempDir, oldDir
}

func cleanupTestEnvironment(t *testing.T, tempDir, oldDir string) {
	t.Logf("Cleaning up test environment: %s", tempDir)

	// Change back to original directory before removing temp dir
	if oldDir != "" {
		if err := os.Chdir(oldDir); err != nil {
			t.Logf("Warning: failed to change back to original directory: %v", err)
		}
	}

	if err := os.RemoveAll(tempDir); err != nil {
		t.Logf("Warning: failed to clean up temp directory: %v", err)
	}
}

func initializeProject(t *testing.T, mvxBinary, tempDir string) {
	t.Logf("Initializing project with mvx init...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, mvxBinary, "init", "--format", "json5")

	startTime := time.Now()
	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)

	t.Logf("mvx init completed in %v", duration)
	t.Logf("mvx init output (%d bytes): %s", len(output), string(output))

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			t.Fatalf("mvx init timed out after 30 seconds")
		}
		t.Fatalf("mvx init failed: %v", err)
	}

	// Verify config file was created
	if _, err := os.Stat(".mvx/config.json5"); err != nil {
		t.Fatalf("Config file not created: %v", err)
	}

	t.Logf("Project initialized successfully")
}

func addJavaTool(t *testing.T, mvxBinary string) {
	t.Logf("Adding Java 17 tool...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, mvxBinary, "tools", "add", "java", "17", "zulu")

	startTime := time.Now()
	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)

	t.Logf("tools add completed in %v", duration)
	t.Logf("tools add output (%d bytes): %s", len(output), string(output))

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			t.Fatalf("tools add timed out after 30 seconds")
		}
		t.Fatalf("tools add failed: %v", err)
	}

	t.Logf("Java tool added successfully")
}

func installJava(t *testing.T, mvxBinary string) {
	t.Logf("Installing Java with setup command...")

	// Pre-installation system check
	t.Logf("PRE-INSTALLATION SYSTEM CHECK:")
	logSystemState(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, mvxBinary, "setup", "--tools-only", "--sequential")
	cmd.Env = append(os.Environ(), "MVX_VERBOSE=true", "MVX_DEBUG=true")

	t.Logf("Starting setup command...")
	startTime := time.Now()
	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)

	t.Logf("Setup command completed in %v", duration)
	t.Logf("Setup output (%d bytes):", len(output))
	t.Logf("=== SETUP OUTPUT START ===")
	t.Logf("%s", string(output))
	t.Logf("=== SETUP OUTPUT END ===")

	if err != nil {
		t.Logf("SETUP COMMAND FAILED!")
		if ctx.Err() == context.DeadlineExceeded {
			t.Fatalf("Setup timed out after 10 minutes")
		}

		if exitError, ok := err.(*exec.ExitError); ok {
			t.Logf("Exit code: %d", exitError.ExitCode())
			if exitError.ProcessState != nil {
				t.Logf("Process state: %v", exitError.ProcessState)
			}
		}

		// Analyze output for specific errors
		analyzeSetupOutput(t, string(output))

		t.Fatalf("Setup failed: %v", err)
	}

	t.Logf("Setup command succeeded")

	// Post-installation system check
	t.Logf("POST-INSTALLATION SYSTEM CHECK:")
	logSystemState(t)
}

func logSystemState(t *testing.T) {
	// Home directory
	if home, err := os.UserHomeDir(); err == nil {
		t.Logf("Home directory: %s", home)

		// Check .mvx directory
		mvxDir := filepath.Join(home, ".mvx")
		if info, err := os.Stat(mvxDir); err == nil {
			t.Logf(".mvx directory exists (size: %d)", info.Size())

			// Check tools directory
			toolsDir := filepath.Join(mvxDir, "tools")
			if entries, err := os.ReadDir(toolsDir); err == nil {
				t.Logf("Tools directory has %d entries", len(entries))
				for _, entry := range entries {
					t.Logf("  - %s (isDir: %v)", entry.Name(), entry.IsDir())
				}
			} else {
				t.Logf("Cannot read tools directory: %v", err)
			}
		} else {
			t.Logf(".mvx directory does not exist: %v", err)
		}
	}

	// Disk space
	if usage, err := getDiskUsage("."); err == nil {
		t.Logf("Disk usage - Total: %d MB, Free: %d MB",
			usage.Total/1024/1024, usage.Free/1024/1024)
	}

	// Process info
	t.Logf("Process ID: %d, UID: %d, GID: %d", os.Getpid(), os.Getuid(), os.Getgid())
}

func analyzeSetupOutput(t *testing.T, output string) {
	t.Logf("ANALYZING SETUP OUTPUT FOR ERRORS:")

	errorPatterns := map[string][]string{
		"Permission": {"permission denied", "access denied", "forbidden"},
		"Network":    {"network", "connection", "timeout", "dns", "unreachable"},
		"NotFound":   {"404", "not found", "no such file"},
		"DiskSpace":  {"no space left", "disk full"},
		"Java":       {"java", "jdk", "openjdk", "zulu"},
	}

	for category, patterns := range errorPatterns {
		for _, pattern := range patterns {
			if strings.Contains(strings.ToLower(output), pattern) {
				t.Logf("ERROR PATTERN DETECTED [%s]: %s", category, pattern)
			}
		}
	}

	if len(output) == 0 {
		t.Logf("CRITICAL: Setup produced no output")
	} else if len(output) < 100 {
		t.Logf("WARNING: Setup produced very short output (%d bytes)", len(output))
	}
}

func verifyJavaInstallation(t *testing.T) {
	t.Logf("Verifying Java installation...")

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Cannot get home directory: %v", err)
	}

	javaDir := filepath.Join(home, ".mvx", "tools", "java")
	if entries, err := os.ReadDir(javaDir); err == nil {
		t.Logf("Found %d Java installations", len(entries))
		for _, entry := range entries {
			t.Logf("  - %s", entry.Name())
		}
	} else {
		t.Fatalf("Java tools directory not found: %v", err)
	}

	t.Logf("Java installation verified")
}

// Use the DiskUsage type and function from integration_test.go to avoid redeclaration
