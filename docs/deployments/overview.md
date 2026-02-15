# Deployment Guide

AssistantKit generates platform-specific plugins from a unified specs directory. This guide covers how to structure your specs and configure deployments for multiple AI assistant platforms.

## Specs Directory Structure

The specs directory contains all your plugin definitions in a platform-agnostic format:

```
specs/
├── plugin.json          # Plugin metadata (required)
├── agents/              # Agent definitions (*.md with YAML frontmatter)
│   ├── coordinator.md
│   ├── reviewer.md
│   └── writer.md
├── commands/            # Slash commands (*.md)
│   └── release.md
├── skills/              # Reusable skills (*.md)
│   ├── code-review.md
│   └── version-analysis.md
├── teams/               # Multi-agent workflows (*.json)
│   └── release-team.json
└── deployments/         # Deployment targets (*.json)
    ├── local.json       # Local development
    └── production.json  # Production deployment
```

## plugin.json

The plugin manifest defines your plugin's identity and MCP server configurations:

```json
{
  "name": "my-plugin",
  "displayName": "My Plugin",
  "version": "1.0.0",
  "description": "A plugin for AI coding assistants",
  "keywords": ["automation", "release"],
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@anthropic-ai/mcp-server-github"]
    }
  }
}
```

### Required Fields

| Field | Description |
|-------|-------------|
| `name` | Unique plugin identifier (lowercase, hyphens) |
| `version` | Semantic version (e.g., "1.0.0") |
| `description` | Brief plugin description |

### Optional Fields

| Field | Description |
|-------|-------------|
| `displayName` | Human-readable name |
| `keywords` | Discovery keywords |
| `mcpServers` | MCP server configurations |

## Agent Definitions

Agents are defined as Markdown files with YAML frontmatter in the `agents/` directory:

```markdown
---
name: release-coordinator
description: Orchestrates software releases
model: sonnet
tools:
  - Read
  - Write
  - Bash
  - Glob
  - Grep
skills:
  - version-analysis
  - commit-classification
---

You are a release coordinator responsible for orchestrating software releases.

## Responsibilities

1. Analyze commits since the last release
2. Determine the appropriate version bump
3. Generate changelog entries
4. Create release tags and notes
```

### Agent Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Agent identifier (use short names, no prefix) |
| `description` | Yes | When to use this agent |
| `model` | No | haiku, sonnet, or opus (default: sonnet) |
| `tools` | No | Available tools |
| `skills` | No | Skill dependencies |

!!! tip "Short Names"
    Use short, canonical names for agents (e.g., `coordinator`, not `myteam_coordinator`). Platform-specific prefixes are applied during generation via deployment configuration.

## Skill Definitions

Skills are reusable capabilities that agents can reference:

```markdown
---
name: version-analysis
description: Analyzes git history to determine semantic version bumps
triggers:
  - analyze version
  - determine version bump
---

# Version Analysis

Analyze commits since the last release to determine the appropriate semantic version bump.

## Rules

- feat: commits → minor bump
- fix: commits → patch bump
- BREAKING CHANGE → major bump
```

## Command Definitions

Commands define slash commands available to users:

```markdown
---
name: release
description: Execute full release workflow
arguments:
  - version
dependencies:
  - version-analysis
---

# Release Command

Execute the complete release workflow for the specified version.

## Usage

```
/release v1.2.3
```

## Steps

1. Validate the version format
2. Run pre-release checks
3. Generate changelog
4. Create git tag
5. Push to remote
```

## Deployment Configuration

Deployment files in `deployments/` define where and how plugins are generated:

```json
{
  "team": "my-team",
  "targets": [
    {
      "name": "local-claude",
      "platform": "claude-code",
      "output": ".claude/plugins/my-team"
    },
    {
      "name": "local-kiro",
      "platform": "kiro-cli",
      "output": "plugins/kiro",
      "kiroCli": {
        "prefix": "myteam_"
      }
    },
    {
      "name": "local-gemini",
      "platform": "gemini-cli",
      "output": "plugins/gemini"
    }
  ]
}
```

### Target Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Target identifier |
| `platform` | Yes | Target platform (see below) |
| `output` | Yes | Output directory (relative to `--output` flag) |

### Supported Platforms

| Platform | Description |
|----------|-------------|
| `claude-code` | Claude Code plugins with commands, skills, agents |
| `kiro-cli` | Kiro CLI agents with steering files |
| `gemini-cli` | Gemini CLI extensions in TOML format |

## Validation

AssistantKit validates your specs before generation:

```bash
# Standalone validation
assistantkit validate --specs=specs

# Validation runs automatically with generate
assistantkit generate --specs=specs
```

### Validation Checks

| Check | Description |
|-------|-------------|
| plugin.json | Required fields, valid JSON |
| agents | Valid frontmatter, required fields |
| skills | Valid skill definitions |
| commands | Valid command definitions |
| skill-refs | Agent skill references resolve |
| teams | DAG acyclicity, agent references |
| deployments | Valid platforms, no path conflicts |

## Generation

Generate plugins for all targets in a deployment:

```bash
# Using local deployment (default)
assistantkit generate --specs=specs

# Using a specific deployment target
assistantkit generate --specs=specs --target=production

# Skip validation if needed
assistantkit generate --specs=specs --skip-validate
```

### Output Structure

Each platform receives a complete plugin in its native format. See platform-specific guides for details:

- [Kiro CLI Deployment](kiro-cli.md)
- [Claude Code Deployment](claude-code.md) (coming soon)
- [Gemini CLI Deployment](gemini-cli.md) (coming soon)

## Multi-Team Deployments

When multiple teams share an output directory (common with Kiro CLI), use prefixes to avoid naming conflicts:

```json
{
  "targets": [
    {
      "name": "release-team",
      "platform": "kiro-cli",
      "output": "~/.kiro/agents",
      "kiroCli": { "prefix": "rel_" }
    },
    {
      "name": "review-team",
      "platform": "kiro-cli",
      "output": "~/.kiro/agents",
      "kiroCli": { "prefix": "rev_" }
    }
  ]
}
```

This generates:

- `rel_coordinator.json`, `rel_reviewer.json`
- `rev_coordinator.json`, `rev_reviewer.json`

## Next Steps

- [Kiro CLI Deployment](kiro-cli.md) — Detailed Kiro deployment walkthrough
- [Validate Command](../cli/validate.md) — Validation reference
- [Generate Command](../cli/generate-plugins.md) — Generation reference
