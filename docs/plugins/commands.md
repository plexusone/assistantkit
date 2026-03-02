# Commands

Commands are slash commands that users invoke in their AI assistant.

## Canonical Format

```go
type Command struct {
    Name         string   `yaml:"name"`
    Description  string   `yaml:"description"`
    Prompt       string   `yaml:"prompt"`
    AllowedTools []string `yaml:"allowed_tools,omitempty"`
    Model        string   `yaml:"model,omitempty"`
}
```

## Creating a Command

### Using Go

```go
package main

import (
    "github.com/plexusone/assistantkit/commands/core"
)

func main() {
    cmd := core.Command{
        Name:        "build",
        Description: "Build the project",
        Prompt: `Build the project using the appropriate build system.
Detect the project type and run the correct build command:
- Go: go build ./...
- Node.js: npm run build
- Python: python setup.py build`,
        AllowedTools: []string{"Bash", "Read"},
    }
}
```

### Using Markdown

Create `commands/build.md`:

```markdown
---
name: build
description: Build the project
allowed_tools:
  - Bash
  - Read
---

Build the project using the appropriate build system.
Detect the project type and run the correct build command:

- Go: `go build ./...`
- Node.js: `npm run build`
- Python: `python setup.py build`
```

## Command Options

| Field | Description | Required |
|-------|-------------|----------|
| `name` | Command name (used as `/name`) | Yes |
| `description` | Short description shown in help | Yes |
| `prompt` | Instructions for the AI | Yes |
| `allowed_tools` | Tools the command can use | No |
| `model` | Preferred model (sonnet, opus, haiku) | No |

## Tool Restrictions

You can limit which tools a command can use:

```yaml
allowed_tools:
  - Read      # Read files
  - Write     # Write files
  - Edit      # Edit files
  - Bash      # Run shell commands
  - Glob      # Find files by pattern
  - Grep      # Search file contents
```

!!! warning "Security"
    Commands with `Bash` access can execute arbitrary shell commands.
    Consider limiting tools for commands that process untrusted input.

## Assistant-Specific Output

### Claude Code

```markdown
---
name: build
description: Build the project
allowed_tools:
  - Bash
  - Read
---

Build the project using the appropriate build system...
```

### Gemini CLI

```yaml
name: build
description: Build the project
prompt: |
  Build the project using the appropriate build system...
```

### OpenAI Codex

```markdown
---
name: build
description: Build the project
---

Build the project using the appropriate build system...
```

## Examples

### Test Command

```markdown
---
name: test
description: Run project tests
allowed_tools:
  - Bash
  - Read
---

Run the project's test suite:

1. Detect the project type
2. Run appropriate test command
3. Report results clearly

For Go: `go test -v ./...`
For Node.js: `npm test`
For Python: `pytest -v`
```

### Deploy Command

```markdown
---
name: deploy
description: Deploy to production
allowed_tools:
  - Bash
  - Read
model: opus
---

Deploy the application to production:

1. Verify all tests pass
2. Check for uncommitted changes
3. Run the deployment script
4. Verify deployment succeeded

IMPORTANT: Always confirm with user before deploying.
```
