---
title: Commands
description: Learn about built-in and custom commands in mvx
layout: page
---

# Commands

mvx provides both built-in commands for tool management and the ability to define custom commands specific to your project.

## Built-in Commands

### Project Management

```bash
# Initialize mvx in current directory
mvx init

# Show mvx version
mvx version

# Show help
mvx help
mvx --help
```

### Tool Management

```bash
# Install all configured tools
./mvx setup

# List all supported tools
./mvx tools list

# Search for tools
./mvx tools search java
./mvx tools search maven

# Show available versions for a tool
./mvx tools versions java

# Install specific tool
./mvx tools install java
./mvx tools install maven

# Show tool information
./mvx tools info java

# Verify tool installation
./mvx tools verify java

# Uninstall tool
./mvx tools uninstall java
```

### Environment Management

```bash
# Show environment information
./mvx env

# Show tool paths
./mvx env paths

# Export environment variables
./mvx env export

# Clean tool cache
./mvx clean cache

# Clean all tools
./mvx clean tools
```

## Custom Commands

Define custom commands in your `.mvx/config.json5` file. These become available as top-level commands.

### Basic Custom Commands

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

Usage:
```bash
./mvx build
./mvx test
./mvx clean
```

### Multi-Step Commands

Commands can execute multiple steps:

```json5
{
  commands: {
    "full-build": {
      description: "Complete build with quality checks",
      script: [
        "echo 'Starting full build...'",
        "mvn clean",
        "mvn compile",
        "mvn test",
        "mvn package",
        "echo 'Build completed!'"
      ]
    }
  }
}
```

### Platform-Specific Commands

> **Note**: Platform-specific script syntax is not yet implemented. See [issue #21](https://github.com/gnodet/mvx/issues/21) for planned cross-platform script support.

For now, use platform detection within scripts:

```json5
{
  commands: {
    "open-logs": {
      description: "Open log directory",
      script: "if [[ \"$OSTYPE\" == \"msys\" || \"$OSTYPE\" == \"win32\" ]]; then explorer logs; elif [[ \"$OSTYPE\" == \"darwin\"* ]]; then open logs; else xdg-open logs; fi"
    }
  }
}
```

### Commands with Working Directory

Execute commands in specific directories:

```json5
{
  commands: {
    "build-frontend": {
      description: "Build frontend application",
      script: "npm run build",
      workingDirectory: "frontend"
    },
    "build-backend": {
      description: "Build backend services",
      script: "mvn clean install",
      workingDirectory: "backend"
    }
  }
}
```

### Commands with Environment Variables

Set environment variables for commands:

```json5
{
  commands: {
    "build-prod": {
      description: "Build for production",
      script: "mvn clean install",
      environment: {
        MAVEN_OPTS: "-Xmx2g",
        SPRING_PROFILES_ACTIVE: "production"
      }
    },
    "dev-server": {
      description: "Start development server",
      script: "npm run dev",
      environment: {
        NODE_ENV: "development",
        PORT: "3000"
      }
    }
  }
}
```

## Command Hooks

Run commands before or after built-in mvx commands:

```json5
{
  hooks: {
    "pre-setup": {
      description: "Prepare environment before setup",
      script: "echo 'Preparing environment...'"
    },
    "post-setup": {
      description: "Verify installation after setup",
      script: [
        "echo 'Verifying tools...'",
        "./mvx tools verify java",
        "./mvx tools verify maven"
      ]
    },
    "pre-build": {
      description: "Pre-build checks",
      script: "echo 'Running pre-build checks...'"
    },
    "post-build": {
      description: "Post-build actions",
      script: "echo 'Build completed successfully!'"
    }
  }
}
```

Available hook points:
- `pre-setup` / `post-setup`
- `pre-build` / `post-build` (for custom build commands)
- `pre-test` / `post-test` (for custom test commands)
- `pre-clean` / `post-clean`

## Command Overrides

Override built-in mvx commands with custom implementations:

```json5
{
  commands: {
    // Override the built-in 'setup' command
    setup: {
      description: "Custom setup with additional steps",
      script: [
        "echo 'Running custom setup...'",
        "mvx tools install",
        "echo 'Installing additional dependencies...'",
        "npm install -g typescript",
        "echo 'Setup completed!'"
      ]
    },
    
    // Override the built-in 'clean' command
    clean: {
      description: "Custom clean with extra cleanup",
      script: [
        "mvx clean tools",
        "rm -rf node_modules",
        "rm -rf target",
        "echo 'Deep clean completed!'"
      ]
    }
  }
}
```

## Command Examples

### Java/Maven Project

```json5
{
  commands: {
    compile: {
      description: "Compile source code",
      script: "mvn compile"
    },
    test: {
      description: "Run unit tests",
      script: "mvn test"
    },
    "integration-test": {
      description: "Run integration tests",
      script: "mvn verify -P integration-tests"
    },
    package: {
      description: "Package application",
      script: "mvn package -DskipTests"
    },
    run: {
      description: "Run Spring Boot application",
      script: "mvn spring-boot:run"
    },
    "docker-build": {
      description: "Build Docker image",
      script: "docker build -t myapp:latest ."
    }
  }
}
```

### Go Project

```json5
{
  commands: {
    build: {
      description: "Build Go binary",
      script: "go build -o bin/app ."
    },
    test: {
      description: "Run tests with coverage",
      script: "go test -v -cover ./..."
    },
    "test-race": {
      description: "Run tests with race detection",
      script: "go test -race ./..."
    },
    fmt: {
      description: "Format Go code",
      script: "go fmt ./..."
    },
    vet: {
      description: "Run go vet",
      script: "go vet ./..."
    },
    mod: {
      description: "Download dependencies",
      script: "go mod download"
    },
    run: {
      description: "Run the application",
      script: "go run ."
    }
  }
}
```

### Node.js Project

```json5
{
  commands: {
    install: {
      description: "Install dependencies",
      script: "npm install"
    },
    build: {
      description: "Build for production",
      script: "npm run build"
    },
    dev: {
      description: "Start development server",
      script: "npm run dev"
    },
    test: {
      description: "Run tests",
      script: "npm test"
    },
    "test-watch": {
      description: "Run tests in watch mode",
      script: "npm run test:watch"
    },
    lint: {
      description: "Run ESLint",
      script: "npm run lint"
    },
    "lint-fix": {
      description: "Fix ESLint issues",
      script: "npm run lint:fix"
    }
  }
}
```

### Full-Stack Project

```json5
{
  commands: {
    "install-all": {
      description: "Install all dependencies",
      script: [
        "cd frontend && npm install",
        "cd backend && mvn dependency:resolve"
      ]
    },
    "build-all": {
      description: "Build frontend and backend",
      script: [
        "cd frontend && npm run build",
        "cd backend && mvn clean install"
      ]
    },
    "dev-frontend": {
      description: "Start frontend dev server",
      script: "npm run dev",
      workingDirectory: "frontend"
    },
    "dev-backend": {
      description: "Start backend dev server",
      script: "mvn spring-boot:run",
      workingDirectory: "backend"
    },
    "test-all": {
      description: "Run all tests",
      script: [
        "cd frontend && npm test",
        "cd backend && mvn test"
      ]
    }
  }
}
```

## Command Best Practices

1. **Use descriptive names**: `build-frontend` instead of `bf`
2. **Add descriptions**: Help team members understand what commands do
3. **Group related commands**: Use consistent naming patterns
4. **Keep commands simple**: Break complex workflows into multiple commands
5. **Use working directories**: Keep commands focused on specific parts of your project
6. **Document complex commands**: Use comments in JSON5 format
7. **Test on all platforms**: Ensure commands work on Windows, macOS, and Linux

## Command Discovery

```bash
# List all available commands
./mvx help

# Show command details
./mvx help build
./mvx help test
```

## Next Steps

- [Learn about configuration](/configuration)
- [Explore supported tools](/tools)
- [Check out examples](https://github.com/gnodet/mvx/tree/main/examples)
