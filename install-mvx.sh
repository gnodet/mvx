#!/bin/bash

##############################################################################
# mvx Installer
#
# This script downloads and installs mvx bootstrap files into your project.
# Run this once to set up mvx, then commit the files to your repo.
##############################################################################

set -e

echo "Installing mvx..."

# Allow override with command line argument or environment variable
REQUESTED_VERSION="${1:-$MVX_VERSION}"

if [ -n "$REQUESTED_VERSION" ]; then
    echo "🔧 Using specified version: $REQUESTED_VERSION"
    BOOTSTRAP_VERSION="$REQUESTED_VERSION"

    # If dev version is specified, use development version
    if [ "$REQUESTED_VERSION" = "dev" ]; then
        echo "📦 Installing development version"
    fi
else
    # Get latest release version
    echo "🔍 Getting latest release version..."
    LATEST_RELEASE=$(curl -s https://api.github.com/repos/gnodet/mvx/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "$LATEST_RELEASE" ]; then
        echo "❌ Failed to get latest release, falling back to development version"
        BOOTSTRAP_VERSION="dev"
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
if [ "$BOOTSTRAP_VERSION" = "dev" ]; then
    echo "📝 Setting version to 'dev' in mvx.properties"
    sed -i.bak "s/^mvxVersion=.*/mvxVersion=dev/" .mvx/mvx.properties && rm -f .mvx/mvx.properties.bak

    echo ""
    echo "⚠️  Development version requires local build:"
    echo "   Since you're using the development version, you'll need to build mvx locally"
    echo "   or the bootstrap will fail when trying to download a 'dev' binary."
    echo ""
    echo "   To build locally:"
    echo "     git clone https://github.com/gnodet/mvx.git"
    echo "     cd mvx && ./mvx build"
    echo ""
    echo "   The build automatically installs to ~/.mvx/dev/mvx and all projects"
    echo "   using 'mvxVersion=dev' will automatically use this binary."
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
echo "  - Install development version: curl -fsSL ... | MVX_VERSION=dev bash"
echo "  - Install specific version: curl -fsSL ... | MVX_VERSION=v0.1.0 bash"
echo "  - Or with command line argument: curl -fsSL ... | bash -s dev"
echo ""
echo "For more information, see: https://github.com/gnodet/mvx/blob/main/BOOTSTRAP.md"
