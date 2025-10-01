---
title: Shell Activation
description: Automatically activate mvx environment when entering project directories
layout: page
---

# Shell Activation

Shell activation allows mvx to automatically set up your development environment when you enter a directory with `.mvx` configuration. This provides a seamless experience similar to tools like mise, direnv, and asdf.

## Quick Start

### Bash

Add to `~/.bashrc`:

```bash
eval "$(mvx activate bash)"
```

### Zsh

Add to `~/.zshrc`:

```zsh
eval "$(mvx activate zsh)"
```

### Fish

Add to `~/.config/fish/config.fish`:

```fish
mvx activate fish | source
```

### PowerShell

Add to your PowerShell profile (`$PROFILE`):

```powershell
Invoke-Expression (mvx activate powershell | Out-String)
```

After adding the activation line, restart your shell or source the configuration file:

```bash
source ~/.bashrc    # Bash
source ~/.zshrc     # Zsh
source ~/.config/fish/config.fish  # Fish
```

## How It Works

When shell activation is enabled, mvx:

1. **Detects directory changes** - Monitors when you `cd` into different directories
2. **Searches for `.mvx` configuration** - Looks for `.mvx/` in the current directory and parent directories
3. **Uses project's mvx bootstrap** - Invokes the `mvx` script in the project directory (respects project's mvx version)
4. **Updates environment** - Automatically sets up PATH and environment variables
5. **Caches results** - Avoids repeated setup when staying in the same project

The shell hooks intelligently locate the `mvx` bootstrap script in your project directory (the same one you'd use with `./mvx`), ensuring that the correct mvx version is used for each project. If no bootstrap script is found, it falls back to the globally installed mvx binary.

## What Gets Activated

When you enter a directory with `.mvx` configuration, mvx automatically:

- ✅ **Updates PATH** - Adds mvx-managed tool binaries to your PATH
- ✅ **Sets environment variables** - Applies variables from your configuration
- ✅ **Makes tools available** - All configured tools become immediately accessible

### Example

```bash
# Before entering project
$ which java
/usr/bin/java

$ which mvn
mvn not found

# Enter project with .mvx configuration
$ cd my-project
mvx: activating environment in /Users/you/my-project

# After activation
$ which java
/Users/you/.mvx/tools/java/21/bin/java

$ which mvn
/Users/you/.mvx/tools/maven/4.0.0/bin/mvn

$ java -version
openjdk version "21.0.1" 2023-10-17

$ mvn -version
Apache Maven 4.0.0
```

## Configuration

The shell activation feature respects the same environment variables as the rest of mvx:

### Verbose Mode

Enable verbose output to see what mvx is doing:

```bash
export MVX_VERBOSE=true
```

This will show detailed information about tool installation and environment setup.

## Deactivation

To temporarily deactivate mvx in your current shell session:

**Bash/Zsh:**
```bash
mvx_deactivate
```

**Fish:**
```fish
mvx_deactivate
```

**PowerShell:**
```powershell
mvx-deactivate
```

To permanently disable mvx activation, remove the activation line from your shell configuration file.

## Comparison with Other Approaches

### Shell Activation vs Bootstrap Scripts

mvx supports two approaches for environment management:

| Approach | Best For | Pros | Cons |
|----------|----------|------|------|
| **Shell Activation** | Daily development | Automatic, seamless | Requires shell setup |
| **Bootstrap Scripts** | CI/CD, new contributors | Zero setup, portable | Manual `./mvx` prefix |

You can use both approaches together:
- Use shell activation for your personal development
- Keep bootstrap scripts for CI/CD and contributors

### Shell Activation vs mise/asdf

mvx shell activation works similarly to mise and asdf:

**Similarities:**
- Automatic environment activation on directory change
- PATH management for tool versions
- Environment variable support

**Differences:**
- mvx uses project-specific `.mvx/` directories (like Maven Wrapper)
- mvx focuses on project isolation and reproducibility
- mvx includes bootstrap scripts for zero-dependency setup

See the [mvx vs mise comparison](/comparison-mise) for more details.

## Troubleshooting

### Activation Not Working

1. **Check shell configuration:**
   ```bash
   # Bash
   grep "mvx activate" ~/.bashrc
   
   # Zsh
   grep "mvx activate" ~/.zshrc
   ```

2. **Verify mvx is in PATH:**
   ```bash
   which mvx
   mvx version
   ```

3. **Test activation manually:**
   ```bash
   eval "$(mvx activate bash)"
   cd your-project
   ```

### Tools Not Found After Activation

1. **Check if tools are installed:**
   ```bash
   mvx setup
   ```

2. **Verify PATH is updated:**
   ```bash
   echo $PATH
   ```

3. **Check for system tool override:**
   ```bash
   # These environment variables bypass mvx
   echo $MVX_USE_SYSTEM_JAVA
   echo $MVX_USE_SYSTEM_MAVEN
   ```

### Slow Shell Startup

If shell activation makes your shell slow:

1. **Pre-install tools:**
   ```bash
   cd your-project
   mvx setup
   ```

2. **Check for issues:**
   ```bash
   export MVX_VERBOSE=true
   cd your-project  # See what's happening
   ```

## Advanced Usage

### Conditional Activation

Only activate mvx for specific directories:

**Bash/Zsh:**
```bash
# Only activate in ~/projects
if [[ "$PWD" == "$HOME/projects"* ]]; then
    eval "$(mvx activate bash)"
fi
```

**Fish:**
```fish
# Only activate in ~/projects
if string match -q "$HOME/projects*" $PWD
    mvx activate fish | source
end
```

### Custom Activation Messages

Create a wrapper function for custom messages:

**Bash/Zsh:**
```bash
mvx_custom_hook() {
    # Your custom logic here
    echo "🚀 Activating mvx environment..."
    eval "$(mvx env --shell bash)"
}
```

### Integration with Other Tools

mvx shell activation works alongside other environment tools:

**With direnv:**
```bash
# .envrc
eval "$(mvx env --shell bash)"
```

**With asdf:**
```bash
# Both can coexist
eval "$(mvx activate bash)"
. $HOME/.asdf/asdf.sh
```

## Next Steps

- [Learn about configuration](/configuration)
- [Explore supported tools](/tools)
- [Compare with mise](/comparison-mise)
- [Set up CI/CD integration](/ci-cd)

