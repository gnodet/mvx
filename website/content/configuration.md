---
title: Configuration
description: Complete guide to configuring mvx for your projects
layout: page
---

# Configuration Guide

mvx uses a JSON5 configuration file (`.mvx/config.json5`) to define your project's tools, commands, and settings. JSON5 allows comments and unquoted keys for better readability.

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
    }
  },
  

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
    }
  }
}
```

### Cross-Platform Scripts

mvx provides powerful cross-platform script support with two approaches:

#### Platform-Specific Scripts

Define different scripts for different operating systems:

```json5
{
  commands: {
    "start-db": {
      description: "Start database service",
      script: {
        windows: "net start postgresql",
        linux: "sudo systemctl start postgresql",
        darwin: "brew services start postgresql",
        unix: "echo 'Please start PostgreSQL manually'",  // Fallback for Unix-like systems
        default: "echo 'Platform not supported'"          // Final fallback
      }
    }
  }
}
```

**Platform Resolution Order:**
1. Exact platform match (`windows`, `linux`, `darwin`)
2. Platform family (`unix` for Linux/macOS)
3. `default` fallback
4. Error if no match found

#### Cross-Platform Interpreter (mvx-shell)

Use the built-in `mvx-shell` interpreter for truly portable scripts:

```json5
{
  commands: {
    "build-all": {
      description: "Build all modules",
      script: "cd frontend && npm run build && cd ../backend && mvn clean install -DskipTests",
      interpreter: "mvx-shell"  // Cross-platform interpreter
    },

    "open-results": {
      description: "Open build results",
      script: "open target/",     // Works on Windows, macOS, and Linux
      interpreter: "mvx-shell"
    },

    "setup-dev": {
      description: "Setup development environment",
      script: "mkdir -p logs temp && copy .env.example .env",
      interpreter: "mvx-shell"
    }
  }
}
```

**mvx-shell Commands:**
- `cd <dir>` - Change directory (cross-platform)
- `mkdir <dir>` - Create directories (with `-p` behavior)
- `copy <src> <dst>` - Copy files
- `rm <path>` - Remove files/directories
- `echo <text>` - Print text
- `open <path>` - Open files/directories (platform-appropriate)
- `<tool> <args>` - Execute any external command

**Command Chaining:**
- `&&` - Execute next command only if previous succeeded

#### Mixed Approach

Combine both approaches for maximum flexibility:

```json5
{
  commands: {
    "dev-setup": {
      description: "Setup development environment",
      script: {
        windows: {
          script: "mkdir logs && copy config\\dev.properties config\\app.properties",
          interpreter: "mvx-shell"
        },
        unix: {
          script: "mkdir -p logs && cp config/dev.properties config/app.properties",
          interpreter: "native"  // Use system shell
        }
      }
    }
  }
}
```

#### Interpreter Options

- **`native`** (default): Use system shell (`/bin/bash` on Unix, `cmd` on Windows)
- **`mvx-shell`**: Use built-in cross-platform interpreter

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

### Cross-Platform Hooks

Hooks also support cross-platform scripts and interpreters:

```json5
{
  commands: {
    "test": {
      description: "Run tests with cross-platform setup and cleanup",
      pre: {
        script: "mkdir -p test-results && echo Preparing test environment",
        interpreter: "mvx-shell"
      },
      post: {
        script: "echo Tests completed && open test-results/",
        interpreter: "mvx-shell"
      }
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
