package tools

import "time"

// Download Configuration Constants
const (
	// File size limits
	DefaultMinFileSize = 1024       // 1KB minimum file size
	DefaultMaxFileSize = 2147483648 // 2GB maximum file size

	// Timeout defaults
	DefaultDownloadTimeout = 600 * time.Second // 10 minutes
	DefaultRegistryTimeout = 120 * time.Second // 2 minutes
	DefaultChecksumTimeout = 120 * time.Second // 2 minutes
	DefaultTLSTimeout      = 120 * time.Second // 2 minutes
	DefaultResponseTimeout = 120 * time.Second // 2 minutes
	DefaultIdleTimeout     = 90 * time.Second  // 90 seconds

	// Retry configuration
	DefaultMaxRetries = 3
	DefaultRetryDelay = 2 * time.Second

	// Concurrency limits
	DefaultMaxConcurrent = 3
	MaxRedirects         = 10
)

// API Base URLs
const (
	FoojayDiscoAPIBase = "https://api.foojay.io/disco/v3.0"
	GitHubAPIBase      = "https://api.github.com"
	NodeJSDistBase     = "https://nodejs.org/dist"
	GoDevAPIBase       = "https://go.dev/dl"
	ApacheMavenBase    = "https://archive.apache.org/dist/maven"
	ApacheDistBase     = "https://dist.apache.org/repos/dist/release/maven"
)

// Environment Variable Names
const (
	EnvVerbose           = "MVX_VERBOSE"
	EnvDownloadTimeout   = "MVX_DOWNLOAD_TIMEOUT"
	EnvRegistryTimeout   = "MVX_REGISTRY_TIMEOUT"
	EnvChecksumTimeout   = "MVX_CHECKSUM_TIMEOUT"
	EnvTLSTimeout        = "MVX_TLS_TIMEOUT"
	EnvResponseTimeout   = "MVX_RESPONSE_TIMEOUT"
	EnvIdleTimeout       = "MVX_IDLE_TIMEOUT"
	EnvMaxRetries        = "MVX_MAX_RETRIES"
	EnvRetryDelay        = "MVX_RETRY_DELAY"
	EnvParallelDownloads = "MVX_PARALLEL_DOWNLOADS"
	EnvNoColor           = "MVX_NO_COLOR"
)

// File Extensions
const (
	ExtZip   = ".zip"
	ExtTarGz = ".tar.gz"
	ExtTarXz = ".tar.xz"
	ExtTgz   = ".tgz"
)

// Archive Types
const (
	ArchiveTypeZip   = "zip"
	ArchiveTypeTarGz = "tar.gz"
	ArchiveTypeTarXz = "tar.xz"
)

// Content Types
const (
	ContentTypeApplication = "application"
	ContentTypeOctetStream = "application/octet-stream"
)

// Tool Names (for consistency)
const (
	ToolJava  = "java"
	ToolMaven = "maven"
	ToolMvnd  = "mvnd"
	ToolNode  = "node"
	ToolGo    = "go"
)

// Platform Strings
const (
	PlatformLinuxX64    = "linux-x64"
	PlatformLinuxArm64  = "linux-arm64"
	PlatformDarwinX64   = "darwin-x64"
	PlatformDarwinArm64 = "darwin-arm64"
	PlatformWinX64      = "win-x64"
)

// Binary Names
const (
	BinaryJava  = "java"
	BinaryMaven = "mvn"
	BinaryMvnd  = "mvnd"
	BinaryNode  = "node"
	BinaryGo    = "go"
)
