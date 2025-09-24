---
title: Configuration
description: Complete guide to configuring mvx for your projects
layout: page
---

# Configuration Guide

mvx uses a JSON5 configuration file (`.mvx/config.json5`) to define your project's tools, commands, and settings. JSON5 allows comments, unquoted keys, trailing commas, and line continuation for better readability and maintainability.

## Configuration File Structure

```json5
{
  // Project metadata
  project: {
    name: "my-project",
    description: "Optional project description"
  },

  // Tool definitions
  tools: {
    maven: { version: "3.9.6" },
    java: { version: "21" },
    go: { version: "1.21.0" },
    node: { version: "20.0.0" }
  },

  // Custom commands
  commands: {
    build: {
      description: "Build the project",
      script: "mvn clean install"
    },
    test: {
      description: "Run tests",
      script: "mvn test"
    },
    // Multi-line scripts with line continuation
    "complex-build": {
      description: "Complex build with multiple steps",
      script: "mvn clean compile && \
               mvn test && \
               mvn package"
    }
  }
}
```

## JSON5 Features

mvx supports all standard JSON5 features for enhanced configuration readability:

### Comments
```json5
{
  // Single-line comments
  tools: {
    /* Multi-line comments
       for detailed explanations */
    maven: { version: "3.9.6" }
  }
}
```

### Unquoted Keys and Trailing Commas
```json5
{
  tools: {
    maven: { version: "3.9.6" },
    java: { version: "21" },  // Trailing comma allowed
  },
}
```

### Line Continuation
Use backslash (`\`) for multi-line strings:

```json5
{
  commands: {
    deploy: {
      description: "Deploy with multiple steps",
      script: "echo 'Starting deployment...' && \
               mvn clean package && \
               docker build -t myapp . && \
               docker push myapp:latest"
    }
  }
}
```

## Project Section

The `project` section contains metadata about your project:

```json5
{
  project: {
    name: "my-awesome-project",           // Required: project name
    description: "An awesome project",    // Optional: project description
    version: "1.0.0"                     // Optional: project version
  }
}
```

## Tools Section

The `tools` section defines which development tools your project needs:

### Basic Tool Configuration

```json5
{
  tools: {
    maven: { version: "3.9.6" },
    java: { version: "21" },
    go: { version: "1.21.0" },
    node: { version: "20.0.0" }
  }
}
```

### Advanced Tool Configuration

```json5
{
  tools: {
    java: {
      version: "21",
      distribution: "temurin",  // Optional: specify JDK distribution
      arch: "x64"               // Optional: specify architecture
    },
    maven: {
      version: "3.9.6",
      settings: "custom-settings.xml",  // Optional: custom settings file
      checksum: {                       // Optional: checksum verification
        type: "sha512",                 // sha256 or sha512
        required: true                  // Fail on verification errors
      }
    }
  }
}
```

### Security Configuration

Enable checksum verification for enhanced security:

```json5
{
  tools: {
    maven: {
      version: "3.9.6",
      checksum: {
        required: true  // Mandatory checksum verification
      }
    },
    mvnd: {
      version: "1.0.2",
      checksum: {
        type: "sha512",
        value: "abc123def456...",  // Custom checksum
        required: true
      }
    }
  }
}
```

## Maven Integration

mvx provides enhanced Maven wrapper functionality with transparent argument passing:

### Direct Maven Usage

You can use Maven directly through mvx without defining custom commands:

```bash
# All Maven commands work naturally
./mvx mvn clean install
./mvx mvn -V                    # Show version
./mvx mvn -X test              # Debug mode
./mvx mvn -Pproduction package # Profile activation

# Combine mvx global flags with Maven flags
./mvx --verbose mvn -V          # mvx verbose + Maven version
./mvx --quiet mvn test          # mvx quiet mode + Maven test
```

### Maven Configuration Options

```json5
{
  tools: {
    maven: {
      version: "3.9.6",                    // Maven version
      settings: "custom-settings.xml",     // Custom settings file (optional)
      checksum: {                          // Security verification (optional)
        required: true                     // Fail on checksum errors
      }
    }
  }
}
```

### Maven vs Custom Commands

You can use both approaches in the same project:

```json5
{
  tools: {
    maven: { version: "3.9.6" },
    java: { version: "21" }
  },
  commands: {
    // Custom commands for common workflows
    build: {
      description: "Standard build",
      script: "mvn clean install"
    },
    "quick-build": {
      description: "Build without tests",
      script: "mvn clean install -DskipTests"
    }
  }
}
```

**Usage:**
```bash
# Use custom commands for common workflows
./mvx build
./mvx quick-build

# Use Maven directly for ad-hoc commands
./mvx mvn dependency:tree
./mvx mvn versions:display-updates
./mvx mvn -Pintegration-tests verify
```

### Backward Compatibility

Scripts using the `--` separator continue to work:

```bash
# Old syntax (still works with migration warnings)
./mvx mvn -- -V
./mvx mvn -- clean install

# New syntax (recommended)
./mvx mvn -V
./mvx mvn clean install
```

## Commands Section

Define custom commands that become available as `./mvx <command>`:

### Basic Commands

```json5
{
  commands: {
    build: {
      description: "Build the project",
      script: "mvn clean install"
    },
    test: {
      description: "Run tests",
      script: "mvn test"
    },
    clean: {
      description: "Clean build artifacts",
      script: "mvn clean"
    }
  }
}
```

### Multi-Step Commands

```json5
{
  commands: {
    "full-build": {
      description: "Complete build with quality checks",
      script: [
        "mvn clean",
        "mvn compile",
        "mvn test",
        "mvn package"
      ]
    },
    // Alternative: single script with line continuation
    "full-build-single": {
      description: "Complete build as single command",
      script: "mvn clean && \
               mvn compile && \
               mvn test && \
               mvn package"
    }
  }
}
```

### Platform-Specific Commands

> **Note**: Platform-specific script syntax is not yet implemented. See [issue #21](https://github.com/gnodet/mvx/issues/21) for planned cross-platform script support.

For now, use conditional logic within scripts:

```json5
{
  commands: {
    "start-db": {
      description: "Start database (cross-platform)",
      script: "if [[ \"$OSTYPE\" == \"msys\" ]]; then ./start-db.bat; else ./start-db.sh; fi"
    }
  }
}
```

## Command Hooks

You can add pre and post hooks to built-in mvx commands by defining them within the command configuration:

```json5
{
  commands: {
    // Add hooks to the built-in 'build' command
    build: {
      description: "Build with custom hooks",
      pre: "echo 'Starting build...'",
      post: "echo 'Build completed!'"
    },

    // Add hooks to the built-in 'test' command
    test: {
      description: "Test with verification",
      pre: "echo 'Preparing test environment...'",
      post: [
        "echo 'Tests completed!'",
        "echo 'Generating reports...'"
      ]
    }
  }
}
```

Available hook points for built-in commands:
- `build` - Build command hooks
- `test` - Test command hooks
- `setup` - Setup command hooks
- `clean` - Clean command hooks

## Command Overrides

You can override built-in mvx commands:

```json5
{
  commands: {
    // Override the built-in 'setup' command
    setup: {
      description: "Custom setup with additional steps",
      script: [
        "echo 'Running custom setup...'",
        "mvx tools install",
        "echo 'Setup completed!'"
      ]
    }
  }
}
```

## Environment Variables

### Custom Environment Variables

Set environment variables for your commands:

```json5
{
  environment: {
    JAVA_OPTS: "-Xmx2g -Xms1g",
    MAVEN_OPTS: "-Xmx1g",
    NODE_ENV: "development"
  },
  commands: {
    build: {
      description: "Build with custom environment",
      script: "mvn clean install",
      environment: {
        MAVEN_OPTS: "-Xmx2g"  // Override global setting
      }
    }
  }
}
```

### mvx System Environment Variables

mvx recognizes several environment variables to control its behavior:

#### Timeout Configuration

Configure timeouts for downloads and network operations (useful for slow networks or CI/CD):

```bash
# TLS handshake timeout (default: 2 minutes)
export MVX_TLS_TIMEOUT="5m"

# Response header timeout (default: 2 minutes)
export MVX_RESPONSE_TIMEOUT="3m"

# Idle connection timeout (default: 90 seconds)
export MVX_IDLE_TIMEOUT="2m"

# Overall download timeout (default: 10 minutes)
export MVX_DOWNLOAD_TIMEOUT="20m"

# Checksum verification timeout (default: 2 minutes)
export MVX_CHECKSUM_TIMEOUT="10m"

# API registry timeout (default: 2 minutes)
export MVX_REGISTRY_TIMEOUT="8m"
```

#### Download Retry Configuration

Configure retry behavior for failed downloads:

```bash
# Maximum number of download retries (default: 3)
export MVX_MAX_RETRIES="5"

# Delay between retry attempts (default: 2 seconds)
export MVX_RETRY_DELAY="5s"
```

**When to use timeout configuration:**
- **Slow networks**: Increase timeouts in environments with poor connectivity
- **CI/CD systems**: Configure longer timeouts for reliable builds
- **Corporate networks**: Handle proxy delays and security scanning
- **Apache servers**: Some Apache servers (like archive.apache.org) can be slow

#### Development Version Control

Control which version of mvx to use:

```bash
# Use development version (built locally as mvx-dev)
export MVX_VERSION=dev

# Use specific release version
export MVX_VERSION=0.3.0

# Use latest stable version (default)
export MVX_VERSION=latest
```

**Development Binary Priority:**
- Development binaries (`mvx-dev`) are only used when `MVX_VERSION=dev` is explicitly set
- Specific versions (e.g., `MVX_VERSION=0.3.0`) always take precedence over local development binaries
- This ensures consistent behavior across environments

**Development Binary Locations:**
1. Project-specific: `./mvx-dev` (current directory)
2. Global development: `~/.mvx/dev/mvx` (shared across projects)

#### Version Overrides

Override tool versions specified in your configuration file using environment variables:

```bash
# Override Java version (uses Java 21 instead of config version)
export MVX_JAVA_VERSION=21

# Override Maven version (uses Maven 3.9.6 instead of config version)
export MVX_MAVEN_VERSION=3.9.6

# Override Go version
export MVX_GO_VERSION=1.21.0

# Override Node.js version
export MVX_NODE_VERSION=20.0.0

# Override Python version
export MVX_PYTHON_VERSION=3.12.0

# Use multiple overrides together
export MVX_JAVA_VERSION=21
export MVX_MAVEN_VERSION=4.0.0-rc-4
mvx setup
```

**How version overrides work:**
- ✅ **Takes precedence**: Environment variables override configuration file settings
- ✅ **Temporary**: Only affects the current command execution
- ✅ **Flexible**: Supports both exact versions (e.g., "21.0.2") and version patterns (e.g., "21")
- ✅ **Verbose logging**: Use `MVX_VERBOSE=true` to see when overrides are active
- ✅ **Version resolution**: Patterns like "21" are resolved to latest available (e.g., "21.0.2")

**Use cases:**
- **Testing**: Test your project with different tool versions
- **CI/CD**: Use specific versions in different pipeline stages
- **Debugging**: Quickly switch versions to isolate issues
- **Team coordination**: Temporarily align on specific versions

#### Other System Variables

```bash
# Force use of system-installed tools (no version checking, no fallback)
export MVX_USE_SYSTEM_JAVA=true   # Implemented
export MVX_USE_SYSTEM_MAVEN=true  # Implemented
export MVX_USE_SYSTEM_NODE=true   # Coming soon
export MVX_USE_SYSTEM_GO=true     # Coming soon
export MVX_USE_SYSTEM_PYTHON=true # Coming soon

# Control parallel downloads (default: 4)
export MVX_PARALLEL_DOWNLOADS=2

# Enable verbose logging
export MVX_VERBOSE=true

# Disable color output
export MVX_NO_COLOR=true
```

## Working Directory

Specify working directory for commands:

```json5
{
  commands: {
    "build-frontend": {
      description: "Build frontend",
      script: "npm run build",
      workingDirectory: "frontend"
    },
    "build-backend": {
      description: "Build backend",
      script: "mvn clean install",
      workingDirectory: "backend"
    }
  }
}
```

## Configuration Examples

### Full-Stack Project

```json5
{
  project: {
    name: "fullstack-app",
    description: "Full-stack application with React frontend and Spring Boot backend"
  },
  tools: {
    java: { version: "21" },
    maven: { version: "3.9.6" },
    node: { version: "20.0.0" }
  },
  commands: {
    "build-all": {
      description: "Build frontend and backend",
      script: [
        "cd frontend && npm run build",
        "cd backend && mvn clean install"
      ]
    },
    "dev-frontend": {
      description: "Start frontend development server",
      script: "npm run dev",
      workingDirectory: "frontend"
    },
    "dev-backend": {
      description: "Start backend development server",
      script: "mvn spring-boot:run",
      workingDirectory: "backend"
    }
  }
}
```

### Microservices Project

```json5
{
  project: {
    name: "microservices-platform"
  },
  tools: {
    java: { version: "21" },
    maven: { version: "3.9.6" },
    go: { version: "1.21.0" }
  },
  commands: {
    "build-all": {
      description: "Build all services",
      script: [
        "cd user-service && mvn clean install",
        "cd order-service && mvn clean install",
        "cd notification-service && go build"
      ]
    },
    "test-all": {
      description: "Test all services",
      script: [
        "cd user-service && mvn test",
        "cd order-service && mvn test",
        "cd notification-service && go test ./..."
      ]
    }
  }
}
```

## Best Practices

1. **Use descriptive command names**: `build-frontend` instead of `bf`
2. **Add descriptions**: Help team members understand what commands do
3. **Group related commands**: Use consistent naming patterns
4. **Version lock tools**: Specify exact versions for reproducibility
5. **Document complex commands**: Use comments in JSON5 format
6. **Test on all platforms**: Ensure commands work on Windows, macOS, and Linux

## Next Steps

- [Explore supported tools](/tools)
- [Learn about custom commands](/commands)
- [Check out examples on GitHub](https://github.com/gnodet/mvx/tree/main/examples)
