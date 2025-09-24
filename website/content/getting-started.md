---
title: Getting Started
description: Learn how to install and use mvx in your projects
layout: page
---

# Getting Started with mvx

mvx is a universal build environment bootstrap tool that automatically manages development tools for your projects. Think of it as "Maven Wrapper for the modern era" - but it works with any technology stack.

## Installation

### Quick Install (Recommended)

Download and install mvx globally:

```bash
curl -fsSL https://raw.githubusercontent.com/gnodet/mvx/main/install-mvx.sh | sh
```

This installs mvx to `~/.mvx/bin/mvx` and adds it to your PATH.

### Development Version

Install the latest development version (may be unstable):

```bash
curl -fsSL https://raw.githubusercontent.com/gnodet/mvx/main/install-mvx.sh | MVX_VERSION=dev bash
```

### Manual Installation

1. Download the latest release for your platform from [GitHub Releases](https://github.com/gnodet/mvx/releases)
2. Extract and place the binary in your PATH
3. Make it executable: `chmod +x mvx`

## First Steps

### 1. Initialize Your Project

```bash
# Create bootstrap scripts
mvx init

# This creates:
# - ./mvx (Unix/Linux/macOS)
# - ./mvx.cmd (Windows)
# - .mvx/config.json5 (configuration)
```

### 2. Configure Your Tools

#### Option A: Add Tools Using Commands (Recommended)

Use the `mvx tools add` command to easily add tools:

```bash
# Add tools one by one
mvx tools add java 21
mvx tools add maven 3.9.6
mvx tools add node lts

# Add Java with specific distribution
mvx tools add java 17 zulu
```

#### Option B: Edit Configuration Manually

Edit `.mvx/config.json5` to specify your project's requirements:

```json5
{
  project: {
    name: "my-project"
  },
  tools: {
    maven: {
      version: "3.9.6"
    },
    java: {
      version: "21"
    }
  },
  commands: {
    build: {
      description: "Build the project",
      script: "mvn clean install"
    },
    test: {
      description: "Run tests",
      script: "mvn test"
    }
  }
}
```

### 3. Setup Your Environment

```bash
# Install all configured tools
./mvx setup

# This downloads and installs:
# - Maven 3.9.6
# - Java 21
# - Sets up environment variables
```

### 4. Start Building

```bash
# Run your custom commands
./mvx build
./mvx test

# Use Maven directly with natural syntax
./mvx mvn -V                    # Show Maven version
./mvx mvn clean install         # Standard Maven build
./mvx mvn -X test              # Debug mode testing

# Combine mvx and Maven flags
./mvx --verbose mvn -V          # mvx verbose + Maven version

# Or use built-in commands
./mvx tools list
./mvx version
```

## Common Examples

### Java/Maven Project

```json5
{
  project: {
    name: "my-java-app"
  },
  tools: {
    maven: { version: "3.9.6" },
    java: { version: "21" }
  },
  commands: {
    build: {
      description: "Build with tests",
      script: "mvn clean install"
    },
    "quick-build": {
      description: "Build without tests",
      script: "mvn clean install -DskipTests"
    },
    run: {
      description: "Run the application",
      script: "mvn spring-boot:run"
    }
  }
}
```

**Usage Examples:**
```bash
# Use custom commands
./mvx build
./mvx quick-build
./mvx run

# Or use Maven directly with natural syntax
./mvx mvn -V                    # Show Maven version
./mvx mvn clean install         # Standard build
./mvx mvn -Pproduction package  # Production build with profile
./mvx mvn spring-boot:run       # Run Spring Boot app

# Combine mvx and Maven flags for debugging
./mvx --verbose mvn -X test     # mvx verbose + Maven debug mode
```

### Secure Production Project

For production environments, enable checksum verification:

```json5
{
  project: {
    name: "secure-production-app"
  },
  tools: {
    maven: {
      version: "3.9.6",
      checksum: {
        required: true  // Fail if checksum verification fails
      }
    },
    java: { version: "21" },
    mvnd: {
      version: "1.0.2",
      checksum: {
        required: true  // Enhanced security for build tools
      }
    }
  },
  commands: {
    "secure-build": {
      description: "Build with security verification",
      script: "mvn clean verify \
               -Dsecurity.check=true \
               -Dchecksum.verify=true && \
               echo Security verification complete"
    }
  }
}
```

### Go Project

```json5
{
  project: {
    name: "my-go-app"
  },
  tools: {
    go: { version: "1.21.0" }
  },
  commands: {
    build: {
      description: "Build Go binary with optimization",
      script: "mkdir -p bin && \
               go build -ldflags='-s -w' -o bin/app . && \
               echo Binary built successfully"
    },
    test: {
      description: "Run tests with coverage",
      script: "go test -v -cover ./... && \
               go test -race ./... && \
               echo All tests passed"
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
  project: {
    name: "my-node-app"
  },
  tools: {
    node: { version: "20.0.0" }
  },
  commands: {
    build: {
      description: "Build for production",
      script: "npm ci && \
               npm run build && \
               echo Production build complete"
    },
    dev: {
      description: "Start development server with hot reload",
      script: "npm install && \
               npm run dev && \
               echo Development server started"
    },
    test: {
      description: "Run comprehensive tests",
      script: "npm run test:unit && \
               npm run test:integration && \
               echo All tests completed"
    }
  }
}
```

## Cross-Platform Scripts

mvx provides powerful cross-platform script support, making your commands work seamlessly across Windows, Linux, and macOS:

### Platform-Specific Scripts

```json5
{
  commands: {
    "start-db": {
      description: "Start database service",
      script: {
        windows: "net start postgresql",
        linux: "sudo systemctl start postgresql",
        darwin: "brew services start postgresql",
        default: "echo 'Please start PostgreSQL manually'"
      }
    }
  }
}
```

### Cross-Platform Interpreter

```json5
{
  commands: {
    "build-all": {
      description: "Build all modules",
      script: "cd frontend && npm run build && cd ../backend && mvn clean install || echo Build failed"
      // No interpreter specified - automatically uses mvx-shell (cross-platform)
    }
  }
}
```

**Learn more**: [Cross-Platform Scripts Guide](/cross-platform-scripts)

## Key Benefits

- **🚀 Zero Dependencies**: No need to install tools on your system
- **🌍 Cross-Platform**: Works on Linux, macOS, and Windows
- **🔧 Universal Tools**: One tool to manage Maven, Go, Node.js, and more
- **📦 Version Isolation**: Each project specifies its own tool versions
- **⚡ Fast Setup**: New team members can start building immediately

## Next Steps

- [Learn about configuration options](/configuration)
- [Write cross-platform scripts](/cross-platform-scripts)
- [Explore supported tools](/tools)
- [Discover custom commands](/commands)
- [Use version overrides for testing](/configuration#version-overrides)
- [Check out the GitHub repository](https://github.com/gnodet/mvx)

## Need Help?

- 📖 [Configuration Guide](/configuration)
- 🌍 [Cross-Platform Scripts](/cross-platform-scripts)
- 🔧 [Supported Tools](/tools)
- 💬 [GitHub Discussions](https://github.com/gnodet/mvx/discussions)
- 🐛 [Report Issues](https://github.com/gnodet/mvx/issues)
