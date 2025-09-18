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
  
  // Command hooks (optional)
  hooks: {
    "pre-build": {
      description: "Run before build",
      script: "echo 'Starting build...'"
    },
    "post-build": {
      description: "Run after build",
      script: "echo 'Build completed!'"
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

### Platform-Specific Commands

```json5
{
  commands: {
    "start-db": {
      description: "Start database",
      script: {
        windows: "start-db.bat",
        unix: "./start-db.sh"
      }
    }
  }
}
```

## Hooks Section

Hooks allow you to run commands before or after built-in mvx commands:

```json5
{
  hooks: {
    "pre-setup": {
      description: "Prepare environment",
      script: "echo 'Preparing environment...'"
    },
    "post-setup": {
      description: "Verify installation",
      script: "echo 'Verifying tools...'"
    },
    "pre-build": {
      description: "Pre-build checks",
      script: "echo 'Running pre-build checks...'"
    }
  }
}
```

Available hook points:
- `pre-setup` / `post-setup`
- `pre-build` / `post-build`
- `pre-test` / `post-test`
- `pre-clean` / `post-clean`

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
