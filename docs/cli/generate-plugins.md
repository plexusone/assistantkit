# Generate Command

The `generate` command creates platform-specific plugins from a unified specs directory.

!!! note "v0.10.0 Update"
    As of v0.10.0, the `generate` command automatically validates specs before generation. Use `--skip-validate` to bypass validation if needed.

## Synopsis

```bash
assistantkit generate [flags]
```

## Description

This command reads plugin definitions from a unified specs directory and generates complete platform-specific plugins for each deployment target. Each target receives agents, commands, skills, and plugin manifest.

### Automatic Validation

Before generation, the command validates the specs directory to catch errors early:

```bash
assistantkit generate --specs=specs

=== AssistantKit Generator ===
Specs directory: /path/to/specs
Target: local
Output directory: .

Validating specs...
✓ Validation passed (6 agents, 4 skills, 2 commands)

Team: my-team
Loaded: 2 commands, 4 skills, 6 agents

Generated targets:
  - local-kiro: plugins/kiro
  - local-claude: .claude/agents

Done!
```

Validation checks include:

- **plugin.json** — required fields present, valid JSON
- **agents** — valid YAML frontmatter with required fields
- **skills** — valid skill definitions
- **commands** — valid command definitions
- **skill-refs** — agent skill references resolve to existing skills
- **teams** — DAG acyclicity, agent references, input `from` references
- **deployments** — valid platforms, no output path conflicts

Use `--skip-validate` to bypass validation if needed.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--specs` | `specs` | Path to unified specs directory |
| `--target` | `local` | Deployment target (looks for `specs/deployments/<target>.json`) |
| `--output` | `.` | Output base directory for relative paths |
| `--skip-validate` | `false` | Skip validation before generation |

## Supported Platforms

- **claude-code**: Claude Code plugins (`.claude-plugin/`, commands/, skills/, agents/)
- **kiro-cli**: Kiro IDE Powers (POWER.md + mcp.json) or Kiro Agents (agents/*.json)
- **gemini-cli**: Gemini CLI extensions (gemini-extension.json, commands/, agents/)

## Specs Directory Structure

The unified specs directory should contain:

```
specs/
├── plugin.json          # Plugin metadata
├── agents/              # Agent definitions (*.md with YAML frontmatter)
│   ├── coordinator.md
│   └── writer.md
├── commands/            # Command definitions (*.md or *.json)
│   └── release.md
├── skills/              # Skill definitions (*.md or *.json)
│   └── review.md
├── teams/               # Team workflow definitions (optional)
│   └── my-team.json
└── deployments/         # Deployment configurations
    ├── local.json       # Local development (default)
    └── production.json  # Production deployment
```

### plugin.json

The plugin metadata file defines the plugin name, version, keywords, and MCP server configurations:

```json
{
  "name": "my-plugin",
  "displayName": "My Plugin",
  "version": "1.0.0",
  "description": "A plugin for AI assistants",
  "keywords": ["keyword1", "keyword2"],
  "mcpServers": {
    "my-server": {
      "command": "my-mcp-server",
      "args": []
    }
  }
}
```

### agents/*.md

Agent definitions using multi-agent-spec format with YAML frontmatter:

```markdown
---
name: release-coordinator
description: Orchestrates software releases
model: sonnet
tools: [Read, Write, Bash, Glob, Grep]
skills: [version-analysis, commit-classification]
---

You are a release coordinator agent responsible for...
```

### commands/*.md

Command definitions for slash commands:

```markdown
---
name: release
description: Execute full release workflow
arguments: [version]
dependencies: [version-analysis]
---

# Release Command

When executing a release, follow these steps...
```

### skills/*.md

Skill definitions for reusable capabilities:

```markdown
---
name: code-review
description: Reviews code for best practices
triggers: [review code, check code]
---

# Code Review Skill

When reviewing code, analyze for...
```

### deployments/*.json

Deployment configurations defining output targets:

```json
{
  "team": "my-team",
  "targets": [
    {
      "name": "local-claude",
      "platform": "claude-code",
      "output": "plugins/claude"
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

#### Kiro Prefix Configuration

For Kiro CLI targets, you can specify a prefix in the `kiroCli` field. This prefix is applied to:

- Agent JSON filenames: `myteam_pm.json`
- Agent name field: `"name": "myteam_pm"`
- Steering filenames: `myteam_code-review.md`
- README examples: `kiro-cli chat --agent myteam_pm`

This enables multiple teams to share a single Kiro agents directory without naming conflicts. Define the prefix once in the deployment config—agent source files use short names (e.g., `pm`), and the prefix is applied during generation.

## Generated Output

Each deployment target receives a complete plugin:

### Claude Code (`claude-code`)

```
plugins/claude/
├── .claude-plugin/
│   └── plugin.json       # Claude plugin manifest
├── commands/
│   └── release.md        # Command instructions
├── skills/
│   └── code-review/
│       └── SKILL.md      # Skill instructions
└── agents/
    └── release-coordinator.md  # Agent definition
```

### Kiro CLI (`kiro-cli`)

```
plugins/kiro/
├── POWER.md              # Power description (or agents/*.json)
├── mcp.json              # MCP server configuration
└── steering/
    └── code-review.md    # Steering files from skills
```

### Gemini CLI (`gemini-cli`)

```
plugins/gemini/
├── gemini-extension.json # Extension manifest
├── commands/
│   └── release.toml      # Command in TOML format
└── agents/
    └── release-coordinator.toml  # Agent in TOML format
```

## Examples

Generate plugins using defaults:

```bash
assistantkit generate
```

Use a specific deployment target:

```bash
assistantkit generate --target=production
```

Generate with custom directories:

```bash
assistantkit generate --specs=my-specs --target=local --output=/path/to/output
```

## Deprecated Subcommands

The following subcommands are deprecated and will show warnings when used:

| Deprecated | Replacement |
|------------|-------------|
| `generate plugins` | `generate --specs=... --target=...` |
| `generate agents` | `generate --specs=... --target=...` |
| `generate all` | `generate --specs=... --target=...` |
| `generate deployment` | `generate --specs=... --target=...` |

## See Also

- [Validate Command](validate.md) - Standalone validation command
- [Plugin Structure](../plugins/structure.md) - Learn about plugin components
- [Commands](../plugins/commands.md) - Command definition details
- [Skills](../plugins/skills.md) - Skill definition details
- [Agents](../plugins/agents.md) - Agent definition details
