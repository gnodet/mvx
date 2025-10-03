package tools

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/gnodet/mvx/pkg/config"
)

func TestChecksumVerifier_VerifyFile(t *testing.T) {
	// Create a temporary file with known content
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "Hello, mvx checksum verification!"

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Calculate expected checksum
	hasher := sha256.New()
	hasher.Write([]byte(testContent))
	expectedChecksum := hex.EncodeToString(hasher.Sum(nil))

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	verifier := NewChecksumVerifier(manager)

	tests := []struct {
		name        string
		checksum    ChecksumInfo
		expectError bool
	}{
		{
			name: "valid checksum",
			checksum: ChecksumInfo{
				Type:  SHA256,
				Value: expectedChecksum,
			},
			expectError: false,
		},
		{
			name: "invalid checksum",
			checksum: ChecksumInfo{
				Type:  SHA256,
				Value: "invalid_checksum",
			},
			expectError: true,
		},
		{
			name: "empty checksum",
			checksum: ChecksumInfo{
				Type:  SHA256,
				Value: "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := verifier.VerifyFile(testFile, tt.checksum)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestChecksumVerifier_calculateChecksum(t *testing.T) {
	// Create a temporary file with known content
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "Test content for checksum calculation"

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	verifier := NewChecksumVerifier(manager)

	// Calculate checksum
	checksum, err := verifier.calculateChecksum(testFile, SHA256)
	if err != nil {
		t.Fatalf("Failed to calculate checksum: %v", err)
	}

	// Verify the checksum is correct
	hasher := sha256.New()
	hasher.Write([]byte(testContent))
	expectedChecksum := hex.EncodeToString(hasher.Sum(nil))

	if checksum != expectedChecksum {
		t.Errorf("Checksum mismatch: expected %s, got %s", expectedChecksum, checksum)
	}
}

func TestChecksumVerifier_parseChecksumFile(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	verifier := NewChecksumVerifier(manager)

	tests := []struct {
		name        string
		content     string
		filename    string
		expected    string
		expectError bool
	}{
		{
			name:     "standard format",
			content:  "abc123def456  test.zip\n789ghi012jkl  other.zip\n",
			filename: "test.zip",
			expected: "abc123def456",
		},
		{
			name:     "binary mode format",
			content:  "abc123def456 *test.zip\n789ghi012jkl *other.zip\n",
			filename: "test.zip",
			expected: "abc123def456",
		},
		{
			name:        "file not found",
			content:     "abc123def456  test.zip\n789ghi012jkl  other.zip\n",
			filename:    "missing.zip",
			expectError: true,
		},
		{
			name:     "multiple spaces",
			content:  "abc123def456    test.zip\n",
			filename: "test.zip",
			expected: "abc123def456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := verifier.parseChecksumFile(tt.content, tt.filename)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestJavaToolChecksum(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	javaTool := NewJavaTool(manager)

	// Test the GetChecksum method (should return error since Java checksums are handled during installation)
	cfg := config.ToolConfig{Distribution: "temurin"}
	checksum, err := javaTool.GetChecksum("21", cfg, "test-file.tar.gz")
	if err == nil {
		t.Errorf("Expected error for Java checksum since it's handled during installation, but got checksum: %v", checksum)
	}
	t.Logf("Java checksum correctly returns error (handled during installation): %v", err)
}

func TestNodeToolChecksum(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	nodeTool := NewNodeTool(manager)

	// Test with a known Node.js version (this is a real API call)
	nodeCfg := config.ToolConfig{}
	checksum, err := nodeTool.GetChecksum("22.14.0", nodeCfg, "node-v22.14.0-linux-x64.tar.xz")
	if err != nil {
		t.Logf("Node.js checksum fetch failed (expected in CI): %v", err)
		return // Skip test if API is not accessible
	}

	if checksum.Type != SHA256 {
		t.Errorf("Expected SHA256 checksum type, got %s", checksum.Type)
	}
	if checksum.Value == "" {
		t.Errorf("Expected non-empty checksum value")
	}
	// Verify the checksum looks like a valid SHA256 (64 hex characters)
	if len(checksum.Value) != 64 {
		t.Errorf("Expected 64-character SHA256 checksum, got %d characters: %s", len(checksum.Value), checksum.Value)
	}
	t.Logf("Node.js 22.14.0 checksum value: %s", checksum.Value)
}
