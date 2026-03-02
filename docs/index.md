# AI Assist Kit

A Go library for building plugins for AI coding assistants.

## Overview

AI Assist Kit provides a unified way to create plugins that work across multiple AI coding assistants:

- **Claude Code** - Anthropic's CLI assistant
- **Gemini CLI** - Google's CLI assistant
- **OpenAI Codex CLI** - OpenAI's CLI assistant
- **AWS Kiro** - Amazon's AI coding assistant

## Features

- **Canonical Plugin Format** - Define plugins once, generate for multiple assistants
- **Commands** - Custom slash commands for your workflows
- **Skills** - Reusable prompt templates and capabilities
- **Agents** - Specialized sub-agents for complex tasks
- **MCP Server Configuration** - Model Context Protocol server management
- **Marketplace Publishing** - Automated submission to official marketplaces

## Quick Example

Define a canonical command:

```go
package main

import (
    "github.com/plexusone/assistantkit/commands/core"
)

func main() {
    cmd := core.Command{
        Name:        "greet",
        Description: "Greet the user",
        Prompt:      "Say hello to the user in a friendly way.",
    }
    // Generate for Claude, Gemini, etc.
}
```

## Supported Assistants

| Assistant | Commands | Skills | Agents | MCP |
|-----------|----------|--------|--------|-----|
| Claude Code | Yes | Yes | Yes | Yes |
| Gemini CLI | Yes | Yes | No | Yes |
| OpenAI Codex | Yes | No | No | Yes |
| AWS Kiro | Yes | Yes | Yes | Yes |

## Getting Started

1. [Install the library](getting-started/installation.md)
2. [Follow the quick start guide](getting-started/quickstart.md)
3. [Learn about plugin structure](plugins/structure.md)

## License

MIT License
