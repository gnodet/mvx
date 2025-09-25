package tools

import (
	"fmt"
)

// ToolError represents a standardized error for tool operations
type ToolError struct {
	Tool    string // Tool name (e.g., "java", "maven")
	Version string // Tool version (e.g., "17", "4.0.0")
	Op      string // Operation (e.g., "install", "verify", "download")
	Err     error  // Underlying error
}

// Error implements the error interface
func (e *ToolError) Error() string {
	if e.Version != "" {
		return fmt.Sprintf("%s %s %s failed: %v", e.Tool, e.Version, e.Op, e.Err)
	}
	return fmt.Sprintf("%s %s failed: %v", e.Tool, e.Op, e.Err)
}

// Unwrap returns the underlying error for error unwrapping
func (e *ToolError) Unwrap() error {
	return e.Err
}

// NewToolError creates a new ToolError
func NewToolError(tool, version, op string, err error) *ToolError {
	return &ToolError{
		Tool:    tool,
		Version: version,
		Op:      op,
		Err:     err,
	}
}

// InstallError creates a standardized installation error
func InstallError(tool, version string, err error) *ToolError {
	return NewToolError(tool, version, "install", err)
}

// VerifyError creates a standardized verification error
func VerifyError(tool, version string, err error) *ToolError {
	return NewToolError(tool, version, "verify", err)
}

// DownloadError creates a standardized download error
func DownloadError(tool, version string, err error) *ToolError {
	return NewToolError(tool, version, "download", err)
}

// ListVersionsError creates a standardized list versions error
func ListVersionsError(tool string, err error) *ToolError {
	return NewToolError(tool, "", "list versions", err)
}

// PathError creates a standardized path resolution error
func PathError(tool, version string, err error) *ToolError {
	return NewToolError(tool, version, "path resolution", err)
}
