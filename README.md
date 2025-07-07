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

## 🔧 Core Principles

### 1. **Self-Contained**
- Single shell script with no external dependencies
- Downloads and manages tools as needed
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

```
project/
├── mvx                    # Main executable script
├── mvx.bat               # Windows batch equivalent  
├── .mvx/
│   ├── config.yml        # Project configuration
│   ├── tools/            # Tool definitions and cache
│   └── commands/         # Custom command scripts
└── your-project-files...
```

## 📋 Planned Features

### Tool Management
- [x] Maven wrapper functionality (baseline)
- [ ] Java version management (OpenJDK, GraalVM, etc.)
- [ ] Node.js and npm/yarn support
- [ ] Python and pip/poetry support
- [ ] Go toolchain support
- [ ] Rust and Cargo support
- [ ] Custom tool definitions

### Command System
- [ ] Configurable command definitions
- [ ] Built-in help and documentation
- [ ] Command aliases and shortcuts
- [ ] Conditional commands (platform/environment specific)
- [ ] Command composition and pipelines

### Environment Management
- [ ] Environment variable configuration
- [ ] PATH management
- [ ] Shell integration (bash, zsh, fish)
- [ ] IDE configuration generation (VS Code, IntelliJ)

### Advanced Features
- [ ] Parallel tool downloads
- [ ] Checksum verification for security
- [ ] Offline mode support
- [ ] CI/CD integration helpers
- [ ] Container/Docker support
- [ ] Build metrics and performance tracking

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

Here's what a `.mvx/config.yml` might look like for a typical Java project:

```yaml
project:
  name: "my-awesome-app"
  description: "A full-stack application"

tools:
  java:
    version: "21"
    distribution: "temurin"
  maven:
    version: "4.0.0"
  node:
    version: "20.x"
    required_for: ["frontend", "docs"]

environment:
  JAVA_OPTS: "-Xmx2g -XX:+UseG1GC"
  APP_ENV: "development"

commands:
  build:
    description: "Build the entire application"
    script: "./mvnw clean install"

  frontend:
    description: "Build frontend assets"
    script: "npm run build"
    working_dir: "frontend"
    requires: ["node"]

  demo:
    description: "Run application demos"
    script: "java -cp target/classes com.example.Demo"
    args:
      - name: "type"
        description: "Demo type (web, cli, batch)"
        default: "web"

  dev:
    description: "Start development environment"
    script: |
      ./mvx frontend &
      ./mvnw spring-boot:run
```

## 🚦 Current Status

**Early Development** - This project is in the conceptual and early implementation phase.

We're starting with the Maven ecosystem (building on Maven Wrapper's success) and expanding from there. The goal is to create a tool that feels familiar to Maven users but works universally.

### Roadmap

**Phase 1: Maven Foundation** (Q1 2025)
- [ ] Enhanced Maven wrapper with tool management
- [ ] Basic Java version detection and management
- [ ] Simple command configuration system

**Phase 2: Multi-Tool Support** (Q2 2025)
- [ ] Node.js and npm integration
- [ ] Python and pip support
- [ ] Environment variable management

**Phase 3: Advanced Features** (Q3 2025)
- [ ] IDE integration
- [ ] CI/CD helpers
- [ ] Performance optimizations

## 🤝 Contributing

This project is just getting started! We're looking for:

- **Feedback** on the overall vision and goals
- **Use case examples** from your projects
- **Tool integration ideas** for different ecosystems
- **Implementation contributions** as we build it out

## 📚 Inspiration

mvx builds on the success of:
- **Maven Wrapper** - Proved that self-contained bootstrap works
- **asdf/mise** - Demonstrated multi-tool version management
- **just/task** - Showed the value of simple command runners
- **direnv** - Pioneered automatic environment management

## 📄 License

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.

---

**Note**: This is an early-stage project. The API and features described above are subject to change as we develop and refine the tool based on community feedback and real-world usage.
