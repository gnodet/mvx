---
title: About mvx
description: Learn about the mvx project, its goals, and the team behind it
layout: page
---

# About mvx

mvx is a universal build environment bootstrap tool that simplifies development environment setup across different programming languages and platforms. Think of it as "Maven Wrapper for the modern era" - but it works with any technology stack.

## The Problem

Modern software development involves multiple programming languages, build tools, and runtime environments. Setting up a consistent development environment across team members and CI/CD systems is challenging:

- **"Works on my machine"** syndrome
- **Complex setup instructions** for new team members
- **Version mismatches** between development and production
- **Tool installation conflicts** on developer machines
- **CI/CD environment drift** from local development

## The Solution

mvx solves these problems by providing:

- **Zero-dependency bootstrap**: No need to pre-install tools
- **Automatic tool management**: Downloads and configures tools automatically
- **Project-specific isolation**: Each project uses its own tool versions
- **Cross-platform consistency**: Works the same on Linux, macOS, and Windows
- **Simple configuration**: One JSON5 file defines your entire environment

## Key Features

### 🚀 Zero Dependencies
No need to install Java, Maven, Node.js, Go, or any other tools on your system. mvx downloads and manages everything automatically.

### 🌍 Cross-Platform
Works seamlessly on Linux, macOS, and Windows with support for both x86_64 and ARM64 architectures.

### 🔧 Universal Tool Support
Supports Maven, Go, Node.js, Java, Python, Rust, and more. One tool to manage your entire development environment.

### ⚡ Simple Configuration
Define your tools and commands in a simple JSON5 configuration file. Custom commands become top-level commands.

### 🏗️ Project Isolation
Tools are installed globally in `~/.mvx/` but each project specifies its own tool versions. No conflicts between projects or with system installations.

## Use Cases

### Enterprise Development Teams
- Eliminate environment setup complexity
- Ensure consistent builds across all team members
- Reduce onboarding time for new developers
- Standardize development environments

### CI/CD Pipelines
- Reproducible builds in any environment
- No more dependency installation scripts
- Consistent tool versions between local and CI
- Faster pipeline execution with tool caching

### Open Source Projects
- Lower the barrier for contributors
- One-command setup for any technology stack
- Clear documentation of required tools
- Platform-independent development

### Educational Projects
- Students focus on learning, not environment setup
- Works on any school computer or personal device
- Consistent experience across different platforms
- Easy to share and reproduce projects

## Design Philosophy

### Simplicity First
mvx prioritizes ease of use over advanced features. The goal is to make development environment setup as simple as possible.

### Convention Over Configuration
Sensible defaults reduce the need for extensive configuration. Most projects can get started with minimal setup.

### Platform Agnostic
mvx works the same way on all platforms. No platform-specific scripts or workarounds needed.

### Project-Centric
Tools are managed per project, not globally. This prevents conflicts and ensures reproducibility.

### Minimal Footprint
mvx itself is a single binary with no dependencies. Tool installations are isolated and can be easily removed.

## Project History

mvx was created by [Guillaume Nodet](https://github.com/gnodet) to address the challenges of managing development environments across multiple projects and teams.

## Technology Stack

mvx is built with:

- **Go**: For cross-platform compatibility and single-binary distribution
- **JSON5**: For human-friendly configuration files
- **Shell Scripts**: For platform-specific bootstrap scripts

## Contributing

mvx is an open source project and welcomes contributions:

- **Bug Reports**: [GitHub Issues](https://github.com/gnodet/mvx/issues)
- **Feature Requests**: [GitHub Discussions](https://github.com/gnodet/mvx/discussions)
- **Code Contributions**: [Pull Requests](https://github.com/gnodet/mvx/pulls)
- **Documentation**: Help improve guides and examples

## License

mvx is released under the [Apache License 2.0](https://github.com/gnodet/mvx/blob/main/LICENSE).

## Comparison with Other Tools

Wondering how mvx compares to other tool managers like mise, asdf, or direnv?

- **[mvx vs mise](/comparison-mise)**: Detailed comparison with mise (formerly rtx)

## Community

- **GitHub**: [gnodet/mvx](https://github.com/gnodet/mvx)
- **Discussions**: [GitHub Discussions](https://github.com/gnodet/mvx/discussions)
- **Issues**: [GitHub Issues](https://github.com/gnodet/mvx/issues)

---

*mvx - Universal Build Environment Bootstrap*  
*Created with ❤️ by [Guillaume Nodet](https://github.com/gnodet)*
