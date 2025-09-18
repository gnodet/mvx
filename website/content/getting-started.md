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
curl -fsSL https://github.com/gnodet/mvx/raw/main/install.sh | sh
```

This installs mvx to `~/.mvx/bin/mvx` and adds it to your PATH.

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
      description: "Build Go binary",
      script: "go build -o bin/app ."
    },
    test: {
      description: "Run tests with coverage",
      script: "go test -v -cover ./..."
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
      script: "npm run build"
    },
    dev: {
      description: "Start development server",
      script: "npm run dev"
    },
    test: {
      description: "Run tests",
      script: "npm test"
    }
  }
}
```

## Key Benefits

- **🚀 Zero Dependencies**: No need to install tools on your system
- **🌍 Cross-Platform**: Works on Linux, macOS, and Windows
- **🔧 Universal Tools**: One tool to manage Maven, Go, Node.js, and more
- **📦 Version Isolation**: Each project specifies its own tool versions
- **⚡ Fast Setup**: New team members can start building immediately

## Next Steps

- [Learn about configuration options](/configuration)
- [Explore supported tools](/tools)
- [Discover custom commands](/commands)
- [Check out the GitHub repository](https://github.com/gnodet/mvx)

## Need Help?

- 📖 [Configuration Guide](/configuration)
- 🔧 [Supported Tools](/tools)
- 💬 [GitHub Discussions](https://github.com/gnodet/mvx/discussions)
- 🐛 [Report Issues](https://github.com/gnodet/mvx/issues)
