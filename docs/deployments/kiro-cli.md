# Kiro CLI Deployment

This guide walks through deploying agents to [Kiro CLI](https://kiro.dev/docs/cli/) using AssistantKit.

## Overview

Kiro CLI uses a flat agent directory structure where all agents live in `~/.kiro/agents/`. Unlike Claude Code which supports namespaced directories, Kiro requires unique agent names across all installed agents.

AssistantKit addresses this with **single-source prefix configuration**: define the prefix once in your deployment config, and it's applied to all generated agent names, filenames, and documentation.

## Project vs User-Level Agents

Kiro CLI and Claude Code handle agent storage differently:

| Aspect | Claude Code | Kiro CLI |
|--------|-------------|----------|
| **Project agents** | `.claude/agents/*.md` ✅ | Not supported ❌ |
| **User agents** | `~/.claude/agents/*.md` | `~/.kiro/agents/*.json` |
| **Project context** | `CLAUDE.md` | `.kiro/steering/*.md` |
| **Auto-discovery** | Yes (scans project) | No (user dir only) |
| **Namespacing** | Directory-based | Prefix-based |

### Why Kiro Works This Way

Kiro CLI reads agents **only** from the user-level `~/.kiro/agents/` directory. It does not scan project directories for agent definitions. This design means:

1. **All agents share one namespace** — Without prefixes, an agent named `coordinator` in Project A would conflict with `coordinator` in Project B
2. **Installation is always required** — You can't just define agents in your project and use them; they must be copied to `~/.kiro/`
3. **Prefixes prevent conflicts** — Using `rel_coordinator` vs `rev_coordinator` lets multiple teams coexist

### Project-Level Context (Steering Files)

While agent *definitions* must live in `~/.kiro/agents/`, you can still use **project-level context** via steering files:

```
my-project/
└── .kiro/
    └── steering/
        ├── project-guidelines.md
        └── coding-standards.md
```

Reference them in your agent's `resources` field:

```json
{
  "name": "rel_coordinator",
  "resources": [
    "file://.kiro/steering/**/*.md",
    "file://README.md"
  ]
}
```

This gives agents project-specific context while the agent definitions remain in the global directory.

!!! note "Claude Code Alternative"
    If you prefer project-scoped agents without installation, consider using Claude Code instead. Claude Code automatically discovers agents in `.claude/agents/` within your project, with no copy step required.

## Installation

Kiro CLI reads agents from `~/.kiro/agents/` and steering files from `~/.kiro/steering/`. Generated files must be installed to these directories before Kiro can use them.

### Option A: Generate then Copy (Recommended for Distribution)

Generate to a project directory, then copy to Kiro's config:

```bash
# Generate to plugins/kiro/
assistantkit generate --specs=specs --target=local

# Install to Kiro
cp plugins/kiro/agents/*.json ~/.kiro/agents/
cp plugins/kiro/steering/*.md ~/.kiro/steering/
```

This approach is best when:

- You want to review generated files before installing
- You're distributing agents to others (commit `plugins/kiro/` to git)
- You have multiple projects and want to manage installations manually

### Option B: Direct Install (Recommended for Personal Use)

Set `output` to `~/.kiro` in your deployment config to install directly:

```json
{
  "team": "release-team",
  "targets": [
    {
      "name": "local-kiro",
      "platform": "kiro-cli",
      "output": "~/.kiro",
      "kiroCli": {
        "prefix": "rel_"
      }
    }
  ]
}
```

Then generate:

```bash
assistantkit generate --specs=specs --target=local
```

Files are written directly to `~/.kiro/agents/` and `~/.kiro/steering/`. No copy step needed.

!!! tip "Output Path Must Be Plugin Root"
    Set `output` to `~/.kiro` (the plugin root), NOT `~/.kiro/agents`. The generator creates the `agents/` and `steering/` subdirectories automatically.

## Generated Output

For a Kiro CLI target, AssistantKit generates:

```
plugins/kiro/
├── agents/
│   ├── myteam_coordinator.json    # Agent definitions
│   ├── myteam_reviewer.json
│   └── myteam_writer.json
├── steering/
│   ├── myteam_code-review.md      # Skills as steering files
│   └── myteam_version-analysis.md
└── README.md                       # Usage documentation
```

## Step-by-Step Setup

### 1. Create Specs Directory

```bash
mkdir -p specs/{agents,skills,deployments}
```

### 2. Define Plugin Metadata

Create `specs/plugin.json`:

```json
{
  "name": "release-team",
  "displayName": "Release Team",
  "version": "1.0.0",
  "description": "Multi-agent team for release automation"
}
```

### 3. Create Agents

Create agents using **short, canonical names** (no prefix):

`specs/agents/coordinator.md`:

```markdown
---
name: coordinator
description: Orchestrates release workflows by delegating to specialized agents
model: sonnet
tools:
  - Read
  - Write
  - Bash
  - Glob
  - Grep
skills:
  - version-analysis
---

You are a release coordinator agent.

## Responsibilities

1. Analyze the release scope
2. Delegate tasks to specialized agents
3. Aggregate results and create release artifacts

## Available Sub-Agents

- **reviewer**: Code review and quality checks
- **writer**: Documentation and changelog generation

## Workflow

When asked to prepare a release:

1. Use version-analysis skill to determine version bump
2. Delegate code review to reviewer agent
3. Delegate changelog to writer agent
4. Create git tag and release notes
```

`specs/agents/reviewer.md`:

```markdown
---
name: reviewer
description: Reviews code changes for quality and security
model: sonnet
tools:
  - Read
  - Grep
  - Glob
skills:
  - code-review
---

You are a code reviewer agent specializing in quality and security analysis.

## Review Checklist

- [ ] No hardcoded secrets or credentials
- [ ] Proper error handling
- [ ] Test coverage for new code
- [ ] No breaking API changes without version bump
```

`specs/agents/writer.md`:

```markdown
---
name: writer
description: Generates documentation, changelogs, and release notes
model: sonnet
tools:
  - Read
  - Write
  - Glob
---

You are a technical writer agent.

## Responsibilities

1. Generate changelog entries from commits
2. Write release notes
3. Update documentation
```

### 4. Create Skills

Skills become Kiro steering files:

`specs/skills/version-analysis.md`:

```markdown
---
name: version-analysis
description: Analyzes commits to determine semantic version bumps
---

# Version Analysis

Analyze git commits since the last release tag to determine the appropriate semantic version bump.

## Conventional Commits Mapping

| Commit Type | Version Bump |
|-------------|--------------|
| `feat:` | Minor |
| `fix:` | Patch |
| `docs:` | Patch |
| `BREAKING CHANGE` | Major |

## Process

1. Find the latest release tag: `git describe --tags --abbrev=0`
2. List commits since tag: `git log <tag>..HEAD --oneline`
3. Parse commit messages for conventional commit types
4. Determine highest priority bump (major > minor > patch)
```

`specs/skills/code-review.md`:

```markdown
---
name: code-review
description: Reviews code for best practices and security
---

# Code Review

Review code changes for quality, security, and maintainability.

## Security Checks

- No hardcoded secrets (API keys, passwords, tokens)
- No SQL injection vulnerabilities
- Proper input validation
- Secure authentication handling

## Quality Checks

- Proper error handling (no silent failures)
- Adequate test coverage
- Clear naming conventions
- No dead code or unused imports
```

### 5. Configure Deployment

Create `specs/deployments/local.json`:

```json
{
  "team": "release-team",
  "targets": [
    {
      "name": "local-kiro",
      "platform": "kiro-cli",
      "output": "plugins/kiro",
      "kiroCli": {
        "prefix": "rel_"
      }
    }
  ]
}
```

### 6. Validate and Generate

```bash
# Validate specs
assistantkit validate --specs=specs

# Generate Kiro agents
assistantkit generate --specs=specs --target=local
```

### 7. Install Agents

Install generated files to Kiro's config directory:

```bash
# Create directories if needed
mkdir -p ~/.kiro/agents ~/.kiro/steering

# Copy agents
cp plugins/kiro/agents/*.json ~/.kiro/agents/

# Copy steering files (if any)
cp plugins/kiro/steering/*.md ~/.kiro/steering/
```

!!! tip "Direct Installation"
    To skip the copy step, set `"output": "~/.kiro"` in your deployment config. See [Installation](#installation) for details.

### 8. Use Agents

```bash
# Start with coordinator
kiro-cli chat --agent rel_coordinator

# Or use a specific agent
kiro-cli chat --agent rel_reviewer
```

## Prefix Configuration

The `kiroCli.prefix` field is applied to:

| Item | Without Prefix | With `rel_` Prefix |
|------|----------------|-------------------|
| Agent filename | `coordinator.json` | `rel_coordinator.json` |
| Agent name field | `"name": "coordinator"` | `"name": "rel_coordinator"` |
| Steering filename | `version-analysis.md` | `rel_version-analysis.md` |
| README examples | `--agent coordinator` | `--agent rel_coordinator` |

### Why Prefixes Matter

Kiro CLI stores all agents in a single directory (`~/.kiro/agents/`). Without prefixes, agents from different teams would conflict:

```
~/.kiro/agents/
├── coordinator.json    # Which team's coordinator?
├── reviewer.json       # Conflict!
└── writer.json
```

With prefixes:

```
~/.kiro/agents/
├── rel_coordinator.json    # Release team
├── rel_reviewer.json
├── rev_coordinator.json    # Review team
├── rev_reviewer.json
└── prd_coordinator.json    # Product team
```

## Multi-Team Example

Deploy multiple teams to the same Kiro installation:

`specs/deployments/local.json`:

```json
{
  "team": "release-team",
  "targets": [
    {
      "name": "local-kiro",
      "platform": "kiro-cli",
      "output": "~/.kiro",
      "kiroCli": {
        "prefix": "rel_"
      }
    }
  ]
}
```

In another repository for your review team:

```json
{
  "team": "review-team",
  "targets": [
    {
      "name": "local-kiro",
      "platform": "kiro-cli",
      "output": "~/.kiro",
      "kiroCli": {
        "prefix": "rev_"
      }
    }
  ]
}
```

## Tool Mapping

AssistantKit maps canonical tool names to Kiro equivalents:

| Canonical | Kiro |
|-----------|------|
| Read | read |
| Write | write |
| Edit | write |
| Bash | shell |
| Glob | glob |
| Grep | grep |
| WebSearch | web_search |
| WebFetch | web_fetch |

## Model Mapping

| Canonical | Kiro |
|-----------|------|
| haiku | claude-haiku |
| sonnet | claude-sonnet-4 |
| opus | claude-opus-4 |

## Steering Files

Skills are converted to Kiro steering files and placed in the `steering/` directory. Reference them in your agent's `resources` field:

```json
{
  "name": "rel_coordinator",
  "resources": [
    "file://.kiro/steering/rel_*.md"
  ]
}
```

!!! note "Manual Resource Configuration"
    Currently, AssistantKit generates steering files but doesn't automatically add them to agent `resources`. You may need to manually add resource references after generation, or use a glob pattern in your agent config.

## Sub-Agent Spawning

Kiro CLI supports spawning sub-agents within a chat session:

```
> Use the rel_reviewer agent to check the authentication module
> Use the rel_writer agent to generate the changelog
```

The coordinator agent's instructions should reference other agents by their prefixed names when describing delegation patterns.

## Generated README

AssistantKit generates a README with:

- Agent descriptions and purposes
- Usage examples with correct prefixed names
- Installation instructions
- Coordinator detection (if an agent has "coordinator" in the name)

Example generated content:

```markdown
# Release Team Agents

## Quick Start

# Use the coordinator for orchestrated workflows
kiro-cli chat --agent rel_coordinator

## Agents

| Agent | Description |
|-------|-------------|
| rel_coordinator | Orchestrates release workflows |
| rel_reviewer | Reviews code changes |
| rel_writer | Generates documentation |
```

## Troubleshooting

### Agent Not Found

If `kiro-cli chat --agent myagent` fails:

1. Verify the agent JSON exists in `~/.kiro/agents/`
2. Check the filename matches the agent name
3. Ensure JSON is valid: `jq . ~/.kiro/agents/myagent.json`

### Steering Files Not Loaded

If steering file content isn't available to the agent:

1. Verify files exist in `~/.kiro/steering/`
2. Check agent's `resources` array includes the correct path
3. Use glob patterns for multiple files: `file://.kiro/steering/rel_*.md`

### Prefix Mismatch

If generated agents have wrong prefixes:

1. Check `kiroCli.prefix` in deployment config
2. Ensure agent source files use short names (no prefix)
3. Re-run generation: `assistantkit generate --specs=specs`

## See Also

- [Deployment Overview](overview.md) — Common deployment concepts
- [Kiro CLI Reference](../assistants/kiro.md) — Kiro CLI features and configuration
- [Generate Command](../cli/generate-plugins.md) — CLI reference
