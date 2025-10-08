package tools

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// ChecksumType represents the type of checksum algorithm
type ChecksumType string

const (
	// SHA256 represents SHA-256 checksum
	SHA256 ChecksumType = "sha256"
	// SHA512 represents SHA-512 checksum
	SHA512 ChecksumType = "sha512"
)

// ChecksumInfo contains checksum information for a file
type ChecksumInfo struct {
	Type     ChecksumType `json:"type" yaml:"type"`
	Value    string       `json:"value" yaml:"value"`
	URL      string       `json:"url,omitempty" yaml:"url,omitempty"`
	Filename string       `json:"filename,omitempty" yaml:"filename,omitempty"`
}

// ChecksumVerifier handles checksum verification for downloaded files
type ChecksumVerifier struct {
	manager *Manager
}

// NewChecksumVerifier creates a new checksum verifier
func NewChecksumVerifier(manager *Manager) *ChecksumVerifier {
	return &ChecksumVerifier{
		manager: manager,
	}
}

// VerifyFile verifies a file against the provided checksum information
func (cv *ChecksumVerifier) VerifyFile(filePath string, checksum ChecksumInfo) error {
	if checksum.Value == "" && checksum.URL == "" {
		return fmt.Errorf("no checksum value or URL provided")
	}

	// Get expected checksum value
	expectedChecksum := checksum.Value
	if expectedChecksum == "" && checksum.URL != "" {
		var err error
		expectedChecksum, err = cv.fetchChecksumFromURL(checksum.URL, checksum.Filename)
		if err != nil {
			return fmt.Errorf("failed to fetch checksum from URL: %w", err)
		}
	}

	// Calculate actual checksum
	actualChecksum, err := cv.calculateChecksum(filePath, checksum.Type)
	if err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	// Compare checksums (case-insensitive)
	if !strings.EqualFold(expectedChecksum, actualChecksum) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

// calculateChecksum calculates the checksum of a file
func (cv *ChecksumVerifier) calculateChecksum(filePath string, checksumType ChecksumType) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	switch checksumType {
	case SHA256:
		hasher := sha256.New()
		if _, err := io.Copy(hasher, file); err != nil {
			return "", fmt.Errorf("failed to read file: %w", err)
		}
		return hex.EncodeToString(hasher.Sum(nil)), nil
	case SHA512:
		hasher := sha512.New()
		if _, err := io.Copy(hasher, file); err != nil {
			return "", fmt.Errorf("failed to read file: %w", err)
		}
		return hex.EncodeToString(hasher.Sum(nil)), nil
	default:
		return "", fmt.Errorf("unsupported checksum type: %s", checksumType)
	}
}

// fetchChecksumFromURL fetches checksum from a remote URL
func (cv *ChecksumVerifier) fetchChecksumFromURL(url, filename string) (string, error) {
	resp, err := cv.manager.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch checksum URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("checksum URL returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read checksum response: %w", err)
	}

	content := string(body)

	// If filename is specified, parse the checksum file format
	if filename != "" {
		checksum, err := cv.parseChecksumFile(content, filename)
		if err != nil {
			// Add debug information for checksum parsing failures
			fmt.Printf("⚠️  Debug: Failed to parse checksum for file '%s' from URL '%s'\n", filename, url)
			fmt.Printf("   Content preview (first 200 chars): %s\n", truncateString(content, 200))
			return "", fmt.Errorf("failed to parse checksum file: %w", err)
		}
		return checksum, nil
	}

	// Otherwise, assume the entire content is the checksum
	return strings.TrimSpace(content), nil
}

// parseChecksumFile parses a checksum file and extracts the checksum for a specific filename
// Supports formats like: "checksum  filename" or "checksum *filename"
func (cv *ChecksumVerifier) parseChecksumFile(content, filename string) (string, error) {
	lines := strings.Split(content, "\n")
	var candidateFiles []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on whitespace
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		checksum := parts[0]
		fileInLine := parts[1]

		// Remove leading asterisk if present (binary mode indicator)
		if strings.HasPrefix(fileInLine, "*") {
			fileInLine = fileInLine[1:]
		}

		// Store candidate files for debugging
		candidateFiles = append(candidateFiles, fileInLine)

		// Check if this line matches our filename (exact match)
		if fileInLine == filename {
			return checksum, nil
		}

		// Also try matching just the basename
		if filepath.Base(fileInLine) == filename {
			return checksum, nil
		}

		// Try matching without path separators
		if strings.Contains(fileInLine, filename) {
			return checksum, nil
		}
	}

	// If we get here, no match was found - provide helpful debug info
	fmt.Printf("⚠️  Debug: Available files in checksum file: %v\n", candidateFiles)
	fmt.Printf("   Looking for: %s\n", filename)
	return "", fmt.Errorf("checksum not found for file %s", filename)
}

// truncateString truncates a string to the specified length with ellipsis
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// VerifyFileWithWarning verifies a file with checksum, but only warns if verification fails
func (cv *ChecksumVerifier) VerifyFileWithWarning(filePath string, checksum ChecksumInfo) {
	if err := cv.VerifyFile(filePath, checksum); err != nil {
		fmt.Printf("  ⚠️  Checksum verification failed: %v\n", err)
		fmt.Printf("      File: %s\n", filePath)
		fmt.Printf("      This could indicate a corrupted download or security issue.\n")
	} else {
		fmt.Printf("  ✅ Checksum verified successfully\n")
	}
}
