package tools

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gnodet/mvx/pkg/config"
)

func TestUseSystemTool(t *testing.T) {
	// Test when MVX_USE_SYSTEM_JAVA is not set
	os.Unsetenv("MVX_USE_SYSTEM_JAVA")
	if useSystemTool("java") {
		t.Error("useSystemTool('java') should return false when MVX_USE_SYSTEM_JAVA is not set")
	}

	// Test when MVX_USE_SYSTEM_JAVA is set to false
	os.Setenv("MVX_USE_SYSTEM_JAVA", "false")
	if useSystemTool("java") {
		t.Error("useSystemTool('java') should return false when MVX_USE_SYSTEM_JAVA=false")
	}

	// Test when MVX_USE_SYSTEM_JAVA is set to true
	os.Setenv("MVX_USE_SYSTEM_JAVA", "true")
	if !useSystemTool("java") {
		t.Error("useSystemTool('java') should return true when MVX_USE_SYSTEM_JAVA=true")
	}

	// Test Maven tool
	os.Unsetenv("MVX_USE_SYSTEM_MAVEN")
	if useSystemTool("maven") {
		t.Error("useSystemTool('maven') should return false when MVX_USE_SYSTEM_MAVEN is not set")
	}

	os.Setenv("MVX_USE_SYSTEM_MAVEN", "true")
	if !useSystemTool("maven") {
		t.Error("useSystemTool('maven') should return true when MVX_USE_SYSTEM_MAVEN=true")
	}

	// Clean up
	os.Unsetenv("MVX_USE_SYSTEM_JAVA")
	os.Unsetenv("MVX_USE_SYSTEM_MAVEN")
}

func TestUseSystemJava(t *testing.T) {
	// Test when MVX_USE_SYSTEM_JAVA is not set
	os.Unsetenv("MVX_USE_SYSTEM_JAVA")
	if useSystemJava() {
		t.Error("useSystemJava() should return false when MVX_USE_SYSTEM_JAVA is not set")
	}

	// Test when MVX_USE_SYSTEM_JAVA is set to false
	os.Setenv("MVX_USE_SYSTEM_JAVA", "false")
	if useSystemJava() {
		t.Error("useSystemJava() should return false when MVX_USE_SYSTEM_JAVA=false")
	}

	// Test when MVX_USE_SYSTEM_JAVA is set to true
	os.Setenv("MVX_USE_SYSTEM_JAVA", "true")
	if !useSystemJava() {
		t.Error("useSystemJava() should return true when MVX_USE_SYSTEM_JAVA=true")
	}

	// Clean up
	os.Unsetenv("MVX_USE_SYSTEM_JAVA")
}

func TestJavaSystemDetector(t *testing.T) {
	detector := &JavaSystemDetector{}

	// Save original JAVA_HOME
	originalJavaHome := os.Getenv("JAVA_HOME")
	defer func() {
		if originalJavaHome != "" {
			os.Setenv("JAVA_HOME", originalJavaHome)
		} else {
			os.Unsetenv("JAVA_HOME")
		}
	}()

	// Test when JAVA_HOME is not set
	os.Unsetenv("JAVA_HOME")
	_, err := detector.GetSystemHome()
	if err == nil {
		t.Error("GetSystemHome() should return error when JAVA_HOME is not set")
	}

	// Test when JAVA_HOME points to non-existent directory
	os.Setenv("JAVA_HOME", "/non/existent/path")
	_, err = detector.GetSystemHome()
	if err == nil {
		t.Error("GetSystemHome() should return error when JAVA_HOME points to non-existent directory")
	}
}

func TestJavaSystemDetectorVersion(t *testing.T) {
	detector := &JavaSystemDetector{}

	// This test requires a real Java installation, so we'll skip it if JAVA_HOME is not set
	javaHome := os.Getenv("JAVA_HOME")
	if javaHome == "" {
		t.Skip("Skipping test because JAVA_HOME is not set")
	}

	// Check if Java executable exists
	javaExe := filepath.Join(javaHome, "bin", "java")
	if runtime.GOOS == "windows" {
		javaExe += ".exe"
	}

	if _, err := os.Stat(javaExe); err != nil {
		t.Skip("Skipping test because Java executable not found at JAVA_HOME")
	}

	version, err := detector.GetSystemVersion(javaHome)
	if err != nil {
		t.Errorf("GetSystemVersion() failed: %v", err)
	}

	if version == "" {
		t.Error("GetSystemVersion() returned empty version")
	}

	t.Logf("Detected Java version: %s", version)
}

func TestJavaVersionCompatibility(t *testing.T) {
	detector := &JavaSystemDetector{}

	tests := []struct {
		systemVersion    string
		requestedVersion string
		expected         bool
	}{
		{"21", "21", true},
		{"17", "17", true},
		{"11", "11", true},
		{"8", "8", true},
		{"21", "17", false},
		{"17", "21", false},
		{"11", "8", false},
	}

	for _, test := range tests {
		result := detector.IsVersionCompatible(test.systemVersion, test.requestedVersion)
		if result != test.expected {
			t.Errorf("IsVersionCompatible(%s, %s) = %v, expected %v",
				test.systemVersion, test.requestedVersion, result, test.expected)
		}
	}
}

func TestJavaToolWithSystemJava(t *testing.T) {
	// Save original environment variables
	originalUseSystemJava := os.Getenv("MVX_USE_SYSTEM_JAVA")
	originalJavaHome := os.Getenv("JAVA_HOME")
	defer func() {
		if originalUseSystemJava != "" {
			os.Setenv("MVX_USE_SYSTEM_JAVA", originalUseSystemJava)
		} else {
			os.Unsetenv("MVX_USE_SYSTEM_JAVA")
		}
		if originalJavaHome != "" {
			os.Setenv("JAVA_HOME", originalJavaHome)
		} else {
			os.Unsetenv("JAVA_HOME")
		}
	}()

	// Create a mock manager
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	javaTool := &JavaTool{manager: manager}

	// Test with MVX_USE_SYSTEM_JAVA=false (default behavior)
	os.Unsetenv("MVX_USE_SYSTEM_JAVA")
	os.Unsetenv("JAVA_HOME")

	cfg := config.ToolConfig{
		Version:      "11", // Use version 11 which won't be installed
		Distribution: "temurin",
	}

	// IsInstalled should return false when no Java is installed and MVX_USE_SYSTEM_JAVA is not set
	if javaTool.IsInstalled("11", cfg) {
		t.Error("IsInstalled should return false when no Java is installed")
	}

	// Test with MVX_USE_SYSTEM_JAVA=true but no JAVA_HOME
	os.Setenv("MVX_USE_SYSTEM_JAVA", "true")
	os.Unsetenv("JAVA_HOME")

	// IsInstalled should still return false when JAVA_HOME is not set
	if javaTool.IsInstalled("11", cfg) {
		t.Error("IsInstalled should return false when MVX_USE_SYSTEM_JAVA=true but JAVA_HOME is not set")
	}

	// Test with MVX_USE_SYSTEM_JAVA=true and JAVA_HOME set but version mismatch
	os.Setenv("MVX_USE_SYSTEM_JAVA", "true")
	os.Setenv("JAVA_HOME", "/usr/lib/jvm/java-21-openjdk-amd64")

	// IsInstalled should return false when system Java version doesn't match
	if javaTool.IsInstalled("11", cfg) {
		t.Error("IsInstalled should return false when system Java version doesn't match requested version")
	}
}

func TestJavaGetFileExtension(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	javaTool := &JavaTool{manager: manager}

	tests := []struct {
		url      string
		expected string
	}{
		{"https://example.com/jdk-21.tar.gz", ".tar.gz"},
		{"https://example.com/jdk-21.tgz", ".tgz"},
		{"https://example.com/jdk-21.dmg", ".dmg"},
		{"https://example.com/jdk-21.zip", ".zip"},
		{"https://example.com/jdk-21.tar.xz", ".tar.xz"},
		{"https://example.com/jdk-21", ".tar.gz"}, // default
		{"https://example.com/jdk-21.unknown", ".tar.gz"}, // default
	}

	for _, test := range tests {
		result := javaTool.getFileExtension(test.url)
		if result != test.expected {
			t.Errorf("getFileExtension(%s) = %s, expected %s", test.url, result, test.expected)
		}
	}
}

func TestJavaHasJdkStructure(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	javaTool := &JavaTool{manager: manager}

	// Create a temporary directory structure
	tmpDir := t.TempDir()

	// Test directory without JDK structure
	if javaTool.hasJdkStructure(tmpDir) {
		t.Error("hasJdkStructure should return false for directory without JDK structure")
	}

	// Create JDK structure
	jdkDir := filepath.Join(tmpDir, "Contents", "Home", "bin")
	if err := os.MkdirAll(jdkDir, 0755); err != nil {
		t.Fatalf("Failed to create JDK directory structure: %v", err)
	}

	// Create java executable
	javaExe := filepath.Join(jdkDir, "java")
	if err := os.WriteFile(javaExe, []byte("#!/bin/bash\necho java"), 0755); err != nil {
		t.Fatalf("Failed to create java executable: %v", err)
	}

	// Test directory with JDK structure
	if !javaTool.hasJdkStructure(tmpDir) {
		t.Error("hasJdkStructure should return true for directory with JDK structure")
	}
}
