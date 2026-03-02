# AWS Kiro

[AWS Kiro](https://kiro.dev/) is Amazon's AI coding assistant with agentic capabilities available as both an IDE extension and CLI.

## Kiro CLI

Kiro CLI is the command-line interface that provides:

- **Custom Agents**: Define specialized agents for workflows
- **Sub-Agents**: Spawn parallel agents for complex tasks
- **MCP Integration**: Connect external tools via Model Context Protocol
- **Steering Files**: Guide agent behavior with context documents
- **Smart Hooks**: Automate workflows with lifecycle triggers

### Installation

```bash
# Install via npm
npm install -g @anthropic-ai/kiro-cli

# Or via Homebrew
brew install kiro-cli
```

### Starting Kiro CLI

```bash
# Default agent
kiro-cli

# With specific agent
kiro-cli --agent release-coordinator
```

## Custom Agents

Agents are JSON files stored at `~/.kiro/agents/{agent-name}.json`.

### Agent Configuration

```json
{
  "name": "my-agent",
  "description": "A custom agent for my workflow",
  "tools": ["read", "write", "shell"],
  "allowedTools": ["read", "shell"],
  "resources": ["file://README.md", "file://.kiro/steering/**/*.md"],
  "prompt": "You are a helpful coding assistant",
  "model": "claude-sonnet-4"
}
```

### Configuration Fields

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Agent identifier (required) |
| `description` | string | Agent purpose (required) |
| `tools` | array | Available tool capabilities |
| `allowedTools` | array | Tools that run without prompting |
| `resources` | array | Context files (glob patterns supported) |
| `prompt` | string | System instructions for the agent |
| `model` | string | Claude model to use |
| `mcpServers` | object | MCP server configurations |
| `includeMcpJson` | boolean | Inherit global MCP config |

### Creating Agents

**Interactive CLI:**

```bash
kiro-cli agent create --name my-agent
```

**Within chat session:**

```
> /agent generate
```

### Agent Management

```bash
# Swap agents mid-session
> /agent swap release-qa

# List available agents
> /agent list
```

## Sub-Agents

Kiro CLI supports spawning sub-agents for complex tasks with parallel execution.

### How Sub-Agents Work

1. **Task Assignment**: Describe a task and Kiro determines if sub-agents are appropriate
2. **Initialization**: Sub-agent created with context from agent configuration
3. **Autonomous Execution**: Independent task completion
4. **Progress Updates**: Live notifications of work status
5. **Result Return**: Completed results returned to main agent

### Spawning Sub-Agents

```
> Use the security-scanner agent to audit the authentication module
> Use the qa-agent to verify test coverage
```

Sub-agents can run in parallel for faster completion.

### Sub-Agent Tool Access

| Available | Not Available |
|-----------|---------------|
| read | web_search |
| write | web_fetch |
| shell | introspect |
| MCP tools | thinking |
| | todo_list |

## Steering Files

Steering files provide context and guidelines for agents.

### Structure

```
.kiro/
└── steering/
    ├── project-guidelines.md
    ├── coding-standards.md
    └── release-process.md
```

### Referencing in Agents

```json
{
  "name": "my-agent",
  "resources": [
    "file://.kiro/steering/**/*.md",
    "file://README.md"
  ]
}
```

!!! note "Resource Prefix"
    Resources require the `file://` prefix and support glob patterns.

## MCP Server Configuration

### Global Config

`~/.kiro/mcp.json`:

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

### Per-Agent Config

```json
{
  "name": "my-agent",
  "mcpServers": {
    "custom-server": {
      "command": "./my-server",
      "args": ["--config", "config.json"]
    }
  },
  "includeMcpJson": true
}
```

## Tool Reference

| Tool | Description |
|------|-------------|
| `read` | Read file contents |
| `write` | Write/create files |
| `shell` | Execute shell commands |
| `glob` | Find files by pattern |
| `grep` | Search file contents |
| `web_search` | Search the web |
| `web_fetch` | Fetch web content |

## Model Options

| Model ID | Description |
|----------|-------------|
| `claude-sonnet-4` | Balanced performance (default) |
| `claude-opus-4` | Most capable |
| `claude-haiku` | Fast, lightweight |

## Converting from Canonical

aiassistkit provides adapters for converting between canonical and Kiro formats:

```go
import (
    "github.com/plexusone/assistantkit/agents/core"
    "github.com/plexusone/assistantkit/agents/kiro"
)

// Create canonical agent
canonical := &core.Agent{
    Name:         "scanner",
    Description:  "Security scanner",
    Instructions: "You are a security expert...",
    Model:        "sonnet",
    Tools:        []string{"Read", "Grep", "Bash"},
}

// Convert to Kiro format
adapter := &kiro.Adapter{}
kiroAgent := adapter.FromCore(canonical)
// Result:
// {
//   "name": "scanner",
//   "model": "claude-sonnet-4",
//   "tools": ["read", "grep", "shell"]
// }

// Write to user agents directory
kiro.WriteUserAgent(canonical)
```

## Tool Mapping

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
| sonnet | claude-sonnet-4 |
| opus | claude-opus-4 |
| haiku | claude-haiku |

## Kiro Powers vs Kiro CLI

Kiro has two extension mechanisms:

| Feature | Kiro Powers | Kiro CLI Agents |
|---------|-------------|-----------------|
| Format | MCP server bundle | JSON config |
| Sub-agents | No | Yes |
| Parallel execution | No | Yes |
| Marketplace | kiro.dev/powers | Manual install |
| Use case | IDE extensions | CLI workflows |

For multi-agent workflows like release automation, Kiro CLI agents are recommended over Powers.

## Example: Release Agents

A set of release-focused agents:

```bash
# Install release agents
cp release-coordinator.json ~/.kiro/agents/
cp release-qa.json ~/.kiro/agents/
cp release-security.json ~/.kiro/agents/

# Use coordinator
kiro-cli --agent release-coordinator

# Spawn sub-agents for review
> Use the release-qa agent to verify tests
> Use the release-security agent to audit dependencies
```

## Sources

- [Kiro CLI Documentation](https://kiro.dev/docs/cli/)
- [Custom Agents Guide](https://kiro.dev/docs/cli/custom-agents/)
- [Sub-Agents Reference](https://kiro.dev/docs/cli/chat/subagents/)
