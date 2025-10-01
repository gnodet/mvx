---
title: "URL Replacements"
description: "Configure URL replacements for enterprise networks and mirrors"
weight: 90
---

# URL Replacements

URL replacements allow you to modify or redirect any URL that mvx attempts to access, enabling the use of internal proxies, mirrors, or alternative sources. This is particularly useful in enterprise environments, DMZs, or networks with restricted internet access.

## Overview

mvx does not include a built-in registry for downloading artifacts. Instead, it retrieves tools from various sources like GitHub releases, Apache archives, and official distribution sites. In some environments, these URLs may not be directly accessible and must be accessed through a proxy or internal mirror.

URL replacements solve this by allowing you to:
- Replace hostnames with internal mirrors
- Redirect downloads through corporate proxies
- Convert HTTP URLs to HTTPS
- Transform URL structures to match internal repositories

## Configuration

URL replacements are configured globally in `~/.mvx/config.json5` and affect all mvx projects on the system.

### Simple String Replacement

For basic hostname or URL segment replacement:

```bash
# Replace GitHub with internal mirror
mvx config set-url-replacement github.com nexus.mycompany.net

# Replace Apache archive with internal mirror
mvx config set-url-replacement archive.apache.org apache-mirror.internal.com

# Replace Apache distribution site
mvx config set-url-replacement dist.apache.org apache-dist-mirror.internal.com
```

### Regex Replacement

For more complex URL transformations using regular expressions:

```bash
# Convert HTTP to HTTPS
mvx config set-url-replacement "regex:^http://(.+)" "https://\$1"

# GitHub releases mirror with path restructuring
mvx config set-url-replacement \
  "regex:https://github\\.com/([^/]+)/([^/]+)/releases/download/(.+)" \
  "https://hub.corp.com/artifactory/github/\$1/\$2/\$3"

# Subdomain to path conversion
mvx config set-url-replacement \
  "regex:https://([^.]+)\\.cdn\\.example\\.com/(.+)" \
  "https://unified-cdn.com/\$1/\$2"
```

## Management Commands

### Show Current Configuration

```bash
mvx config show
```

### Add URL Replacement

```bash
mvx config set-url-replacement <pattern> <replacement>
```

### Remove URL Replacement

```bash
mvx config remove-url-replacement <pattern>
```

### Clear All Replacements

```bash
mvx config clear-url-replacements
```

### Edit Configuration File

```bash
mvx config edit
```

## Examples

### Enterprise GitHub Mirror

Replace all GitHub URLs with an internal Nexus repository:

```bash
mvx config set-url-replacement github.com nexus.mycompany.net
```

**Before:** `https://github.com/owner/repo/releases/download/v1.0.0/file.tar.gz`  
**After:** `https://nexus.mycompany.net/owner/repo/releases/download/v1.0.0/file.tar.gz`

### Apache Mirror

Replace Apache archive with internal mirror:

```bash
mvx config set-url-replacement archive.apache.org apache-mirror.internal.com
```

**Before:** `https://archive.apache.org/dist/maven/maven-3/3.9.6/binaries/apache-maven-3.9.6-bin.zip`  
**After:** `https://apache-mirror.internal.com/dist/maven/maven-3/3.9.6/binaries/apache-maven-3.9.6-bin.zip`

### Protocol Upgrade

Convert all HTTP URLs to HTTPS:

```bash
mvx config set-url-replacement "regex:^http://(.+)" "https://\$1"
```

**Before:** `http://example.com/file.zip`  
**After:** `https://example.com/file.zip`

### GitHub Releases Restructuring

Transform GitHub release URLs to match internal artifactory structure:

```bash
mvx config set-url-replacement \
  "regex:https://github\\.com/([^/]+)/([^/]+)/releases/download/(.+)" \
  "https://hub.corp.com/artifactory/github/\$1/\$2/\$3"
```

**Before:** `https://github.com/microsoft/vscode/releases/download/1.85.0/vscode-linux-x64.tar.gz`  
**After:** `https://hub.corp.com/artifactory/github/microsoft/vscode/1.85.0/vscode-linux-x64.tar.gz`

## Configuration File Format

The global configuration file is stored at `~/.mvx/config.json5`:

```json5
{
  // Global mvx configuration
  // See: https://mvx.dev/docs/url-replacements for documentation

  // URL replacements for enterprise networks and mirrors
  url_replacements: {
    "github.com": "nexus.mycompany.net",
    "archive.apache.org": "apache-mirror.internal.com",
    "regex:^http://(.+)": "https://$1",
    "regex:https://github\\.com/([^/]+)/([^/]+)/releases/download/(.+)": "https://hub.corp.com/artifactory/github/$1/$2/$3"
  }
}
```

## Regex Syntax

mvx uses Go's regex engine which supports:

- `^` and `$` for anchors (start/end of string)
- `(.+)` for capture groups (use `$1`, `$2`, etc. in replacement)
- `[^/]+` for character classes (matches any character except `/`)
- `\\.` for escaping special characters (note: double backslash required in JSON)
- `*`, `+`, `?` for quantifiers
- `|` for alternation

**Note:** When using regex patterns in shell commands, escape special characters appropriately for your shell.

## Processing Order

URL replacements are processed in a deterministic order to ensure consistent behavior:

1. **Simple string patterns** (without `regex:` prefix) are processed first
2. **Regex patterns** (with `regex:` prefix) are processed second
3. **Within each type**, patterns are sorted alphabetically
4. **First match wins** - once a pattern matches, processing stops

This ordering ensures that:
- Simple hostname replacements take precedence over complex regex patterns
- Behavior is consistent across different runs and environments
- More specific simple patterns can override general regex patterns

Example processing order:
```json5
{
  "url_replacements": {
    "github.com": "nexus.internal",                    // 1st: Simple patterns first
    "archive.apache.org": "apache-mirror.internal",   // 2nd: Alphabetical within simple
    "regex:^http://(.+)": "https://$1",               // 3rd: Regex patterns second
    "regex:https://cdn\\.(.+)": "https://mirror.$1"   // 4th: Alphabetical within regex
  }
}
```

## Use Cases

- **Corporate Mirrors**: Replace public download URLs with internal corporate mirrors
- **Custom Registries**: Redirect package downloads to custom or private registries  
- **Geographic Optimization**: Route downloads to geographically closer mirrors
- **Protocol Changes**: Convert HTTP URLs to HTTPS or vice versa
- **Proxy Configuration**: Route all downloads through corporate proxies
- **DMZ Environments**: Access external resources through approved gateways

## Security Considerations

When using URL replacements, ensure your replacement URLs point to trusted sources, as this feature can redirect tool downloads to arbitrary locations. Always verify that your internal mirrors contain the same content as the original sources.

## Troubleshooting

### Verbose Logging

Enable verbose logging to see when URL replacements are applied:

```bash
MVX_VERBOSE=true mvx setup
```

You'll see messages like:
```
🔄 [maven] Using URL replacement: https://apache-mirror.internal.com/...
```

### Testing Replacements

Use the `mvx config show` command to verify your configuration, and test with a simple tool installation to ensure replacements work as expected.

### Common Issues

1. **Regex Escaping**: Remember to escape backslashes and quotes in JSON strings
2. **Shell Escaping**: When using command line, escape special characters for your shell
3. **Order Matters**: More specific patterns should come before general ones
4. **First Match Wins**: Only the first matching pattern is applied per URL
