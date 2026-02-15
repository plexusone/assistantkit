# Validate Command

The `validate` command performs static validation on specs directories before generation.

## Synopsis

```bash
assistantkit validate [flags]
```

## Description

This command validates plugin definitions in a specs directory without generating any output. It checks for common errors and configuration issues that would cause generation to fail.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--specs` | `specs` | Path to unified specs directory |

## Validation Checks

The validator performs the following checks:

- **plugin.json** — required fields present, valid JSON
- **agents** — valid YAML frontmatter with required fields (name, description)
- **skills** — valid skill definitions
- **commands** — valid command definitions
- **skill-refs** — agent skill references resolve to existing skills
- **teams** — DAG acyclicity, agent references, input `from` references
- **deployments** — valid platforms, no output path conflicts

## Output

```bash
assistantkit validate --specs=specs

=== AssistantKit Validator ===
Specs directory: /path/to/specs

✓ plugin.json    my-team v1.0.0
✓ agents         6 agents
✓ skills         4 skills
✓ commands       2 commands
✓ skill-refs     8 references resolve
✓ teams          1 teams, 6 steps, 4 phases
✓ deployments    1 deployments, 2 targets

✓ Validation passed (6 agents, 4 skills, 2 commands)
```

### Error Output

When validation fails, errors are displayed with context:

```bash
assistantkit validate --specs=specs

=== AssistantKit Validator ===
Specs directory: /path/to/specs

✓ plugin.json    my-team v1.0.0
✓ agents         6 agents
✓ skills         4 skills
✗ skill-refs     agent 'writer' references unknown skill: nonexistent
✗ teams          step 'review' references unknown agent: reviewer

✗ Validation failed with 2 error(s)
```

## Team DAG Validation

For team workflow files, the validator checks:

- **DAG acyclicity** — detects circular dependencies in `depends_on`
- **Agent references** — verifies each step's agent exists in `agents/`
- **Input references** — validates `from` fields use correct `step.output` format
- **Phase counting** — reports parallel execution phases from topological sort

## Integration with Generate

The `generate` command automatically runs validation before generation. Use `--skip-validate` to bypass:

```bash
# Normal flow: validate then generate
assistantkit generate --specs=specs

# Skip validation
assistantkit generate --specs=specs --skip-validate
```

## Examples

Validate using defaults:

```bash
assistantkit validate
```

Validate a specific specs directory:

```bash
assistantkit validate --specs=my-specs
```

## See Also

- [Generate Command](generate-plugins.md) - Generate plugins from specs
- [Plugin Structure](../plugins/structure.md) - Learn about plugin components
- [Agents](../plugins/agents.md) - Agent definition details
