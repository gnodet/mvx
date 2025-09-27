---
title: MVX Interpreters
description: Choose the best execution environment for your commands with multiple interpreter support
layout: page
---

# MVX Interpreters

MVX supports multiple interpreters for executing commands, allowing you to choose the best execution environment for your specific needs.

## Available Interpreters

### 1. Native Interpreter (`native`)

The native interpreter uses the system's default shell (bash on Unix systems, cmd on Windows) to execute commands. This is the recommended interpreter for commands that:

- Use external tools and binaries
- Require complex shell features like pipes, redirects, or advanced scripting
- Need access to system-specific functionality

**Example:**
```json5
{
  commands: {
    "build": {
      description: "Build the project",
      script: "go build -o dist/myapp .",
      interpreter: "native"
    }
  }
}
```

### 2. MVX Shell Interpreter (`mvx-shell`)

The mvx-shell interpreter is a cross-platform command interpreter built into MVX. It provides:

- Cross-platform compatibility
- Built-in commands (cd, echo, mkdir, rm, cp)
- Environment variable expansion
- Basic command chaining with `&&`, `||`, and `;`

**Example:**
```json5
{
  commands: {
    "setup": {
      description: "Setup project directories",
      script: "mkdir -p dist && echo 'Directories created'",
      interpreter: "mvx-shell"
    }
  }
}
```

## Interpreter Selection

### Automatic Selection

MVX automatically selects the appropriate interpreter based on your script configuration:

- **Simple string scripts** default to `mvx-shell` for cross-platform compatibility
- **Platform-specific scripts** default to `native` for system integration

### Explicit Selection

You can explicitly specify the interpreter using the `interpreter` field:

```json5
{
  commands: {
    "format": {
      description: "Format code",
      script: "go fmt ./...",
      interpreter: "native"  // Explicitly use native interpreter
    },
    "clean": {
      description: "Clean build artifacts",
      script: "rm -rf dist && echo 'Cleaned'",
      interpreter: "mvx-shell"  // Explicitly use mvx-shell
    }
  }
}
```

## Environment Variables

Both interpreters support environment variables with different capabilities:

### Global Environment Variables

Set in the `environment` section of your configuration:

```json5
{
  environment: {
    BUILD_MODE: "production",
    LOG_LEVEL: "info"
  }
}
```

### Command-Specific Environment Variables

Set per command using the `environment` field:

```json5
{
  commands: {
    "test": {
      description: "Run tests",
      script: "go test ./...",
      interpreter: "native",
      environment: {
        GO_ENV: "test",
        VERBOSE: "true"
      }
    }
  }
}
```

### Variable Expansion

Both interpreters support comprehensive variable expansion with multiple syntax options:

#### Supported Syntax

- **Simple Variables**: `$VARIABLE_NAME`
- **Braced Variables**: `$\{VARIABLE_NAME\}` (useful when followed by other characters)

#### MVX-Shell Variable Expansion

The mvx-shell interpreter provides comprehensive variable expansion for all commands:

```json5
{
  commands: {
    "setup-project": {
      description: "Setup project structure",
      script: "mkdir $PROJECT_NAME && cd $PROJECT_NAME && echo 'Created ${PROJECT_NAME}_config.json'",
      environment: {
        PROJECT_NAME: "myapp"
      }
    },
    "deploy": {
      description: "Deploy application",
      script: "echo 'Deploying $PROJECT_NAME in $BUILD_MODE mode to ${DEPLOY_ENV}_server'",
      environment: {
        PROJECT_NAME: "myapp",
        BUILD_MODE: "production",
        DEPLOY_ENV: "staging"
      }
    }
  }
}
```

#### Variable Expansion Features

- **Command Names**: Variables are expanded in command names
- **Arguments**: Variables are expanded in all command arguments
- **Built-in Commands**: All mvx-shell built-in commands support variable expansion
- **External Commands**: Variables are expanded before executing external commands
- **Environment Override**: Command-specific environment variables override global ones

#### Examples

```bash
# These all work with variable expansion in mvx-shell:
mkdir $PROJECT_DIR                    # Create directory with variable name
echo "Hello $USER_NAME"              # Print with variable expansion
rm ${TEMP_DIR}/cache                 # Remove with braced variable syntax
cd $PROJECT_DIR && echo "In project" # Chain commands with variables
```

## PATH Management

MVX automatically manages the PATH environment variable to ensure mvx-managed tools are available:

### Tool PATH Integration

When you have tools configured in your `.mvx/config.json5`:

```json5
{
  tools: {
    go: { version: "1.24.2" },
    java: { version: "17", distribution: "zulu" },
    maven: { version: "3.9.11" }
  }
}
```

MVX automatically prepends the tool bin directories to PATH:
- `~/.mvx/tools/go/1.24.2/go/bin`
- `~/.mvx/tools/java/17-zulu/bin`
- `~/.mvx/tools/maven/3.9.11/bin`

### Interpreter-Specific PATH Handling

- **Native interpreter**: Full PATH integration with system shell
- **MVX-shell interpreter**: Limited PATH support for external commands

For commands that use external tools (like `go`, `mvn`, `node`), it's recommended to use the `native` interpreter to ensure proper PATH resolution.

## Best Practices

### When to Use Native Interpreter

Use `interpreter: "native"` for commands that:

- Execute external tools (`go`, `mvn`, `npm`, etc.)
- Use complex shell features (pipes, redirects, loops)
- Require system-specific functionality
- Need reliable PATH resolution

```json5
{
  commands: {
    "build": {
      script: "go build -o dist/app .",
      interpreter: "native"
    },
    "test": {
      script: "go test -v ./...",
      interpreter: "native"
    },
    "deps": {
      script: "go mod download && go mod tidy",
      interpreter: "native"
    }
  }
}
```

### When to Use MVX-Shell Interpreter

Use `interpreter: "mvx-shell"` for commands that:

- Use only built-in commands
- Need cross-platform compatibility
- Perform simple file operations
- Don't require external tools

```json5
{
  commands: {
    "clean": {
      script: "rm -rf dist && mkdir -p dist",
      interpreter: "mvx-shell"
    },
    "setup": {
      script: "mkdir -p src test docs && echo 'Project structure created'",
      interpreter: "mvx-shell"
    }
  }
}
```

## Debugging

### Verbose Logging

Enable verbose logging to debug interpreter and PATH issues:

```bash
MVX_VERBOSE=true mvx your-command
```

This will show:
- Environment setup details
- PATH modifications
- Interpreter selection
- Command execution details

### Common Issues

1. **"executable file not found" errors**: Usually indicates PATH issues. Use `interpreter: "native"` for external tools.

2. **Cross-platform compatibility**: Use `mvx-shell` for simple commands that need to work across different operating systems.

3. **Environment variable not found**: Check that variables are properly defined in global or command-specific environment sections.

## Migration Guide

If you're experiencing PATH issues with existing commands, add `interpreter: "native"` to commands that use external tools:

```json5
// Before (may have PATH issues)
{
  "fmt": {
    "script": "go fmt ./..."
  }
}

// After (fixed)
{
  "fmt": {
    "script": "go fmt ./...",
    "interpreter": "native"
  }
}
```
