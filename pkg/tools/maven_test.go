package tools

import (
	"os"
	"testing"

	"github.com/gnodet/mvx/pkg/config"
)

func TestUseSystemMaven(t *testing.T) {
	// Test when MVX_USE_SYSTEM_MAVEN is not set
	os.Unsetenv("MVX_USE_SYSTEM_MAVEN")
	if useSystemMaven() {
		t.Error("useSystemMaven() should return false when MVX_USE_SYSTEM_MAVEN is not set")
	}

	// Test when MVX_USE_SYSTEM_MAVEN is set to false
	os.Setenv("MVX_USE_SYSTEM_MAVEN", "false")
	if useSystemMaven() {
		t.Error("useSystemMaven() should return false when MVX_USE_SYSTEM_MAVEN=false")
	}

	// Test when MVX_USE_SYSTEM_MAVEN is set to true
	os.Setenv("MVX_USE_SYSTEM_MAVEN", "true")
	if !useSystemMaven() {
		t.Error("useSystemMaven() should return true when MVX_USE_SYSTEM_MAVEN=true")
	}

	// Clean up
	os.Unsetenv("MVX_USE_SYSTEM_MAVEN")
}

func TestMavenSystemDetector(t *testing.T) {
	detector := &MavenSystemDetector{}

	// Test GetSystemHome when no Maven environment variables are set
	os.Unsetenv("MAVEN_HOME")
	os.Unsetenv("M2_HOME")

	_, err := detector.GetSystemHome()
	// This might succeed if mvn is in PATH, so we don't assert failure
	if err != nil {
		t.Logf("Maven not found in system (expected): %v", err)
	}
}

func TestMavenToolWithSystemMaven(t *testing.T) {
	// Save original environment variables
	originalUseSystemMaven := os.Getenv("MVX_USE_SYSTEM_MAVEN")
	originalMavenHome := os.Getenv("MAVEN_HOME")
	originalM2Home := os.Getenv("M2_HOME")
	defer func() {
		if originalUseSystemMaven != "" {
			os.Setenv("MVX_USE_SYSTEM_MAVEN", originalUseSystemMaven)
		} else {
			os.Unsetenv("MVX_USE_SYSTEM_MAVEN")
		}
		if originalMavenHome != "" {
			os.Setenv("MAVEN_HOME", originalMavenHome)
		} else {
			os.Unsetenv("MAVEN_HOME")
		}
		if originalM2Home != "" {
			os.Setenv("M2_HOME", originalM2Home)
		} else {
			os.Unsetenv("M2_HOME")
		}
	}()

	// Create a mock manager
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	mavenTool := &MavenTool{manager: manager}

	// Test with MVX_USE_SYSTEM_MAVEN=false (default behavior)
	os.Unsetenv("MVX_USE_SYSTEM_MAVEN")
	os.Unsetenv("MAVEN_HOME")
	os.Unsetenv("M2_HOME")

	cfg := config.ToolConfig{
		Version: "3.9.6",
	}

	// IsInstalled should return false when no Maven is installed and MVX_USE_SYSTEM_MAVEN is not set
	if mavenTool.IsInstalled("3.9.6", cfg) {
		t.Error("IsInstalled should return false when no Maven is installed")
	}

	// Test with MVX_USE_SYSTEM_MAVEN=true but no MAVEN_HOME or M2_HOME
	os.Setenv("MVX_USE_SYSTEM_MAVEN", "true")
	os.Unsetenv("MAVEN_HOME")
	os.Unsetenv("M2_HOME")

	// IsInstalled should return false when Maven environment variables are not set
	// (unless mvn is found in PATH with a compatible version)
	if mavenTool.IsInstalled("3.9.6", cfg) {
		t.Log("Maven found in PATH with compatible version")
	}
}
