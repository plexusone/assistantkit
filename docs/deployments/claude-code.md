# Claude Code Deployment

This guide walks through deploying plugins to [Claude Code](https://docs.anthropic.com/en/docs/claude-code) using AssistantKit.

## Overview

Claude Code plugins use a namespaced directory structure where each plugin has its own folder containing commands, skills, and agents. Unlike Kiro CLI's flat structure, Claude Code naturally isolates plugins, so prefixes aren't needed.

Claude Code plugins support:

- **Commands** — Slash commands (`/release`, `/check`)
- **Skills** — Reusable capabilities agents can reference
- **Agents** — Sub-agents spawned via the Task tool
- **Hooks** — Lifecycle event handlers
- **MCP Servers** — External tool integrations

## Generated Output

For a Claude Code target, AssistantKit generates:

```
plugins/claude/
├── .claude-plugin/
│   └── plugin.json           # Plugin manifest
├── CLAUDE.md                  # Context file (optional)
├── commands/
│   └── release.md            # Slash commands
├── skills/
│   └── version-analysis/
│       └── SKILL.md          # Skills in subdirectories
└── agents/
    ├── coordinator.md        # Sub-agent definitions
    ├── reviewer.md
    └── writer.md
```

## Step-by-Step Setup

### 1. Create Specs Directory

```bash
mkdir -p specs/{agents,skills,commands,deployments}
```

### 2. Define Plugin Metadata

Create `specs/plugin.json`:

```json
{
  "name": "release-automation",
  "displayName": "Release Automation",
  "version": "1.0.0",
  "description": "Multi-agent release automation for software projects",
  "keywords": ["release", "automation", "changelog"],
  "author": {
    "name": "Your Name",
    "url": "https://github.com/yourname"
  },
  "repository": "https://github.com/yourname/release-automation",
  "license": "MIT"
}
```

### 3. Create Agents

Agents are Markdown files with YAML frontmatter:

`specs/agents/coordinator.md`:

```markdown
---
name: release-coordinator
description: Orchestrates release workflows by delegating to specialized agents
model: sonnet
tools:
  - Read
  - Write
  - Bash
  - Glob
  - Grep
  - Task
skills:
  - version-analysis
  - commit-classification
---

You are a release coordinator responsible for orchestrating software releases.

## Your Role

Coordinate the release process by:

1. Analyzing the release scope using version-analysis skill
2. Delegating code review to the qa agent
3. Delegating documentation to the docs agent
4. Aggregating results and creating release artifacts

## Spawning Sub-Agents

Use the Task tool to delegate work:

```
Task(subagent_type="qa", prompt="Review changes since v1.0.0")
Task(subagent_type="docs", prompt="Generate changelog for v1.1.0")
```

## Parallel Execution

For independent tasks, spawn multiple agents in parallel:

```
> Run qa and security agents in parallel to review this release
```
```

`specs/agents/qa.md`:

```markdown
---
name: qa
description: Quality assurance agent for code review and test validation
model: sonnet
tools:
  - Read
  - Grep
  - Glob
  - Bash
skills:
  - code-review
---

You are a QA agent specializing in code quality and test coverage.

## Responsibilities

1. Review code changes for quality issues
2. Verify test coverage for new functionality
3. Check for security vulnerabilities
4. Validate error handling

## Review Process

1. Identify changed files using git diff
2. Read and analyze each changed file
3. Check for corresponding test files
4. Report findings with severity levels
```

`specs/agents/docs.md`:

```markdown
---
name: docs
description: Documentation agent for changelogs and release notes
model: sonnet
tools:
  - Read
  - Write
  - Glob
  - Grep
skills:
  - commit-classification
---

You are a documentation agent specializing in release documentation.

## Responsibilities

1. Generate changelog entries from commits
2. Write release notes
3. Update API documentation
4. Maintain README accuracy

## Changelog Format

Use conventional changelog format:

```markdown
## [1.1.0] - 2024-01-15

### Added
- New feature description

### Fixed
- Bug fix description
```
```

### 4. Create Skills

Skills become reusable instructions that agents can reference:

`specs/skills/version-analysis.md`:

```markdown
---
name: version-analysis
description: Analyzes git history to determine semantic version bumps
triggers:
  - analyze version
  - determine version bump
  - what version should this be
---

# Version Analysis

Analyze commits since the last release to determine the appropriate semantic version bump.

## Conventional Commits Mapping

| Commit Type | Version Bump | Example |
|-------------|--------------|---------|
| `feat:` | Minor | feat: add user authentication |
| `fix:` | Patch | fix: resolve login timeout |
| `docs:` | Patch | docs: update API reference |
| `perf:` | Patch | perf: optimize query performance |
| `BREAKING CHANGE` | Major | feat!: redesign API endpoints |

## Analysis Steps

1. Find latest release tag:
   ```bash
   git describe --tags --abbrev=0
   ```

2. List commits since tag:
   ```bash
   git log <tag>..HEAD --oneline
   ```

3. Parse each commit for conventional commit type
4. Return highest priority bump needed
```

`specs/skills/commit-classification.md`:

```markdown
---
name: commit-classification
description: Classifies commits for changelog generation
triggers:
  - classify commits
  - categorize changes
---

# Commit Classification

Classify commits according to conventional commits specification.

## Categories

| Category | Commit Types | Changelog Section |
|----------|--------------|-------------------|
| Features | feat | Added |
| Bug Fixes | fix | Fixed |
| Performance | perf | Performance |
| Documentation | docs | Documentation |
| Breaking | feat!, fix!, BREAKING CHANGE | Breaking Changes |

## Output Format

Return classified commits as structured data:

```json
{
  "breaking": ["commit message"],
  "features": ["commit message"],
  "fixes": ["commit message"],
  "other": ["commit message"]
}
```
```

`specs/skills/code-review.md`:

```markdown
---
name: code-review
description: Reviews code for quality, security, and best practices
triggers:
  - review code
  - check code quality
---

# Code Review

Review code changes for quality, security, and maintainability.

## Security Checklist

- [ ] No hardcoded secrets (API keys, passwords, tokens)
- [ ] No SQL injection vulnerabilities
- [ ] Proper input validation and sanitization
- [ ] Secure authentication and authorization
- [ ] No sensitive data in logs

## Quality Checklist

- [ ] Proper error handling (no silent failures)
- [ ] Adequate test coverage
- [ ] Clear naming conventions
- [ ] No dead code or unused imports
- [ ] Consistent code style

## Review Output

Provide findings with severity:

- **Critical**: Security vulnerabilities, data loss risks
- **High**: Bugs, missing error handling
- **Medium**: Code quality issues, missing tests
- **Low**: Style issues, minor improvements
```

### 5. Create Commands

Commands define slash commands available to users:

`specs/commands/release.md`:

```markdown
---
name: release
description: Execute full release workflow for $ARGUMENTS
allowed_tools:
  - Read
  - Write
  - Bash
  - Glob
  - Grep
  - Task
---

# Release Command

Execute the complete release workflow for the specified version.

## Usage

```
/release v1.2.0
/release patch
/release minor
```

## Workflow

1. **Validate** — Check version format and git status
2. **Analyze** — Determine changes since last release
3. **Review** — Spawn QA agent to review changes
4. **Document** — Spawn docs agent for changelog
5. **Tag** — Create and push git tag
6. **Publish** — Create GitHub release

## Pre-Release Checks

Before releasing, verify:

- Working tree is clean
- All tests pass
- No uncommitted changes
- Branch is up to date with remote
```

`specs/commands/check.md`:

```markdown
---
name: check
description: Run pre-release validation checks
allowed_tools:
  - Read
  - Bash
  - Glob
  - Grep
---

# Check Command

Run validation checks before releasing.

## Usage

```
/check
```

## Checks Performed

1. **Git Status** — Working tree clean
2. **Tests** — All tests passing
3. **Lint** — No linting errors
4. **Dependencies** — No security vulnerabilities
5. **Documentation** — README and docs up to date
```

### 6. Configure Deployment

Create `specs/deployments/local.json`:

```json
{
  "team": "release-automation",
  "targets": [
    {
      "name": "local-claude",
      "platform": "claude-code",
      "output": "plugins/claude"
    }
  ]
}
```

For deploying to a project directory:

```json
{
  "team": "release-automation",
  "targets": [
    {
      "name": "project-claude",
      "platform": "claude-code",
      "output": "."
    }
  ]
}
```

This generates directly into the current project:

```
./
├── .claude-plugin/
│   └── plugin.json
├── commands/
├── skills/
└── agents/
```

### 7. Validate and Generate

```bash
# Validate specs
assistantkit validate --specs=specs

# Generate Claude Code plugin
assistantkit generate --specs=specs --target=local
```

### 8. Install Plugin

```bash
# Add plugin from local path
claude plugin add plugins/claude

# Or if generated in project root
claude plugin add .
```

### 9. Use Plugin

```bash
# Start Claude Code in your project
claude

# Use commands
> /release v1.2.0
> /check

# Sub-agents are spawned automatically by coordinator
```

## Plugin Structure Details

### .claude-plugin/plugin.json

The generated manifest includes paths to all components:

```json
{
  "name": "release-automation",
  "version": "1.0.0",
  "description": "Multi-agent release automation",
  "commands": "./commands/",
  "skills": "./skills/",
  "agents": "./agents/"
}
```

### Commands Directory

Each command becomes a Markdown file:

```
commands/
├── release.md
└── check.md
```

### Skills Directory

Skills use subdirectory structure with SKILL.md:

```
skills/
├── version-analysis/
│   └── SKILL.md
├── commit-classification/
│   └── SKILL.md
└── code-review/
    └── SKILL.md
```

### Agents Directory

Agents are Markdown files:

```
agents/
├── release-coordinator.md
├── qa.md
└── docs.md
```

## Sub-Agent System

Claude Code's Task tool spawns sub-agents defined in your plugin:

```
Task(subagent_type="qa", prompt="Review the authentication changes")
```

### Built-in Sub-Agent Types

| Type | Description |
|------|-------------|
| `Bash` | Command execution specialist |
| `general-purpose` | Multi-step task handling |
| `Explore` | Fast codebase exploration |
| `Plan` | Implementation planning |

### Custom Sub-Agents

Your plugin's agents are available by name:

```
Task(subagent_type="qa", prompt="...")
Task(subagent_type="docs", prompt="...")
Task(subagent_type="release-coordinator", prompt="...")
```

### Parallel Execution

Spawn multiple agents simultaneously:

```
> Run qa and security in parallel to review changes
```

The coordinator agent receives results from both when complete.

## Tool Reference

Claude Code tools available to agents:

| Tool | Description |
|------|-------------|
| `Read` | Read file contents |
| `Write` | Create or overwrite files |
| `Edit` | Edit files in place |
| `Bash` | Execute shell commands |
| `Glob` | Find files by pattern |
| `Grep` | Search file contents |
| `Task` | Spawn sub-agents |
| `WebFetch` | Fetch web content |
| `WebSearch` | Search the web |

## Model Options

| Model | Description | Use Case |
|-------|-------------|----------|
| `haiku` | Fast, lightweight | Quick tasks, simple queries |
| `sonnet` | Balanced performance | Most development tasks |
| `opus` | Most capable | Complex reasoning, architecture |

## CLAUDE.md Context File

Add a `CLAUDE.md` to your plugin for persistent context:

`specs/context/CLAUDE.md`:

```markdown
# Release Automation Plugin

This plugin provides multi-agent release automation.

## Available Commands

- `/release <version>` — Execute full release workflow
- `/check` — Run pre-release validation

## Agents

- **release-coordinator** — Orchestrates the release process
- **qa** — Code review and test validation
- **docs** — Changelog and documentation

## Dependencies

- `git` — Version control
- `gh` — GitHub CLI (for releases)
```

## Multi-Target Deployment

Deploy to both Claude Code and Kiro CLI:

```json
{
  "team": "release-automation",
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
        "prefix": "rel_"
      }
    }
  ]
}
```

The same specs generate platform-native output for each target.

## Hooks Configuration

Add lifecycle hooks to your plugin:

`specs/hooks/hooks.json`:

```json
{
  "pre_tool_use": [
    {
      "matcher": "Bash",
      "command": "scripts/validate-command.sh",
      "description": "Validate shell commands before execution"
    }
  ],
  "post_tool_use": [
    {
      "matcher": "Write",
      "command": "scripts/log-file-changes.sh",
      "description": "Log file modifications"
    }
  ]
}
```

## Publishing to Marketplace

Submit your plugin to the Claude Plugins marketplace:

1. Ensure plugin meets quality guidelines
2. Add comprehensive documentation
3. Submit PR to `anthropics/claude-plugins-official`

See [Claude Marketplace](../publishing/claude-marketplace.md) for detailed instructions.

## Troubleshooting

### Plugin Not Found

If `claude plugin add` fails:

1. Verify `.claude-plugin/plugin.json` exists
2. Check JSON is valid: `jq . .claude-plugin/plugin.json`
3. Ensure all referenced directories exist

### Command Not Available

If `/mycommand` isn't recognized:

1. Check command file exists in `commands/`
2. Verify frontmatter has `name` field
3. Reload plugin: `claude plugin update plugin-name`

### Agent Not Spawning

If `Task(subagent_type="myagent")` fails:

1. Verify agent file exists in `agents/`
2. Check agent's `name` in frontmatter matches
3. Ensure agent file has valid YAML frontmatter

### Skill Not Found

If an agent references a missing skill:

1. Check skill directory exists: `skills/skill-name/SKILL.md`
2. Verify skill's `name` in frontmatter
3. Run validation: `assistantkit validate --specs=specs`

## See Also

- [Deployment Overview](overview.md) — Common deployment concepts
- [Claude Code Reference](../assistants/claude.md) — Claude Code features
- [Generate Command](../cli/generate-plugins.md) — CLI reference
- [Claude Marketplace](../publishing/claude-marketplace.md) — Publishing guide
