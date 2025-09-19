package tools

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
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

	verifier := NewChecksumVerifier()

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

	verifier := NewChecksumVerifier()

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
	verifier := NewChecksumVerifier()

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

func TestChecksumRegistry_GetChecksum(t *testing.T) {
	registry := NewChecksumRegistry()

	// Test known checksum
	checksum, exists := registry.GetChecksum("maven", "3.9.6", "bin")
	if !exists {
		t.Errorf("Expected to find checksum for maven 3.9.6")
	}
	if checksum.Type != SHA512 {
		t.Errorf("Expected SHA512 checksum type, got %s", checksum.Type)
	}
	if checksum.URL == "" {
		t.Errorf("Expected checksum URL to be set")
	}

	// Test unknown checksum
	_, exists = registry.GetChecksum("unknown", "1.0.0", "bin")
	if exists {
		t.Errorf("Did not expect to find checksum for unknown tool")
	}
}

func TestChecksumRegistry_SupportsChecksumVerification(t *testing.T) {
	registry := NewChecksumRegistry()

	tests := []struct {
		toolName string
		expected bool
	}{
		{"maven", true},
		{"mvnd", true},
		{"go", true},
		{"java", true}, // Java uses Adoptium API
		{"node", true}, // Node.js uses SHASUMS256.txt
		{"unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.toolName, func(t *testing.T) {
			result := registry.SupportsChecksumVerification(tt.toolName)
			if result != tt.expected {
				t.Errorf("Expected %v for %s, got %v", tt.expected, tt.toolName, result)
			}
		})
	}
}

func TestChecksumRegistry_GetJavaChecksumFromAPI(t *testing.T) {
	registry := NewChecksumRegistry()

	// Test with a known Java version (this is a real API call)
	checksum, err := registry.GetJavaChecksumFromAPI("21", "x64", "linux")
	if err != nil {
		t.Logf("Java checksum API call failed (expected in CI): %v", err)
		return // Skip test if API is not accessible
	}

	if checksum.Type != SHA256 {
		t.Errorf("Expected SHA256 checksum type, got %s", checksum.Type)
	}
	if checksum.Value == "" {
		t.Errorf("Expected non-empty checksum value")
	}
	t.Logf("Java 21 checksum: %s", checksum.Value)
}

func TestChecksumRegistry_GetNodeChecksumFromSHASUMS(t *testing.T) {
	registry := NewChecksumRegistry()

	// Test with a known Node.js version (this is a real API call)
	checksum, err := registry.GetNodeChecksumFromSHASUMS("22.14.0", "node-v22.14.0-linux-x64.tar.xz")
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
