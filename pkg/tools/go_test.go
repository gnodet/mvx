package tools

import (
	"runtime"
	"testing"
)

func TestGoToolGetDownloadOptions(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	goTool := NewGoTool(manager)
	options := goTool.GetDownloadOptions()

	// Test that download options are returned (FileExtension is used for temp file naming)
	if options.FileExtension == "" {
		t.Errorf("Expected FileExtension to be non-empty")
	}
}

func TestGoToolDownloadURLPlatformSpecific(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	goTool := NewGoTool(manager)

	// Get download URL for a test version
	url := goTool.getDownloadURL("1.21.5")

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

func TestGoToolAutomaticArchiveDetection(t *testing.T) {
	// Test that automatic archive detection works for different file types
	testCases := []struct {
		filename     string
		expectedType string
	}{
		{"go1.21.5.windows-amd64.zip", "zip"},
		{"go1.21.5.linux-amd64.tar.gz", "tar.gz"},
		{"go1.21.5.darwin-arm64.tar.gz", "tar.gz"},
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
