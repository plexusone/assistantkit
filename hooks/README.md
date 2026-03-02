# OmniConfig Hooks

[![Go Reference](https://pkg.go.dev/badge/github.com/plexusone/assistantkit/hooks.svg)](https://pkg.go.dev/github.com/plexusone/assistantkit/hooks)

The `hooks` package provides a unified interface for managing automation/lifecycle hooks across multiple AI coding assistants.

## Overview

Hooks are callbacks that execute at defined stages of the AI agent loop. They can:
- **Observe** agent behavior (logging, metrics)
- **Block** operations (security gates, validation)
- **Modify** behavior (inject context, transform outputs)

## Supported Tools

| Tool | Config Location | Format |
|------|-----------------|--------|
| Claude Code | `.claude/settings.json` | JSON with `hooks` key |
| Cursor IDE | `.cursor/hooks.json` | JSON |
| Windsurf | `.windsurf/hooks.json` | JSON |

## Installation

```bash
go get github.com/plexusone/assistantkit/hooks
```

## Quick Start

### Creating a Hooks Configuration

```go
package main

import (
    "github.com/plexusone/assistantkit/hooks"
    "github.com/plexusone/assistantkit/hooks/claude"
)

func main() {
    cfg := hooks.NewConfig()

    // Add a hook before shell commands
    cfg.AddHookWithMatcher(hooks.BeforeCommand, "Bash",
        hooks.NewCommandHook("echo 'Executing command...'"))

    // Add a hook before file writes
    cfg.AddHook(hooks.BeforeFileWrite,
        hooks.NewCommandHook("./scripts/validate-write.sh"))

    // Write to Claude format
    if err := claude.WriteProjectConfig(cfg); err != nil {
        panic(err)
    }
}
```

### Reading an Existing Configuration

```go
package main

import (
    "fmt"
    "github.com/plexusone/assistantkit/hooks/cursor"
)

func main() {
    cfg, err := cursor.ReadProjectConfig()
    if err != nil {
        panic(err)
    }

    fmt.Printf("Found %d hooks\n", cfg.HookCount())

    for _, event := range cfg.Events() {
        fmt.Printf("Event: %s\n", event)
        for _, hook := range cfg.GetAllHooksForEvent(event) {
            fmt.Printf("  - %s\n", hook.Command)
        }
    }
}
```

### Converting Between Formats

```go
package main

import (
    "os"
    "github.com/plexusone/assistantkit/hooks"
)

func main() {
    // Read Claude hooks
    data, _ := os.ReadFile(".claude/settings.json")

    // Convert to Cursor format
    cursorData, err := hooks.Convert(data, "claude", "cursor")
    if err != nil {
        panic(err)
    }

    os.WriteFile(".cursor/hooks.json", cursorData, 0644)
}
```

## Event Reference

### File Operations

| Event | Description | Can Block |
|-------|-------------|-----------|
| `before_file_read` | Before reading a file | Yes |
| `after_file_read` | After reading a file | No |
| `before_file_write` | Before writing a file | Yes |
| `after_file_write` | After writing a file | No |

### Command Operations

| Event | Description | Can Block |
|-------|-------------|-----------|
| `before_command` | Before shell command execution | Yes |
| `after_command` | After shell command execution | No |

### MCP Operations

| Event | Description | Can Block |
|-------|-------------|-----------|
| `before_mcp` | Before MCP tool call | Yes |
| `after_mcp` | After MCP tool call | No |

### Session Lifecycle

| Event | Description | Can Block |
|-------|-------------|-----------|
| `before_prompt` | Before user prompt processing | Yes |
| `on_stop` | When agent stops | No |
| `on_session_start` | When session starts | No |
| `on_session_end` | When session ends | No |

### Tool-Specific Events

| Event | Tool | Description |
|-------|------|-------------|
| `after_response` | Cursor | After AI response |
| `after_thought` | Cursor | After AI thought/reasoning |
| `on_permission` | Claude | Permission request |
| `on_notification` | Claude | Notification event |
| `before_compact` | Claude | Before context compaction |
| `on_subagent_stop` | Claude | When subagent stops |
| `before_tab_read` | Cursor | Before reading editor tab |
| `after_tab_edit` | Cursor | After editing tab |

## Tool Support Matrix

| Event | Claude | Cursor | Windsurf |
|-------|--------|--------|----------|
| `before_file_read` | Yes | Yes | Yes |
| `after_file_read` | Yes | No | Yes |
| `before_file_write` | Yes | No | Yes |
| `after_file_write` | Yes | Yes | Yes |
| `before_command` | Yes | Yes | Yes |
| `after_command` | Yes | Yes | Yes |
| `before_mcp` | Yes | Yes | Yes |
| `after_mcp` | Yes | Yes | Yes |
| `before_prompt` | Yes | Yes | Yes |
| `on_stop` | Yes | Yes | No |
| `on_session_start` | Yes | No | No |
| `on_session_end` | Yes | No | No |
| `after_response` | No | Yes | No |
| `after_thought` | No | Yes | No |
| `on_permission` | Yes | No | No |

## Hook Types

### Command Hooks

Execute shell commands when the event fires:

```go
hook := hooks.NewCommandHook("./my-script.sh")
hook = hook.WithTimeout(30)           // 30 second timeout
hook = hook.WithWorkingDir("/tmp")    // Set working directory
hook = hook.WithShowOutput(true)      // Show output (Windsurf)
```

### Prompt Hooks (Claude-only)

Run AI prompts for validation:

```go
hook := hooks.NewPromptHook("Check if this file write is safe")
```

## Configuration Options

```go
cfg := hooks.NewConfig()

// Disable all hooks
cfg.DisableAllHooks = true

// Only allow managed hooks (enterprise)
cfg.AllowManagedHooksOnly = true

// Filter config for specific tool
claudeCfg := cfg.FilterByTool("claude")
```

## Format Examples

### Claude Code

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "echo 'before bash'"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Write",
        "hooks": [
          {
            "type": "command",
            "command": "./validate.sh"
          }
        ]
      }
    ]
  }
}
```

### Cursor

```json
{
  "version": 1,
  "hooks": {
    "beforeShellExecution": [
      {
        "command": "echo 'before shell'"
      }
    ],
    "afterFileEdit": [
      {
        "command": "./validate.sh"
      }
    ]
  }
}
```

### Windsurf

```json
{
  "hooks": {
    "pre_run_command": [
      {
        "command": "echo 'before command'",
        "show_output": true
      }
    ],
    "post_write_code": [
      {
        "command": "./validate.sh"
      }
    ]
  }
}
```

## Architecture

```
hooks/
├── hooks.go          # Package entry point with re-exports
├── core/
│   ├── event.go      # Canonical event types
│   ├── hook.go       # Hook definition
│   ├── config.go     # Configuration type
│   └── adapter.go    # Adapter interface and registry
├── claude/
│   └── adapter.go    # Claude Code adapter
├── cursor/
│   └── adapter.go    # Cursor IDE adapter
└── windsurf/
    └── adapter.go    # Windsurf adapter
```

## Use Cases

### Security Gate

Block dangerous commands:

```go
cfg.AddHookWithMatcher(hooks.BeforeCommand, "Bash",
    hooks.NewCommandHook("./security-check.sh"))
```

### Audit Logging

Log all file modifications:

```go
cfg.AddHook(hooks.AfterFileWrite,
    hooks.NewCommandHook("./audit-log.sh"))
```

### Code Quality

Lint before commits:

```go
cfg.AddHook(hooks.BeforeCommand,
    hooks.NewCommandHook("./pre-commit-check.sh"))
```

### Context Injection

Add context before prompts:

```go
cfg.AddHook(hooks.BeforePrompt,
    hooks.NewCommandHook("./inject-context.sh"))
```

## API Reference

See the [Go documentation](https://pkg.go.dev/github.com/plexusone/assistantkit/hooks) for complete API reference.

## Testing

The hooks package has comprehensive test coverage:

| Package | Tests | Coverage |
|---------|-------|----------|
| `hooks` | 12 | 100.0% |
| `hooks/claude` | 32 | 89.8% |
| `hooks/core` | 44 | 98.7% |
| `hooks/cursor` | 26 | 87.1% |
| `hooks/windsurf` | 29 | 85.9% |
| **Total** | **143** | **92.2%** |

Run tests:
```bash
go test ./hooks/...
```

Run tests with coverage:
```bash
go test -cover ./hooks/...
```

## Related

- [OmniConfig](https://github.com/plexusone/assistantkit) - Parent project
- [OmniConfig MCP](../mcp/) - MCP server configuration
- [Claude Code Hooks](https://docs.anthropic.com/en/docs/claude-code/hooks)
- [Cursor Hooks](https://docs.cursor.com/advanced/hooks)

## License

MIT License - see [LICENSE](../LICENSE) for details.
