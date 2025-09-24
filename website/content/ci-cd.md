---
title: CI/CD Integration
description: Complete guide to configuring mvx for your CI/CD pipelines
layout: page
---
# CI/CD Integration

This guide covers how to integrate mvx with various CI/CD platforms, with special focus on optimizing build performance through caching.

## mvx Installation Options

### Repository-Based Installation (Recommended)

The easiest way to use mvx in CI/CD is to include it directly in your repository. This ensures all team members and CI/CD systems use the same mvx version.

**Required files:**
- `./mvx` - Bootstrap script for Unix/Linux/macOS
- `./mvx.cmd` - Bootstrap script for Windows
- `./.mvx/mvx.properties` - Version configuration
- `./.mvx/config.json5` - Project configuration (optional)

**Setup:**
```bash
# Use the installation script (handles all platforms and edge cases)
curl -fsSL https://raw.githubusercontent.com/gnodet/mvx/main/install-mvx.sh | bash

# Commit to repository
git add mvx mvx.cmd .mvx/
git commit -m "Add mvx bootstrap scripts"
```

**Why use the installation script?**
- ✅ **Automatic version detection** - Gets the latest stable release
- ✅ **Robust error handling** - Handles network issues and fallbacks
- ✅ **Cross-platform compatibility** - Works on all Unix-like systems
- ✅ **Proper configuration** - Creates correct `.mvx/mvx.properties` file
- ✅ **Version flexibility** - Supports specific versions: `MVX_VERSION=v0.5.0 bash`

**Benefits:**
- ✅ **Zero installation required** - works immediately in any environment
- ✅ **Version consistency** - all developers and CI use the same mvx version
- ✅ **Offline-friendly** - cached binaries work without internet
- ✅ **Cross-platform** - works on Linux, macOS, and Windows

## GitHub Actions

### Repository-Based Setup (Recommended)

Here's a GitHub Actions workflow using repository-based mvx installation:

```yaml
name: Build with mvx (Repository-Based)
on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      # No setup needed - mvx bootstrap scripts are in the repository
      # (Added using: curl -fsSL https://raw.githubusercontent.com/gnodet/mvx/main/install-mvx.sh | bash)

      - name: Install tools
        run: ./mvx setup

      - name: Build project
        run: ./mvx mvn clean verify
```



### Optimized Setup with Caching

To avoid re-downloading tools on every build, use GitHub Actions cache:

```yaml
name: Build with mvx (Repository-Based + Cached)
on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      # Cache mvx tools to avoid re-downloading
      - name: Cache mvx tools
        uses: actions/cache@v4
        with:
          path: ~/.mvx/tools
          key: mvx-tools-${{ runner.os }}-${{ hashFiles('.mvx/mvx.properties', '.mvx/config.json5') }}
          restore-keys: |
            mvx-tools-${{ runner.os }}-

      # Cache mvx binaries (downloaded by bootstrap script)
      - name: Cache mvx binaries
        uses: actions/cache@v4
        with:
          path: ~/.mvx/versions
          key: mvx-binary-${{ runner.os }}-${{ hashFiles('.mvx/mvx.properties') }}
          restore-keys: |
            mvx-binary-${{ runner.os }}-

      - name: Install tools
        run: ./mvx setup

      - name: Build project
        run: ./mvx mvn clean verify
```

### Advanced Caching Strategy

For even better performance, cache both tools and Maven dependencies:

```yaml
name: Build with mvx (Advanced Caching)
on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup mvx
        run: |
          curl -fsSL https://raw.githubusercontent.com/gnodet/mvx/main/install-mvx.sh | bash
          echo "$HOME/.mvx/bin" >> $GITHUB_PATH
      
      # Cache mvx tools
      - name: Cache mvx tools
        uses: actions/cache@v4
        with:
          path: ~/.mvx/tools
          key: mvx-tools-${{ runner.os }}-${{ hashFiles('.mvx/config.yml') }}
          restore-keys: |
            mvx-tools-${{ runner.os }}-
      
      # Cache Maven dependencies
      - name: Cache Maven dependencies
        uses: actions/cache@v4
        with:
          path: ~/.m2/repository
          key: maven-${{ runner.os }}-${{ hashFiles('**/pom.xml') }}
          restore-keys: |
            maven-${{ runner.os }}-
      
      - name: Install tools
        run: mvx setup
      
      - name: Build project
        run: mvx mvn clean verify
```

### Multi-Platform Builds

For projects that need to build on multiple platforms, repository-based installation is even more convenient:

```yaml
name: Multi-Platform Build (Repository-Based)
on: [push, pull_request]

jobs:
  build:
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]
    runs-on: ${{ matrix.os }}

    steps:
      - uses: actions/checkout@v4

      # Cache mvx tools (platform-specific)
      - name: Cache mvx tools
        uses: actions/cache@v4
        with:
          path: |
            ~/.mvx/tools
            %USERPROFILE%\.mvx\tools
          key: mvx-tools-${{ runner.os }}-${{ hashFiles('.mvx/mvx.properties', '.mvx/config.json5') }}
          restore-keys: |
            mvx-tools-${{ runner.os }}-

      # Cache mvx binaries (platform-specific)
      - name: Cache mvx binaries
        uses: actions/cache@v4
        with:
          path: |
            ~/.mvx/versions
            %USERPROFILE%\.mvx\versions
          key: mvx-binary-${{ runner.os }}-${{ hashFiles('.mvx/mvx.properties') }}
          restore-keys: |
            mvx-binary-${{ runner.os }}-

      - name: Install tools (Unix)
        if: runner.os != 'Windows'
        run: ./mvx setup

      - name: Install tools (Windows)
        if: runner.os == 'Windows'
        run: .\mvx.cmd setup

      - name: Build project (Unix)
        if: runner.os != 'Windows'
        run: ./mvx mvn clean verify

      - name: Build project (Windows)
        if: runner.os == 'Windows'
        run: .\mvx.cmd mvn clean verify
```

## Version Overrides in CI/CD

Use environment variable overrides to test different tool versions without modifying your configuration:

### Matrix Testing with Version Overrides

```yaml
name: Test Multiple Tool Versions
on: [push, pull_request]

jobs:
  test:
    strategy:
      matrix:
        java: [17, 21]
        maven: [3.9.6, 4.0.0-rc-4]
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - name: Cache mvx tools
        uses: actions/cache@v4
        with:
          path: ~/.mvx/tools
          key: mvx-tools-${{ runner.os }}-java${{ matrix.java }}-maven${{ matrix.maven }}

      - name: Test with specific versions
        env:
          MVX_JAVA_VERSION: ${{ matrix.java }}
          MVX_MAVEN_VERSION: ${{ matrix.maven }}
        run: |
          ./mvx setup
          ./mvx test
```

### Environment-Specific Versions

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      # Use Java 17 for tests, Java 21 for production builds
      - name: Run tests
        env:
          MVX_JAVA_VERSION: 17
        run: ./mvx test

      - name: Build for production
        env:
          MVX_JAVA_VERSION: 21
          MVX_MAVEN_VERSION: 4.0.0-rc-4
        run: ./mvx build
```

### Debugging CI Issues

```yaml
      - name: Debug with specific versions
        env:
          MVX_JAVA_VERSION: 21.0.2  # Use exact version for debugging
          MVX_VERBOSE: true          # Enable verbose logging
        run: ./mvx build
```

**Benefits:**
- 🧪 **Test compatibility** across multiple tool versions
- 🚀 **Stage-specific versions** (test vs production)
- 🐛 **Debug issues** with specific tool versions
- 📝 **No config changes** required

## Cache Key Strategies

### Tool Version-Based Caching

Cache based on tool versions in your configuration:

```yaml
- name: Cache mvx tools
  uses: actions/cache@v4
  with:
    path: ~/.mvx/tools
    key: mvx-tools-${{ runner.os }}-java${{ matrix.java }}-maven${{ matrix.maven }}
    restore-keys: |
      mvx-tools-${{ runner.os }}-java${{ matrix.java }}-
      mvx-tools-${{ runner.os }}-
```

### Date-Based Cache Invalidation

For tools that update frequently, include date in cache key:

```yaml
- name: Get current date
  id: date
  run: echo "date=$(date +'%Y-%m-%d')" >> $GITHUB_OUTPUT

- name: Cache mvx tools
  uses: actions/cache@v4
  with:
    path: ~/.mvx/tools
    key: mvx-tools-${{ runner.os }}-${{ steps.date.outputs.date }}
    restore-keys: |
      mvx-tools-${{ runner.os }}-
```

## Other CI/CD Platforms

### GitLab CI

```yaml
stages:
  - build

variables:
  MVX_CACHE_DIR: "$CI_PROJECT_DIR/.mvx"

cache:
  key: mvx-tools-$CI_RUNNER_OS
  paths:
    - .mvx/tools/

build:
  stage: build
  script:
    - curl -fsSL https://raw.githubusercontent.com/gnodet/mvx/main/install-mvx.sh | bash
    - export PATH="$HOME/.mvx/bin:$PATH"
    - mvx setup
    - mvx mvn clean verify
```

### Azure Pipelines

```yaml
trigger:
- main

pool:
  vmImage: 'ubuntu-latest'

variables:
  MVX_CACHE_DIR: $(Pipeline.Workspace)/.mvx

steps:
- task: Cache@2
  inputs:
    key: 'mvx-tools | "$(Agent.OS)" | .mvx/config.yml'
    restoreKeys: |
      mvx-tools | "$(Agent.OS)"
    path: $(MVX_CACHE_DIR)/tools
  displayName: 'Cache mvx tools'

- script: |
    curl -fsSL https://raw.githubusercontent.com/gnodet/mvx/main/install-mvx.sh | bash
    echo "##vso[task.prependpath]$HOME/.mvx/bin"
  displayName: 'Setup mvx'

- script: mvx setup
  displayName: 'Install tools'

- script: mvx mvn clean verify
  displayName: 'Build project'
```

## Performance Tips

### 1. **Selective Tool Installation**
Only install the tools you need:

```yaml
- name: Install specific tools
  run: |
    mvx tools install java --version 21
    mvx tools install maven --version 4.0.0-rc-4
```

### 2. **Parallel Tool Installation**
mvx installs tools in parallel by default, but you can control concurrency:

```yaml
- name: Install tools with custom concurrency
  run: mvx setup --max-concurrent 2
```

### 3. **Use System Tools When Available**
For faster builds, use system tools when they match your requirements:

```yaml
env:
  MVX_USE_SYSTEM_JAVA: true
  MVX_USE_SYSTEM_MAVEN: true
```

### 4. **Cache Validation**
Verify cache effectiveness by checking cache hit rates in your CI logs.

### 5. **Timeout Configuration**
For slow networks or servers (especially Apache servers), configure timeouts via environment variables:

```yaml
env:
  MVX_TLS_TIMEOUT: "5m"           # TLS handshake timeout (default: 2m)
  MVX_RESPONSE_TIMEOUT: "3m"      # Response header timeout (default: 2m)
  MVX_IDLE_TIMEOUT: "2m"          # Idle connection timeout (default: 90s)
  MVX_DOWNLOAD_TIMEOUT: "20m"     # Overall download timeout (default: 10m)
  MVX_CHECKSUM_TIMEOUT: "10m"     # Checksum verification timeout (default: 2m)
  MVX_REGISTRY_TIMEOUT: "8m"      # API registry timeout (default: 2m)
```

**Note**: The `MVX_TLS_TIMEOUT` is particularly important for Apache servers that can take 10+ seconds for TLS handshake.

### 6. **Download Retry Configuration**
For unreliable networks, configure retry behavior:

```yaml
env:
  MVX_MAX_RETRIES: "5"        # Maximum retry attempts (default: 3)
  MVX_RETRY_DELAY: "5s"       # Delay between retries (default: 2s)
```

## Troubleshooting

### Cache Misses
If you're experiencing frequent cache misses:

1. Check that your cache key is stable
2. Verify the cache path is correct
3. Ensure file permissions allow cache restoration

### Tool Download Failures
If tools fail to download in CI:

1. Check network connectivity
2. Verify the tool versions exist
3. Consider using fallback URLs (mvx 0.5.0+ automatically tries archive.apache.org)
4. **Increase timeouts for slow servers** (especially for TLS handshake timeouts):
   ```yaml
   env:
     MVX_TLS_TIMEOUT: "10m"       # For very slow TLS handshakes (Apache servers)
     MVX_IDLE_TIMEOUT: "5m"       # For very slow connection reuse
     MVX_DOWNLOAD_TIMEOUT: "30m"  # For very slow connections
     MVX_CHECKSUM_TIMEOUT: "15m"  # For slow checksum verification
   ```

### Platform-Specific Issues
- **Windows**: Use PowerShell for setup scripts
- **macOS**: Some tools may require additional permissions
- **Linux**: Different distributions may have different requirements

For more troubleshooting tips, see the [GitHub Issues](https://github.com/gnodet/mvx/issues).
