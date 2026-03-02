# Quick Start

This guide walks you through creating a simple plugin that works with multiple AI assistants.

## Step 1: Create Your Project

```bash
mkdir my-plugin
cd my-plugin
go mod init github.com/yourname/my-plugin
go get github.com/plexusone/assistantkit
```

## Step 2: Define a Canonical Command

Create `canonical/commands/greet.md`:

```markdown
---
name: greet
description: Greet the user with a friendly message
---

Say hello to the user in a friendly and welcoming way.
Include their name if provided.
```

## Step 3: Create the Plugin Generator

Create `main.go`:

```go
package main

import (
    "log"
    "os"

    "github.com/plexusone/assistantkit/commands/core"
    "github.com/plexusone/assistantkit/commands/claude"
    "github.com/plexusone/assistantkit/commands/gemini"
    "github.com/plexusone/assistantkit/plugins/core"
)

func main() {
    // Define canonical command
    greetCmd := commandscore.Command{
        Name:        "greet",
        Description: "Greet the user with a friendly message",
        Prompt:      "Say hello to the user in a friendly and welcoming way.",
    }

    // Generate for Claude
    claudeCmd := claude.FromCanonical(greetCmd)
    if err := os.MkdirAll("plugins/claude/commands", 0755); err != nil {
        log.Fatal(err)
    }
    if err := claudeCmd.WriteFile("plugins/claude/commands/greet.md"); err != nil {
        log.Fatal(err)
    }

    // Generate for Gemini
    geminiCmd := gemini.FromCanonical(greetCmd)
    if err := os.MkdirAll("plugins/gemini/commands", 0755); err != nil {
        log.Fatal(err)
    }
    if err := geminiCmd.WriteFile("plugins/gemini/commands/greet.yaml"); err != nil {
        log.Fatal(err)
    }

    log.Println("Plugins generated successfully!")
}
```

## Step 4: Generate Plugins

```bash
go run main.go
```

This creates:

```
plugins/
├── claude/
│   └── commands/
│       └── greet.md
└── gemini/
    └── commands/
        └── greet.yaml
```

## Step 5: Install the Plugin

=== "Claude Code"

    ```bash
    # From local directory
    claude plugin add ./plugins/claude

    # Or from GitHub
    claude plugin add github:yourname/my-plugin/plugins/claude
    ```

=== "Gemini CLI"

    ```bash
    # Copy to Gemini config
    cp -r plugins/gemini ~/.gemini/plugins/my-plugin
    ```

## Step 6: Test Your Plugin

=== "Claude Code"

    ```bash
    claude
    > /greet
    ```

=== "Gemini CLI"

    ```bash
    gemini
    > /greet
    ```

## Next Steps

- [Learn about plugin structure](../plugins/structure.md)
- [Add skills to your plugin](../plugins/skills.md)
- [Create specialized agents](../plugins/agents.md)
- [Publish to marketplaces](../publishing/overview.md)
