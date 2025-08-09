#!/bin/bash

##############################################################################
# mvx Installer
#
# This script downloads and installs mvx bootstrap files into your project.
# Run this once to set up mvx, then commit the files to your repo.
##############################################################################

set -e

echo "Installing mvx..."

# Allow override with environment variable for development
if [ -n "$MVX_VERSION" ]; then
    echo "🔧 Using specified version: $MVX_VERSION"
    BOOTSTRAP_VERSION="$MVX_VERSION"

    # If main branch is specified, use development version
    if [ "$MVX_VERSION" = "main" ]; then
        echo "📦 Installing from main branch (development version)"
    fi
else
    # Get latest release version
    echo "🔍 Getting latest release version..."
    LATEST_RELEASE=$(curl -s https://api.github.com/repos/gnodet/mvx/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "$LATEST_RELEASE" ]; then
        echo "❌ Failed to get latest release, falling back to main branch"
        BOOTSTRAP_VERSION="main"
    else
        echo "📦 Latest release: $LATEST_RELEASE"
        BOOTSTRAP_VERSION="$LATEST_RELEASE"
    fi
fi

BASE_URL="https://raw.githubusercontent.com/gnodet/mvx/${BOOTSTRAP_VERSION}"

# Create .mvx directory
mkdir -p .mvx

# Download mvx files
echo "Downloading mvx (Unix script)..."
curl -fsSL "${BASE_URL}/mvx" -o mvx
chmod +x mvx

echo "Downloading mvx.cmd (Windows script)..."
curl -fsSL "${BASE_URL}/mvx.cmd" -o mvx.cmd

echo "Downloading mvx configuration..."
if ! curl -fsSL "${BASE_URL}/.mvx/mvx.properties" -o .mvx/mvx.properties; then
    echo "⚠️  Configuration file not found in release, creating default..."
    cat > .mvx/mvx.properties << 'EOF'
# mvx Configuration
# This file configures mvx bootstrap behavior

# The version of mvx to download and use
# Can be a specific version (e.g., "1.0.0") or "latest" for the most recent release
# For development, use "dev" when you have a local mvx-dev binary
mvxVersion=latest

# Alternative download URL (optional)
# If not specified, defaults to GitHub releases
# mvxDownloadUrl=https://github.com/gnodet/mvx/releases

# Checksum validation (future feature)
# mvxChecksumUrl=https://github.com/gnodet/mvx/releases/download/v{version}/checksums.txt
# mvxValidateChecksum=true
EOF
fi

# Update the version in the properties file to match the installed version
if [ "$BOOTSTRAP_VERSION" = "main" ]; then
    echo "📝 Setting version to 'dev' for main branch in mvx.properties"
    sed -i.bak "s/^mvxVersion=.*/mvxVersion=dev/" .mvx/mvx.properties && rm -f .mvx/mvx.properties.bak
else
    # Remove 'v' prefix if present for version number
    VERSION_NUMBER=$(echo "$BOOTSTRAP_VERSION" | sed 's/^v//')
    echo "📝 Setting version to $VERSION_NUMBER in mvx.properties"
    sed -i.bak "s/^mvxVersion=.*/mvxVersion=$VERSION_NUMBER/" .mvx/mvx.properties && rm -f .mvx/mvx.properties.bak
fi

echo ""
echo "✅ mvx installed successfully!"
echo ""
echo "Files created:"
echo "  - mvx (Unix/Linux/macOS script)"
echo "  - mvx.cmd (Windows script)"
echo "  - .mvx/mvx.properties (configuration and version specification)"
echo ""
echo "Next steps:"
echo "  1. Run './mvx setup' to install tools"
echo "  2. Run './mvx build' to build your project"
echo "  3. Run './mvx update-bootstrap' to check for updates"
echo "  4. Commit these files to your repository"
echo ""
echo "Advanced usage:"
echo "  - Install development version: MVX_VERSION=main curl -fsSL ... | bash"
echo "  - Install specific version: MVX_VERSION=v0.1.0 curl -fsSL ... | bash"
echo ""
echo "For more information, see: https://github.com/gnodet/mvx/blob/main/BOOTSTRAP.md"
