package tools

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gnodet/mvx/pkg/config"
	"github.com/gnodet/mvx/pkg/util"
)

// DownloadConfig contains configuration for robust downloads with checksum verification
type DownloadConfig struct {
	URL           string
	DestPath      string
	MaxRetries    int
	RetryDelay    time.Duration
	Timeout       time.Duration
	ExpectedType  string // Expected content type
	MinSize       int64  // Minimum expected file size
	MaxSize       int64  // Maximum expected file size
	ValidateMagic bool   // Whether to validate magic bytes
	ToolName      string // Name of the tool being downloaded (for progress reporting)
	Version       string // Tool version for checksum verification
	Config        config.ToolConfig
	Tool          Tool // Tool instance for checksum verification
}

// getTimeoutFromEnv returns a timeout from environment variable or default value
func getTimeoutFromEnv(envVar string, defaultTimeout time.Duration) time.Duration {
	if timeoutStr := os.Getenv(envVar); timeoutStr != "" {
		if timeout, err := time.ParseDuration(timeoutStr); err == nil && timeout > 0 {
			return timeout
		}
	}
	return defaultTimeout
}

// DefaultDownloadConfig returns a default download configuration
func DefaultDownloadConfig(url, destPath string) *DownloadConfig {
	configProvider := NewDownloadConfigProvider(NewEnvironmentConfigProvider())

	return &DownloadConfig{
		URL:           url,
		DestPath:      destPath,
		MaxRetries:    configProvider.GetMaxRetries(),
		RetryDelay:    configProvider.GetRetryDelay(),
		Timeout:       configProvider.GetDownloadTimeout(),
		MinSize:       configProvider.GetMinFileSize(),
		MaxSize:       configProvider.GetMaxFileSize(),
		ValidateMagic: true,
	}
}

// DownloadResult contains information about the download
type DownloadResult struct {
	Size        int64
	ContentType string
	StatusCode  int
	FinalURL    string
}

// RobustDownload performs a robust download with validation and retries
func RobustDownload(config *DownloadConfig) (*DownloadResult, error) {
	// Apply URL replacements if configured
	originalURL := config.URL
	urlReplacer, err := LoadURLReplacer()
	if err != nil {
		util.LogVerbose("Warning: failed to load URL replacements: %v", err)
	} else {
		config.URL = urlReplacer.ApplyReplacements(config.URL)
		if config.URL != originalURL {
			toolPrefix := ""
			if config.ToolName != "" {
				toolPrefix = fmt.Sprintf("[%s] ", config.ToolName)
			}
			fmt.Printf("  🔄 %sUsing URL replacement: %s\n", toolPrefix, getUserFriendlyURL(config.URL))
		}
	}

	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			toolPrefix := ""
			if config.ToolName != "" {
				toolPrefix = fmt.Sprintf("[%s] ", config.ToolName)
			}
			fmt.Printf("  🔄 %sRetry attempt %d/%d after %v...\n", toolPrefix, attempt, config.MaxRetries, config.RetryDelay)
			time.Sleep(config.RetryDelay * time.Duration(attempt)) // Exponential backoff
		}

		result, err := attemptDownload(config)
		if err == nil {
			return result, nil
		}

		lastErr = err
		toolPrefix := ""
		if config.ToolName != "" {
			toolPrefix = fmt.Sprintf("[%s] ", config.ToolName)
		}
		fmt.Printf("  ⚠️  %sDownload attempt %d failed: %v\n", toolPrefix, attempt+1, err)
	}

	return nil, fmt.Errorf("download failed after %d attempts: %w", config.MaxRetries+1, lastErr)
}

// attemptDownload performs a single download attempt
func attemptDownload(config *DownloadConfig) (*DownloadResult, error) {
	// Create temporary file
	tempFile, err := os.CreateTemp("", "mvx-download-*.tmp")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Create HTTP client with granular timeouts for better handling of slow servers
	configProvider := NewDownloadConfigProvider(NewEnvironmentConfigProvider())

	client := &http.Client{
		Transport: &http.Transport{
			TLSHandshakeTimeout:   configProvider.GetTLSTimeout(),
			ResponseHeaderTimeout: configProvider.GetResponseTimeout(),
			IdleConnTimeout:       configProvider.GetIdleTimeout(),
		},
		// Use context timeout instead of global client timeout for better control
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= MaxRedirects {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	// Create request with context timeout for the entire operation
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", config.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set user agent
	req.Header.Set("User-Agent", "mvx/1.0 (https://github.com/gnodet/mvx)")

	// Perform request with progress indication for slow servers
	toolPrefix := ""
	if config.ToolName != "" {
		toolPrefix = fmt.Sprintf("[%s] ", config.ToolName)
	}
	fmt.Printf("  🌐 %sConnecting to server...\n", toolPrefix)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("  📡 %sServer responded, starting download...\n", toolPrefix)

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Validate content type if specified
	if config.ExpectedType != "" {
		contentType := resp.Header.Get("Content-Type")
		if contentType != "" && !strings.Contains(contentType, config.ExpectedType) {
			return nil, fmt.Errorf("unexpected content type: got %s, expected %s", contentType, config.ExpectedType)
		}
	}

	// Check content length if available
	if contentLength := resp.ContentLength; contentLength > 0 {
		if contentLength < config.MinSize {
			return nil, fmt.Errorf("content too small: %d bytes (minimum %d)", contentLength, config.MinSize)
		}
		if contentLength > config.MaxSize {
			return nil, fmt.Errorf("content too large: %d bytes (maximum %d)", contentLength, config.MaxSize)
		}
	}

	// Download with size tracking
	written, err := io.Copy(tempFile, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("download failed: %w", err)
	}

	// Validate downloaded size
	if written < config.MinSize {
		return nil, fmt.Errorf("downloaded file too small: %d bytes (minimum %d)", written, config.MinSize)
	}
	if written > config.MaxSize {
		return nil, fmt.Errorf("downloaded file too large: %d bytes (maximum %d)", written, config.MaxSize)
	}

	// Close temp file before validation
	tempFile.Close()

	// Validate file content if requested
	if config.ValidateMagic {
		if err := validateFileFormat(tempFile.Name(), config.URL); err != nil {
			return nil, fmt.Errorf("file validation failed: %w", err)
		}
	}

	// Verify checksum if tool is available
	if config.Tool != nil {
		// Update config with final URL for better filename detection
		finalConfig := *config
		finalConfig.URL = resp.Request.URL.String()
		if err := verifyChecksum(tempFile.Name(), &finalConfig); err != nil {
			return nil, err
		}
	}

	// Create destination directory
	if err := os.MkdirAll(filepath.Dir(config.DestPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Move to final destination with Windows-specific retry logic
	if err := moveFileWithRetry(tempFile.Name(), config.DestPath); err != nil {
		return nil, fmt.Errorf("failed to move file to destination: %w", err)
	}

	return &DownloadResult{
		Size:        written,
		ContentType: resp.Header.Get("Content-Type"),
		StatusCode:  resp.StatusCode,
		FinalURL:    resp.Request.URL.String(),
	}, nil
}

// validateFileFormat validates the downloaded file format based on magic bytes
func validateFileFormat(filePath, url string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file for validation: %w", err)
	}
	defer file.Close()

	// Read first few bytes for magic number detection
	header := make([]byte, 512)
	n, err := file.Read(header)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read file header: %w", err)
	}
	header = header[:n]

	// Determine expected format from URL
	if strings.HasSuffix(url, ".tar.gz") || strings.HasSuffix(url, ".tgz") {
		return validateTarGz(header)
	} else if strings.HasSuffix(url, ".tar.xz") {
		return validateTarXz(header)
	} else if strings.HasSuffix(url, ".zip") {
		return validateZip(header)
	} else if strings.HasSuffix(url, ".gz") {
		return validateGzip(header)
	}

	// If we can't determine format from URL, try to detect
	return validateAnyArchive(header)
}

// validateTarGz validates tar.gz format
func validateTarGz(header []byte) error {
	// Check for gzip magic bytes (1f 8b)
	if len(header) < 2 || header[0] != 0x1f || header[1] != 0x8b {
		return fmt.Errorf("invalid gzip header: expected 1f 8b, got %02x %02x", header[0], header[1])
	}
	return nil
}

// validateZip validates ZIP format
func validateZip(header []byte) error {
	// Check for ZIP magic bytes (50 4b 03 04 or 50 4b 05 06 or 50 4b 07 08)
	if len(header) < 4 {
		return fmt.Errorf("file too short for ZIP format")
	}
	if header[0] != 0x50 || header[1] != 0x4b {
		return fmt.Errorf("invalid ZIP header: expected 50 4b, got %02x %02x", header[0], header[1])
	}
	return nil
}

// validateGzip validates gzip format
func validateGzip(header []byte) error {
	return validateTarGz(header) // Same magic bytes
}

// validateTarXz validates tar.xz format
func validateTarXz(header []byte) error {
	// Check for XZ magic bytes (fd 37 7a 58 5a 00)
	if len(header) < 6 {
		return fmt.Errorf("file too short for XZ format")
	}
	expected := []byte{0xfd, 0x37, 0x7a, 0x58, 0x5a, 0x00}
	for i, b := range expected {
		if header[i] != b {
			return fmt.Errorf("invalid XZ header: expected %02x at position %d, got %02x", b, i, header[i])
		}
	}
	return nil
}

// validateAnyArchive tries to detect any known archive format
func validateAnyArchive(header []byte) error {
	if len(header) < 4 {
		return fmt.Errorf("file too short to determine format")
	}

	// Check for common archive formats
	if header[0] == 0x1f && header[1] == 0x8b {
		return nil // gzip
	}
	if header[0] == 0x50 && header[1] == 0x4b {
		return nil // ZIP
	}
	if bytes.HasPrefix(header, []byte("ustar")) {
		return nil // tar
	}
	if len(header) >= 6 && header[0] == 0xfd && header[1] == 0x37 && header[2] == 0x7a {
		return nil // XZ
	}

	// Check if it looks like HTML (common error response)
	if bytes.Contains(header[:min(len(header), 100)], []byte("<html")) ||
		bytes.Contains(header[:min(len(header), 100)], []byte("<!DOCTYPE")) {
		return fmt.Errorf("received HTML content instead of binary archive (likely an error page)")
	}

	// Check if it looks like JSON (API error response)
	if bytes.HasPrefix(bytes.TrimSpace(header), []byte("{")) {
		return fmt.Errorf("received JSON content instead of binary archive (likely an API error)")
	}

	return fmt.Errorf("unrecognized file format")
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// DiagnoseDownloadError provides detailed diagnosis of download failures
func DiagnoseDownloadError(url string, err error) string {
	errStr := err.Error()

	// Common network issues
	if strings.Contains(errStr, "connection refused") {
		return fmt.Sprintf("Connection refused to %s. The server may be down or the URL may be incorrect.", url)
	}
	if strings.Contains(errStr, "timeout") {
		return fmt.Sprintf("Download timeout from %s. This may be due to slow network or server issues. Try again later.", url)
	}
	if strings.Contains(errStr, "no such host") {
		return fmt.Sprintf("DNS resolution failed for %s. Check your internet connection and the URL.", url)
	}

	// HTTP errors
	if strings.Contains(errStr, "HTTP 404") {
		return fmt.Sprintf("File not found (404) at %s. The requested version may not be available.", url)
	}
	if strings.Contains(errStr, "HTTP 403") {
		return fmt.Sprintf("Access forbidden (403) to %s. You may need authentication or the file may be restricted.", url)
	}
	if strings.Contains(errStr, "HTTP 500") || strings.Contains(errStr, "HTTP 502") || strings.Contains(errStr, "HTTP 503") {
		return fmt.Sprintf("Server error from %s. The server is experiencing issues. Try again later.", url)
	}

	// Content validation errors
	if strings.Contains(errStr, "gzip: invalid header") {
		return fmt.Sprintf("Invalid gzip file downloaded from %s. The server may have returned an error page instead of the expected file.", url)
	}
	if strings.Contains(errStr, "invalid ZIP header") {
		return fmt.Sprintf("Invalid ZIP file downloaded from %s. The server may have returned an error page instead of the expected file.", url)
	}
	if strings.Contains(errStr, "HTML content") {
		return fmt.Sprintf("Received HTML error page instead of binary file from %s. Check the URL and try again.", url)
	}
	if strings.Contains(errStr, "JSON content") {
		return fmt.Sprintf("Received JSON error response instead of binary file from %s. The API may have returned an error.", url)
	}

	// Size validation errors
	if strings.Contains(errStr, "too small") {
		return fmt.Sprintf("Downloaded file from %s is too small. The download may have been incomplete.", url)
	}
	if strings.Contains(errStr, "too large") {
		return fmt.Sprintf("Downloaded file from %s is too large. This may not be the expected file.", url)
	}

	// Generic fallback
	return fmt.Sprintf("Download failed from %s: %s", url, errStr)
}

// verifyChecksum verifies the checksum of a downloaded file
func verifyChecksum(filePath string, config *DownloadConfig) error {
	if config.Tool == nil {
		return nil // No tool provided, skip verification
	}

	// Get checksum information from config
	var checksumInfo ChecksumInfo
	var hasChecksum bool

	// Check if checksum is provided in configuration
	if config.Config.Checksum != nil {
		checksumType := config.Config.Checksum.Type
		if checksumType == "" {
			checksumType = "sha256" // default
		}

		checksumInfo = ChecksumInfo{
			Type:     ChecksumType(checksumType),
			Value:    config.Config.Checksum.Value,
			URL:      config.Config.Checksum.URL,
			Filename: config.Config.Checksum.Filename,
		}
		hasChecksum = true
	}

	if !hasChecksum {
		// Try to get checksum from tool using dynamic lookup
		// Extract filename from URL, handling redirects and query parameters
		filename := extractFilenameFromURL(config.URL)
		fmt.Printf("  🔍 Attempting to find checksum for file: %s\n", filename)

		// Use tool's GetChecksum method for dynamic checksum resolution
		if dynamicChecksum, err := config.Tool.GetChecksum(config.Version, filename); err == nil {
			checksumInfo = dynamicChecksum
			hasChecksum = true
		} else {
			fmt.Printf("  ⚠️  Tool checksum lookup failed: %v\n", err)
		}

		if !hasChecksum {
			// Fallback to URL patterns for tools with static checksum URLs
			checksumURL := config.Tool.GetChecksumURL(config.Version, filename)
			if checksumURL != "" {
				fmt.Printf("  🔍 Using checksum URL pattern: %s\n", checksumURL)
				checksumInfo = ChecksumInfo{
					Type:     SHA256,
					URL:      checksumURL,
					Filename: filename,
				}
				hasChecksum = true
			}
		}
	}

	if !hasChecksum {
		if config.Tool.SupportsChecksumVerification() {
			fmt.Printf("⚠️  No checksum available for %s %s\n", config.ToolName, config.Version)
			fmt.Printf("   Consider adding checksum verification to your configuration for enhanced security.\n")
		}
		return nil
	}

	// Check if checksum verification is required
	isRequired := config.Config.Checksum != nil && config.Config.Checksum.Required

	// Create checksum verifier
	verifier := NewChecksumVerifier(config.Tool.GetManager())

	if isRequired {
		// Strict verification - fail on error
		if err := verifier.VerifyFile(filePath, checksumInfo); err != nil {
			// Remove the downloaded file on checksum failure
			os.Remove(filePath)
			// Don't panic, return error instead for better error handling
			return fmt.Errorf("checksum verification failed (required): %v", err)
		}
		fmt.Printf("✅ Checksum verified successfully (required)\n")
	} else {
		// Optional verification - warn on error
		verifier.VerifyFileWithWarning(filePath, checksumInfo)
	}

	return nil
}

// extractFilenameFromURL extracts the filename from a URL, handling redirects and query parameters
func extractFilenameFromURL(urlStr string) string {
	// Parse the URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		// Fallback to simple basename if URL parsing fails
		return filepath.Base(urlStr)
	}

	// Get the path component
	path := parsedURL.Path

	// Handle GitHub release asset URLs and other redirect URLs
	if strings.Contains(path, "/") {
		// Check for Content-Disposition filename in query parameters
		if query := parsedURL.Query(); query.Has("response-content-disposition") {
			disposition := query.Get("response-content-disposition")
			if strings.Contains(disposition, "filename=") {
				// Extract filename from Content-Disposition header
				parts := strings.Split(disposition, "filename=")
				if len(parts) > 1 {
					filename := strings.TrimSpace(parts[1])
					// Remove quotes if present
					filename = strings.Trim(filename, `"'`)
					// Remove any additional parameters after semicolon
					if idx := strings.Index(filename, ";"); idx != -1 {
						filename = filename[:idx]
					}
					if filename != "" {
						return filename
					}
				}
			}
		}
	}

	// Fallback to basename of path
	filename := filepath.Base(path)

	// If filename is empty or looks like a redirect artifact, try to extract from URL patterns
	if filename == "" || filename == "." || filename == "redirect" || len(filename) < 3 {
		// For GitHub release assets, try to extract from the URL pattern
		if strings.Contains(urlStr, "github.com") || strings.Contains(urlStr, "githubusercontent.com") {
			// Look for common archive extensions in the URL
			extensions := []string{ExtTarGz, ExtTarXz, ExtZip, ExtTgz}
			for _, ext := range extensions {
				if idx := strings.LastIndex(urlStr, ext); idx != -1 {
					// Find the start of the filename by looking backwards for a path separator
					start := strings.LastIndex(urlStr[:idx], "/")
					if start != -1 {
						return urlStr[start+1 : idx+len(ext)]
					}
				}
			}
		}

		// Final fallback - return a generic name based on the tool
		return "download"
	}

	return filename
}

// moveFileWithRetry moves a file with retry logic for Windows compatibility
func moveFileWithRetry(src, dst string) error {
	// First, try a simple rename
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	// If rename fails, try copy + delete approach (more reliable on Windows)
	return copyAndDelete(src, dst)
}

// copyAndDelete copies a file and then deletes the source
func copyAndDelete(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy content
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		// Clean up partial destination file
		os.Remove(dst)
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// Ensure data is written to disk
	if err := dstFile.Sync(); err != nil {
		os.Remove(dst)
		return fmt.Errorf("failed to sync destination file: %w", err)
	}

	// Close destination file before removing source
	dstFile.Close()

	// Remove source file
	if err := os.Remove(src); err != nil {
		// Don't fail if we can't remove the source - the copy succeeded
		// This is just cleanup
		return nil
	}

	return nil
}
