# Skills

Skills are reusable prompt templates that can be invoked by name or combined with other prompts.

## Canonical Format

```go
type Skill struct {
    Name        string   `yaml:"name"`
    Description string   `yaml:"description"`
    Prompt      string   `yaml:"prompt"`
    Model       string   `yaml:"model,omitempty"`
    Tools       []string `yaml:"tools,omitempty"`
}
```

## Creating a Skill

### Using Go

```go
package main

import (
    "github.com/plexusone/assistantkit/skills/core"
)

func main() {
    skill := core.Skill{
        Name:        "code-review",
        Description: "Review code for quality and best practices",
        Prompt: `Review the provided code for:

## Security
- Check for common vulnerabilities (injection, XSS, etc.)
- Verify proper input validation
- Ensure sensitive data is protected

## Performance
- Identify inefficient algorithms
- Check for unnecessary allocations
- Look for N+1 query patterns

## Maintainability
- Assess code clarity and readability
- Check for proper error handling
- Verify adequate documentation

Provide specific, actionable feedback with code examples.`,
    }
}
```

### Using Markdown

Create `skills/code-review.md`:

```markdown
---
name: code-review
description: Review code for quality and best practices
---

Review the provided code for:

## Security

- Check for common vulnerabilities (injection, XSS, etc.)
- Verify proper input validation
- Ensure sensitive data is protected

## Performance

- Identify inefficient algorithms
- Check for unnecessary allocations
- Look for N+1 query patterns

## Maintainability

- Assess code clarity and readability
- Check for proper error handling
- Verify adequate documentation

Provide specific, actionable feedback with code examples.
```

## Skill Options

| Field | Description | Required |
|-------|-------------|----------|
| `name` | Skill identifier | Yes |
| `description` | Short description | Yes |
| `prompt` | Detailed instructions | Yes |
| `model` | Preferred model | No |
| `tools` | Required tools | No |

## Assistant Support

| Assistant | Skills Support |
|-----------|---------------|
| Claude Code | Yes |
| Gemini CLI | Yes |
| OpenAI Codex | No |
| AWS Kiro | Yes (via steering files) |

## Assistant-Specific Output

### Claude Code

Skills are stored in `skills/` directory:

```markdown
---
name: code-review
description: Review code for quality and best practices
---

Review the provided code for...
```

### Gemini CLI

Skills are YAML files:

```yaml
name: code-review
description: Review code for quality and best practices
prompt: |
  Review the provided code for...
```

### AWS Kiro

Skills map to steering files in `.kiro/steering/`:

```markdown
# Code Review

Review the provided code for...
```

## Examples

### Documentation Skill

```markdown
---
name: document
description: Generate documentation for code
---

Generate comprehensive documentation for the provided code:

1. **Overview** - What does this code do?
2. **Usage** - How to use it with examples
3. **Parameters** - Document all inputs
4. **Returns** - Document outputs
5. **Errors** - What can go wrong
6. **Examples** - Real-world usage

Use appropriate documentation format:
- Go: GoDoc comments
- Python: Docstrings
- JavaScript: JSDoc
```

### Refactor Skill

```markdown
---
name: refactor
description: Suggest code refactoring improvements
---

Analyze the code and suggest refactoring improvements:

## Look For

- Duplicated code that could be extracted
- Long functions that should be split
- Complex conditionals to simplify
- Poor naming that reduces clarity
- Missing abstractions

## Guidelines

- Maintain backward compatibility
- Preserve existing behavior
- Keep changes focused and minimal
- Explain the benefit of each change

Provide before/after code examples.
```

### Security Audit Skill

```markdown
---
name: security-audit
description: Audit code for security vulnerabilities
model: opus
---

Perform a security audit of the provided code:

## OWASP Top 10

- [ ] Injection (SQL, Command, etc.)
- [ ] Broken Authentication
- [ ] Sensitive Data Exposure
- [ ] XML External Entities (XXE)
- [ ] Broken Access Control
- [ ] Security Misconfiguration
- [ ] Cross-Site Scripting (XSS)
- [ ] Insecure Deserialization
- [ ] Using Components with Known Vulnerabilities
- [ ] Insufficient Logging & Monitoring

## Additional Checks

- Hardcoded credentials
- Insecure random number generation
- Path traversal vulnerabilities
- Race conditions

Rate severity: Critical, High, Medium, Low
Provide remediation steps for each finding.
```
