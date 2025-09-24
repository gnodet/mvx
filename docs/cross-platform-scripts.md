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

mvx-shell supports comprehensive command chaining:

```json5
{
  script: "mkdir dist && copy target/*.jar dist/ || echo Build failed; echo Done",
  interpreter: "mvx-shell"
}
```

**Supported Operators:**
- `&&` - Execute next command only if previous succeeded
- `||` - Execute next command only if previous failed
- `;` - Execute commands sequentially regardless of success/failure
- `|` - Simple pipe support (sequential execution)
- `()` - Parentheses for grouping (basic support)

**Examples:**
```json5
{
  // Conditional execution
  script: "mvn test && echo Tests passed || echo Tests failed",
  interpreter: "mvx-shell"
}
```

```json5
{
  // Sequential execution
  script: "echo Starting; mvn clean; mvn compile; echo Done",
  interpreter: "mvx-shell"
}
```

```json5
{
  // Complex chaining
  script: "mkdir -p target && (mvn clean install || echo Build failed) && echo Complete",
  interpreter: "mvx-shell"
}
```

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

mvx uses **intelligent defaults** based on script type for maximum intuitiveness:

### Intelligent Defaults

- **Simple scripts**: Default to `mvx-shell` (cross-platform by nature)
- **Platform-specific scripts**: Default to `native` (platform-specific by nature)

This means you get the most appropriate interpreter automatically:

```json5
{
  commands: {
    // Automatically uses mvx-shell (cross-platform)
    "build": {
      script: "mkdir dist && copy target/*.jar dist/"
    },

    // Automatically uses native (platform-specific)
    "start-db": {
      script: {
        windows: "net start postgresql",
        unix: "sudo systemctl start postgresql"
      }
    }
  }
}
```

### Available Interpreters

#### native

Uses the system's native shell:
- **Unix/Linux/macOS**: `/bin/bash`
- **Windows**: `cmd`

Best for platform-specific commands that leverage system-specific features.

#### mvx-shell

Uses the built-in cross-platform interpreter with portable commands.

Best for cross-platform scripts that should work identically everywhere.

### Override Defaults

You can always override the intelligent defaults when needed:

```json5
{
  commands: {
    // Force native for simple script
    "system-specific": {
      script: "echo $HOME",  // Uses shell variables
      interpreter: "native"
    },

    // Force mvx-shell for platform-specific script
    "portable-platform": {
      script: {
        windows: "echo Starting on Windows",
        unix: "echo Starting on Unix"
      },
      interpreter: "mvx-shell"  // Use portable echo
    }
  }
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

- Pipe operator (`|`) provides sequential execution only (no actual piping yet)
- Parentheses provide basic grouping (no full subshell support)
- No variable substitution
- No conditional execution (`if`/`else` statements)
- No loops or functions

### Future Enhancements

- Full pipe support with input/output redirection
- Complete subshell support with parentheses
- Environment variable substitution
- Conditional execution (`if`/`else`)
- Loop constructs (`for`/`while`)
- Function definitions and calls
- Advanced file operations (glob patterns, permissions)
- File redirection (`>`, `>>`, `<`)
