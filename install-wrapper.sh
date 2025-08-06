#!/bin/bash

##############################################################################
# mvx Wrapper Installer
#
# This script downloads and installs the mvx wrapper files into your project.
# Run this once to set up the wrapper, then commit the files to your repo.
##############################################################################

set -e

WRAPPER_VERSION="main"
BASE_URL="https://raw.githubusercontent.com/gnodet/mvx/${WRAPPER_VERSION}"

echo "Installing mvx wrapper..."

# Create .mvx/wrapper directory
mkdir -p .mvx/wrapper

# Download wrapper files
echo "Downloading mvx (Unix script)..."
curl -fsSL "${BASE_URL}/mvx" -o mvx
chmod +x mvx

echo "Downloading mvx.cmd (Windows script)..."
curl -fsSL "${BASE_URL}/mvx.cmd" -o mvx.cmd

echo "Downloading wrapper configuration..."
curl -fsSL "${BASE_URL}/.mvx/wrapper/mvx-wrapper.properties" -o .mvx/wrapper/mvx-wrapper.properties

echo "Downloading version file..."
curl -fsSL "${BASE_URL}/.mvx/version" -o .mvx/version

echo ""
echo "✅ mvx wrapper installed successfully!"
echo ""
echo "Files created:"
echo "  - mvx (Unix/Linux/macOS script)"
echo "  - mvx.cmd (Windows script)"
echo "  - .mvx/wrapper/mvx-wrapper.properties (configuration)"
echo "  - .mvx/version (version specification)"
echo ""
echo "Next steps:"
echo "  1. Edit .mvx/wrapper/mvx-wrapper.properties or .mvx/version to set your desired mvx version"
echo "  2. Test the wrapper: ./mvx version"
echo "  3. Commit these files to your repository"
echo "  4. Update your documentation to use './mvx' instead of 'mvx'"
echo ""
echo "For more information, see: https://github.com/gnodet/mvx/blob/main/WRAPPER.md"
