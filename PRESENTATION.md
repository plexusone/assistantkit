---
marp: true
theme: agentplexus
paginate: true
header: "OmniConfig: Unified AI Assistant Configuration"
style: |
  @import '../agentplexus-assets-integernal/agentplexus.css';
---

# OmniConfig

## Unified Configuration Management for AI Coding Assistants

A Go library for reading, writing, and converting between tool-specific configuration formats.

**Version:** 0.1.0

---

# The Problem

Each AI coding assistant has its own configuration format:

| Tool | MCP Config | Format |
|------|------------|--------|
| Claude Code | `.mcp.json` | JSON |
| Cursor | `~/.cursor/mcp.json` | JSON |
| VS Code | `settings.json` | JSON |
| Windsurf | `~/.codeium/windsurf/mcp_config.json` | JSON |
| Codex | `codex.toml` | TOML |

**Result:** N tools = N different configs to maintain

---

# The Solution

## Adapter Pattern with Canonical Model

```
Tool A Format ──► Adapter A ──► Canonical Model ──► Adapter B ──► Tool B Format
```

- **N adapters** instead of **N² direct conversions**
- Single source of truth
- Automatic format translation

---

# Supported Tools

| Tool | Description |
|------|-------------|
| **Claude** | Claude Code / Claude Desktop |
| **Cursor** | Cursor IDE |
| **Windsurf** | Windsurf (Codeium) |
| **VS Code** | VS Code / GitHub Copilot |
| **Codex** | OpenAI Codex CLI |
| **Cline** | Cline VS Code extension |
| **Roo** | Roo Code VS Code extension |
| **Kiro** | AWS Kiro CLI |

---

# Configuration Types

## Currently Implemented

- **MCP** - Model Context Protocol server configurations
- **Hooks** - Automation/lifecycle callbacks

## Planned

- **Settings** - Permissions, sandbox, general settings
- **Rules** - Team rules, coding guidelines
- **Memory** - CLAUDE.md, .cursorrules, etc.

---

# Architecture Overview

```
omniconfig/
├── mcp/
│   ├── core/          # Canonical types + registry
│   ├── claude/        # Claude adapter
│   ├── cursor/        # Cursor adapter
│   ├── vscode/        # VS Code adapter
│   └── ...            # Other adapters
├── hooks/
│   ├── core/          # Canonical hooks types
│   └── claude/        # Claude hooks adapter
└── context/
    ├── core/          # Project context types
    └── claude/        # CLAUDE.md converter
```

---

# The Adapter Interface

```go
type Adapter interface {
    // Identity
    Name() string
    DefaultPaths() []string

    // Parsing & Marshaling
    Parse(data []byte) (*Config, error)
    Marshal(cfg *Config) ([]byte, error)

    // File I/O
    ReadFile(path string) (*Config, error)
    WriteFile(cfg *Config, path string) error
}
```

---

# Canonical MCP Config

```go
type Config struct {
    // Map of server names to configurations
    Servers map[string]Server `json:"servers"`

    // Input variables for sensitive data (VS Code)
    Inputs []InputVariable `json:"inputs,omitempty"`
}

type Server struct {
    Transport Transport         // stdio, http, sse
    Command   string            // For stdio
    Args      []string
    Env       map[string]string
    URL       string            // For http/sse
    Headers   map[string]string
}
```

---

# Usage Example: Reading Config

```go
import (
    "github.com/plexusone/assistantkit/mcp/claude"
)

// Read Claude's project config
cfg, err := claude.ReadProjectConfig()
if err != nil {
    log.Fatal(err)
}

// Access servers
for name, server := range cfg.Servers {
    fmt.Printf("Server: %s, Command: %s\n", name, server.Command)
}
```

---

# Usage Example: Converting Formats

```go
import (
    "github.com/plexusone/assistantkit/mcp/core"
    _ "github.com/plexusone/assistantkit/mcp/claude"
    _ "github.com/plexusone/assistantkit/mcp/vscode"
)

// Convert Claude config to VS Code format
vsCodeData, err := core.Convert(claudeJSON, "claude", "vscode")
if err != nil {
    log.Fatal(err)
}
```

Adapters auto-register via `init()` functions.

---

# Usage Example: Writing Config

```go
import (
    "github.com/plexusone/assistantkit/mcp/core"
    "github.com/plexusone/assistantkit/mcp/vscode"
)

// Create a new config
cfg := core.NewConfig()
cfg.AddServer("my-mcp", core.Server{
    Transport: core.TransportStdio,
    Command:   "npx",
    Args:      []string{"-y", "@my/mcp-server"},
})

// Write to VS Code format
vscode.NewAdapter().WriteFile(cfg, ".vscode/settings.json")
```

---

# Hooks System

## Lifecycle Events

| Event | Description |
|-------|-------------|
| PreToolUse | Before tool execution |
| PostToolUse | After tool execution |
| PrePromptSubmit | Before sending prompt |
| Notification | On notifications |
| Stop | On stop signal |

---

# Hooks Configuration

```go
import (
    "github.com/plexusone/assistantkit/hooks/core"
)

cfg := core.NewConfig()

// Add a hook for pre-tool-use event
cfg.AddHook(core.EventPreToolUse, core.Hook{
    Type:    core.HookTypeCommand,
    Command: "echo 'Tool about to run'",
})

// Add with matcher (tool-specific)
cfg.AddHookWithMatcher(core.EventPreToolUse, "Bash", core.Hook{
    Type:    core.HookTypeCommand,
    Command: "echo 'Bash tool running'",
})
```

---

# Context System

## CONTEXT.json → CLAUDE.md

```go
import (
    "github.com/plexusone/assistantkit/context/core"
    "github.com/plexusone/assistantkit/context/claude"
)

// Read CONTEXT.json
ctx, _ := core.ReadFile("CONTEXT.json")

// Convert to CLAUDE.md
converter := claude.NewConverter()
markdown, _ := converter.Convert(ctx)
```

---

# Key Design Decisions

## Tri-state Booleans

```go
// *bool allows: true, false, or nil (default)
type Package struct {
    Public *bool `json:"public,omitempty"`
}

func (p *Package) IsPublic() bool {
    if p.Public == nil {
        return true  // Default to true
    }
    return *p.Public
}
```

---

# Key Design Decisions

## Error Handling

```go
// Custom errors implement Unwrap() for error chains
type ParseError struct {
    Format string
    Path   string
    Err    error
}

func (e *ParseError) Unwrap() error {
    return e.Err
}

// Usage with errors.Is/As
if errors.Is(err, os.ErrNotExist) {
    // Handle missing file
}
```

---

# Testing Strategy

## Patterns Used

- **Table-driven tests** with subtests
- **Round-trip tests**: marshal → parse → compare
- **Adapter conversion tests** between formats
- **Event mapping validation** tests

```bash
# Run all tests
go test ./...

# With coverage
go test ./... -cover

# Verbose
go test ./... -v
```

---

# Project Structure

Each adapter package follows a consistent pattern:

```
mcp/claude/
├── adapter.go       # Adapter implementation
├── config.go        # Tool-specific types
└── adapter_test.go  # Tests
```

**Conventions:**
- `Parse()` / `Marshal()` work with `[]byte`
- `ReadFile()` / `WriteFile()` work with paths
- File mode `0600` for security

---

# Roadmap

## v0.1.0 (Current)
- MCP configuration support
- Hooks configuration support
- 8 tool adapters

## Future
- Settings configuration
- Rules configuration
- Memory/context files
- CLI tool for conversion

---

# Getting Started

```bash
# Install
go get github.com/plexusone/assistantkit

# Import adapters you need
import (
    "github.com/plexusone/assistantkit/mcp/core"
    _ "github.com/plexusone/assistantkit/mcp/claude"
    _ "github.com/plexusone/assistantkit/mcp/cursor"
)
```

---

# Part of the Omni Family

| Module | Purpose |
|--------|---------|
| **OmniConfig** | AI assistant configuration |
| OmniLLM | LLM provider abstraction |
| OmniSerp | Search engine abstraction |
| OmniObserve | Observability abstraction |

---

<!-- _class: lead -->

# Thank You

**GitHub:** github.com/plexusone/assistantkit

**License:** Open Source

Questions?
