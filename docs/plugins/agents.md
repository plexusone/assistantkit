# Agents

Agents are specialized sub-agents that handle specific tasks autonomously.

## Canonical Format

```go
type Agent struct {
    Name         string   `json:"name"`
    Description  string   `json:"description,omitempty"`
    Instructions string   `json:"instructions,omitempty"`
    Model        string   `json:"model,omitempty"`
    Tools        []string `json:"tools,omitempty"`
    Skills       []string `json:"skills,omitempty"`
}
```

## Creating an Agent

### Using Go

```go
package main

import (
    "github.com/plexusone/assistantkit/agents/core"
)

func main() {
    agent := core.Agent{
        Name:        "security-scanner",
        Description: "Scans code for security vulnerabilities",
        Instructions: `You are a security expert specializing in code review.

Your task is to analyze code for security vulnerabilities:

1. Check for OWASP Top 10 vulnerabilities
2. Look for hardcoded secrets
3. Identify insecure configurations
4. Find authentication/authorization issues

Always explain the risk and provide remediation steps.`,
        Model: "sonnet",
        Tools: []string{"Read", "Grep", "Glob"},
    }
}
```

### Using YAML

Create `agents/security-scanner.yaml`:

```yaml
name: security-scanner
description: Scans code for security vulnerabilities
instructions: |
  You are a security expert specializing in code review.

  Your task is to analyze code for security vulnerabilities:

  1. Check for OWASP Top 10 vulnerabilities
  2. Look for hardcoded secrets
  3. Identify insecure configurations
  4. Find authentication/authorization issues

  Always explain the risk and provide remediation steps.
model: sonnet
tools:
  - Read
  - Grep
  - Glob
```

## Agent Options

| Field | Description | Required |
|-------|-------------|----------|
| `name` | Agent identifier | Yes |
| `description` | Short description | Yes |
| `instructions` | Detailed behavior instructions | Yes |
| `model` | Model to use (sonnet, opus, haiku) | No |
| `tools` | Available tools | No |
| `skills` | Skills the agent can use | No |

## Assistant Support

| Assistant | Agents Support |
|-----------|---------------|
| Claude Code | Yes |
| Gemini CLI | No |
| OpenAI Codex | No |
| AWS Kiro | Yes |

## Assistant-Specific Output

### Claude Code

Agents are JSON files in `agents/`:

```json
{
  "name": "security-scanner",
  "description": "Scans code for security vulnerabilities",
  "instructions": "You are a security expert...",
  "model": "sonnet",
  "tools": ["Read", "Grep", "Glob"]
}
```

### AWS Kiro

Agents are stored in `~/.kiro/agents/`:

```json
{
  "name": "security-scanner",
  "description": "Scans code for security vulnerabilities",
  "prompt": "You are a security expert...",
  "model": "claude-sonnet-4",
  "allowedTools": ["read", "shell"]
}
```

## Tool Mapping

Tools are mapped between canonical names and assistant-specific names:

| Canonical | Claude Code | Kiro |
|-----------|-------------|------|
| Read | Read | read |
| Write | Write | write |
| Edit | Edit | edit |
| Bash | Bash | shell |
| Glob | Glob | glob |
| Grep | Grep | grep |

## Model Mapping

| Canonical | Claude Code | Kiro |
|-----------|-------------|------|
| sonnet | sonnet | claude-sonnet-4 |
| opus | opus | claude-opus-4 |
| haiku | haiku | claude-haiku-3.5 |

## Examples

### Documentation Agent

```yaml
name: doc-generator
description: Generates comprehensive documentation
instructions: |
  You are a technical writer specializing in developer documentation.

  Your task is to generate documentation that is:
  - Clear and concise
  - Well-organized
  - Complete with examples
  - Appropriate for the target audience

  Use the project's existing documentation style if present.
model: sonnet
tools:
  - Read
  - Glob
  - Write
```

### Test Generator Agent

```yaml
name: test-generator
description: Generates unit tests for code
instructions: |
  You are a testing expert who writes comprehensive unit tests.

  For each function or method:
  1. Identify edge cases and boundary conditions
  2. Write tests for happy path scenarios
  3. Write tests for error conditions
  4. Ensure good code coverage

  Follow the project's existing test patterns and frameworks.
model: sonnet
tools:
  - Read
  - Write
  - Bash
```

### Performance Analyzer Agent

```yaml
name: perf-analyzer
description: Analyzes code for performance issues
instructions: |
  You are a performance optimization expert.

  Analyze code for:
  - Algorithmic complexity issues (O(n²), O(n!), etc.)
  - Memory leaks and excessive allocations
  - I/O bottlenecks
  - Concurrency issues
  - Database query optimization

  Provide specific recommendations with benchmarks when possible.
model: opus
tools:
  - Read
  - Grep
  - Bash
```

## Invoking Agents

### Claude Code

Agents are invoked via the Task tool:

```
Use the security-scanner agent to analyze the authentication module.
```

### AWS Kiro

Agents are invoked with the `/agent` command:

```
/agent security-scanner analyze src/auth/
```
