package tools

import (
	"runtime"
	"testing"
)

func TestNodeToolGetDownloadOptions(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	nodeTool := NewNodeTool(manager)
	options := nodeTool.GetDownloadOptions()

	// Test that download options are returned (FileExtension is used for temp file naming)
	if options.FileExtension == "" {
		t.Errorf("Expected FileExtension to be non-empty")
	}
}

func TestNodeToolDownloadURLPlatformSpecific(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	nodeTool := NewNodeTool(manager)

	// Get download URL for a test version
	url := nodeTool.getDownloadURL("20.19.5")

	// Verify platform-specific URL behavior
	if runtime.GOOS == "windows" {
		if !endsWith(url, ExtZip) {
			t.Errorf("Expected Windows URL to end with %s, got %s", ExtZip, url)
		}
	} else {
		if !endsWith(url, ExtTarGz) {
			t.Errorf("Expected non-Windows URL to end with %s, got %s", ExtTarGz, url)
		}
	}
}

func TestNodeToolAutomaticArchiveDetection(t *testing.T) {
	// Test that automatic archive detection works for different file types
	testCases := []struct {
		filename     string
		expectedType string
	}{
		{"node-v20.19.5-win-x64.zip", "zip"},
		{"node-v20.19.5-linux-x64.tar.gz", "tar.gz"},
		{"node-v20.19.5-darwin-arm64.tar.gz", "tar.gz"},
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			// Test that detectArchiveType works correctly with the filename directly
			detectedType := detectArchiveType(tc.filename)
			if detectedType != tc.expectedType {
				t.Errorf("Expected archive type %s for %s, got %s", tc.expectedType, tc.filename, detectedType)
			}
		})
	}
}

// Helper function to check if a string ends with a suffix
func endsWith(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}
