---
title: Shell Command
description: Execute shell commands within the mvx environment with access to all managed tools
layout: page
---

# Shell Command

The `mvx shell` command allows you to execute shell commands within the mvx environment, with access to all mvx-managed tools and their environment variables.

## Basic Usage

```bash
# Show help
mvx shell --help

# Execute a simple command
mvx shell 'echo "Hello from mvx shell!"'

# Show environment variables
mvx shell 'echo $JAVA_HOME'
mvx shell 'env | grep JAVA'
```

## Examples with Tools

### Java
```bash
# Show Java version
mvx shell 'java -version'

# Show Java home directory
mvx shell 'echo $JAVA_HOME'

# Compile and run a simple Java program
mvx shell 'echo "public class Hello { public static void main(String[] args) { System.out.println(\"Hello from Java!\"); } }" > Hello.java'
mvx shell 'javac Hello.java && java Hello'
```

### Go
```bash
# Show Go version
mvx shell 'go version'

# Show Go environment
mvx shell 'go env GOROOT'
mvx shell 'go env GOPATH'

# Create and run a simple Go program
mvx shell 'echo "package main\nimport \"fmt\"\nfunc main() { fmt.Println(\"Hello from Go!\") }" > hello.go'
mvx shell 'go run hello.go'
```

### Maven
```bash
# Show Maven version
mvx shell 'mvn --version'

# Show Maven home
mvx shell 'echo $MAVEN_HOME'
```

## Complex Commands

### Command Chaining
```bash
# Multiple commands with &&
mvx shell 'mkdir test-project && cd test-project && pwd'

# Environment variable usage
mvx shell 'echo "Java is at: $JAVA_HOME" && echo "Go is at: $GOROOT"'
```

### File Operations
```bash
# Create directories and files
mvx shell 'mkdir -p src/main/java && echo "Created directory structure"'

# List files with details
mvx shell 'ls -la'

# Copy files
mvx shell 'cp file1.txt file2.txt'
```

## Key Features

1. **Environment Setup**: Automatically sets up PATH and environment variables for all configured tools
2. **Cross-platform**: Uses mvx-shell interpreter for consistent behavior across platforms
3. **Variable Expansion**: Supports `$VAR` and `${VAR}` syntax for environment variables
4. **Built-in Commands**: Includes cross-platform implementations of common commands (cd, echo, mkdir, rm, cp)
5. **Tool Integration**: All mvx-managed tools are available in the PATH

## Configuration

The shell command uses your project's `.mvx/config.json5` file to determine which tools to set up:

```json5
{
  project: {
    name: "my-project"
  },
  tools: {
    java: {
      version: "17",
      distribution: "zulu"
    },
    go: {
      version: "1.24.2"
    }
  }
}
```

## Tips

1. **Use quotes** for commands with arguments that might be interpreted as flags:
   ```bash
   mvx shell 'java -version'  # Good
   mvx shell java -version    # May fail due to flag parsing
   ```

2. **Environment variables** are expanded by the mvx-shell interpreter:
   ```bash
   mvx shell 'echo $JAVA_HOME'  # Works
   mvx shell "echo \$JAVA_HOME" # Also works but requires escaping
   ```

3. **Working directory** changes persist within the same command:
   ```bash
   mvx shell 'cd subdir && pwd'  # Shows subdir path
   ```

4. **Check available tools** by examining the PATH:
   ```bash
   mvx shell 'echo $PATH | tr ":" "\n" | grep mvx'
   ```

## Use Cases

This command is particularly useful for:

- **Testing tool installations**: Verify that tools are properly installed and configured
- **Environment debugging**: Check environment variables and PATH settings
- **Cross-platform scripting**: Run commands that work consistently across different operating systems
- **Tool integration**: Execute commands with the exact same environment that mvx custom commands use
- **Development workflows**: Quick access to mvx-managed tools without leaving your current shell

## Comparison with Custom Commands

While [custom commands](/commands) are great for defining reusable project-specific tasks, the shell command is perfect for:

- One-off commands and experimentation
- Debugging environment issues
- Testing tool configurations
- Running commands interactively

```bash
# Custom command (defined in .mvx/config.json5)
mvx build

# Shell command (ad-hoc execution)
mvx shell 'mvn clean install -DskipTests'
```
