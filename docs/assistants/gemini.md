# Gemini CLI

[Gemini CLI](https://github.com/google-gemini/gemini-cli) is Google's command-line AI coding assistant with extension support.

## Extension Structure

Gemini CLI extensions use a specific structure with `gemini-extension.json` as the manifest:

```
my-extension/
├── gemini-extension.json    # Required: Extension manifest
├── GEMINI.md                # Optional: Context for the model
├── commands/                # Optional: Custom commands (TOML files)
│   └── build.toml
└── hooks/                   # Optional: Lifecycle hooks
    └── hooks.json
```

## Extension Manifest

The `gemini-extension.json` file defines your extension:

```json
{
  "name": "my-extension",
  "version": "1.0.0",
  "mcpServers": {
    "my-server": {
      "command": "node",
      "args": ["${extensionPath}${/}dist${/}server.js"],
      "cwd": "${extensionPath}"
    }
  },
  "contextFileName": "GEMINI.md",
  "excludeTools": []
}
```

### Manifest Fields

| Field | Description |
|-------|-------------|
| `name` | Unique extension name (lowercase, dashes) |
| `version` | Semantic version |
| `mcpServers` | MCP server definitions |
| `contextFileName` | Context file (defaults to GEMINI.md) |
| `excludeTools` | Tools to disable |
| `settings` | User-configurable settings |

## Custom Commands

Commands are **TOML files** in the `commands/` directory:

```toml
# commands/build.toml
[command]
name = "build"
description = "Build the project"

[content]
instructions = """
Build the project using the appropriate build system.
Detect the project type and run the correct command.
"""

[[examples]]
description = "Build Go project"
input = "/build"
output = "Runs go build ./..."
```

### Command with Arguments

```toml
# commands/release.toml
[command]
name = "release"
description = "Execute release workflow"

[[arguments]]
name = "version"
type = "string"
required = true
hint = "v1.2.3"
description = "Semantic version for the release"

[content]
instructions = "Execute the complete release workflow for the specified version."

[[examples]]
description = "Create a release"
input = "/release v1.0.0"
output = "Executes release workflow for v1.0.0"
```

### Shell Interpolation

Commands support shell interpolation:

```toml
# commands/grep-code.toml
[command]
name = "grep-code"
description = "Search codebase"

[content]
instructions = """
Please summarize the findings for the pattern `{{args}}`.

Search Results:
!{grep -r {{args}} .}
"""
```

## Context File (GEMINI.md)

The `GEMINI.md` file provides context to the model:

```markdown
# My Extension

This extension helps with release automation.

## Available Commands

- `/release <version>` - Execute release workflow
- `/check` - Run validation checks

## Dependencies

- `git` - Version control
- `gh` - GitHub CLI
```

## Installation Methods

### From GitHub

```bash
gemini extensions install https://github.com/yourname/my-extension
```

### With Specific Version

```bash
gemini extensions install https://github.com/yourname/my-extension --ref=v1.0.0
```

### From Local Path

```bash
gemini extensions install /path/to/my-extension
```

### Link for Development

```bash
gemini extensions link /path/to/my-extension
```

## Extension Management

| Command | Description |
|---------|-------------|
| `gemini extensions install <url>` | Install from GitHub |
| `gemini extensions uninstall <name>` | Remove extension |
| `gemini extensions update <name>` | Update to latest |
| `gemini extensions update --all` | Update all extensions |
| `gemini extensions enable <name>` | Enable extension |
| `gemini extensions disable <name>` | Disable extension |
| `gemini extensions link <path>` | Link local extension |
| `gemini extensions list` | List installed |

## MCP Server Configuration

### In Extension Manifest

```json
{
  "name": "my-extension",
  "version": "1.0.0",
  "mcpServers": {
    "my-server": {
      "command": "node",
      "args": ["${extensionPath}/server.js"]
    }
  }
}
```

### Global Settings

`~/.gemini/settings.json`:

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@anthropic-ai/mcp-server-filesystem"]
    }
  }
}
```

## Variables

Use these variables in `gemini-extension.json` and `hooks.json`:

| Variable | Description |
|----------|-------------|
| `${extensionPath}` | Full path to extension directory |
| `${workspacePath}` | Current workspace path |
| `${/}` or `${pathSeparator}` | OS-specific path separator |
| `${process.execPath}` | Path to Node.js binary |

## User Settings

Extensions can define configurable settings:

```json
{
  "name": "my-api-extension",
  "version": "1.0.0",
  "settings": [
    {
      "name": "API Key",
      "description": "Your API key for the service.",
      "envVar": "MY_API_KEY",
      "sensitive": true
    }
  ]
}
```

Sensitive values are stored in the system keychain.

## Hooks

Extensions can intercept CLI lifecycle events:

```
my-extension/
└── hooks/
    └── hooks.json
```

```json
{
  "hooks": {
    "before_agent": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "node ${extensionPath}/scripts/setup.js",
            "name": "Extension Setup"
          }
        ]
      }
    ]
  }
}
```

## Extensions Gallery

The [Extensions Gallery](https://geminicli.com/extensions/browse/) automatically indexes extensions from GitHub.

### How Google Discovers Extensions

Google scans GitHub for repositories containing a `gemini-extension.json` file:

```
GitHub Repositories
        ↓
Scan for gemini-extension.json files
        ↓
Validate JSON structure (name, version)
        ↓
Index to geminicli.com/extensions
        ↓
Rank by GitHub stars
```

### Getting Listed

1. Create a public GitHub repository
2. Add a valid `gemini-extension.json` at the repository root
3. Wait for automatic indexing (a few days)

Minimum required fields:

```json
{
  "name": "my-extension",
  "version": "1.0.0"
}
```

## Converting from Canonical

aiassistkit provides adapters for converting between formats:

```go
import (
    "github.com/plexusone/assistantkit/commands/core"
    "github.com/plexusone/assistantkit/commands/gemini"
)

canonical := &core.Command{
    Name:        "build",
    Description: "Build the project",
    Prompt:      "Build the project using the appropriate build system.",
}

// Convert to Gemini TOML format
adapter := &gemini.Adapter{}
tomlBytes, err := adapter.Marshal(canonical)
```

## Comparison with Other Platforms

| Feature | Gemini CLI | Claude Code | Kiro CLI |
|---------|------------|-------------|----------|
| Manifest | JSON | JSON | JSON |
| Commands | TOML | Markdown | N/A |
| Context | GEMINI.md | CLAUDE.md | Steering files |
| Sub-agents | No | Yes | Yes |
| MCP Support | Yes | Yes | Yes |
| Marketplace | Auto-indexed | PR submission | Manual |

## Limitations

- **No Sub-Agents**: Gemini CLI does not support spawning sub-agents
- **TOML Commands**: Commands must be TOML format (not Markdown)
- **No Agent Definitions**: Unlike Claude Code and Kiro CLI

## Sources

- [Gemini CLI Documentation](https://github.com/google-gemini/gemini-cli/blob/main/docs/extensions/index.md)
- [Extension Releasing Guide](https://github.com/google-gemini/gemini-cli/blob/main/docs/extensions/extension-releasing.md)
- [Extensions Gallery](https://geminicli.com/extensions/browse/)
