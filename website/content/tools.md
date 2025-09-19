---
title: Supported Tools
description: Complete list of development tools supported by mvx
layout: page
---

# Supported Tools

mvx supports a wide range of development tools across different programming languages and ecosystems. Each tool is automatically downloaded, installed, and configured for your project.

## Adding Tools to Your Project

The easiest way to add tools to your project is using the `mvx tools add` command:

```bash
# Add Java 21 (default Temurin distribution)
mvx tools add java 21

# Add Java 17 with specific distribution
mvx tools add java 17 zulu

# Add Maven 4.0.0-rc-4
mvx tools add maven 4.0.0-rc-4

# Add Node.js LTS
mvx tools add node lts

# Add Go 1.23.1
mvx tools add go 1.23.1

# Add Python 3.12
mvx tools add python 3.12
```

**Benefits:**
- ✅ **Validates** the tool and version exist
- ✅ **Updates** your `.mvx/config.json5` automatically
- ✅ **Preserves** existing configuration and formatting
- ✅ **Adds comments** and proper JSON5 structure

## Using System Tools

For CI environments, corporate setups, or when you prefer to use existing tool installations, mvx supports using system-installed tools instead of downloading them. This is controlled via environment variables:

```bash
# Enable system tools individually
export MVX_USE_SYSTEM_JAVA=true
export MVX_USE_SYSTEM_MAVEN=true
export MVX_USE_SYSTEM_NODE=true    # Coming soon
export MVX_USE_SYSTEM_GO=true      # Coming soon

./mvx setup
```

**Benefits:**
- ⚡ **Faster builds**: No download time
- 🛡️ **More reliable**: Avoids network issues
- 💾 **Better resource usage**: Uses existing installations
- 🎯 **Selective control**: Enable per tool independently
- 🔒 **Security**: Use centrally managed, approved tool versions

**How it works:**
1. When `MVX_USE_SYSTEM_<TOOL>=true` is set, mvx first tries to use the system installation
2. Validates that the system tool version matches your configuration
3. If compatible, creates a symlink to integrate with mvx's tool management
4. If incompatible or unavailable, falls back to downloading

## Supported Tools

## Java Ecosystem

### Java (OpenJDK)

Automatic installation of OpenJDK distributions.

```json5
{
  tools: {
    java: {
      version: "21",                    // Java version (8, 11, 17, 21, etc.)
      distribution: "temurin",          // Optional: temurin, zulu, corretto
      arch: "x64"                       // Optional: x64, aarch64
    }
  }
}
```

#### Using System Java

For CI environments or when you prefer to use an existing Java installation, you can configure mvx to use the system Java instead of downloading:

```bash
export MVX_USE_SYSTEM_JAVA=true
export JAVA_HOME=/path/to/your/java
./mvx build
```

When `MVX_USE_SYSTEM_JAVA=true` is set:

- ✅ **Uses existing Java**: mvx will use the Java installation from `JAVA_HOME`
- ✅ **Version validation**: Ensures the system Java version matches the requested version
- ✅ **Faster setup**: No time spent downloading Java in CI environments
- ✅ **Fallback behavior**: If system Java is unavailable or incompatible, falls back to downloading

**Requirements:**
- `JAVA_HOME` environment variable must be set
- System Java version must match the major version specified in your configuration
- Java executable must be available at `$JAVA_HOME/bin/java`

**Use Cases:**
- **GitHub Actions**: Runners have Java pre-installed
- **Corporate environments**: Java is centrally managed
- **Offline environments**: Where downloading is restricted
```

**Supported Versions**: 8, 11, 17, 21, 22, 23  
**Supported Distributions**: Eclipse Temurin, Azul Zulu, Amazon Corretto  
**Platforms**: Linux (x64, aarch64), macOS (x64, aarch64), Windows (x64)

### Maven

Apache Maven build automation tool.

```json5
{
  tools: {
    maven: {
      version: "3.9.6",                // Maven version
      settings: "custom-settings.xml"   // Optional: custom settings file
    }
  }
}
```

**Supported Versions**: 3.6.x, 3.8.x, 3.9.x
**Platforms**: All (Java-based)

#### Using System Maven

For CI environments or when you prefer to use an existing Maven installation:

```bash
export MVX_USE_SYSTEM_MAVEN=true
./mvx build
```

When `MVX_USE_SYSTEM_MAVEN=true` is set:

- ✅ **Uses existing Maven**: mvx will use Maven from `MAVEN_HOME`, `M2_HOME`, or PATH
- ✅ **Version validation**: Ensures the system Maven version matches the requested version
- ✅ **Faster setup**: No time spent downloading Maven in CI environments
- ✅ **Fallback behavior**: If system Maven is unavailable or incompatible, falls back to downloading

**Detection Order:**
1. `MAVEN_HOME` environment variable
2. `M2_HOME` environment variable (fallback)
3. `mvn` command in PATH

**Requirements:**
- Maven must be accessible via one of the detection methods above
- System Maven version must match the version specified in your configuration
- Maven executable must be functional (`mvn --version` works)

### Maven Daemon

Faster Maven builds with persistent JVM.

```json5
{
  tools: {
    mvnd: {
      version: "1.0.1"                 // Maven Daemon version
    }
  }
}
```

**Supported Versions**: 0.9.x, 1.0.x  
**Platforms**: Linux (x64, aarch64), macOS (x64, aarch64), Windows (x64)

## Go Ecosystem

### Go

Go programming language compiler and tools.

```json5
{
  tools: {
    go: {
      version: "1.21.0",               // Go version
      modules: true                     // Optional: enable Go modules (default: true)
    }
  }
}
```

**Supported Versions**: 1.19.x, 1.20.x, 1.21.x, 1.22.x, 1.23.x  
**Platforms**: Linux (x64, aarch64), macOS (x64, aarch64), Windows (x64)

## Node.js Ecosystem

### Node.js

JavaScript runtime environment.

```json5
{
  tools: {
    node: {
      version: "20.0.0",               // Node.js version
      npm: "10.0.0"                    // Optional: specific npm version
    }
  }
}
```

**Supported Versions**: 16.x, 18.x, 20.x, 21.x, 22.x  
**Platforms**: Linux (x64, aarch64), macOS (x64, aarch64), Windows (x64)

### Yarn

Alternative package manager for Node.js.

```json5
{
  tools: {
    yarn: {
      version: "1.22.19"               // Yarn version
    }
  }
}
```

**Supported Versions**: 1.22.x, 3.x, 4.x
**Platforms**: All (Node.js-based)

## Python Ecosystem

### Python

Python programming language interpreter with **automatic project isolation**.

```json5
{
  tools: {
    python: {
      version: "3.12.0",                  // Python version
      options: {
        requirements: "requirements.txt", // Optional: auto-install requirements
        venv: "true"                      // Optional: enable virtual environments (default: true)
      }
    }
  }
}
```

**Supported Versions**: 3.8.x, 3.9.x, 3.10.x, 3.11.x, 3.12.x, 3.13.x
**Platforms**: Linux (x64, aarch64), macOS (x64, aarch64), Windows (x64)

**🔒 Project Isolation Features**:
- **Virtual Environments**: Each project gets its own isolated Python environment
- **Package Isolation**: pip packages are installed per-project, not globally
- **Path Management**: Project-specific PYTHONPATH and PATH configuration
- **No Conflicts**: Projects don't interfere with each other or system Python
- **Automatic Setup**: Virtual environments created automatically when needed

### pip & Requirements

Python package management with project isolation.

```json5
{
  tools: {
    python: {
      version: "3.12.0",
      options: {
        requirements: "requirements.txt"  // Auto-install from requirements file
      }
    }
  }
}
```

**Automatic Requirements Installation**:
- Detects `requirements.txt`, `requirements.pip`, or `pyproject.toml`
- Installs packages in project-specific virtual environment
- No global package pollution
- Consistent dependencies across team members

**Environment Variables Set**:
- `VIRTUAL_ENV`: Path to project's virtual environment
- `PYTHONPATH`: Project-specific Python path
- `PYTHONNOUSERSITE`: Disables user site packages for isolation
- `PATH`: Prioritizes virtual environment binaries

## Python Environment Isolation

mvx provides **complete project isolation** for Python, ensuring that:

### 🔒 No Cross-Project Contamination

Each project gets its own isolated environment:

```bash
# Project A uses Python 3.11 with Django
cd /path/to/project-a
./mvx setup  # Creates isolated environment for this project

# Project B uses Python 3.12 with Flask
cd /path/to/project-b
./mvx setup  # Creates separate isolated environment
```

**Directory Structure**:
```
# Global Python installations (shared across projects)
~/.mvx/tools/python/
├── 3.11.0/                    # Python 3.11 installation
│   ├── bin/python3
│   └── lib/
└── 3.12.0/                   # Python 3.12 installation
    ├── bin/python3
    └── lib/

# Project-specific virtual environments (in each project)
project-a/
├── .mvx/
│   ├── config.json5
│   └── venv/                  # Project A's isolated virtual environment
├── requirements.txt
└── src/

project-b/
├── .mvx/
│   ├── config.json5
│   └── venv/                  # Project B's isolated virtual environment
├── requirements.txt
└── src/
```

### 🛡️ System Python Protection

mvx **never modifies** your system Python:
- No global package installations
- No PATH modifications outside mvx
- System Python remains untouched
- No `sudo` required for package management

### 📦 Automatic Dependency Management

```json5
{
  tools: {
    python: {
      version: "3.12",
      options: {
        requirements: "requirements.txt"  // Auto-installed in project venv
      }
    }
  }
}
```

When you run `./mvx setup`:
1. Downloads Python 3.12 to `~/.mvx/tools/python/3.12.0/`
2. Creates project-specific virtual environment
3. Installs requirements.txt packages in the virtual environment
4. Sets up isolated environment variables

### 🔄 Seamless Project Switching

```bash
# Work on Django project
cd ~/projects/django-app
./mvx setup         # Creates .mvx/venv/ with Django
./mvx run server    # Uses Django in isolated environment

# Switch to Flask project
cd ~/projects/flask-api
./mvx setup         # Creates .mvx/venv/ with Flask
./mvx run server    # Uses Flask in different isolated environment
```

Each project maintains its own:
- Python version (shared installation)
- Package versions (in `.mvx/venv/`)
- Environment variables
- Virtual environment (in `.mvx/venv/`)

**Benefits of project-local venvs:**
- ✅ Automatic cleanup when project is deleted
- ✅ Clear ownership and easy discovery
- ✅ Project portability (can move/copy projects)
- ✅ No orphaned virtual environments
- ✅ Follows Python community conventions

## Tool Discovery

### List Available Tools

```bash
# List all supported tools
./mvx tools list

# Search for specific tools
./mvx tools search java
./mvx tools search node
```

### Check Tool Versions

```bash
# List available versions for a tool
./mvx tools versions java
./mvx tools versions maven

# Check currently installed version
./mvx tools info java
```

### Tool Installation

```bash
# Install all configured tools
./mvx setup

# Install specific tool
./mvx tools install java
./mvx tools install maven

# Verify tool installation
./mvx tools verify java
```

## Tool Isolation

mvx manages tools globally while maintaining project-specific configurations:

- **Global installation**: Tools are installed in `~/.mvx/tools/` and shared across projects
- **Version isolation**: Different projects can specify different tool versions in their config
- **No conflicts**: Tools don't interfere with system installations
- **Clean removal**: Delete `~/.mvx/` directory to remove all tools
- **Project configuration**: Each project's `.mvx/config.json5` specifies which tool versions to use

## Custom Tool Paths

Override tool paths if needed:

```json5
{
  tools: {
    java: {
      version: "21",
      path: "/custom/path/to/java"      // Use custom installation
    }
  }
}
```

## Tool Verification

mvx automatically verifies tool installations:

- **Checksum verification**: Downloaded binaries are verified against SHA256/SHA512 checksums
  - ✅ Maven: Uses Apache's official SHA512 checksums
  - ✅ Maven Daemon: Uses Apache's official SHA512 checksums
  - ✅ Java: Uses Adoptium API SHA256 checksums
  - ✅ Node.js: Uses official SHASUMS256.txt files
  - ✅ Go: Framework ready for Go's checksum database
- **Version validation**: Ensures correct version is installed
- **Path resolution**: Verifies tools are accessible
- **Health checks**: Basic functionality tests

### Checksum Configuration

#### Verification Modes

**Optional Verification** (default): Warns on checksum failures but continues installation
```
⚠️  Checksum verification failed: checksum mismatch
   This could indicate a corrupted download or security issue.
```

**Required Verification**: Fails installation on checksum errors
```json5
{
  tools: {
    maven: {
      version: "3.9.6",
      checksum: {
        required: true  // Fail installation if checksum verification fails
      }
    }
  }
}
```

#### Custom Checksums

Provide your own checksums for enhanced security:

```json5
{
  tools: {
    maven: {
      version: "3.9.6",
      checksum: {
        type: "sha512",  // sha256 or sha512
        value: "abc123def456...",
        required: true
      }
    }
  }
}
```

#### Checksum URLs

Use external checksum sources:

```json5
{
  tools: {
    maven: {
      version: "3.9.6",
      checksum: {
        type: "sha512",
        url: "https://archive.apache.org/dist/maven/maven-3/3.9.6/binaries/apache-maven-3.9.6-bin.zip.sha512",
        filename: "apache-maven-3.9.6-bin.zip",
        required: true
      }
    }
  }
}
```

#### Automatic Checksum Sources

mvx automatically fetches checksums from official sources:

- **Maven**: `https://archive.apache.org/dist/maven/maven-\{version}/binaries/\{filename}.sha512`
- **Maven Daemon**: `https://archive.apache.org/dist/maven/mvnd/\{version}/\{filename}.sha512`
- **Java**: Adoptium API at `https://api.adoptium.net/v3/assets/latest/\{version}/hotspot`
- **Node.js**: `https://nodejs.org/dist/v\{version}/SHASUMS256.txt`
- **Go**: Planned integration with Go's checksum database

#### Security Best Practices

1. **Enable required verification** for production:
   ```json5
   {
     tools: {
       maven: { version: "3.9.6", checksum: { required: true } },
       java: { version: "21", checksum: { required: true } },
       node: { version: "22.14.0", checksum: { required: true } }
     }
   }
   ```

2. **Pin specific checksums** for critical tools:
   ```json5
   {
     tools: {
       maven: {
         version: "3.9.6",
         checksum: {
           type: "sha512",
           value: "specific-checksum-here",
           required: true
         }
       }
     }
   }
   ```

3. **Verify checksums manually** for sensitive environments by downloading from official sources

## Adding New Tools

Want to see a tool added to mvx? 

1. [Open an issue](https://github.com/gnodet/mvx/issues) with tool details
2. Check our [contribution guide](https://github.com/gnodet/mvx/blob/main/CONTRIBUTING.md)
3. Submit a pull request with tool implementation

## Next Steps

- [Learn about configuration](/configuration)
- [Explore custom commands](/commands)
- [Check out examples](https://github.com/gnodet/mvx/tree/main/examples)
