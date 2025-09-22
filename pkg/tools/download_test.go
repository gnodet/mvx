package tools

import (
	"os"
	"testing"
)

func TestValidateDmg(t *testing.T) {
	tests := []struct {
		name     string
		header   []byte
		expected bool
	}{
		{
			name:     "Valid DMG with zlib compression (78 01)",
			header:   []byte{0x78, 0x01, 0x00, 0x00},
			expected: true,
		},
		{
			name:     "Valid DMG with zlib compression (78 9c)",
			header:   []byte{0x78, 0x9c, 0x00, 0x00},
			expected: true,
		},
		{
			name:     "Valid DMG with zlib compression (78 da)",
			header:   []byte{0x78, 0xda, 0x00, 0x00},
			expected: true,
		},
		{
			name:     "Valid DMG with koly signature",
			header:   []byte("koly\x00\x00\x00\x00"),
			expected: true,
		},
		{
			name:     "Valid DMG with koly signature in middle",
			header:   []byte("\x00\x00koly\x00\x00"),
			expected: true,
		},
		{
			name:     "Unknown DMG format (should pass)",
			header:   []byte{0x00, 0x01, 0x02, 0x03},
			expected: true,
		},
		{
			name:     "Too short header",
			header:   []byte{0x78},
			expected: false,
		},
		{
			name:     "Empty header",
			header:   []byte{},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateDmg(test.header)
			hasError := err != nil
			
			if test.expected && hasError {
				t.Errorf("validateDmg() failed for %s: %v", test.name, err)
			} else if !test.expected && !hasError {
				t.Errorf("validateDmg() should have failed for %s", test.name)
			}
		})
	}
}

func TestValidateFileFormat(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		header   []byte
		expected bool
	}{
		{
			name:     "DMG file with .dmg extension",
			url:      "https://example.com/jdk-21.dmg",
			header:   []byte{0x78, 0x01, 0x00, 0x00},
			expected: true,
		},
		{
			name:     "DMG file with koly signature",
			url:      "https://example.com/jdk-21.dmg",
			header:   []byte("koly\x00\x00\x00\x00"),
			expected: true,
		},
		{
			name:     "tar.gz file",
			url:      "https://example.com/jdk-21.tar.gz",
			header:   []byte{0x1f, 0x8b, 0x08, 0x00},
			expected: true,
		},
		{
			name:     "ZIP file",
			url:      "https://example.com/jdk-21.zip",
			header:   []byte{0x50, 0x4b, 0x03, 0x04},
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create a temporary file with the test header
			tmpFile := t.TempDir() + "/test-file"
			if err := writeTestFile(tmpFile, test.header); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			err := validateFileFormat(tmpFile, test.url)
			hasError := err != nil
			
			if test.expected && hasError {
				t.Errorf("validateFileFormat() failed for %s: %v", test.name, err)
			} else if !test.expected && !hasError {
				t.Errorf("validateFileFormat() should have failed for %s", test.name)
			}
		})
	}
}

// Helper function to write test files
func writeTestFile(path string, content []byte) error {
	return os.WriteFile(path, content, 0644)
}
