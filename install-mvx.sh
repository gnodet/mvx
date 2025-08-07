#!/bin/bash

##############################################################################
# mvx Installer
#
# This script downloads and installs mvx bootstrap files into your project.
# Run this once to set up mvx, then commit the files to your repo.
##############################################################################

set -e

WRAPPER_VERSION="main"
BASE_URL="https://raw.githubusercontent.com/gnodet/mvx/${WRAPPER_VERSION}"

echo "Installing mvx..."

# Create .mvx directory
mkdir -p .mvx

# Download mvx files
echo "Downloading mvx (Unix script)..."
curl -fsSL "${BASE_URL}/mvx" -o mvx
chmod +x mvx

echo "Downloading mvx.cmd (Windows script)..."
curl -fsSL "${BASE_URL}/mvx.cmd" -o mvx.cmd

echo "Downloading mvx configuration..."
curl -fsSL "${BASE_URL}/.mvx/mvx.properties" -o .mvx/mvx.properties

echo ""
echo "✅ mvx installed successfully!"
echo ""
echo "Files created:"
echo "  - mvx (Unix/Linux/macOS script)"
echo "  - mvx.cmd (Windows script)"
echo "  - .mvx/mvx.properties (configuration and version specification)"
echo ""
echo "Next steps:"
echo "  1. Edit .mvx/mvx.properties to set your desired mvx version"
echo "  2. Test mvx: ./mvx version"
echo "  3. Commit these files to your repository"
echo "  4. Update your documentation to use './mvx' instead of 'mvx'"
echo ""
echo "For more information, see: https://github.com/gnodet/mvx/blob/main/WRAPPER.md"
