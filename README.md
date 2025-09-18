# mvx - Maven eXtended

> A universal build environment bootstrap tool that goes beyond Maven

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
```

No more "works on my machine" - every developer gets the exact same environment.

## 📦 mvx Bootstrap

Just like Maven Wrapper (`mvnw`), mvx provides bootstrap scripts that automatically download and run the appropriate mvx version for your project:

```bash
# Install mvx in your project (one-time setup) - installs latest stable release
curl -fsSL https://raw.githubusercontent.com/gnodet/mvx/main/install-mvx.sh | bash

# For development version (latest features, may be unstable)
MVX_VERSION=main curl -fsSL https://raw.githubusercontent.com/gnodet/mvx/main/install-mvx.sh | bash

# Now anyone can use mvx without installing it
./mvx setup
./mvx build
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

```bash
# Install mvx in your project (latest stable release)
curl -fsSL https://raw.githubusercontent.com/gnodet/mvx/main/install-mvx.sh | bash

# For development version (latest features, may be unstable)
MVX_VERSION=main curl -fsSL https://raw.githubusercontent.com/gnodet/mvx/main/install-mvx.sh | bash

# Use mvx without global installation
./mvx setup
./mvx build
./mvx test
```

### Direct Binary Installation

Download the appropriate binary for your platform from [GitHub Releases](https://github.com/gnodet/mvx/releases):

```bash
# Linux x64
curl -fsSL https://github.com/gnodet/mvx/releases/latest/download/mvx-linux-amd64 -o mvx
chmod +x mvx

# macOS x64
curl -fsSL https://github.com/gnodet/mvx/releases/latest/download/mvx-darwin-amd64 -o mvx
chmod +x mvx

# macOS ARM64 (Apple Silicon)
curl -fsSL https://github.com/gnodet/mvx/releases/latest/download/mvx-darwin-arm64 -o mvx
chmod +x mvx

# Windows x64
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

```bash
# Clone the repository
git clone https://github.com/gnodet/mvx.git
cd mvx

# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test
```

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
- [x] Tool installation and caching
- [x] Environment setup and PATH management
- [x] Version resolution (latest, major.minor, exact versions)

#### Developer Experience

- [x] Shell completion (bash, zsh, fish, powershell)
- [x] Verbose and quiet modes
- [x] Built-in help system
- [x] Command validation and error handling

### 🚧 Planned Features

#### Extended Tool Support

- [ ] Node.js and npm/yarn support
- [ ] Python and pip/poetry support
- [ ] Custom tool definitions and installers

#### Enhanced Commands

- [ ] Command aliases and shortcuts
- [ ] Conditional commands (platform/environment specific)

#### Security & Performance

- [ ] Checksum verification for security
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

**Phase 2: Multi-Tool Support** (Q2 2025)

- [ ] Node.js and npm integration
- [ ] Python and pip support
- [ ] Security improvements (checksum verification)

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
