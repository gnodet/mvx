#!/bin/bash

##############################################################################
# mvx Changelog Generator
#
# This script generates a changelog from git commits between two tags/releases.
# It categorizes commits by type and formats them for GitHub releases.
#
# Usage:
#   ./scripts/generate-changelog.sh [from_tag] [to_tag]
#   ./scripts/generate-changelog.sh v0.6.1 v0.7.0
#   ./scripts/generate-changelog.sh v0.6.1 HEAD
#
# If no arguments provided, it will generate changelog from the latest tag to HEAD
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
    echo -e "${BLUE}ℹ️  $1${NC}" >&2
}

success() {
    echo -e "${GREEN}✅ $1${NC}" >&2
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}" >&2
}

error() {
    echo -e "${RED}❌ $1${NC}" >&2
    exit 1
}

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    error "This script must be run from within a git repository"
fi

# Get the range for changelog generation
FROM_TAG="$1"
TO_TAG="${2:-HEAD}"

# If no FROM_TAG provided, use the latest tag
if [ -z "$FROM_TAG" ]; then
    FROM_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
    if [ -z "$FROM_TAG" ]; then
        error "No tags found in repository. Please provide a from_tag argument."
    fi
    info "Using latest tag: $FROM_TAG"
fi

# Validate that FROM_TAG exists
if ! git rev-parse "$FROM_TAG" >/dev/null 2>&1; then
    error "Tag/commit '$FROM_TAG' does not exist"
fi

# Validate that TO_TAG exists (if not HEAD)
if [ "$TO_TAG" != "HEAD" ] && ! git rev-parse "$TO_TAG" >/dev/null 2>&1; then
    error "Tag/commit '$TO_TAG' does not exist"
fi

info "Generating changelog from $FROM_TAG to $TO_TAG"

# Get commit range
if [ "$FROM_TAG" = "$TO_TAG" ]; then
    error "FROM_TAG and TO_TAG cannot be the same"
fi

# Get commits in the range (excluding the FROM_TAG commit itself)
COMMITS=$(git log --pretty=format:"%H|%s|%an|%ad" --date=short "${FROM_TAG}..${TO_TAG}" --reverse)

if [ -z "$COMMITS" ]; then
    warning "No commits found between $FROM_TAG and $TO_TAG"
    echo "No changes to report."
    exit 0
fi

# Initialize arrays for different types of changes
declare -a FEATURES=()
declare -a FIXES=()
declare -a DOCS=()
declare -a IMPROVEMENTS=()
declare -a BREAKING=()
declare -a OTHER=()

# Process each commit
while IFS='|' read -r hash subject author date; do
    # Skip empty lines
    [ -z "$hash" ] && continue
    
    # Categorize commits based on conventional commit format and keywords
    if [[ "$subject" =~ ^feat(\(.+\))?!?: ]] || [[ "$subject" =~ [Aa]dd.*support ]] || [[ "$subject" =~ [Ii]mplement ]] || [[ "$subject" =~ [Ii]ntroduce ]]; then
        # Check for breaking changes
        if [[ "$subject" =~ ! ]] || [[ "$subject" =~ [Bb]reaking ]] || [[ "$subject" =~ BREAKING ]]; then
            BREAKING+=("$subject")
        else
            FEATURES+=("$subject")
        fi
    elif [[ "$subject" =~ ^fix(\(.+\))?: ]] || [[ "$subject" =~ [Ff]ix ]] || [[ "$subject" =~ [Bb]ug ]] || [[ "$subject" =~ [Rr]esolve ]]; then
        FIXES+=("$subject")
    elif [[ "$subject" =~ ^docs?(\(.+\))?: ]] || [[ "$subject" =~ [Dd]ocs?: ]] || [[ "$subject" =~ [Dd]ocumentation ]] || [[ "$subject" =~ [Rr]eadme ]] || [[ "$subject" =~ website ]]; then
        DOCS+=("$subject")
    elif [[ "$subject" =~ ^(refactor|perf|style|test)(\(.+\))?: ]] || [[ "$subject" =~ [Ee]nhance ]] || [[ "$subject" =~ [Ii]mprove ]] || [[ "$subject" =~ [Oo]ptimize ]] || [[ "$subject" =~ [Uu]pdate ]] || [[ "$subject" =~ [Rr]efactor ]]; then
        IMPROVEMENTS+=("$subject")
    else
        OTHER+=("$subject")
    fi
done <<< "$COMMITS"

# Generate the changelog
echo "## 🚀 What's New"
echo

# Breaking Changes (highest priority)
if [ ${#BREAKING[@]} -gt 0 ]; then
    echo "### ⚠️ Breaking Changes"
    echo
    for item in "${BREAKING[@]}"; do
        # Clean up the commit message
        clean_msg=$(echo "$item" | sed -E 's/^[a-z]+(\([^)]+\))?!?: //')
        # Capitalize first letter
        clean_msg="$(echo "${clean_msg:0:1}" | tr '[:lower:]' '[:upper:]')${clean_msg:1}"
        echo "- $clean_msg"
    done
    echo
fi

# Features
if [ ${#FEATURES[@]} -gt 0 ]; then
    echo "### ✨ New Features"
    echo
    for item in "${FEATURES[@]}"; do
        # Clean up the commit message
        clean_msg=$(echo "$item" | sed -E 's/^feat(\([^)]+\))?: //')
        # Capitalize first letter
        clean_msg="$(echo "${clean_msg:0:1}" | tr '[:lower:]' '[:upper:]')${clean_msg:1}"
        echo "- $clean_msg"
    done
    echo
fi

# Bug Fixes
if [ ${#FIXES[@]} -gt 0 ]; then
    echo "### 🐛 Bug Fixes"
    echo
    for item in "${FIXES[@]}"; do
        # Clean up the commit message
        clean_msg=$(echo "$item" | sed -E 's/^fix(\([^)]+\))?: //')
        # Capitalize first letter
        clean_msg="$(echo "${clean_msg:0:1}" | tr '[:lower:]' '[:upper:]')${clean_msg:1}"
        echo "- $clean_msg"
    done
    echo
fi

# Improvements
if [ ${#IMPROVEMENTS[@]} -gt 0 ]; then
    echo "### 🔧 Improvements"
    echo
    for item in "${IMPROVEMENTS[@]}"; do
        # Clean up the commit message
        clean_msg=$(echo "$item" | sed -E 's/^(refactor|perf|style|test)(\([^)]+\))?: //')
        # Capitalize first letter
        clean_msg="$(echo "${clean_msg:0:1}" | tr '[:lower:]' '[:upper:]')${clean_msg:1}"
        echo "- $clean_msg"
    done
    echo
fi

# Documentation
if [ ${#DOCS[@]} -gt 0 ]; then
    echo "### 📚 Documentation"
    echo
    for item in "${DOCS[@]}"; do
        # Clean up the commit message
        clean_msg=$(echo "$item" | sed -E 's/^docs?(\([^)]+\))?: //')
        # Capitalize first letter
        clean_msg="$(echo "${clean_msg:0:1}" | tr '[:lower:]' '[:upper:]')${clean_msg:1}"
        echo "- $clean_msg"
    done
    echo
fi

# Other changes
if [ ${#OTHER[@]} -gt 0 ]; then
    echo "### 🔄 Other Changes"
    echo
    for item in "${OTHER[@]}"; do
        # Just capitalize first letter
        clean_msg="$(echo "${item:0:1}" | tr '[:lower:]' '[:upper:]')${item:1}"
        echo "- $clean_msg"
    done
    echo
fi

# Add footer with commit range info
echo "---"
echo
echo "**Full Changelog**: https://github.com/gnodet/mvx/compare/${FROM_TAG}...${TO_TAG}"

success "Changelog generated successfully"
