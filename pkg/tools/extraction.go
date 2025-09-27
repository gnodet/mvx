package tools

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// detectSingleTopLevelDirectory checks if all files in a ZIP archive are under a single top-level directory
// Returns the directory prefix to strip, or empty string if no stripping should be done
func detectSingleTopLevelDirectory(files []*zip.File) string {
	if len(files) == 0 {
		return ""
	}

	var topLevelDir string
	for _, file := range files {
		// Skip empty entries
		if file.Name == "" {
			continue
		}

		// Get the first path component
		parts := strings.Split(file.Name, "/")
		if len(parts) == 0 {
			return "" // No directory structure
		}

		firstComponent := parts[0]
		if firstComponent == "" {
			return "" // Absolute path, don't strip
		}

		if topLevelDir == "" {
			topLevelDir = firstComponent
		} else if topLevelDir != firstComponent {
			return "" // Multiple top-level directories, don't strip
		}
	}

	// Ensure the top-level directory actually exists as a directory entry
	topLevelDirPath := topLevelDir + "/"
	for _, file := range files {
		if file.Name == topLevelDirPath && file.FileInfo().IsDir() {
			return topLevelDirPath
		}
	}

	// If we don't find the directory entry, still strip if all files are under the same prefix
	if topLevelDir != "" {
		return topLevelDir + "/"
	}

	return ""
}

// detectSingleTopLevelDirectoryTar checks if all files in a tar archive are under a single top-level directory
// Returns the directory prefix to strip, or empty string if no stripping should be done
func detectSingleTopLevelDirectoryTar(headers []*tar.Header) string {
	if len(headers) == 0 {
		return ""
	}

	var topLevelDir string
	for _, header := range headers {
		// Skip empty entries
		if header.Name == "" {
			continue
		}

		// Get the first path component
		parts := strings.Split(header.Name, "/")
		if len(parts) == 0 {
			return "" // No directory structure
		}

		firstComponent := parts[0]
		if firstComponent == "" {
			return "" // Absolute path, don't strip
		}

		if topLevelDir == "" {
			topLevelDir = firstComponent
		} else if topLevelDir != firstComponent {
			return "" // Multiple top-level directories, don't strip
		}
	}

	// Ensure the top-level directory actually exists as a directory entry
	topLevelDirPath := topLevelDir + "/"
	for _, header := range headers {
		if header.Name == topLevelDirPath && header.Typeflag == tar.TypeDir {
			return topLevelDirPath
		}
	}

	// If we don't find the directory entry, still strip if all files are under the same prefix
	if topLevelDir != "" {
		return topLevelDir + "/"
	}

	return ""
}

// extractZipFile extracts a zip file to the destination directory
func extractZipFile(src, dest string) error {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("failed to open ZIP archive: %w", err)
	}
	defer reader.Close()

	// Create destination directory
	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Check if archive contains a single top-level directory
	stripPrefix := detectSingleTopLevelDirectory(reader.File)

	// Extract files
	for _, file := range reader.File {
		// Skip the top-level directory if we're stripping it
		relativePath := file.Name
		if stripPrefix != "" {
			if !strings.HasPrefix(file.Name, stripPrefix) {
				continue
			}
			relativePath = strings.TrimPrefix(file.Name, stripPrefix)
			if relativePath == "" {
				continue // Skip the directory itself
			}
		}

		targetPath := filepath.Join(dest, relativePath)

		// Security check: ensure the file path is within destDir
		if !strings.HasPrefix(targetPath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path in ZIP: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			// Create directory
			if err := os.MkdirAll(targetPath, file.FileInfo().Mode()); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", targetPath, err)
			}
		} else {
			// Extract file
			if err := extractSingleZipFile(file, targetPath); err != nil {
				return fmt.Errorf("failed to extract file %s: %w", targetPath, err)
			}
		}
	}

	return nil
}

// extractSingleZipFile extracts a single file from ZIP archive
func extractSingleZipFile(file *zip.File, targetPath string) error {
	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}

	// Open file in ZIP
	reader, err := file.Open()
	if err != nil {
		return err
	}
	defer reader.Close()

	// Ensure we have write permissions for the file
	mode := file.FileInfo().Mode()
	if mode&0200 == 0 {
		mode |= 0200 // Add write permission for owner
	}

	// Create target file
	targetFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	// Copy content
	_, err = io.Copy(targetFile, reader)
	return err
}

// extractTarGzFile extracts a tar.gz file to the destination directory
func extractTarGzFile(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	// Create gzip reader
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	// Create tar reader
	tarReader := tar.NewReader(gzReader)

	// Create destination directory
	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// First pass: collect all headers to detect single top-level directory
	var headers []*tar.Header
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}
		headers = append(headers, header)
	}

	// Detect if we should strip a single top-level directory
	stripPrefix := detectSingleTopLevelDirectoryTar(headers)

	// Reopen the file for second pass
	file.Close()
	file, err = os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to reopen archive: %w", err)
	}
	defer file.Close()

	gzReader, err = gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	tarReader = tar.NewReader(gzReader)

	// Second pass: extract files
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		// Skip the top-level directory if we're stripping it
		relativePath := header.Name
		if stripPrefix != "" {
			if !strings.HasPrefix(header.Name, stripPrefix) {
				continue
			}
			relativePath = strings.TrimPrefix(header.Name, stripPrefix)
			if relativePath == "" {
				continue // Skip the directory itself
			}
		}

		targetPath := filepath.Join(dest, relativePath)

		// Security check: ensure the file path is within destDir
		if !strings.HasPrefix(targetPath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path in tar: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", targetPath, err)
			}
		case tar.TypeReg:
			// Extract regular file
			if err := extractSingleTarFile(tarReader, targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to extract file %s: %w", targetPath, err)
			}
		case tar.TypeSymlink:
			// Create symlink, handling existing symlinks
			if err := createSymlinkSafely(header.Linkname, targetPath); err != nil {
				return fmt.Errorf("failed to create symlink %s: %w", targetPath, err)
			}
		default:
			// Skip other file types (char devices, block devices, etc.)
			logVerbose("Skipping unsupported file type %d for %s", header.Typeflag, header.Name)
		}
	}

	return nil
}

// extractSingleTarFile extracts a single file from tar reader
func extractSingleTarFile(tarReader *tar.Reader, targetPath string, mode os.FileMode) error {
	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}

	// Ensure we have write permissions for the file
	if mode&0200 == 0 {
		mode |= 0200 // Add write permission for owner
	}

	// Create file
	file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer file.Close()

	// Copy content
	_, err = io.Copy(file, tarReader)
	return err
}

// createSymlinkSafely creates a symlink, handling existing files/symlinks
func createSymlinkSafely(linkname, targetPath string) error {
	// Check if target already exists
	if _, err := os.Lstat(targetPath); err == nil {
		// Target exists, check if it's already the correct symlink
		if existingLink, err := os.Readlink(targetPath); err == nil {
			if existingLink == linkname {
				// Already the correct symlink, nothing to do
				logVerbose("Symlink %s already exists with correct target %s", targetPath, linkname)
				return nil
			}
			// Different symlink target, remove and recreate
			logVerbose("Removing existing symlink %s (target: %s) to create new one (target: %s)", targetPath, existingLink, linkname)
		} else {
			// Not a symlink, but some other file/directory exists
			logVerbose("Removing existing file/directory %s to create symlink (target: %s)", targetPath, linkname)
		}

		// Remove existing file/symlink/directory
		if err := os.RemoveAll(targetPath); err != nil {
			return fmt.Errorf("failed to remove existing file %s: %w", targetPath, err)
		}
	}

	// Create the symlink
	return os.Symlink(linkname, targetPath)
}

// extractTarXzFile extracts a tar.xz file using system tar command
func extractTarXzFile(src, dest string) error {
	// Create destination directory
	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Use system tar command for tar.xz files
	cmd := exec.Command("tar", "-xJf", src, "-C", dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract tar.xz file: %w", err)
	}

	return nil
}

// detectArchiveType detects the archive type from file extension
func detectArchiveType(filename string) string {
	filename = strings.ToLower(filename)

	if strings.HasSuffix(filename, ".tar.gz") || strings.HasSuffix(filename, ".tgz") {
		return "tar.gz"
	}
	if strings.HasSuffix(filename, ".tar.xz") {
		return "tar.xz"
	}
	if strings.HasSuffix(filename, ".zip") {
		return "zip"
	}
	if strings.HasSuffix(filename, ".gz") {
		return "tar.gz" // Assume tar.gz for .gz files
	}

	// Default fallback
	return "tar.gz"
}

// ExtractArchive extracts an archive file automatically detecting the type
func ExtractArchive(src, dest string) error {
	archiveType := detectArchiveType(src)

	switch archiveType {
	case "zip":
		return extractZipFile(src, dest)
	case "tar.gz":
		return extractTarGzFile(src, dest)
	case "tar.xz":
		return extractTarXzFile(src, dest)
	default:
		return fmt.Errorf("unsupported archive type: %s", archiveType)
	}
}
