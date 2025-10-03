package tools

import "time"

// Download Configuration Constants
const (
	// File size limits - very permissive since we rely on checksums for validation
	DefaultMinFileSize = 1024       // 1KB minimum file size (just to catch empty files)
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
	GoGithubAPIBase    = "https://api.github.com/repos/golang/go"
	ApacheMavenBase    = "https://archive.apache.org/dist/maven"
	ApacheDistBase     = "https://dist.apache.org/repos/dist/release/maven"
)

// Environment Variable Names
const (
	// MVX Configuration Environment Variables
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

	// Tool Home Directory Environment Variables
	EnvJavaHome  = "JAVA_HOME"
	EnvMavenHome = "MAVEN_HOME"
	EnvMvndHome  = "MVND_HOME"
	EnvNodeHome  = "NODE_HOME"
	EnvGoRoot    = "GOROOT"
	EnvGoPath    = "GOPATH"
)

// File Extensions
const (
	ExtExe   = ".exe"
	ExtCmd   = ".cmd"
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
