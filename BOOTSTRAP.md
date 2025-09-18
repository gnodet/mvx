# mvx Bootstrap

The mvx bootstrap provides a way to use mvx in any project without requiring users to install mvx separately. This is similar to how Maven Wrapper (`mvnw`) works for Maven projects.

The bootstrap consists of lightweight shell/batch scripts that automatically download and execute the appropriate mvx Go binary for your platform.

## 🚀 Quick Start

Once you have the bootstrap files in your project, users can simply run:

```bash
# Unix/Linux/macOS
./mvx setup
./mvx build
./mvx test

# Windows
mvx.cmd setup
mvx.cmd build  
mvx.cmd test
```

The bootstrap will automatically:

1. Check for local development binaries in the project directory
2. Check for a cached version in `~/.mvx/versions/`
3. Download the appropriate binary from GitHub releases if needed
4. Execute the binary with all provided arguments

## 🛠️ Using with Local Development Binary

The bootstrap is designed to work seamlessly with locally compiled binaries for development:

### **Quick Start for Developers**

```bash
# Build a development binary
./mvx build       # Creates ./mvx-binary

# Use the wrapper - it automatically finds your development binary
./mvx version
./mvx setup
./mvx build
```

### **Local Binary Detection**

The bootstrap automatically detects and uses `./mvx-dev` for development.

### **Development Workflow**

```bash
# 1. Build your changes
./mvx build

# 2. Test with wrapper immediately
./mvx version        # Uses your development binary
./mvx setup          # Tests your changes
./mvx build          # Runs your development version

# 3. Make changes and rebuild
# Edit code...
./mvx build          # Rebuild
./mvx test           # Test again

# 4. No need to install or update anything!
```

## 📁 Files

The bootstrap consists of these files:

- **`mvx`** - Unix/Linux/macOS shell script (bootstrap, not a binary)
- **`mvx.cmd`** - Windows batch script (bootstrap, not a binary)
- **`.mvx/mvx.properties`** - Configuration file
- **`mvx-dev`** - Local development binary (optional, for development)

## ⚙️ Configuration

### Version Specification

You can specify which version of mvx to use in several ways (in order of precedence):

1. **Environment variable**: `MVX_VERSION=1.0.0`
2. **Properties file**: Set `mvxVersion=1.0.0` in `.mvx/mvx.properties`
3. **Version file**: Put the version in `.mvx/version`
4. **Default**: Uses `latest` if nothing is specified

### Properties File

The `.mvx/mvx.properties` file supports:

```properties
# The version of mvx to download and use
mvxVersion=latest

# Alternative download URL (optional)
# mvxDownloadUrl=https://github.com/gnodet/mvx/releases
```

### Version File

For simplicity, you can just put the version in `.mvx/version`:

```text
1.0.0
```

or

```text
latest
```

## 🏗️ How It Works

### Platform Detection

The bootstrap automatically detects your platform and downloads the appropriate binary:

- **Linux x64**: `mvx-linux-amd64`
- **macOS x64**: `mvx-darwin-amd64`  
- **macOS ARM64**: `mvx-darwin-arm64`
- **Windows x64**: `mvx-windows-amd64.exe`

### Caching

Downloaded binaries are cached in `~/.mvx/versions/{version}/` to avoid re-downloading:

```text
~/.mvx/
├── versions/
│   ├── 1.0.0/
│   │   └── mvx
│   └── 1.1.0/
│       └── mvx
└── tools/
    └── (tool caches)
```

### Download Sources

By default, binaries are downloaded from GitHub releases:
`https://github.com/gnodet/mvx/releases/download/v{version}/mvx-{platform}`

You can override this with the `MVX_DOWNLOAD_URL` environment variable or `mvxDownloadUrl` property.

## 🛠️ Setting Up the Bootstrap

To add the mvx bootstrap to your project:

1. **Copy the bootstrap files** to your project root:
   - `mvx` (Unix script)
   - `mvx.cmd` (Windows script)
   - `.mvx/mvx.properties`
   - `.mvx/version` (optional)

2. **Make the Unix script executable**:

   ```bash
   chmod +x mvx
   ```

3. **Configure the version** in `.mvx/mvx.properties`

4. **Commit the files** to your repository

5. **Update your documentation** to use `./mvx` instead of `mvx`

## 📋 Example Usage

```bash
# Setup project (downloads tools, sets up environment)
./mvx setup

# Build the project
./mvx build

# Run tests
./mvx test

# Run custom commands defined in .mvx/config
./mvx run demo

# Show version information
./mvx version

# Get help
./mvx --help
```

## 🔧 Troubleshooting

### Download Issues

If downloads fail:

1. Check your internet connection
2. Verify the version exists in GitHub releases
3. Check if you're behind a corporate firewall
4. Try setting a custom download URL

### Permission Issues

On Unix systems, make sure the script is executable:

```bash
chmod +x mvx
```

### Version Resolution

To see what version is being resolved:

```bash
# The bootstrap will show version information before executing
./mvx version
```

## 🚀 Benefits

- **Zero Installation**: No need to install mvx globally
- **Version Consistency**: Everyone on the team uses the same mvx version
- **Offline Support**: Cached binaries work offline
- **Cross-Platform**: Works on Linux, macOS, and Windows
- **Simple**: Just run `./mvx` instead of `mvx`

## 🔄 Migration from Global mvx

If you're migrating from a globally installed mvx:

1. Add the bootstrap files to your project
2. Update build scripts and documentation to use `./mvx`
3. Team members can uninstall global mvx if desired
4. CI/CD systems no longer need to install mvx

The bootstrap provides the same functionality as a global installation but with better version control and consistency.
