# AI Assist Kit - Technical Requirements Document

## Architecture Overview

AI Assist Kit follows the adapter pattern with canonical types at the core and tool-specific adapters at the edges.

```
┌─────────────────────────────────────────────────────────────────┐
│                         aiassistkit                              │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐            │
│  │ plugins │  │commands │  │ skills  │  │ agents  │  NEW       │
│  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘            │
│       │            │            │            │                  │
│  ┌────┴────────────┴────────────┴────────────┴────┐            │
│  │                    core/                        │            │
│  │  Plugin, Command, Skill, Agent canonical types  │            │
│  └────┬────────────┬────────────┬────────────┬────┘            │
│       │            │            │            │                  │
│  ┌────┴────┐  ┌────┴────┐  ┌────┴────┐  ┌────┴────┐            │
│  │ claude/ │  │ gemini/ │  │ codex/  │  │ cursor/ │  Adapters  │
│  └─────────┘  └─────────┘  └─────────┘  └─────────┘            │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────┐  ┌─────────┐  ┌─────────┐                         │
│  │   mcp   │  │  hooks  │  │ context │  EXISTING               │
│  └─────────┘  └─────────┘  └─────────┘                         │
└─────────────────────────────────────────────────────────────────┘
```

## Package Structure

```
aiassistkit/
├── aiassistkit.go              # Umbrella package, version info
├── go.mod
│
├── plugins/                    # NEW: Plugin/extension manifests
│   ├── plugins.go              # Package umbrella with re-exports
│   ├── core/
│   │   ├── plugin.go           # Plugin struct
│   │   ├── adapter.go          # Adapter interface & registry
│   │   └── errors.go           # Error types
│   ├── claude/
│   │   ├── adapter.go          # .claude-plugin/plugin.json
│   │   └── config.go           # Claude-specific types
│   ├── gemini/
│   │   ├── adapter.go          # gemini-extension.json
│   │   └── config.go           # Gemini-specific types
│   └── schema/
│       └── plugin.schema.json  # JSON Schema
│
├── commands/                   # NEW: Commands/prompts
│   ├── commands.go             # Package umbrella
│   ├── core/
│   │   ├── command.go          # Command struct
│   │   ├── adapter.go          # Adapter interface & registry
│   │   └── errors.go           # Error types
│   ├── claude/
│   │   └── adapter.go          # commands/*.md (Markdown + YAML)
│   ├── gemini/
│   │   └── adapter.go          # commands/*.toml (TOML)
│   ├── codex/
│   │   └── adapter.go          # prompts/*.md (Markdown + YAML)
│   └── schema/
│       └── command.schema.json # JSON Schema
│
├── skills/                     # NEW: Skill definitions
│   ├── skills.go               # Package umbrella
│   ├── core/
│   │   ├── skill.go            # Skill struct
│   │   ├── adapter.go          # Adapter interface & registry
│   │   └── errors.go           # Error types
│   ├── claude/
│   │   └── adapter.go          # skills/*/SKILL.md
│   ├── codex/
│   │   └── adapter.go          # skills/*/SKILL.md
│   └── schema/
│       └── skill.schema.json   # JSON Schema
│
├── agents/                     # NEW: Agent definitions
│   ├── agents.go               # Package umbrella
│   ├── core/
│   │   ├── agent.go            # Agent struct
│   │   ├── adapter.go          # Adapter interface & registry
│   │   └── errors.go           # Error types
│   ├── claude/
│   │   └── adapter.go          # agents/*.md
│   └── schema/
│       └── agent.schema.json   # JSON Schema
│
├── mcp/                        # EXISTING: MCP configurations
├── hooks/                      # EXISTING: Lifecycle hooks
├── context/                    # EXISTING: Project context
│
└── cmd/
    └── aiassistkit/            # CLI tool
        └── main.go
```

## Core Types

### Plugin (plugins/core/plugin.go)

```go
// Plugin represents a canonical plugin/extension definition
type Plugin struct {
    // Metadata
    Name        string `json:"name"`
    Version     string `json:"version"`
    Description string `json:"description"`
    Author      string `json:"author,omitempty"`
    License     string `json:"license,omitempty"`
    Repository  string `json:"repository,omitempty"`
    Homepage    string `json:"homepage,omitempty"`

    // Components
    Commands []string `json:"commands,omitempty"` // Paths to command specs
    Skills   []string `json:"skills,omitempty"`   // Paths to skill specs
    Agents   []string `json:"agents,omitempty"`   // Paths to agent specs
    Hooks    string   `json:"hooks,omitempty"`    // Path to hooks spec

    // Dependencies
    Dependencies []Dependency `json:"dependencies,omitempty"`

    // MCP Servers (for Gemini)
    MCPServers map[string]MCPServer `json:"mcp_servers,omitempty"`
}

type Dependency struct {
    Name     string `json:"name"`
    Command  string `json:"command,omitempty"`  // CLI command to check
    Optional bool   `json:"optional,omitempty"`
}

type MCPServer struct {
    Command string            `json:"command"`
    Args    []string          `json:"args,omitempty"`
    Env     map[string]string `json:"env,omitempty"`
}
```

### Command (commands/core/command.go)

```go
// Command represents a canonical command/prompt definition
type Command struct {
    // Metadata
    Name        string `json:"name"`
    Description string `json:"description"`

    // Arguments
    Arguments []Argument `json:"arguments,omitempty"`

    // Content
    Instructions string `json:"instructions"` // The prompt/instructions

    // Process steps (for documentation)
    Process []string `json:"process,omitempty"`

    // Dependencies
    Dependencies []string `json:"dependencies,omitempty"`
}

type Argument struct {
    Name        string `json:"name"`
    Type        string `json:"type"`                  // string, number, boolean
    Required    bool   `json:"required,omitempty"`
    Default     string `json:"default,omitempty"`
    Pattern     string `json:"pattern,omitempty"`     // Regex validation
    Hint        string `json:"hint,omitempty"`        // User hint
    Description string `json:"description,omitempty"`
}
```

### Skill (skills/core/skill.go)

```go
// Skill represents a canonical skill definition
type Skill struct {
    // Metadata
    Name        string `json:"name"`
    Description string `json:"description"`

    // Content
    Instructions string `json:"instructions"` // The skill instructions

    // Resources
    Scripts    []string `json:"scripts,omitempty"`    // Script files
    References []string `json:"references,omitempty"` // Reference docs
    Assets     []string `json:"assets,omitempty"`     // Template files

    // Invocation
    Triggers []string `json:"triggers,omitempty"` // Keywords that invoke
}
```

### Agent (agents/core/agent.go)

```go
// Agent represents a canonical agent definition
type Agent struct {
    // Metadata
    Name        string `json:"name"`
    Description string `json:"description"`

    // Configuration
    Model  string   `json:"model,omitempty"`  // sonnet, opus, haiku
    Tools  []string `json:"tools,omitempty"`  // Allowed tools
    Skills []string `json:"skills,omitempty"` // Skills to load

    // Content
    Instructions string `json:"instructions"` // Agent instructions

    // Behavior
    MaxTurns      int  `json:"max_turns,omitempty"`
    AllowParallel bool `json:"allow_parallel,omitempty"`
}
```

## Adapter Interface

Each package follows the same adapter pattern:

```go
// Adapter converts between canonical and tool-specific formats
type Adapter interface {
    // Name returns the adapter identifier (e.g., "claude", "gemini")
    Name() string

    // DefaultPaths returns default file paths for this tool
    DefaultPaths() []string

    // Parse converts tool-specific bytes to canonical type
    Parse(data []byte) (*Config, error)

    // Marshal converts canonical type to tool-specific bytes
    Marshal(cfg *Config) ([]byte, error)

    // ReadFile reads from path and returns canonical type
    ReadFile(path string) (*Config, error)

    // WriteFile writes canonical type to path
    WriteFile(cfg *Config, path string) error
}
```

## Output Formats

### Claude Command (commands/claude/)

Output: `commands/{name}.md`

```markdown
---
description: Execute full release workflow
---

# Release

Execute a full release workflow with validation, changelog generation, and git tagging.

## Usage

```
/release-agent:release <version>
```

## Arguments

- **version** (required): Semantic version (e.g., v1.2.3)

## Process

1. Run validation checks
2. Generate changelog
3. Create and push git tag
```

### Gemini Command (commands/gemini/)

Output: `commands/{group}/{name}.toml`

```toml
[command]
name = "release"
description = "Execute full release workflow"

[[arguments]]
name = "version"
type = "string"
required = true
hint = "Semantic version (e.g., v1.2.3)"

[content]
instructions = """
Execute a full release workflow with validation, changelog generation, and git tagging.

Process:
1. Run validation checks
2. Generate changelog
3. Create and push git tag
"""
```

### Codex Prompt (commands/codex/)

Output: `prompts/{name}.md`

```markdown
---
description: Execute full release workflow
argument-hint: VERSION=<semver>
---

Execute a full release workflow with validation, changelog generation, and git tagging.

Arguments:
- $VERSION: Semantic version (e.g., v1.2.3)

Process:
1. Run validation checks
2. Generate changelog
3. Create and push git tag
```

## JSON Schemas

### plugin.schema.json

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://github.com/plexusone/assistantkit/plugins/schema/plugin.schema.json",
  "title": "AI Assist Kit Plugin",
  "description": "Canonical plugin/extension definition",
  "type": "object",
  "required": ["name", "version", "description"],
  "properties": {
    "name": {
      "type": "string",
      "pattern": "^[a-z][a-z0-9-]*$",
      "description": "Plugin identifier (lowercase, hyphens)"
    },
    "version": {
      "type": "string",
      "pattern": "^\\d+\\.\\d+\\.\\d+",
      "description": "Semantic version"
    },
    "description": {
      "type": "string",
      "maxLength": 200
    },
    "commands": {
      "type": "array",
      "items": {"type": "string"}
    },
    "skills": {
      "type": "array",
      "items": {"type": "string"}
    },
    "agents": {
      "type": "array",
      "items": {"type": "string"}
    },
    "dependencies": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["name"],
        "properties": {
          "name": {"type": "string"},
          "command": {"type": "string"},
          "optional": {"type": "boolean"}
        }
      }
    }
  }
}
```

## Registry Pattern

Each package uses a registry for adapter management:

```go
var (
    mu       sync.RWMutex
    adapters = make(map[string]Adapter)
)

// Register adds an adapter to the registry
func Register(adapter Adapter) {
    mu.Lock()
    defer mu.Unlock()
    adapters[adapter.Name()] = adapter
}

// GetAdapter returns an adapter by name
func GetAdapter(name string) (Adapter, bool) {
    mu.RLock()
    defer mu.RUnlock()
    adapter, ok := adapters[name]
    return adapter, ok
}

// AdapterNames returns all registered adapter names
func AdapterNames() []string {
    mu.RLock()
    defer mu.RUnlock()
    names := make([]string, 0, len(adapters))
    for name := range adapters {
        names = append(names, name)
    }
    sort.Strings(names)
    return names
}
```

Adapters self-register via init():

```go
func init() {
    core.Register(&claudeAdapter{})
}
```

## Error Handling

Custom error types with unwrapping:

```go
type ParseError struct {
    Format string
    Path   string
    Err    error
}

func (e *ParseError) Error() string {
    return fmt.Sprintf("parse %s (%s): %v", e.Format, e.Path, e.Err)
}

func (e *ParseError) Unwrap() error {
    return e.Err
}
```

## File Permissions

Secure defaults for generated files:

```go
const DefaultFileMode fs.FileMode = 0600 // User read/write only
const DefaultDirMode fs.FileMode = 0700  // User read/write/execute only
```

## Testing Strategy

1. **Unit Tests**: Each adapter has round-trip tests
2. **Integration Tests**: Cross-adapter conversion tests
3. **Schema Validation**: JSON Schema validation tests
4. **Golden Files**: Expected output comparisons

## Dependencies

- Go 1.23+ (for modern features)
- `github.com/pelletier/go-toml/v2` (Codex/Gemini TOML)
- No other external dependencies (stdlib only where possible)

## Performance Considerations

- Lazy adapter loading via init()
- Minimal memory allocation in hot paths
- Streaming for large file operations

## Security Considerations

- No code execution from parsed configs
- Path traversal prevention
- Secure file permissions by default
- No network operations in core library
