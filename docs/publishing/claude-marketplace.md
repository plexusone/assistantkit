# Claude Marketplace

The [Claude Code Official Marketplace](https://github.com/anthropics/claude-plugins-official) hosts community plugins for Claude Code.

## Repository Structure

```
anthropics/claude-plugins-official/
├── official_plugins/     # Anthropic-maintained plugins
└── external_plugins/     # Community plugins
    └── your-plugin/
        ├── .claude-plugin/
        │   └── plugin.json   # Consolidated config (MCP + hooks embedded)
        ├── commands/         # Optional: Slash commands
        ├── skills/           # Optional: Skills
        ├── agents/           # Optional: Agents
        └── README.md         # Required: Documentation
```

## Requirements

Your plugin must have:

- `.claude-plugin/plugin.json` - Plugin manifest (with embedded MCP and hooks)
- `README.md` - Documentation

## Generating with Bundle API

Use the assistantkit bundle package to generate properly formatted plugin files:

```go
package main

import (
    "log"

    "github.com/plexusone/assistantkit/bundle"
    "github.com/plexusone/assistantkit/hooks/core"
)

func main() {
    // Create bundle
    b := bundle.New("my-plugin", "1.0.0", "A helpful plugin")
    b.Plugin.Author = "Your Name"
    b.Plugin.License = "MIT"
    b.Plugin.Repository = "https://github.com/yourname/my-plugin"

    // Add MCP server
    b.AddMCPServer("my-server", bundle.MCPServer{
        Command: "./my-server",
        Env: map[string]string{
            "API_KEY": "${MY_API_KEY}",
        },
    })

    // Add hooks
    cfg := bundle.NewHooksConfig()
    cfg.AddHook(core.OnStop, core.Hook{
        Type:   "prompt",
        Prompt: "Task stopped. Consider follow-up actions.",
    })
    b.SetHooks(cfg)

    // Add skill
    skill := bundle.NewSkill("my-skill", "Skill description")
    skill.Instructions = "# My Skill\n\nInstructions here..."
    b.AddSkill(skill)

    // Add command
    cmd := bundle.NewCommand("my-cmd", "Command description")
    cmd.Instructions = "Command instructions..."
    b.AddCommand(cmd)

    // Generate for Claude Code
    if err := b.Generate("claude", "./output"); err != nil {
        log.Fatalf("Generate failed: %v", err)
    }
}
```

This generates:

```
output/
├── .claude-plugin/
│   └── plugin.json    # Consolidated with MCP + hooks embedded
├── commands/
│   └── my-cmd.md
└── skills/
    └── my-skill/
        └── SKILL.md
```

## Manual Submission

### 1. Fork the Repository

Fork [anthropics/claude-plugins-official](https://github.com/anthropics/claude-plugins-official)

### 2. Add Your Plugin

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/claude-plugins-official
cd claude-plugins-official

# Create branch
git checkout -b add-my-plugin

# Copy your plugin
cp -r /path/to/your/plugin external_plugins/my-plugin
```

### 3. Create Pull Request

Push and create a PR to `anthropics/claude-plugins-official`.

## Automated Submission

Use AI Assist Kit to automate the PR creation:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/plexusone/assistantkit/publish/claude"
    "github.com/plexusone/assistantkit/publish/core"
)

func main() {
    ctx := context.Background()
    token := os.Getenv("GITHUB_TOKEN")

    publisher := claude.NewPublisher(token)

    // Validate plugin first
    if err := publisher.Validate("./plugins/claude"); err != nil {
        log.Fatalf("Validation failed: %v", err)
    }

    // Submit to marketplace
    result, err := publisher.Publish(ctx, core.PublishOptions{
        PluginDir:  "./plugins/claude",
        PluginName: "my-plugin",
        Title:      "Add my-plugin to marketplace",
        Verbose:    true,
    })
    if err != nil {
        log.Fatalf("Publish failed: %v", err)
    }

    fmt.Printf("PR created: %s\n", result.PRURL)
}
```

## Publish Options

| Option | Description | Required |
|--------|-------------|----------|
| `PluginDir` | Path to plugin directory | Yes |
| `PluginName` | Plugin name in marketplace | Yes |
| `Title` | PR title | No |
| `Body` | PR description | No |
| `Branch` | Branch name | No |
| `ForkOwner` | GitHub username for fork | No |
| `DryRun` | Simulate without creating PR | No |
| `Verbose` | Print progress | No |

## Generated PR

The automated publisher creates a PR with:

```markdown
## Summary

Adding the **my-plugin** plugin to the Claude Code marketplace.

### Description

[Extracted from README.md]

### Checklist

- [ ] Plugin has `.claude-plugin/plugin.json`
- [ ] Plugin has `README.md` with documentation
- [ ] All commands/skills/agents are documented
- [ ] No security issues or sensitive data
- [ ] Tested locally with Claude Code

---

*Submitted via [aiassistkit](https://github.com/plexusone/assistantkit) publish tool*
```

## Validation

Before publishing, validate your plugin:

```go
publisher := claude.NewPublisher(token)
err := publisher.Validate("./plugins/claude")
if err != nil {
    // Handle validation error
    if validationErr, ok := err.(*core.ValidationError); ok {
        fmt.Printf("Missing files: %v\n", validationErr.Missing)
    }
}
```

## Required Files

| File | Description |
|------|-------------|
| `.claude-plugin/plugin.json` | Plugin metadata |
| `README.md` | Documentation |

## plugin.json Format

The consolidated format embeds MCP servers and hooks directly in `plugin.json`:

```json
{
  "name": "my-plugin",
  "version": "1.0.0",
  "description": "A helpful plugin for developers",
  "author": "Your Name",
  "repository": "https://github.com/yourname/my-plugin",
  "license": "MIT",
  "mcpServers": {
    "my-server": {
      "command": "./my-server",
      "args": ["--mode", "production"],
      "env": {
        "API_KEY": "${MY_API_KEY}"
      }
    }
  },
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "echo 'Running command...'"
          }
        ]
      }
    ],
    "Stop": [
      {
        "hooks": [
          {
            "type": "prompt",
            "prompt": "The task has stopped. Consider if follow-up is needed."
          }
        ]
      }
    ]
  },
  "commands": "./commands/",
  "skills": "./skills/",
  "agents": "./agents/"
}
```

### plugin.json Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Plugin identifier (lowercase, dashes) |
| `version` | string | Yes | Semantic version (e.g., `1.0.0`) |
| `description` | string | Yes | Brief description |
| `author` | string | No | Author name or organization |
| `license` | string | No | License identifier (e.g., `MIT`) |
| `repository` | string | No | GitHub repository URL |
| `homepage` | string | No | Project homepage URL |
| `mcpServers` | object | No | MCP server configurations |
| `hooks` | object | No | Event hooks (PreToolUse, PostToolUse, Stop, etc.) |
| `commands` | string | No | Path to commands directory |
| `skills` | string | No | Path to skills directory |
| `agents` | string | No | Path to agents directory |

### MCP Server Configuration

```json
{
  "mcpServers": {
    "server-name": {
      "command": "./path/to/binary",
      "args": ["--flag", "value"],
      "env": {
        "VAR_NAME": "${ENV_VAR}"
      },
      "cwd": "/optional/working/dir",
      "disabled": false
    }
  }
}
```

### Hooks Configuration

Hooks are triggered on specific events:

| Event | Description |
|-------|-------------|
| `PreToolUse` | Before tool execution (can use `matcher` to filter) |
| `PostToolUse` | After tool execution |
| `Stop` | When agent stops |
| `Notification` | When notifications occur |
| `SubagentStop` | When subagent stops |

Hook types:

- `command` - Execute a shell command
- `prompt` - Inject a prompt for the model to consider

## Dry Run

Test the publish process without creating a PR:

```go
result, err := publisher.Publish(ctx, core.PublishOptions{
    PluginDir:  "./plugins/claude",
    PluginName: "my-plugin",
    DryRun:     true,
    Verbose:    true,
})
// result.Status = "Dry run completed - no PR created"
```

## Error Handling

```go
result, err := publisher.Publish(ctx, opts)
if err != nil {
    switch e := err.(type) {
    case *core.ValidationError:
        fmt.Printf("Missing: %v\n", e.Missing)
    case *core.ForkError:
        fmt.Printf("Fork failed: %v\n", e.Err)
    case *core.BranchError:
        fmt.Printf("Branch failed: %v\n", e.Err)
    case *core.CommitError:
        fmt.Printf("Commit failed: %v\n", e.Err)
    case *core.PRError:
        fmt.Printf("PR failed: %v\n", e.Err)
    case *core.AuthError:
        fmt.Printf("Auth failed: %s\n", e.Message)
    default:
        fmt.Printf("Unknown error: %v\n", err)
    }
}
```

## GitHub Token

Create a token at [github.com/settings/tokens](https://github.com/settings/tokens) with:

- `repo` - Full control of private repositories
- `workflow` - Update GitHub Action workflows (if needed)

```bash
export GITHUB_TOKEN=ghp_xxxxxxxxxxxx
```

## After Submission

1. **Wait for Review** - Anthropic team reviews submissions
2. **Address Feedback** - Make requested changes
3. **Merge** - Once approved, your plugin is published
4. **Announce** - Let users know about your plugin!
