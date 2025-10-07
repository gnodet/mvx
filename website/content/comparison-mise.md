---
title: mvx vs mise
description: Comparing mvx and mise - when to choose each tool for your development environment
layout: page
---

# mvx vs mise: A Comparison

Both **mvx** and **[mise](https://mise.jdx.dev)** are excellent tools for managing development environments and tool versions. While they share similar goals, they take different approaches and excel in different scenarios.

## Quick Comparison

| Feature | mvx | mise |
|---------|-----|------|
| **Installation** | Bootstrap scripts (no global install needed) | Global installation required |
| **Primary Focus** | Project-centric, zero-dependency bootstrap | Universal tool version manager |
| **Configuration** | JSON5/YAML in `.mvx/` directory | TOML in `.mise.toml` or `.tool-versions` |
| **Language** | Go | Rust |
| **Maturity** | New (2025) | Established (formerly rtx) |
| **Plugin Ecosystem** | Built-in tools only | Extensive plugin ecosystem |
| **Shell Integration** | Full shell activation support | Full shell activation support |
| **Task Runner** | Built-in with mvx-shell | Built-in task runner |
| **Environment Variables** | Supported | Supported with direnv replacement |

## Philosophy & Approach

### mvx: Bootstrap-First, Project-Centric

mvx follows the **Maven Wrapper philosophy**: every project should be self-contained and require zero external dependencies to get started.

**Key Principles:**
- **Zero dependencies**: Clone and run `./mvx setup` - no prior installation needed
- **Project isolation**: Each project has its own `.mvx/` directory with configuration
- **Reproducibility**: Bootstrap scripts ensure everyone gets the exact same environment
- **Simplicity**: Focused feature set with sensible defaults

**Best for:**
- Projects that need to work "out of the box" for new contributors
- Teams that want guaranteed reproducibility across environments
- CI/CD pipelines where you can't rely on pre-installed tools
- Projects that need to work without internet access (after initial setup)

### mise: Universal Tool Manager

mise is a **comprehensive tool version manager** that replaces asdf, nvm, pyenv, rbenv, and similar tools with a single unified solution.

**Key Principles:**
- **Universal**: One tool to manage all language runtimes and tools
- **Extensible**: Large plugin ecosystem for any tool imaginable
- **Shell integration**: Automatic environment activation when changing directories
- **Feature-rich**: Extensive configuration options and backends

**Best for:**
- Developers who work across many projects and want unified tool management
- Teams already using asdf who want better performance
- Projects that need tools not built into mvx
- Developers who want automatic shell activation

## Feature Comparison

### Tool Management

**mvx:**
```json5
{
  tools: {
    java: { version: "21", distribution: "temurin" },
    maven: { version: "4.0.0" },
    node: { version: "22" },
    go: { version: "1.24" },
    python: { version: "3.12" }
  }
}
```

- Built-in support for Java, Maven, Node.js, Go, Python
- Multiple Java distributions (Temurin, Zulu, GraalVM, etc.)
- Project-specific virtual environments for Python
- Tools cached in `~/.mvx/tools/`

**mise:**
```toml
[tools]
java = "21"
maven = "4.0.0"
node = "22"
go = "1.24"
python = "3.12"
```

- Hundreds of tools via plugins and backends
- Support for asdf, cargo, npm, pipx, ubi, and more backends
- Global and project-specific tool versions
- Tools cached in `~/.local/share/mise/`

### Bootstrap & Installation

**mvx:**
```bash
# Install bootstrap scripts in your project
curl -fsSL https://raw.githubusercontent.com/gnodet/mvx/main/install-mvx.sh | bash

# Now anyone can clone and run
git clone your-project
cd your-project
./mvx setup  # Downloads mvx binary and installs tools
./mvx build  # Just works
```

**Benefits:**
- No global installation required
- Works offline after initial setup
- Bootstrap scripts committed to repository
- Guaranteed version consistency

**mise:**
```bash
# Install mise globally
curl https://mise.run | sh

# Activate in shell
echo 'eval "$(mise activate bash)"' >> ~/.bashrc

# Use in projects
cd your-project
mise install  # Installs tools from .mise.toml
```

**Benefits:**
- One-time global installation
- Automatic shell activation
- Works across all projects
- Faster for developers working on many projects

### Cross-Platform Scripts

**mvx** includes a built-in cross-platform shell interpreter (`mvx-shell`):

```json5
{
  commands: {
    "setup-dev": {
      description: "Setup development environment",
      script: "mkdir -p logs temp && copy .env.example .env",
      interpreter: "mvx-shell"  // Works on Windows, macOS, Linux
    }
  }
}
```

**Built-in commands:** `cd`, `mkdir`, `copy`, `rm`, `echo`, `open`
**Benefits:** Write once, run anywhere without platform-specific scripts

**mise** uses native shell scripts with platform detection:

```toml
[tasks.setup-dev]
run = """
mkdir -p logs temp
cp .env.example .env
"""
```

**Benefits:** Full shell power, but may need platform-specific handling

### Task Running

**mvx:**
```json5
{
  commands: {
    build: {
      description: "Build the project",
      script: "mvn clean install",
      pre: { script: "echo Starting build..." },
      post: { script: "echo Build complete!" }
    }
  }
}
```

**mise:**
```toml
[tasks.build]
description = "Build the project"
run = "mvn clean install"
depends = ["clean"]
```

Both support:
- Custom task definitions
- Task dependencies
- Environment variables per task
- Multi-step commands

### Environment Variables

**mvx:**
```json5
{
  environment: {
    JAVA_OPTS: "-Xmx2g",
    APP_ENV: "development"
  },
  commands: {
    test: {
      environment: {
        TEST_MODE: "integration"
      }
    }
  }
}
```

**mise:**
```toml
[env]
JAVA_OPTS = "-Xmx2g"
APP_ENV = "development"

[tasks.test.env]
TEST_MODE = "integration"
```

Both support global and command-specific environment variables.

### CI/CD Integration

**mvx:**
```yaml
# GitHub Actions
- name: Build with mvx
  run: |
    ./mvx setup
    ./mvx build
```

**Benefits:**
- No installation step needed
- Bootstrap handles everything
- Can use system tools with `MVX_USE_SYSTEM_*` flags

**mise:**
```yaml
# GitHub Actions
- uses: jdx/mise-action@v2
- run: mise install
- run: mise run build
```

**Benefits:**
- Dedicated GitHub Action
- Faster with caching
- Well-established CI patterns

## Unique Features

### mvx Unique Features

1. **Zero-Dependency Bootstrap**
   - No global installation required
   - Bootstrap scripts committed to repository
   - Works immediately after `git clone`

2. **mvx-shell Cross-Platform Interpreter**
   - Write scripts that work on Windows, macOS, and Linux
   - No need for platform-specific script variants
   - Built-in commands for common operations

3. **Project Isolation by Default**
   - Each project has its own `.mvx/` directory
   - Configuration and cache are project-specific
   - No global state to manage

4. **Maven Ecosystem Integration**
   - Seamless Maven and Maven Daemon support
   - Natural Maven flag passing
   - Built for Java/Maven projects

### mise Unique Features

1. **Extensive Plugin Ecosystem**
   - Hundreds of tools available
   - Multiple backends (asdf, cargo, npm, pipx, ubi, etc.)
   - Community-contributed plugins

2. **Shell Activation**
   - Automatic environment setup when entering directories
   - Replaces direnv for environment management
   - Seamless shell integration

3. **Advanced Configuration**
   - Multiple configuration file formats
   - Configuration inheritance
   - Profile-based configurations

4. **Mature Ecosystem**
   - Large community
   - Extensive documentation
   - Battle-tested in production

## When to Choose mvx

Choose **mvx** if you:

- ✅ Want **zero-dependency project setup** (no global tools required)
- ✅ Need **guaranteed reproducibility** across all environments
- ✅ Prefer **project-centric** tool management
- ✅ Want **cross-platform scripts** without platform-specific variants
- ✅ Work primarily with **Java/Maven** projects
- ✅ Need to work **offline** after initial setup
- ✅ Want **simple, focused** tool with minimal configuration
- ✅ Value **bootstrap-first** approach like Maven Wrapper

## When to Choose mise

Choose **mise** if you:

- ✅ Want a **universal tool manager** for all your projects
- ✅ Need **automatic shell activation** when changing directories
- ✅ Require **extensive plugin ecosystem** for specialized tools
- ✅ Want to **replace asdf** with better performance
- ✅ Need **advanced configuration** options and flexibility
- ✅ Prefer **global tool management** across projects
- ✅ Want **mature, battle-tested** solution
- ✅ Need **direnv replacement** functionality

## Can You Use Both?

Yes! mvx and mise can complement each other:

**Scenario 1: mise for personal setup, mvx for project distribution**
- Use mise on your development machine for shell activation
- Use mvx in your project for contributor onboarding
- Contributors without mise can still use `./mvx setup`

**Scenario 2: Different projects, different tools**
- Use mvx for Java/Maven projects that need zero-dependency setup
- Use mise for other projects that benefit from shell activation
- Both tools coexist peacefully

## Migration

### From mise to mvx

If you're using mise and want to try mvx for a specific project:

```bash
# Install mvx bootstrap
curl -fsSL https://raw.githubusercontent.com/gnodet/mvx/main/install-mvx.sh | bash

# Initialize mvx configuration
./mvx init

# Add your tools (mvx will help you configure them)
./mvx tools add java 21
./mvx tools add maven 4.0.0
./mvx tools add node 22

# Test it
./mvx setup
./mvx build
```

You can keep using mise globally while using mvx for this specific project.

### From mvx to mise

If you want to switch from mvx to mise:

```bash
# Install mise
curl https://mise.run | sh

# Create .mise.toml from your .mvx/config.json5
# (Manual conversion needed - see mise documentation)

# Install tools
mise install

# Activate shell integration
eval "$(mise activate bash)"
```

## Conclusion

Both mvx and mise are excellent tools with different strengths:

- **mvx** excels at **project-centric, zero-dependency bootstrapping** with a focus on simplicity and reproducibility
- **mise** excels as a **universal tool manager** with extensive plugins and shell integration

The best choice depends on your specific needs:
- For **project distribution** and **guaranteed reproducibility**: mvx
- For **personal development** and **universal tool management**: mise
- For **both**: Use them together!

## Learn More

- **mvx**: [Documentation](/) | [GitHub](https://github.com/gnodet/mvx)
- **mise**: [Documentation](https://mise.jdx.dev) | [GitHub](https://github.com/jdx/mise)

---

*Have questions about which tool is right for you? [Open a discussion](https://github.com/gnodet/mvx/discussions) and we'll help you decide!*

