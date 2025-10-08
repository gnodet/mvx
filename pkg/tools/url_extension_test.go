package tools

import (
	"testing"
)

// TestURLExtensionDetection tests the fix for the Windows archive extraction issue
// The fix ensures temporary files are created with the correct extension based on the download URL
func TestURLExtensionDetection(t *testing.T) {
	// Test the extension detection logic from Download method
	testCases := []struct {
		name        string
		downloadURL string
		expectedExt string
	}{
		{
			name:        "Windows Node.js ZIP",
			downloadURL: "https://nodejs.org/dist/v20.19.5/node-v20.19.5-win-x64.zip",
			expectedExt: ".zip",
		},
		{
			name:        "Linux Node.js tar.gz",
			downloadURL: "https://nodejs.org/dist/v20.19.5/node-v20.19.5-linux-x64.tar.gz",
			expectedExt: ".tar.gz",
		},
		{
			name:        "Windows Go ZIP",
			downloadURL: "https://go.dev/dl/go1.21.5.windows-amd64.zip",
			expectedExt: ".zip",
		},
		{
			name:        "Linux Go tar.gz",
			downloadURL: "https://go.dev/dl/go1.21.5.linux-amd64.tar.gz",
			expectedExt: ".tar.gz",
		},
		{
			name:        "tar.xz archive",
			downloadURL: "https://example.com/tool-1.0.0.tar.xz",
			expectedExt: ".tar.xz",
		},
		{
			name:        "Java Windows ZIP (GitHub redirect)",
			downloadURL: "https://github.com/adoptium/temurin22-binaries/releases/download/jdk-22.0.2%2B9/OpenJDK22U-jdk_x64_windows_hotspot_22.0.2_9.zip",
			expectedExt: ".zip",
		},
		{
			name:        "Java Linux tar.gz (GitHub redirect)",
			downloadURL: "https://github.com/adoptium/temurin22-binaries/releases/download/jdk-22.0.2%2B9/OpenJDK22U-jdk_x64_linux_hotspot_22.0.2_9.tar.gz",
			expectedExt: ".tar.gz",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Use the actual detection function from Download method
			fileExtension := detectFileExtensionFromURL(tc.downloadURL)

			if fileExtension != tc.expectedExt {
				t.Errorf("Expected extension %s for URL %s, got %s", tc.expectedExt, tc.downloadURL, fileExtension)
			}

			t.Logf("✅ %s: %s -> %s", tc.name, tc.downloadURL, fileExtension)
		})
	}
}
