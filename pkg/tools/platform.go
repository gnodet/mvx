package tools

import (
	"fmt"
	"runtime"
)

// PlatformInfo contains platform detection information
type PlatformInfo struct {
	OS   string // Operating system (linux, darwin, windows)
	Arch string // Architecture (amd64, arm64, 386)
}

// GetPlatformInfo returns the current platform information
func GetPlatformInfo() PlatformInfo {
	return PlatformInfo{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
}

// PlatformMapper provides platform-specific string generation for different tools
type PlatformMapper struct {
	platform PlatformInfo
}

// NewPlatformMapper creates a new platform mapper
func NewPlatformMapper() *PlatformMapper {
	return &PlatformMapper{
		platform: GetPlatformInfo(),
	}
}

// GetGenericPlatform returns a generic platform string (os-arch)
func (pm *PlatformMapper) GetGenericPlatform() string {
	return fmt.Sprintf("%s-%s", pm.platform.OS, pm.platform.Arch)
}

// GetOS returns the current operating system
func (pm *PlatformMapper) GetOS() string {
	return pm.platform.OS
}

// GetArch returns the current architecture
func (pm *PlatformMapper) GetArch() string {
	return pm.platform.Arch
}

// MapArchitecture maps Go architecture names to target naming convention
func (pm *PlatformMapper) MapArchitecture(mapping map[string]string) string {
	if mapped, exists := mapping[pm.platform.Arch]; exists {
		return mapped
	}
	return pm.platform.Arch // fallback to original
}

// MapOS maps Go OS names to target naming convention
func (pm *PlatformMapper) MapOS(mapping map[string]string) string {
	if mapped, exists := mapping[pm.platform.OS]; exists {
		return mapped
	}
	return pm.platform.OS // fallback to original
}

// IsWindows returns true if the current platform is Windows
func (pm *PlatformMapper) IsWindows() bool {
	return pm.platform.OS == "windows"
}

// IsUnix returns true if the current platform is Unix-like (Linux, macOS, etc.)
func (pm *PlatformMapper) IsUnix() bool {
	return pm.platform.OS != "windows"
}

// IsMacOS returns true if the current platform is macOS
func (pm *PlatformMapper) IsMacOS() bool {
	return pm.platform.OS == "darwin"
}

// IsLinux returns true if the current platform is Linux
func (pm *PlatformMapper) IsLinux() bool {
	return pm.platform.OS == "linux"
}

// IsARM64 returns true if the current architecture is ARM64
func (pm *PlatformMapper) IsARM64() bool {
	return pm.platform.Arch == "arm64"
}

// IsAMD64 returns true if the current architecture is AMD64
func (pm *PlatformMapper) IsAMD64() bool {
	return pm.platform.Arch == "amd64"
}
