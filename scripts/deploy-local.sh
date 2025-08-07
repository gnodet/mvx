#!/bin/bash

##############################################################################
# Deploy Local mvx Binary
#
# This script helps deploy your locally built mvx binary to other projects
# for testing and development.
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

# Check arguments
if [ $# -lt 1 ]; then
    echo "Usage: $0 <target-directory> [method]"
    echo ""
    echo "Methods:"
    echo "  copy     - Copy binary to target directory (default)"
    echo "  symlink  - Create symlink to binary in target directory"
    echo "  wrapper  - Install wrapper and copy binary"
    echo ""
    echo "Examples:"
    echo "  $0 ~/projects/my-java-app"
    echo "  $0 ~/projects/my-java-app symlink"
    echo "  $0 ~/projects/my-java-app wrapper"
    exit 1
fi

TARGET_DIR="$1"
METHOD="${2:-copy}"

# Validate target directory
if [ ! -d "$TARGET_DIR" ]; then
    error "Target directory does not exist: $TARGET_DIR"
fi

# Get absolute paths
MVX_ROOT="$(pwd)"
TARGET_DIR="$(cd "$TARGET_DIR" && pwd)"

info "Deploying local mvx binary to: $TARGET_DIR"
info "Method: $METHOD"

# Build the binary if it doesn't exist or is older than source
if [ ! -f "mvx-dev" ] || [ "main.go" -nt "mvx-dev" ]; then
    info "Building development mvx binary..."
    make dev
    success "Built mvx-dev"
fi

case "$METHOD" in
    "copy")
        info "Copying binary to target directory..."
        cp mvx-dev "$TARGET_DIR/mvx-dev"
        success "Copied mvx-dev to $TARGET_DIR"
        ;;

    "symlink")
        info "Creating symlink in target directory..."
        ln -sf "$MVX_ROOT/mvx-dev" "$TARGET_DIR/mvx-dev"
        success "Created symlink $TARGET_DIR/mvx-dev -> $MVX_ROOT/mvx-dev"
        ;;

    "wrapper")
        info "Installing wrapper in target directory..."
        cd "$TARGET_DIR"

        # Install mvx if not present
        if [ ! -f "mvx" ]; then
            curl -fsSL https://raw.githubusercontent.com/gnodet/mvx/main/install-mvx.sh | bash
            success "Installed mvx"
        else
            info "mvx already present"
        fi

        # Copy binary and set version to dev
        cp "$MVX_ROOT/mvx-dev" "./mvx-dev"
        echo "dev" > .mvx/version
        success "Copied binary and configured for development"
        cd "$MVX_ROOT"
        ;;
        
    *)
        error "Unknown method: $METHOD (use: copy, symlink, or wrapper)"
        ;;
esac

# Test the deployment
info "Testing deployment..."
cd "$TARGET_DIR"

if [ -f "./mvx" ]; then
    # Test with wrapper
    if ./mvx version >/dev/null 2>&1; then
        success "Deployment successful! Test with: cd $TARGET_DIR && ./mvx version"
    else
        warning "Deployment completed but test failed"
    fi
elif [ -f "./mvx-dev" ]; then
    # Test direct binary
    if ./mvx-dev version >/dev/null 2>&1; then
        success "Deployment successful! Test with: cd $TARGET_DIR && ./mvx-dev version"
    else
        warning "Deployment completed but test failed"
    fi
else
    warning "No mvx binary found in target directory"
fi

cd "$MVX_ROOT"

echo ""
info "Next steps:"
echo "  cd $TARGET_DIR"
if [ -f "$TARGET_DIR/mvx" ]; then
    echo "  ./mvx init     # Initialize project configuration"
    echo "  ./mvx setup    # Set up development environment"
    echo "  ./mvx build    # Build the project"
else
    echo "  ./mvx-dev init     # Initialize project configuration"
    echo "  ./mvx-dev setup    # Set up development environment"
    echo "  ./mvx-dev build    # Build the project"
fi
