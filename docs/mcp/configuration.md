# MCP Server Configuration

The [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) allows AI assistants to connect to external tools and data sources.

## Overview

MCP servers provide:

- **Tools**: Functions the AI can call
- **Resources**: Data sources the AI can access
- **Prompts**: Pre-defined prompt templates

## Configuration Format

All supported assistants use similar JSON configuration:

```json
{
  "mcpServers": {
    "server-name": {
      "command": "executable",
      "args": ["arg1", "arg2"],
      "env": {
        "API_KEY": "your-key"
      }
    }
  }
}
```

## Assistant-Specific Locations

| Assistant | Config File |
|-----------|-------------|
| Claude Code | `~/.claude.json` |
| Gemini CLI | `~/.gemini/settings.json` |
| OpenAI Codex | `~/.codex/config.json` |
| AWS Kiro | `~/.kiro/mcp.json` |

## Canonical Format

AI Assist Kit provides a canonical MCP configuration format:

```go
type MCPServerConfig struct {
    Command string            `json:"command"`
    Args    []string          `json:"args,omitempty"`
    Env     map[string]string `json:"env,omitempty"`
}

type MCPConfig struct {
    Servers map[string]MCPServerConfig `json:"mcpServers"`
}
```

## Common MCP Servers

### Filesystem Server

Access local files:

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@anthropic-ai/mcp-server-filesystem", "/path/to/allowed/dir"]
    }
  }
}
```

### GitHub Server

Access GitHub repositories:

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@anthropic-ai/mcp-server-github"],
      "env": {
        "GITHUB_TOKEN": "your-token"
      }
    }
  }
}
```

### PostgreSQL Server

Access PostgreSQL databases:

```json
{
  "mcpServers": {
    "postgres": {
      "command": "npx",
      "args": ["-y", "@anthropic-ai/mcp-server-postgres"],
      "env": {
        "DATABASE_URL": "postgresql://user:pass@localhost/db"
      }
    }
  }
}
```

### Slack Server

Access Slack workspaces:

```json
{
  "mcpServers": {
    "slack": {
      "command": "npx",
      "args": ["-y", "@anthropic-ai/mcp-server-slack"],
      "env": {
        "SLACK_TOKEN": "xoxb-your-token"
      }
    }
  }
}
```

## Converting Between Assistants

```go
import (
    "github.com/plexusone/assistantkit/mcp/core"
    "github.com/plexusone/assistantkit/mcp/claude"
    "github.com/plexusone/assistantkit/mcp/kiro"
)

// Define canonical config
config := core.MCPConfig{
    Servers: map[string]core.MCPServerConfig{
        "filesystem": {
            Command: "npx",
            Args:    []string{"-y", "@anthropic-ai/mcp-server-filesystem"},
        },
    },
}

// Convert to Claude format
claudeConfig := claude.FromCanonical(config)

// Convert to Kiro format
kiroConfig := kiro.FromCanonical(config)
```

## Per-Agent MCP Configuration

Some assistants support per-agent MCP configuration:

### AWS Kiro

```json
{
  "name": "database-agent",
  "description": "Agent with database access",
  "mcpServers": {
    "postgres": {
      "command": "npx",
      "args": ["-y", "@anthropic-ai/mcp-server-postgres"]
    }
  },
  "includeMcpJson": true
}
```

The `includeMcpJson: true` option merges the global `~/.kiro/mcp.json` configuration.

## Security Considerations

!!! warning "Sensitive Data"
    Never commit MCP configurations with secrets to version control.
    Use environment variables or secret management tools.

### Using Environment Variables

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@anthropic-ai/mcp-server-github"],
      "env": {
        "GITHUB_TOKEN": "${GITHUB_TOKEN}"
      }
    }
  }
}
```

### Restricting Access

Limit filesystem access to specific directories:

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": [
        "-y",
        "@anthropic-ai/mcp-server-filesystem",
        "/home/user/projects",
        "--read-only"
      ]
    }
  }
}
```

## Troubleshooting

### Server Not Starting

1. Check the command exists: `which npx`
2. Verify package is installed: `npx -y @anthropic-ai/mcp-server-filesystem --help`
3. Check logs for errors

### Permission Denied

1. Verify the directory/resource is accessible
2. Check environment variables are set correctly
3. Ensure tokens have required scopes

### Connection Timeout

1. Check network connectivity
2. Verify firewall rules
3. Increase timeout settings if available
