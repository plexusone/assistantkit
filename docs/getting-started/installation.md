# Installation

## Requirements

- Go 1.21 or later
- One or more AI coding assistants installed:
    - [Claude Code](https://docs.anthropic.com/en/docs/claude-code)
    - [Gemini CLI](https://github.com/google-gemini/gemini-cli)
    - [OpenAI Codex CLI](https://github.com/openai/codex)
    - [AWS Kiro](https://kiro.dev/)

## Install the Library

```bash
go get github.com/plexusone/assistantkit
```

## Install the CLI Tool (Optional)

To use the CLI tool for generating plugins from canonical specs:

```bash
go install github.com/plexusone/assistantkit/cmd/assistantkit@latest
```

See [Generate Plugins](../cli/generate-plugins.md) for CLI usage.

## Verify Installation

```go
package main

import (
    "fmt"
    "github.com/plexusone/assistantkit/plugins/core"
)

func main() {
    plugin := core.Plugin{
        Name:        "test-plugin",
        Version:     "1.0.0",
        Description: "Test plugin",
    }
    fmt.Printf("Plugin: %s v%s\n", plugin.Name, plugin.Version)
}
```

## Optional Dependencies

For marketplace publishing:

```bash
go get github.com/grokify/gogithub
```

## Project Structure

A typical project using aiassistkit:

```
my-plugin/
├── go.mod
├── go.sum
├── main.go                 # Plugin generator
├── canonical/
│   ├── plugin.yaml         # Canonical plugin definition
│   ├── commands/
│   │   └── greet.md
│   └── skills/
│       └── review.md
└── plugins/
    ├── claude/             # Generated Claude plugin
    ├── gemini/             # Generated Gemini plugin
    └── codex/              # Generated Codex plugin
```
