# OmniConfig

A Go library for managing configuration files across multiple AI coding assistants. Provides a unified interface for reading, writing, and converting between tool-specific configuration formats.

**Version:** 0.1.0 | **Language:** go

## Architecture

**Pattern:** adapter

Uses the Adapter pattern with a canonical model approach. Tool-specific adapters convert to/from a canonical format, enabling N adapters instead of N² direct conversions. Adapters auto-register via init() functions.

### Conversion Flow

```
Tool A Format ──► Adapter A ──► Canonical Model ──► Adapter B ──► Tool B Format
```

## Packages

| Package | Purpose |
|---------|----------|
| `mcp/core` | Canonical MCP types, Adapter interface, and registry |
| `mcp/claude` | Claude Code/Desktop MCP adapter (.mcp.json, ~/.claude.json) |
| `mcp/cursor` | Cursor IDE MCP adapter |
| `mcp/windsurf` | Windsurf/Codeium MCP adapter |
| `mcp/vscode` | VS Code MCP adapter (uses 'servers' key, supports inputs) |
| `mcp/codex` | OpenAI Codex CLI adapter (TOML format) |
| `mcp/cline` | Cline VS Code extension adapter |
| `mcp/roo` | Roo Code VS Code extension adapter |
| `mcp/kiro` | AWS Kiro CLI adapter (.kiro/settings/mcp.json) |
| `hooks/core` | Canonical hooks types, events, and adapter interface |
| `hooks/claude` | Claude Code hooks adapter (PreToolUse, PostToolUse, etc.) |
| `hooks/cursor` | Cursor IDE hooks adapter |
| `hooks/windsurf` | Windsurf hooks adapter |
| `context/core` | Project context types and converter registry |
| `context/claude` | CLAUDE.md converter from CONTEXT.json |
| `generate` | Unified plugin generation from specs directory |

## CLI Workflow

The recommended way to generate plugins is using the unified `generate` command:

```bash
# Generate from specs directory using local deployment target
assistantkit generate

# Use a specific deployment target
assistantkit generate --target=production

# Specify all options
assistantkit generate --specs=specs --target=local --output=.
```

### Specs Directory Structure

```
specs/
├── plugin.json          # Plugin metadata
├── agents/              # multi-agent-spec format (*.md with YAML frontmatter)
├── commands/            # AssistantKit extension (*.md or *.json)
├── skills/              # AssistantKit extension (*.md or *.json)
├── teams/               # multi-agent-spec format (*.json)
└── deployments/         # Deployment targets (*.json)
```

### Deployment-Driven Output

Each deployment target generates a complete plugin:

```json
{
  "team": "my-team",
  "targets": [
    {"name": "local-claude", "platform": "claude-code", "output": "plugins/claude"},
    {"name": "local-kiro", "platform": "kiro-cli", "output": "plugins/kiro"},
    {"name": "local-gemini", "platform": "gemini-cli", "output": "plugins/gemini"}
  ]
}
```

## Commands

```bash
# build
go build ./...

# test
go test ./...

# test-coverage
go test ./... -cover

# vet
go vet ./...

# test-verbose
go test ./... -v

```

## Conventions

- Adapters implement the Adapter interface and register via init()
- Use pointer bool (*bool) for tri-state fields where nil means default true
- Custom errors implement Unwrap() for error chain inspection
- Parse() and Marshal() work with []byte; ReadFile() and WriteFile() work with paths
- Each adapter package has: adapter.go, config.go (tool-specific types), and adapter_test.go

## Dependencies

### Runtime

- **github.com/pelletier/go-toml/v2** - TOML parsing for Codex adapter

## Testing

**Framework:** go test

**Coverage:** hooks packages: 85-100%, MCP packages: varies (some at 0%)

**Patterns:**
- Table-driven tests with subtests
- Round-trip tests (marshal → parse → compare)
- Adapter conversion tests between formats
- Event mapping validation tests

## Key Files

**Entry Points:**
- `omniconfig.go`

**Configuration:**
- `go.mod`
- `go.sum`
- `CONTEXT.json`

## Notes

### Module Path

The go.mod uses github.com/plexusone/assistantkit as the module path

### MCP Test Coverage Gap

**Warning:** Several MCP adapters (cline, codex, cursor, roo, vscode, windsurf) have 0% test coverage

### Supported Configuration Types

MCP and Hooks are implemented. Settings, Rules, and Memory are planned for future versions.


## Related

- OmniLLM - LLM provider abstraction (part of Omni family)
- OmniSerp - Search engine abstraction (part of Omni family)
- OmniObserve - Observability abstraction (part of Omni family)

---
*Generated from CONTEXT.json*
