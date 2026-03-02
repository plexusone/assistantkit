---
marp: true
theme: default
paginate: true
backgroundColor: #fff
style: |
  section {
    font-family: 'Helvetica Neue', Arial, sans-serif;
  }
  h1 {
    color: #2563eb;
  }
  h2 {
    color: #1e40af;
  }
  code {
    background-color: #f1f5f9;
    border-radius: 4px;
    padding: 2px 6px;
  }
  pre {
    background-color: #1e293b;
    border-radius: 8px;
  }
  table {
    font-size: 0.85em;
  }
  .columns {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
  }
---

# OmniConfig Hooks

## Unified Hook Management for AI Coding Assistants

---

# What Are Hooks?

Hooks are **automation callbacks** that execute at defined stages of the AI agent loop.

They enable you to:

- **Observe** - Log activity, collect metrics
- **Block** - Security gates, validation checks
- **Modify** - Inject context, transform outputs

---

# The Problem

Each AI assistant has its own hook format:

| Tool | Config File | Key Differences |
|------|-------------|-----------------|
| Claude Code | `.claude/settings.json` | Uses `PreToolUse`/`PostToolUse` with matchers |
| Cursor IDE | `.cursor/hooks.json` | Uses `beforeShellExecution`, `afterFileEdit` |
| Windsurf | `.windsurf/hooks.json` | Uses `pre_run_command`, `post_write_code` |

**Managing hooks across tools is tedious and error-prone.**

---

# The Solution: OmniConfig Hooks

A **unified interface** for hook configuration:

```go
import "github.com/plexusone/assistantkit/hooks"

cfg := hooks.NewConfig()
cfg.AddHook(hooks.BeforeCommand,
    hooks.NewCommandHook("./security-check.sh"))

// Write to any format
claude.WriteProjectConfig(cfg)  // Claude
cursor.WriteProjectConfig(cfg)  // Cursor
windsurf.WriteProjectConfig(cfg) // Windsurf
```

---

# Canonical Event Types

## File Operations

| Event | Description | Can Block |
|-------|-------------|-----------|
| `before_file_read` | Before reading a file | Yes |
| `after_file_read` | After reading a file | No |
| `before_file_write` | Before writing a file | Yes |
| `after_file_write` | After writing a file | No |

---

# Canonical Event Types

## Command & MCP Operations

| Event | Description | Can Block |
|-------|-------------|-----------|
| `before_command` | Before shell execution | Yes |
| `after_command` | After shell execution | No |
| `before_mcp` | Before MCP tool call | Yes |
| `after_mcp` | After MCP tool call | No |

---

# Canonical Event Types

## Session Lifecycle

| Event | Description | Can Block |
|-------|-------------|-----------|
| `before_prompt` | Before prompt processing | Yes |
| `on_stop` | When agent stops | No |
| `on_session_start` | Session begins | No |
| `on_session_end` | Session ends | No |

---

# Tool Support Matrix

| Event | Claude | Cursor | Windsurf |
|-------|:------:|:------:|:--------:|
| `before_file_read` | ✅ | ✅ | ✅ |
| `after_file_read` | ✅ | ❌ | ✅ |
| `before_file_write` | ✅ | ❌ | ✅ |
| `after_file_write` | ✅ | ✅ | ✅ |
| `before_command` | ✅ | ✅ | ✅ |
| `before_mcp` | ✅ | ✅ | ✅ |
| `on_session_start` | ✅ | ❌ | ❌ |
| `after_response` | ❌ | ✅ | ❌ |
| `on_permission` | ✅ | ❌ | ❌ |

---

# Hook Types

## Command Hooks

Execute shell commands:

```go
hook := hooks.NewCommandHook("./my-script.sh")
hook = hook.WithTimeout(30)        // 30 second timeout
hook = hook.WithWorkingDir("/tmp") // Working directory
```

## Prompt Hooks (Claude-only)

Run AI validation prompts:

```go
hook := hooks.NewPromptHook("Is this file write safe?")
```

---

# Creating Hooks

```go
package main

import (
    "github.com/plexusone/assistantkit/hooks"
    "github.com/plexusone/assistantkit/hooks/claude"
)

func main() {
    cfg := hooks.NewConfig()

    // Hook with matcher (tool-specific)
    cfg.AddHookWithMatcher(hooks.BeforeCommand, "Bash",
        hooks.NewCommandHook("echo 'Running bash...'"))

    // Generic hook (all tools)
    cfg.AddHook(hooks.BeforeFileWrite,
        hooks.NewCommandHook("./validate.sh"))

    claude.WriteProjectConfig(cfg)
}
```

---

# Converting Between Formats

```go
package main

import (
    "os"
    "github.com/plexusone/assistantkit/hooks"
)

func main() {
    // Read Claude config
    data, _ := os.ReadFile(".claude/settings.json")

    // Convert to Cursor format
    cursorData, _ := hooks.Convert(data, "claude", "cursor")
    os.WriteFile(".cursor/hooks.json", cursorData, 0644)

    // Convert to Windsurf format
    windsurfData, _ := hooks.Convert(data, "claude", "windsurf")
    os.WriteFile(".windsurf/hooks.json", windsurfData, 0644)
}
```

---

# Format: Claude Code

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
    ]
  }
}
```

Uses `PreToolUse`/`PostToolUse` with **matchers** for tool filtering.

---

# Format: Cursor IDE

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

Uses **camelCase** event names like `beforeShellExecution`.

---

# Format: Windsurf

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

Uses **snake_case** event names with `show_output` option.

---

# Use Case: Security Gate

Block dangerous commands before execution:

```go
cfg.AddHookWithMatcher(hooks.BeforeCommand, "Bash",
    hooks.NewCommandHook("./security-check.sh"))
```

**security-check.sh:**
```bash
#!/bin/bash
if echo "$COMMAND" | grep -qE "rm -rf|sudo|curl.*\|.*sh"; then
    echo "BLOCKED: Dangerous command detected"
    exit 1
fi
```

---

# Use Case: Audit Logging

Log all file modifications for compliance:

```go
cfg.AddHook(hooks.AfterFileWrite,
    hooks.NewCommandHook("./audit-log.sh"))
```

**audit-log.sh:**
```bash
#!/bin/bash
echo "$(date -Iseconds) | WRITE | $FILE_PATH" >> /var/log/ai-audit.log
```

---

# Use Case: Code Quality

Run linting before commits:

```go
cfg.AddHook(hooks.BeforeCommand,
    hooks.NewCommandHook("./pre-commit-check.sh"))
```

**pre-commit-check.sh:**
```bash
#!/bin/bash
if [[ "$COMMAND" == *"git commit"* ]]; then
    golint ./... || exit 1
    go test ./... || exit 1
fi
```

---

# Architecture

```
hooks/
├── hooks.go          # Package entry point
├── core/
│   ├── event.go      # 20+ canonical events
│   ├── hook.go       # Hook types (command, prompt)
│   ├── config.go     # Configuration management
│   └── adapter.go    # Adapter interface & registry
├── claude/
│   └── adapter.go    # Claude Code adapter
├── cursor/
│   └── adapter.go    # Cursor IDE adapter
└── windsurf/
    └── adapter.go    # Windsurf adapter
```

---

# Key Features

- **Canonical Types** - One format to rule them all
- **Bidirectional Conversion** - Any format to any format
- **Tool Filtering** - Automatically filter unsupported events
- **Type Safety** - Compile-time validation
- **Extensible** - Easy to add new adapters
- **Well Tested** - 92.2% code coverage

---

# Test Coverage

| Package | Tests | Coverage |
|---------|-------|----------|
| `hooks` | 12 | 100.0% |
| `hooks/claude` | 32 | 89.8% |
| `hooks/core` | 44 | 98.7% |
| `hooks/cursor` | 26 | 87.1% |
| `hooks/windsurf` | 29 | 85.9% |
| **Total** | **143** | **92.2%** |

---

# API Overview

```go
// Create configuration
cfg := hooks.NewConfig()

// Add hooks
cfg.AddHook(event, hook)
cfg.AddHookWithMatcher(event, matcher, hook)

// Query hooks
cfg.GetHooks(event) []HookEntry
cfg.GetAllHooksForEvent(event) []Hook
cfg.Events() []Event
cfg.HookCount() int

// Convert
hooks.Convert(data, "claude", "cursor")
```

---

# Getting Started

## Install

```bash
go get github.com/plexusone/assistantkit/hooks
```

## Import

```go
import (
    "github.com/plexusone/assistantkit/hooks"
    "github.com/plexusone/assistantkit/hooks/claude"
    "github.com/plexusone/assistantkit/hooks/cursor"
    "github.com/plexusone/assistantkit/hooks/windsurf"
)
```

---

# Part of OmniConfig

**OmniConfig** - Unified AI assistant configuration management

| Module | Description | Status |
|--------|-------------|--------|
| `mcp/` | MCP server configurations | ✅ Available |
| `hooks/` | Automation/lifecycle hooks | ✅ Available |
| `settings/` | Permissions, sandbox settings | Coming soon |
| `rules/` | Team rules, guidelines | Coming soon |
| `memory/` | CLAUDE.md, .cursorrules | Coming soon |

---

# Resources

- **GitHub**: [github.com/plexusone/assistantkit](https://github.com/plexusone/assistantkit)
- **Go Docs**: [pkg.go.dev/github.com/plexusone/assistantkit/hooks](https://pkg.go.dev/github.com/plexusone/assistantkit/hooks)
- **Claude Code Hooks**: [docs.anthropic.com](https://docs.anthropic.com/en/docs/claude-code/hooks)
- **Cursor Hooks**: [docs.cursor.com](https://docs.cursor.com/advanced/hooks)

---

# Thank You

## Questions?

<br>

**OmniConfig Hooks** - Write once, deploy everywhere.

```bash
go get github.com/plexusone/assistantkit/hooks
```
