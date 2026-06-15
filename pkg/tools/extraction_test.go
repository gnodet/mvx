package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectArchiveType(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{"tar.gz extension", "archive.tar.gz", "tar.gz"},
		{"tgz extension", "archive.tgz", "tar.gz"},
		{"tar.xz extension", "archive.tar.xz", "tar.xz"},
		{"zip extension", "archive.zip", "zip"},
		{"gz extension", "archive.gz", "tar.gz"},
		{"uppercase ZIP", "Archive.ZIP", "zip"},
		{"no extension", "archive", ""},
		{"tmp extension", "mvx-download-12345.tmp", ""},
		{"random extension", "file.dat", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := detectArchiveType(tc.filename)
			if result != tc.expected {
				t.Errorf("detectArchiveType(%q) = %q, want %q", tc.filename, result, tc.expected)
			}
		})
	}
}

func TestDetectArchiveTypeFromContent(t *testing.T) {
	tests := []struct {
		name     string
		header   []byte
		expected string
	}{
		{"ZIP magic bytes", []byte{0x50, 0x4b, 0x03, 0x04, 0x00, 0x00}, "zip"},
		{"GZIP magic bytes", []byte{0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00}, "tar.gz"},
		{"XZ magic bytes", []byte{0xfd, 0x37, 0x7a, 0x58, 0x5a, 0x00}, "tar.xz"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.tmp")
			if err := os.WriteFile(tmpFile, tc.header, 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			result, err := detectArchiveTypeFromContent(tmpFile)
			if err != nil {
				t.Fatalf("detectArchiveTypeFromContent() returned error: %v", err)
			}
			if result != tc.expected {
				t.Errorf("detectArchiveTypeFromContent() = %q, want %q", result, tc.expected)
			}
		})
	}
}

func TestDetectArchiveTypeFromContentUnrecognized(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.tmp")
	if err := os.WriteFile(tmpFile, []byte("not an archive"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := detectArchiveTypeFromContent(tmpFile)
	if err == nil {
		t.Error("Expected error for unrecognized content, got nil")
	}
}

func TestExtractArchiveFallsBackToContentDetection(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a minimal ZIP file with a .tmp extension (simulating a temp download)
	// ZIP files start with PK (0x50 0x4b) magic bytes
	// This is a minimal valid empty ZIP (end-of-central-directory record)
	emptyZip := []byte{
		0x50, 0x4b, 0x05, 0x06, // End of central directory signature
		0x00, 0x00, // Number of this disk
		0x00, 0x00, // Disk where central directory starts
		0x00, 0x00, // Number of central directory records on this disk
		0x00, 0x00, // Total number of central directory records
		0x00, 0x00, 0x00, 0x00, // Size of central directory
		0x00, 0x00, 0x00, 0x00, // Offset of start of central directory
		0x00, 0x00, // Comment length
	}

	tmpFile := filepath.Join(tmpDir, "java-download-12345.tmp")
	if err := os.WriteFile(tmpFile, emptyZip, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	destDir := filepath.Join(tmpDir, "extracted")

	// This should detect ZIP from magic bytes and extract successfully
	err := ExtractArchive(tmpFile, destDir)
	if err != nil {
		t.Fatalf("ExtractArchive() with .tmp ZIP file failed: %v", err)
	}

	// Verify destination directory was created
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		t.Error("Destination directory was not created")
	}
}
