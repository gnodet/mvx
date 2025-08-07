#!/bin/bash

##############################################################################
# mvx Release Script
#
# This script helps create releases by:
# 1. Building all platform binaries
# 2. Generating checksums
# 3. Creating and pushing a git tag
# 4. GitHub Actions will automatically create the release
##############################################################################

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

success() {
    echo -e "${GREEN}✅ $1${NC}"
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

error() {
    echo -e "${RED}❌ $1${NC}"
    exit 1
}

# Check if we're in the right directory
if [ ! -f "go.mod" ] || [ ! -f "main.go" ]; then
    error "This script must be run from the mvx project root directory"
fi

# Check if git is clean
if [ -n "$(git status --porcelain)" ]; then
    warning "Git working directory is not clean. Uncommitted changes:"
    git status --short
    echo
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        error "Aborted"
    fi
fi

# Get version from user
if [ -z "$1" ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 v1.0.0"
    exit 1
fi

VERSION="$1"

# Validate version format
if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+)?$ ]]; then
    error "Version must be in format vX.Y.Z or vX.Y.Z-suffix (e.g., v1.0.0, v1.0.0-rc1)"
fi

info "Preparing release $VERSION"

# Check if tag already exists
if git tag -l | grep -q "^$VERSION$"; then
    error "Tag $VERSION already exists"
fi

# Run tests
info "Running tests..."
if ! make test; then
    error "Tests failed"
fi
success "Tests passed"

# Build all platforms
info "Building binaries for all platforms..."
if ! make clean; then
    error "Clean failed"
fi

if ! make release-build; then
    error "Build failed"
fi
success "Built binaries for all platforms"

# Show what was built
info "Built artifacts:"
ls -la dist/

# Verify binaries work (only test compatible ones)
info "Testing built binaries..."

# Detect current platform
case "$(uname -s)" in
    Linux*)     platform_os="linux" ;;
    Darwin*)    platform_os="darwin" ;;
    CYGWIN*|MINGW*|MSYS*) platform_os="windows" ;;
    *)          platform_os="unknown" ;;
esac

case "$(uname -m)" in
    x86_64|amd64)   platform_arch="amd64" ;;
    arm64|aarch64)  platform_arch="arm64" ;;
    *)              platform_arch="unknown" ;;
esac

# Test the binary for current platform
if [ "$platform_os" != "unknown" ] && [ "$platform_arch" != "unknown" ]; then
    if [ "$platform_os" = "windows" ]; then
        test_binary="dist/mvx-${platform_os}-${platform_arch}.exe"
    else
        test_binary="dist/mvx-${platform_os}-${platform_arch}"
    fi

    if [ -f "$test_binary" ]; then
        info "Testing $test_binary..."
        if ! "$test_binary" version >/dev/null 2>&1; then
            error "Binary $test_binary failed to run"
        fi
        success "Binary test passed for current platform"
    else
        warning "Binary for current platform ($platform_os-$platform_arch) not found"
    fi
else
    warning "Could not detect platform for binary testing"
fi

# Verify all binaries exist
info "Verifying all expected binaries exist..."
expected_binaries=(
    "dist/mvx-linux-amd64"
    "dist/mvx-linux-arm64"
    "dist/mvx-darwin-amd64"
    "dist/mvx-darwin-arm64"
    "dist/mvx-windows-amd64.exe"
)

for binary in "${expected_binaries[@]}"; do
    if [ ! -f "$binary" ]; then
        error "Expected binary $binary not found"
    fi
done
success "All expected binaries found"

# Create and push tag
info "Creating git tag $VERSION..."
git tag -a "$VERSION" -m "Release $VERSION"

info "Pushing tag to origin..."
git push origin "$VERSION"

success "Tag $VERSION created and pushed"

echo
info "Release process initiated!"
echo
echo "Next steps:"
echo "1. GitHub Actions will automatically build and create the release"
echo "2. Check the Actions tab: https://github.com/gnodet/mvx/actions"
echo "3. The release will be available at: https://github.com/gnodet/mvx/releases/tag/$VERSION"
echo
echo "To test the release once it's published:"
echo "  curl -fsSL https://raw.githubusercontent.com/gnodet/mvx/main/install-mvx.sh | bash"
echo "  ./mvx version"
