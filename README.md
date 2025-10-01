# mvx - Maven eXtended

> A universal build environment bootstrap tool that goes beyond Maven

📖 **[Documentation & Website](https://gnodet.github.io/mvx/)** | 🚀 **[Getting Started](https://gnodet.github.io/mvx/getting-started/)** | 🛠️ **[Tools](https://gnodet.github.io/mvx/tools/)** | ⚙️ **[Interpreters](https://gnodet.github.io/mvx/interpreters/)**

## 🎯 Goals

**mvx** aims to solve the fundamental challenge of project setup and build environment management. While Maven Wrapper revolutionized Maven distribution, modern development requires more:

- **Zero-dependency bootstrapping** - No external tools required to get started
- **Universal tool management** - Handle Java, Node.js, Python, and other runtimes
- **Simple command interface** - Abstract complex build commands into intuitive actions
- **Cross-platform compatibility** - Works seamlessly on Linux, macOS, and Windows
- **Project-specific environments** - Each project gets exactly what it needs

## 🚀 Vision

Imagine cloning any project and running:

```bash
./mvx setup    # Installs all required tools automatically
./mvx build    # Builds the project with the right environment
./mvx test     # Runs tests with proper configuration
./mvx demo     # Launches project-specific demos or examples

# Or use tools directly with natural syntax
./mvx mvn -V clean install    # Maven with version flag
./mvx --verbose mvn -X test   # mvx verbose + Maven debug

# Execute shell commands in mvx environment
./mvx shell 'echo $JAVA_HOME'  # Show Java home with mvx tools
./mvx shell 'java -version'    # Run Java with mvx environment
```

No more "works on my machine" - every developer gets the exact same environment.

## 🔧 Maven Integration

mvx provides seamless Maven integration with transparent argument passing:

```bash
# All Maven flags work naturally - no special syntax needed
./mvx mvn -V                    # Show Maven version
./mvx mvn -X clean install      # Debug mode with clean install
./mvx mvn -Pproduction package  # Activate profile and package

# Combine mvx global flags with Maven flags
./mvx --verbose mvn -V          # mvx verbose output + Maven version
./mvx --quiet mvn test          # mvx quiet mode + Maven test

# Backward compatibility maintained
./mvx mvn -- -V                 # Still works (with helpful migration warning)
```

**Key Benefits:**
- **🎯 Natural syntax**: Use Maven flags exactly as you would with `mvn`
- **🔄 Transparent wrapper**: mvx acts like `mvnw` but with enhanced tool management
- **⚡ No learning curve**: Existing Maven knowledge applies directly
- **🛡️ Backward compatible**: Existing scripts continue to work
- **🏢 Enterprise ready**: URL replacements for corporate networks and mirrors

## 📦 mvx Bootstrap

Just like Maven Wrapper (`mvnw`), mvx provides bootstrap scripts that automatically download and run the appropriate mvx version for your project:

**Install latest stable release:**

```bash
curl -fsSL https://raw.githubusercontent.com/gnodet/mvx/main/install-mvx.sh | bash
```

**Install development version (latest features, may be unstable):**

```bash
curl -fsSL https://raw.githubusercontent.com/gnodet/mvx/main/install-mvx.sh | MVX_VERSION=dev bash
```

**Use mvx without global installation:**

```bash
./mvx setup
```

```bash
./mvx build
```

```bash
./mvx test
```

The bootstrap automatically:

- Downloads the correct mvx Go binary for your project
- Caches binaries to avoid re-downloading
- Works on Linux, macOS, and Windows
- Requires no global installation
- Creates lightweight shell/batch scripts in your project

See [BOOTSTRAP.md](BOOTSTRAP.md) for detailed documentation.

## 📦 Installation

### Using Bootstrap Scripts (Recommended)

The easiest way to use mvx is via the bootstrap scripts:

**Install latest stable release:**

```bash
curl -fsSL https://raw.githubusercontent.com/gnodet/mvx/main/install-mvx.sh | bash
```

**Install development version:**

```bash
curl -fsSL https://raw.githubusercontent.com/gnodet/mvx/main/install-mvx.sh | MVX_VERSION=dev bash
```

**Use mvx without global installation:**

```bash
./mvx setup
```

```bash
./mvx build
```

```bash
./mvx test
```

### Direct Binary Installation

Download the appropriate binary for your platform from [GitHub Releases](https://github.com/gnodet/mvx/releases):

**Linux x64:**

```bash
curl -fsSL https://github.com/gnodet/mvx/releases/latest/download/mvx-linux-amd64 -o mvx
chmod +x mvx
```

**macOS x64 (Intel):**

```bash
curl -fsSL https://github.com/gnodet/mvx/releases/latest/download/mvx-darwin-amd64 -o mvx
chmod +x mvx
```

**macOS ARM64 (Apple Silicon):**

```bash
curl -fsSL https://github.com/gnodet/mvx/releases/latest/download/mvx-darwin-arm64 -o mvx
chmod +x mvx
```

**Windows x64:**

```bash
curl -fsSL https://github.com/gnodet/mvx/releases/latest/download/mvx-windows-amd64.exe -o mvx.exe
```

### Supported Platforms

| OS | Architecture | Binary | Notes |
|---|---|---|---|
| Linux | x64 | `mvx-linux-amd64` | Static binary, no dependencies |
| Linux | ARM64 | `mvx-linux-arm64` | Static binary, no dependencies |
| macOS | x64 | `mvx-darwin-amd64` | Intel Macs |
| macOS | ARM64 | `mvx-darwin-arm64` | Apple Silicon Macs |
| Windows | x64 | `mvx-windows-amd64.exe` | Static binary |

All binaries are statically linked and have no external dependencies.

### Building from Source

Requirements: Go 1.24+

**Clone the repository:**

```bash
git clone https://github.com/gnodet/mvx.git
cd mvx
```

**Build for current platform:**

```bash
./mvx build
```

**Build for all platforms:**

```bash
./mvx build-all
```

**Run tests:**

```bash
./mvx test
```

## 🔄 Shell Activation

For a seamless development experience, enable shell activation to automatically set up your environment when entering project directories:

**Bash** - Add to `~/.bashrc`:
```bash
eval "$(mvx activate bash)"
```

**Zsh** - Add to `~/.zshrc`:
```bash
eval "$(mvx activate zsh)"
```

**Fish** - Add to `~/.config/fish/config.fish`:
```bash
mvx activate fish | source
```

With shell activation enabled, tools become available automatically:

```bash
cd my-project
# mvx: activating environment in /Users/you/my-project

java -version  # Uses mvx-managed Java
mvn -version   # Uses mvx-managed Maven
```

**Learn more**: See the [Shell Activation Guide](https://gnodet.github.io/mvx/shell-activation/) for detailed documentation.

## 🎯 Shell Completion

mvx supports shell completion for commands and arguments across multiple shells (bash, zsh, fish, powershell):

### Zsh Completion

**For current session:**

```bash
source <(./mvx completion zsh)
```

**For permanent setup (recommended):**

```bash
# Create completion directory if it doesn't exist
mkdir -p ~/.zsh/completions

# Generate completion script
./mvx completion zsh > ~/.zsh/completions/_mvx

# Add to ~/.zshrc (if not already there)
echo 'fpath=(~/.zsh/completions $fpath)' >> ~/.zshrc
echo 'autoload -U compinit && compinit' >> ~/.zshrc

# Reload shell
source ~/.zshrc
```

### Bash Completion

**For current session:**

```bash
source <(./mvx completion bash)
```

**For permanent setup:**

```bash
# Add to ~/.bashrc
echo 'source <(./mvx completion bash)' >> ~/.bashrc
source ~/.bashrc
```

### Other Shells

mvx also supports completion for:

- **Fish**: `./mvx completion fish`
- **PowerShell**: `./mvx completion powershell`

See `./mvx completion [shell] --help` for setup instructions.

### Test Completion

```bash
./mvx <TAB>          # Shows available commands
./mvx build <TAB>    # Shows build options
./mvx setup <TAB>    # Shows setup flags
```

## 🔧 Core Principles

### 1. **Self-Contained**

- Bootstrap scripts with no external dependencies
- Downloads and manages Go binaries and tools as needed
- Caches everything locally for offline work

### 2. **Configurable**

- Project-specific tool versions and commands
- Environment variable management
- Extensible command system

### 3. **Universal**

- Started with Maven but works with any build system
- Language-agnostic tool management
- Ecosystem-aware (npm, pip, cargo, etc.)

### 4. **Developer-Friendly**

- Intuitive command names and help system
- Rich debugging and verbose modes
- IDE integration support

## 🏗️ Architecture

mvx uses a bootstrap system similar to Maven Wrapper, providing zero-dependency project setup:

```text
~/.mvx/                           # Global cache directory
├── versions/                     # Cached mvx versions
│   ├── 1.0.0/
│   │   ├── mvx                  # Go binary (Unix/Linux/macOS)
│   │   └── mvx.exe              # Go binary (Windows)
│   └── 1.1.0/
├── tools/                        # Downloaded tools cache
│   ├── maven/
│   │   ├── 3.9.6/
│   │   └── 4.0.0/
│   ├── mvnd/
│   │   └── 1.0.2/
│   └── java/
│       ├── temurin-21/
│       └── graalvm-21/
└── config/                       # Global configuration

project/                          # Project directory
├── mvx                          # Bootstrap script (Unix/Linux/macOS)
├── mvx.cmd                      # Bootstrap script (Windows)
├── mvx-dev                      # Local development binary (optional)
├── .mvx/
│   ├── mvx.properties           # Bootstrap configuration
│   ├── config.json5             # Project configuration (JSON5) - planned
│   ├── config.yml               # Or YAML format - planned
│   └── local/                   # Project-specific cache - planned
└── your-project-files...
```

### Bootstrap System

The bootstrap scripts (`mvx` and `mvx.cmd`) are **shell/batch scripts** (not binaries) that automatically:

- Detect your platform and architecture
- Check for local development binaries first (`mvx-dev`, `mvx-dev.exe`)
- Download and cache the appropriate mvx **Go binary** version
- Execute commands with the correct binary
- Provide self-update capabilities

**Key distinction**: The `mvx` and `mvx.cmd` files in your project are lightweight bootstrap scripts, while the actual mvx functionality is provided by Go binaries that are downloaded and cached automatically.

## 📋 Features

### ✅ Implemented Features

#### Bootstrap & Distribution

- [x] Cross-platform bootstrap scripts (Unix/Windows)
- [x] Automatic binary download and caching
- [x] Version management via `.mvx/mvx.properties`
- [x] Local development binary support (`mvx-dev`)
- [x] Self-update capabilities (`mvx update-bootstrap`)
- [x] Platform detection (Linux, macOS, Windows, ARM64/x64)

#### Core Commands

- [x] `mvx version` - Version information and diagnostics
- [x] `mvx init` - Initialize mvx configuration in projects
- [x] `mvx setup` - Install tools and configure environment
- [x] `mvx build` - Execute configured build commands
- [x] `mvx test` - Execute configured test commands
- [x] `mvx run` - Execute custom commands from configuration
- [x] `mvx shell` - Execute shell commands in mvx environment
- [x] `mvx tools` - Tool management and discovery
- [x] `mvx info` - Detailed command information

#### Configuration System

- [x] JSON5 configuration format support
- [x] YAML configuration format support
- [x] Project-specific tool definitions
- [x] Custom command definitions with arguments
- [x] Multiline script support
- [x] Command-specific environment variables
- [x] Working directory specification
- [x] Tool requirement validation
- [x] Global environment variable management
- [x] Configurable tool versions

#### Tool Management

- [x] **Java Development Kit** - Multiple distributions (Temurin, GraalVM, Oracle, Corretto, Liberica, Zulu, Microsoft)
- [x] **Apache Maven** - All versions (3.x, 4.x including pre-releases)
- [x] **Maven Daemon (mvnd)** - High-performance Maven alternative
- [x] **Node.js** - All LTS and current versions with npm/yarn support
- [x] **Go** - All stable versions from golang.org
- [x] **Python** - All stable versions (3.8+) with pip support
- [x] Tool installation and caching
- [x] Environment setup and PATH management
- [x] Version resolution (latest, major.minor, exact versions)

#### Developer Experience

- [x] Shell completion (bash, zsh, fish, powershell)
- [x] Verbose and quiet modes
- [x] Built-in help system
- [x] Command validation and error handling

#### Command Execution & Interpreters

- [x] **Multiple interpreter support** - Choose between native shell and cross-platform mvx-shell
- [x] **Automatic PATH management** - mvx-managed tools automatically available in PATH
- [x] **Environment variable support** - Global and command-specific environment variables
- [x] **Intelligent interpreter selection** - Automatic selection based on script complexity
- [x] **Cross-platform compatibility** - Commands work consistently across operating systems

#### Enterprise & Network Support

- [x] **URL replacements** - Redirect downloads through corporate proxies, mirrors, or alternative sources
- [x] **Global configuration** - System-wide settings for enterprise environments
- [x] **Regex-based URL transformations** - Advanced URL rewriting for complex enterprise setups

### 🚧 Planned Features

#### Extended Tool Support

- [x] **Node.js and npm/yarn support** ✅ **IMPLEMENTED**
- [x] **Python and pip/poetry support** ✅ **IMPLEMENTED**
- [ ] Custom tool definitions and installers

#### Enhanced Commands

- [ ] Command aliases and shortcuts
- [ ] Conditional commands (platform/environment specific)

#### Security & Performance

- [x] **Checksum verification for security** ✅ **IMPLEMENTED**
  - SHA256/SHA512 verification for Maven, Maven Daemon, Java, Node.js, and Go
  - Optional and required verification modes
  - Support for custom checksums and checksum URLs
  - Automatic fetching from official sources (Apache, Adoptium, Node.js, etc.)
- [ ] Parallel tool downloads

## 🛠️ Implementation

**Language:** Go (single binary, cross-platform)
**Configuration:** JSON5 and YAML support (inspired by [Maven Mason](https://github.com/maveniverse/mason))
**Installation:** Single command that downloads the binary to your project

### Configuration Format Detection

mvx automatically detects the configuration format:

- `.mvx/config.json5` → JSON5 format
- `.mvx/config.yml` or `.mvx/config.yaml` → YAML format
- Falls back to JSON5 if both exist

## 🎯 Use Cases

### For Project Maintainers

- Eliminate "how to build" documentation
- Ensure consistent development environments
- Simplify onboarding for new contributors
- Reduce support burden for environment issues

### For Developers

- One command to set up any project
- No need to install project-specific tools globally
- Consistent experience across different projects
- Easy switching between project environments

### For Teams

- Standardized development workflows
- Reproducible builds across environments
- Simplified CI/CD setup
- Better collaboration with consistent tooling

## 🚀 CI/CD Integration

mvx is designed to work seamlessly in CI/CD environments. For faster builds and better reliability, you can configure mvx to use pre-installed tools instead of downloading them.

### Using System Tools in CI

When running in CI environments like GitHub Actions, the runners often have tools like Java and Maven pre-installed. You can configure mvx to use system tools instead of downloading them:

```bash
# Use system Java
export MVX_USE_SYSTEM_JAVA=true

# Use system Maven
export MVX_USE_SYSTEM_MAVEN=true

# Use system Node.js (when implemented)
export MVX_USE_SYSTEM_NODE=true

./mvx setup
./mvx build
```

**Supported Tools:**
- ✅ **Java**: Uses `JAVA_HOME` environment variable
- ✅ **Maven**: Uses `MAVEN_HOME`, `M2_HOME`, or finds `mvn` in PATH
- 🚧 **Node.js**: Coming soon
- 🚧 **Go**: Coming soon

**Benefits:**
- ⚡ **Faster builds**: No time spent downloading tools
- 🛡️ **More reliable**: Avoids network/download issues
- 💾 **Better resource usage**: Uses existing installations
- 🎯 **Selective control**: Enable/disable per tool independently

**Example GitHub Actions workflow:**

```yaml
name: Build
on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up JDK 21
        uses: actions/setup-java@v4
        with:
          java-version: '21'
          distribution: 'temurin'

      - name: Set up Maven
        uses: actions/setup-maven@v4
        with:
          maven-version: '3.9.6'

      - name: Build with mvx
        env:
          MVX_USE_SYSTEM_JAVA: true
          MVX_USE_SYSTEM_MAVEN: true
        run: |
          ./mvx setup
          ./mvx build
```

This approach works with any CI system that provides Java pre-installed or allows you to install it separately.

## 💡 Example Configuration

mvx supports both **JSON5** and **YAML** configuration formats, inspired by [Maven Mason](https://github.com/maveniverse/mason).

### JSON5 Configuration (`.mvx/config.json5`)

```json5
{
  // mvx configuration for my-awesome-app
  // See: https://mvx.dev/docs/config for full reference

  project: {
    name: "my-awesome-app",
    description: "A full-stack application",
  },

  tools: {
    // Java 21 required for virtual threads
    java: {
      version: "21",
      distribution: "temurin",  // Consistent across team
    },

    maven: {
      version: "4.0.0",
    },

    // Maven Daemon for faster builds
    mvnd: {
      version: "1.0.2",
    },

    // Python for scripting and automation (with project isolation)
    python: {
      version: "3.12",
      options: {
        requirements: "requirements.txt", // Auto-install project dependencies
      },
    },
  },

  environment: {
    // Increase memory for large builds
    JAVA_OPTS: "-Xmx2g -XX:+UseG1GC",
    APP_ENV: "development",
  },

  commands: {
    build: {
      description: "Build the entire application",
      script: "./mvnw clean install",
    },

    demo: {
      description: "Run application demos",
      script: `
        # Launch with proper classpath and options
        java -cp target/classes \\
             -Xmx1g \\
             com.example.Demo
      `,
      args: [
        {
          name: "type",
          description: "Demo type (web, cli, batch)",
          default: "web",
        },
      ],
    },
  },
}
```

### YAML Configuration (`.mvx/config.yml`)

```yaml
# mvx configuration for my-awesome-app
# See: https://mvx.dev/docs/config for full reference

project:
  name: my-awesome-app
  description: A full-stack application

tools:
  # Java 21 required for virtual threads
  java:
    version: "21"
    distribution: temurin  # Consistent across team

  maven:
    version: "4.0.0"

  # Maven Daemon for faster builds
  mvnd:
    version: "1.0.2"

  # Python for scripting and automation (with project isolation)
  python:
    version: "3.12"
    options:
      requirements: "requirements.txt"  # Auto-install project dependencies

environment:
  # Increase memory for large builds
  JAVA_OPTS: "-Xmx2g -XX:+UseG1GC"
  APP_ENV: development

commands:
  build:
    description: Build the entire application
    script: ./mvnw clean install

  demo:
    description: Run application demos
    script: |
      # Launch with proper classpath and options
      java -cp target/classes \
           -Xmx1g \
           com.example.Demo
    args:
      - name: type
        description: "Demo type (web, cli, batch)"
        default: web
```

## 🌍 Cross-Platform Scripts

mvx provides powerful cross-platform script support with two approaches:

### Platform-Specific Scripts

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

### Cross-Platform Interpreter (mvx-shell)

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
- `||` - Execute next command only if previous failed
- `;` - Execute commands sequentially regardless of success/failure
- `|` - Simple pipe support (sequential execution for now)
- `()` - Parentheses for grouping (basic support)

### Mixed Approach

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

### Interpreter Options

**Intelligent Defaults:**
- **Simple scripts**: Default to `mvx-shell` (cross-platform by nature)
- **Platform-specific scripts**: Default to `native` (platform-specific by nature)

**Available Interpreters:**
- **`native`**: Use system shell (`/bin/bash` on Unix, `cmd` on Windows)
- **`mvx-shell`**: Use built-in cross-platform interpreter

**Examples:**
```json5
{
  // This defaults to mvx-shell (cross-platform)
  script: "mkdir dist && copy target/*.jar dist/"
}
```

```json5
{
  // This defaults to native (platform-specific)
  script: {
    windows: "net start postgresql",
    unix: "sudo systemctl start postgresql"
  }
}
```

### Built-in Command Hooks

Cross-platform scripts work with built-in command hooks too:

```json5
{
  commands: {
    "test": {
      description: "Run tests with setup and cleanup",
      pre: {
        script: "mkdir -p test-results && echo Preparing tests",
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

## 🚦 Current Status

**Early Development** - This project is in the conceptual and early implementation phase.

We're starting with the Maven ecosystem (building on Maven Wrapper's success) and expanding from there. The goal is to create a tool that feels familiar to Maven users but works universally.

### Current Implementation

The project currently includes:

- **Bootstrap Scripts**: `mvx` (Unix/Linux/macOS) and `mvx.cmd` (Windows) - shell/batch scripts that download and execute the appropriate Go binary
- **Development Binary**: `mvx-dev` - a local Go binary for development (ARM64 macOS in this repository)
- **Configuration**: `.mvx/mvx.properties` - bootstrap configuration file
- **Installer**: `install-mvx.sh` - script to install bootstrap files in any project

The bootstrap system is fully functional and provides:

- Automatic platform detection
- Binary caching in `~/.mvx/versions/`
- Version management via `.mvx/mvx.properties`
- Self-update capabilities (`./mvx update-bootstrap`)
- Development binary support for local testing

### Roadmap

**Phase 1: Maven Foundation** (Q1 2025) ✅ **COMPLETED**

- [x] Enhanced Maven bootstrap with tool management
- [x] Java version detection and management (multiple distributions)
- [x] Command configuration system (JSON5/YAML)

**Phase 2: Multi-Tool Support** ✅ **COMPLETED**

- [x] Node.js and npm integration ✅ **IMPLEMENTED**
- [x] Python and pip support ✅ **IMPLEMENTED**
- [x] Security improvements (checksum verification) ✅ **IMPLEMENTED**

## 🤝 Contributing

This project is just getting started! We're looking for:

- **Feedback** on the overall vision and goals
- **Use case examples** from your projects
- **Tool integration ideas** for different ecosystems
- **Implementation contributions** as we build it out

## 📚 Inspiration

mvx builds on the success of:

- **Maven Wrapper** - Proved that self-contained bootstrap works
- **Maven Mason** - Demonstrated multi-format configuration support
- **asdf/mise** - Demonstrated multi-tool version management
- **just/task** - Showed the value of simple command runners
- **direnv** - Pioneered automatic environment management

## 📄 License

Licensed under the Eclipse Public License, Version 2.0. See [LICENSE](LICENSE) for details.

---

**Note**: This is an early-stage project. The API and features described above are subject to change as we develop and refine the tool based on community feedback and real-world usage.
