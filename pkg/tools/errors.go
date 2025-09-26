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

// ConfigurationError creates a standardized configuration error
func ConfigurationError(tool, version string, err error) *ToolError {
	return NewToolError(tool, version, "configuration", err)
}

// SystemToolError creates a standardized system tool error
func SystemToolError(tool string, err error) *ToolError {
	return NewToolError(tool, "", "system tool detection", err)
}

// URLGenerationError creates a standardized URL generation error
func URLGenerationError(tool, version string, err error) *ToolError {
	return NewToolError(tool, version, "URL generation", err)
}

// RegistryError creates a standardized registry operation error
func RegistryError(tool string, err error) *ToolError {
	return NewToolError(tool, "", "registry operation", err)
}

// EnvironmentError creates a standardized environment setup error
func EnvironmentError(tool, version string, err error) *ToolError {
	return NewToolError(tool, version, "environment setup", err)
}

// WrapError wraps an error with tool context if it's not already a ToolError
func WrapError(tool, version, operation string, err error) error {
	if err == nil {
		return nil
	}

	// If it's already a ToolError, return as-is
	if _, ok := err.(*ToolError); ok {
		return err
	}

	return NewToolError(tool, version, operation, err)
}

// IsToolError checks if an error is a ToolError
func IsToolError(err error) bool {
	_, ok := err.(*ToolError)
	return ok
}

// GetToolFromError extracts the tool name from a ToolError
func GetToolFromError(err error) string {
	if toolErr, ok := err.(*ToolError); ok {
		return toolErr.Tool
	}
	return ""
}

// GetOperationFromError extracts the operation from a ToolError
func GetOperationFromError(err error) string {
	if toolErr, ok := err.(*ToolError); ok {
		return toolErr.Op
	}
	return ""
}
