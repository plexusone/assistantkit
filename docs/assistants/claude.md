# Claude Code

[Claude Code](https://docs.anthropic.com/en/docs/claude-code) is Anthropic's official CLI for Claude with plugin and sub-agent support.

## Plugin Structure

Claude Code plugins use a specific structure with `.claude-plugin/plugin.json` as the manifest:

```
my-plugin/
├── .claude-plugin/
│   └── plugin.json          # Required: Plugin manifest
├── CLAUDE.md                # Optional: Context for the model
├── commands/                # Optional: Slash commands (Markdown)
│   └── build.md
├── skills/                  # Optional: Reusable skills (Markdown)
│   └── review/
│       └── SKILL.md
├── agents/                  # Optional: Sub-agent definitions (Markdown)
│   └── scanner.md
└── hooks/                   # Optional: Lifecycle hooks
    └── hooks.json
```

## Plugin Manifest

The `.claude-plugin/plugin.json` file defines your plugin:

```json
{
  "name": "my-plugin",
  "version": "1.0.0",
  "description": "A helpful plugin for developers",
  "author": {
    "name": "Your Name",
    "url": "https://github.com/yourname"
  },
  "homepage": "https://github.com/yourname/my-plugin",
  "repository": "https://github.com/yourname/my-plugin",
  "license": "MIT",
  "keywords": ["automation", "development"],
  "commands": "./commands/",
  "skills": "./skills/",
  "agents": "./agents/",
  "hooks": "./hooks/hooks.json"
}
```

### Manifest Fields

| Field | Description |
|-------|-------------|
| `name` | Unique plugin name |
| `version` | Semantic version |
| `description` | Plugin description |
| `author` | Author information |
| `commands` | Path to commands directory |
| `skills` | Path to skills directory |
| `agents` | Path to agents directory |
| `hooks` | Path to hooks configuration |

## Commands

Commands are **Markdown files** with YAML frontmatter:

```markdown
---
name: build
description: Build the project
allowed_tools:
  - Bash
  - Read
---

Build the project using the appropriate build system.
Detect the project type and run the correct command.
```

### Command with Arguments

```markdown
---
name: release
description: Execute release workflow for $ARGUMENTS
---

# Release

Execute the complete release workflow for the specified version.

## Usage

```
/release v1.2.3
```
```

## Skills

Skills are reusable capabilities in `skills/skill-name/SKILL.md`:

```markdown
---
name: code-review
description: Review code for best practices
---

# Code Review Skill

Review the provided code for:

- Security vulnerabilities
- Performance issues
- Code style and maintainability
- Error handling
```

## Agents

Agents are sub-agent definitions that can be spawned via the Task tool:

```markdown
---
name: security-scanner
description: Scan code for vulnerabilities. Use when reviewing security.
model: sonnet
tools: Read, Grep, Glob, Bash
skills: code-review
---

# Security Scanner Agent

You are a security expert specializing in code review.

## Your Responsibilities

1. Check for hardcoded secrets
2. Review authentication implementations
3. Identify injection vulnerabilities
4. Assess data validation
```

### Agent Fields

| Field | Description |
|-------|-------------|
| `name` | Agent identifier |
| `description` | When to use this agent |
| `model` | haiku, sonnet, or opus |
| `tools` | Comma-separated tool list |
| `skills` | Comma-separated skill dependencies |

## Sub-Agent System

Claude Code's Task tool enables spawning specialized sub-agents:

```
Task(subagent_type="Plan", prompt="Design the implementation")
Task(subagent_type="Explore", prompt="Find all API endpoints")
```

### Built-in Sub-Agent Types

| Type | Description |
|------|-------------|
| `Bash` | Command execution specialist |
| `general-purpose` | Multi-step task handling |
| `Explore` | Fast codebase exploration |
| `Plan` | Implementation planning |

### Custom Sub-Agents

Define custom agents in `agents/` directory and reference by name:

```
Task(subagent_type="security-scanner", prompt="Audit the auth module")
```

### Parallel Execution

Multiple sub-agents can run in parallel:

```
> Run security-scanner and qa-agent in parallel to review this code
```

## Installation Methods

### From Local Path

```bash
claude plugin add /path/to/my-plugin
```

### From GitHub

```bash
claude plugin add github:owner/repo/path/to/plugin
```

### Plugin Management

```bash
# List installed plugins
claude plugin list

# Remove a plugin
claude plugin remove plugin-name

# Update a plugin
claude plugin update plugin-name
```

## Context File (CLAUDE.md)

The `CLAUDE.md` file provides persistent context:

```markdown
# My Plugin

This plugin automates release workflows.

## Available Commands

- `/release <version>` - Execute release
- `/check` - Run validation

## Dependencies

- `git` - Version control
- `gh` - GitHub CLI
```

## MCP Server Configuration

### Global Config

`~/.claude.json`:

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

### Per-Project Config

`.claude/settings.json`:

```json
{
  "mcpServers": {
    "custom-server": {
      "command": "./server.js",
      "args": ["--config", "config.json"]
    }
  }
}
```

## Hooks

Plugins can intercept lifecycle events:

```json
{
  "pre_prompt_submit": [
    {
      "command": "scripts/validate.sh",
      "description": "Validate before submission"
    }
  ],
  "post_tool_execution": [
    {
      "command": "scripts/log-tool.sh",
      "description": "Log tool usage"
    }
  ]
}
```

## Available Tools

| Tool | Description |
|------|-------------|
| `Read` | Read file contents |
| `Write` | Write/create files |
| `Edit` | Edit files in place |
| `Bash` | Execute shell commands |
| `Glob` | Find files by pattern |
| `Grep` | Search file contents |
| `Task` | Spawn sub-agents |
| `WebFetch` | Fetch web content |
| `WebSearch` | Search the web |
| `TodoWrite` | Manage task lists |

## Models

| Model | Description | Use Case |
|-------|-------------|----------|
| `haiku` | Fast, lightweight | Quick tasks, simple queries |
| `sonnet` | Balanced performance | Most development tasks |
| `opus` | Most capable | Complex reasoning, architecture |

## Converting from Canonical

aiassistkit provides adapters for format conversion:

```go
import (
    "github.com/plexusone/assistantkit/commands/core"
    "github.com/plexusone/assistantkit/commands/claude"
)

canonical := &core.Command{
    Name:        "build",
    Description: "Build the project",
    Prompt:      "Build the project using the appropriate build system.",
}

// Convert to Claude Markdown format
adapter := &claude.Adapter{}
mdBytes, err := adapter.Marshal(canonical)
```

## Comparison with Other Platforms

| Feature | Claude Code | Gemini CLI | Kiro CLI |
|---------|-------------|------------|----------|
| Manifest | JSON | JSON | JSON |
| Commands | Markdown | TOML | N/A |
| Skills | Markdown | N/A | N/A |
| Agents | Markdown | N/A | JSON |
| Sub-agents | Yes (Task tool) | No | Yes |
| Context | CLAUDE.md | GEMINI.md | Steering files |
| MCP Support | Yes | Yes | Yes |
| Marketplace | PR submission | Auto-indexed | Manual |

## Plugin Marketplace

Claude plugins can be submitted to the official marketplace via PR to `anthropics/claude-plugins-official`.

See [Claude Marketplace](../publishing/claude-marketplace.md) for detailed submission instructions.

## Sources

- [Claude Code Documentation](https://docs.anthropic.com/en/docs/claude-code)
- [Plugin Development Guide](https://docs.anthropic.com/en/docs/claude-code/plugins)
- [MCP Server Configuration](https://docs.anthropic.com/en/docs/claude-code/mcp)
