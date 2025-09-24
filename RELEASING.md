# Releasing mvx

This document describes the process for creating a new release of mvx.

## Prerequisites

Before creating a release, ensure you have:

1. **Git repository access**: Push access to the main repository
2. **Clean working directory**: No uncommitted changes
3. **All tests passing**: Run `./mvx test` to verify
4. **Updated documentation**: Ensure all changes are documented
5. **Version decision**: Decide on the version number following [Semantic Versioning](https://semver.org/)

## Version Numbering

mvx follows [Semantic Versioning](https://semver.org/):

- **MAJOR** version (e.g., `v1.0.0` → `v2.0.0`): Breaking changes
- **MINOR** version (e.g., `v1.0.0` → `v1.1.0`): New features, backward compatible
- **PATCH** version (e.g., `v1.0.0` → `v1.0.1`): Bug fixes, backward compatible

## Release Process

### 1. Prepare the Release

```bash
# Ensure you're on the main branch
git checkout main
git pull origin main

# Verify the working directory is clean
git status

# Run tests to ensure everything works
./mvx test

# Optional: Preview the changelog for the upcoming release
./mvx changelog  # Shows changes since last tag
```

### 2. Create the Release

Use the release script to automate the process:

```bash
# Create a new release (replace with your version)
./scripts/release.sh v1.2.3
```

The release script will:
1. ✅ Validate the version format
2. ✅ Check that the tag doesn't already exist
3. ✅ Run tests to ensure everything works
4. ✅ Build binaries for all platforms
5. ✅ Generate checksums
6. ✅ Create and push the git tag
7. ✅ Trigger GitHub Actions to create the release

### 3. Monitor the Release

After running the release script:

1. **Check GitHub Actions**: Visit [Actions tab](https://github.com/gnodet/mvx/actions) to monitor the release workflow
2. **Verify the release**: Once complete, check the [Releases page](https://github.com/gnodet/mvx/releases)
3. **Test the release**: Download and test the released binaries

## What Happens Automatically

When you push a tag, GitHub Actions automatically:

1. **Builds binaries** for all supported platforms (Linux, macOS, Windows) and architectures (amd64, arm64)
2. **Generates checksums** for all binaries
3. **Creates release notes** with:
   - Automatic changelog generated from git commits since the last release
   - Installation instructions
   - Asset table with download links
   - Checksum verification instructions
4. **Publishes the release** on GitHub

## Changelog Generation

The release process automatically generates a changelog by analyzing git commits between releases. Commits are categorized as:

- **✨ New Features**: `feat:` prefix or keywords like "add", "implement", "introduce"
- **🐛 Bug Fixes**: `fix:` prefix or keywords like "fix", "bug", "resolve"
- **🔧 Improvements**: `refactor:`, `perf:`, `style:`, `test:` prefixes or keywords like "enhance", "improve", "optimize", "update"
- **📚 Documentation**: `docs:` prefix or keywords like "docs", "documentation", "readme", "website"
- **⚠️ Breaking Changes**: `!` in commit message or "breaking" keyword
- **🔄 Other Changes**: Everything else

### Writing Good Commit Messages

To ensure good changelog generation, write descriptive commit messages:

```bash
# Good examples
git commit -m "feat: add Node.js tool support"
git commit -m "fix: resolve TLS timeout issues in downloads"
git commit -m "docs: update installation instructions"
git commit -m "refactor: improve Maven version resolution"

# Less ideal (but still works)
git commit -m "Add support for Node.js"
git commit -m "Fix timeout bug"
```

## Manual Changelog Generation

You can generate a changelog manually for any range:

```bash
# Between two tags
./mvx changelog v1.0.0 v1.1.0

# From a tag to current HEAD
./mvx changelog v1.0.0 HEAD

# From latest tag to HEAD (default)
./mvx changelog
```

## Testing a Release

After a release is published, test it:

```bash
# Test the wrapper installation
curl -fsSL https://raw.githubusercontent.com/gnodet/mvx/main/install-wrapper.sh | bash
./mvx version

# Test direct binary download
curl -fsSL https://github.com/gnodet/mvx/releases/download/v1.2.3/mvx-linux-amd64 -o mvx-test
chmod +x mvx-test
./mvx-test version

# Verify checksums
curl -fsSL https://github.com/gnodet/mvx/releases/download/v1.2.3/checksums.txt
sha256sum -c checksums.txt
```

## Troubleshooting

### Release Script Fails

If `./scripts/release.sh` fails:

1. **Check git status**: Ensure working directory is clean
2. **Verify tests**: Run `./mvx test` manually
3. **Check version format**: Must be `vX.Y.Z` format
4. **Verify tag doesn't exist**: `git tag -l | grep v1.2.3`

### GitHub Actions Fails

If the GitHub Actions workflow fails:

1. **Check the Actions tab**: Look for error messages
2. **Common issues**:
   - Network timeouts during downloads
   - Build failures due to code issues
   - Permission issues with GitHub token

### Release Notes Are Wrong

If the generated changelog is incorrect:

1. **Check commit messages**: Ensure they follow conventions
2. **Manual edit**: Edit the release on GitHub after it's created
3. **Improve commit messages**: For future releases

## Emergency Procedures

### Delete a Bad Release

If you need to delete a release:

```bash
# Delete the tag locally and remotely
git tag -d v1.2.3
git push origin :refs/tags/v1.2.3

# Delete the GitHub release manually via the web interface
```

### Hotfix Release

For urgent fixes:

1. Create a hotfix branch from the release tag
2. Apply the minimal fix
3. Create a patch release (e.g., `v1.2.3` → `v1.2.4`)
4. Follow the normal release process

## Post-Release Tasks

After a successful release:

1. **Update documentation**: If needed for the new version
2. **Announce the release**: In relevant channels/communities
3. **Update dependent projects**: If you maintain projects that use mvx
4. **Monitor for issues**: Watch for bug reports related to the new release

## Release Checklist

- [ ] Working directory is clean (`git status`)
- [ ] All tests pass (`./mvx test`)
- [ ] Version number decided (following semver)
- [ ] Release script executed (`./scripts/release.sh vX.Y.Z`)
- [ ] GitHub Actions completed successfully
- [ ] Release published on GitHub
- [ ] Release tested with wrapper and direct download
- [ ] Checksums verified
- [ ] Documentation updated if needed

---

For questions or issues with the release process, please [open an issue](https://github.com/gnodet/mvx/issues) or [start a discussion](https://github.com/gnodet/mvx/discussions).
