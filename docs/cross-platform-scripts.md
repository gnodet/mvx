# Cross-Platform Scripts

mvx provides comprehensive cross-platform script support, allowing you to write commands that work seamlessly across Windows, Linux, and macOS.

## Overview

There are two main approaches to cross-platform scripts in mvx:

1. **Platform-Specific Scripts**: Define different scripts for different operating systems
2. **Cross-Platform Interpreter (mvx-shell)**: Use a built-in interpreter that provides portable commands

## Platform-Specific Scripts

### Basic Syntax

```json5
{
  commands: {
    "command-name": {
      description: "Command description",
      script: {
        windows: "Windows-specific command",
        linux: "Linux-specific command", 
        darwin: "macOS-specific command",
        unix: "Unix-like systems fallback",
        default: "Final fallback for any platform"
      }
    }
  }
}
```

### Platform Resolution

mvx resolves platform-specific scripts in the following order:

1. **Exact platform match**: `windows`, `linux`, `darwin`
2. **Platform family**: `unix` (matches Linux and macOS)
3. **Default fallback**: `default`
4. **Error**: If no match is found

### Examples

#### Database Service Management

```json5
{
  commands: {
    "start-db": {
      description: "Start PostgreSQL database",
      script: {
        windows: "net start postgresql",
        linux: "sudo systemctl start postgresql",
        darwin: "brew services start postgresql",
        default: "echo 'Please start PostgreSQL manually'"
      }
    },
    
    "stop-db": {
      description: "Stop PostgreSQL database", 
      script: {
        windows: "net stop postgresql",
        unix: "sudo systemctl stop postgresql || brew services stop postgresql",
        default: "echo 'Please stop PostgreSQL manually'"
      }
    }
  }
}
```

#### File Operations

```json5
{
  commands: {
    "clean-temp": {
      description: "Clean temporary files",
      script: {
        windows: "rmdir /s /q temp 2>nul || echo No temp directory",
        unix: "rm -rf temp || echo No temp directory"
      }
    }
  }
}
```

## Cross-Platform Interpreter (mvx-shell)

The `mvx-shell` interpreter provides a set of built-in commands that work identically across all platforms.

### Basic Usage

```json5
{
  commands: {
    "build-all": {
      description: "Build all modules",
      script: "cd frontend && npm run build && cd ../backend && mvn clean install",
      interpreter: "mvx-shell"
    }
  }
}
```

### Built-in Commands

#### Directory Operations

- `cd <directory>` - Change current directory
- `mkdir <directory>` - Create directory (creates parent directories as needed)

```json5
{
  script: "cd src && mkdir -p build/output",
  interpreter: "mvx-shell"
}
```

#### File Operations

- `copy <source> <destination>` - Copy files
- `rm <path>` - Remove files or directories

```json5
{
  script: "copy README.md dist/ && rm temp/",
  interpreter: "mvx-shell"
}
```

#### Output

- `echo <text>` - Print text to console

```json5
{
  script: "echo Starting build process",
  interpreter: "mvx-shell"
}
```

#### Platform-Specific Operations

- `open <path>` - Open file or directory with default application
  - Windows: Uses `explorer`
  - macOS: Uses `open`
  - Linux: Uses `xdg-open`

```json5
{
  script: "open target/site/",
  interpreter: "mvx-shell"
}
```

#### External Commands

Any command not recognized as a built-in is executed as an external command:

```json5
{
  script: "mvn clean install && npm test",
  interpreter: "mvx-shell"
}
```

### Command Chaining

mvx-shell supports command chaining with `&&`:

```json5
{
  script: "mkdir dist && copy target/*.jar dist/ && echo Build complete",
  interpreter: "mvx-shell"
}
```

**Note**: Additional operators (`||`, `;`, pipes) are planned for future releases.

## Mixed Approach

You can combine platform-specific scripts with interpreter selection:

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
          interpreter: "native"
        }
      }
    }
  }
}
```

## Interpreter Options

### native (default)

Uses the system's native shell:
- **Unix/Linux/macOS**: `/bin/bash`
- **Windows**: `cmd`

```json5
{
  script: "echo Hello World",
  interpreter: "native"  // Optional, this is the default
}
```

### mvx-shell

Uses the built-in cross-platform interpreter:

```json5
{
  script: "mkdir dist && copy *.jar dist/",
  interpreter: "mvx-shell"
}
```

## Built-in Command Integration

Cross-platform scripts work seamlessly with built-in command hooks:

### Pre/Post Hooks

```json5
{
  commands: {
    "test": {
      description: "Run tests with setup and cleanup",
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

### Command Overrides

```json5
{
  commands: {
    "build": {
      description: "Custom build process",
      script: "echo Starting custom build && mvn clean install && echo Build complete",
      interpreter: "mvx-shell",
      override: true
    }
  }
}
```

## Best Practices

### 1. Choose the Right Approach

- **Use platform-specific scripts** when you need to leverage platform-specific tools or commands
- **Use mvx-shell** for simple file operations and cross-platform portability
- **Mix both approaches** when you need platform-specific logic with portable operations

### 2. Fallback Strategy

Always provide fallbacks for maximum compatibility:

```json5
{
  script: {
    windows: "specific-windows-command",
    unix: "specific-unix-command", 
    default: "echo 'Please run manually: <instructions>'"
  }
}
```

### 3. Error Handling

Use command chaining to handle errors gracefully:

```json5
{
  script: "mkdir dist && copy target/*.jar dist/ || echo 'Build failed'",
  interpreter: "mvx-shell"
}
```

### 4. Documentation

Always include clear descriptions for your commands:

```json5
{
  commands: {
    "deploy": {
      description: "Deploy application to staging environment",
      script: "...",
      interpreter: "mvx-shell"
    }
  }
}
```

## Examples

See the [examples/cross-platform-config.json5](../examples/cross-platform-config.json5) file for comprehensive examples of cross-platform script usage.

## Limitations

### Current Limitations

- Command chaining only supports `&&` (more operators planned)
- No variable substitution (planned for future releases)
- No conditional execution (planned for future releases)
- No loops or functions (planned for future releases)

### Future Enhancements

- Support for `||`, `;`, and pipe operators
- Environment variable substitution
- Conditional execution (`if`/`else`)
- Loop constructs (`for`/`while`)
- Function definitions and calls
- Advanced file operations (glob patterns, permissions)
