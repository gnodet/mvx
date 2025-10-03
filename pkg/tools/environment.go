package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnodet/mvx/pkg/util"
)

// EnvironmentManager provides safe environment variable management with special PATH handling
type EnvironmentManager struct {
	envVars  map[string]string
	pathDirs []string
}

// NewEnvironmentManager creates a new environment manager
func NewEnvironmentManager() *EnvironmentManager {
	return &EnvironmentManager{
		envVars:  make(map[string]string),
		pathDirs: []string{},
	}
}

// NewEnvironmentManagerFromMap creates a new environment manager from an existing map
func NewEnvironmentManagerFromMap(envVars map[string]string) *EnvironmentManager {
	em := NewEnvironmentManager()

	for key, value := range envVars {
		if key == "PATH" {
			// Parse PATH into directories
			if value != "" {
				em.pathDirs = strings.Split(value, string(os.PathListSeparator))
			}
		} else {
			em.envVars[key] = value
		}
	}

	return em
}

// SetEnv sets an environment variable (panics if key is "PATH")
func (em *EnvironmentManager) SetEnv(key, value string) {
	if key == "PATH" {
		panic("Cannot set PATH directly, use AddToPath() or AppendToPath() instead")
	}
	em.envVars[key] = value
	util.LogVerbose("Set environment variable %s=%s", key, value)
}

// GetEnv gets an environment variable
func (em *EnvironmentManager) GetEnv(key string) (string, bool) {
	if key == "PATH" {
		return em.GetPath(), true
	}
	value, exists := em.envVars[key]
	return value, exists
}

// AddToPath prepends a directory to PATH if not already present
func (em *EnvironmentManager) AddToPath(dir string) {
	if dir == "" {
		return
	}

	// Clean the path
	dir = filepath.Clean(dir)

	// Check if already present
	for _, existing := range em.pathDirs {
		if existing == dir {
			util.LogVerbose("Directory %s already in PATH", dir)
			return
		}
	}

	// Prepend to PATH
	em.pathDirs = append([]string{dir}, em.pathDirs...)
	util.LogVerbose("Added directory to PATH: %s", dir)
}

// AppendToPath appends a directory to PATH if not already present
func (em *EnvironmentManager) AppendToPath(dir string) {
	if dir == "" {
		return
	}

	// Clean the path
	dir = filepath.Clean(dir)

	// Check if already present
	for _, existing := range em.pathDirs {
		if existing == dir {
			util.LogVerbose("Directory %s already in PATH", dir)
			return
		}
	}

	// Append to PATH
	em.pathDirs = append(em.pathDirs, dir)
	util.LogVerbose("Appended directory to PATH: %s", dir)
}

// GetPath returns the constructed PATH string
func (em *EnvironmentManager) GetPath() string {
	return strings.Join(em.pathDirs, string(os.PathListSeparator))
}

// ToMap converts the environment manager to a map[string]string
func (em *EnvironmentManager) ToMap() map[string]string {
	result := make(map[string]string)

	// Copy all environment variables
	for key, value := range em.envVars {
		result[key] = value
	}

	// Add PATH
	if len(em.pathDirs) > 0 {
		result["PATH"] = em.GetPath()
	}

	return result
}

// ToSlice converts the environment manager to []string in "KEY=VALUE" format
func (em *EnvironmentManager) ToSlice() []string {
	var result []string

	// Add all environment variables
	for key, value := range em.envVars {
		result = append(result, fmt.Sprintf("%s=%s", key, value))
	}

	// Add PATH
	if len(em.pathDirs) > 0 {
		result = append(result, fmt.Sprintf("PATH=%s", em.GetPath()))
	}

	return result
}
