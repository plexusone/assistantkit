# Enhancement: `assistantkit validate` and `assistantkit test`

## Problem

`assistantkit` is codegen-only. There's no way to verify that a specs directory is correct before generating plugins. Errors surface at runtime when agents fail to orchestrate, not at build time.

## Status

### ✅ Implemented

- **`assistantkit validate`** — Static validation of specs directory
- **`generate` + `validate` integration** — Generate runs validate first, `--skip-validate` to bypass
- **Single-source prefix for Kiro** — Prefix defined once in deployment config, applied everywhere

### ⏳ Remaining

- **`assistantkit test`** — Dry-run orchestration with mock agent outputs (deferred)
- **Conditional execution rules** — `run_if` expressions validation (not yet implemented)

---

## `assistantkit validate` (IMPLEMENTED)

Static validation of specs directory. Catches errors before `generate`.

Checks:
- ✅ **plugin.json** — required fields present, valid JSON
- ✅ **DAG acyclicity** — team workflow `depends_on` references form a valid DAG (no cycles)
- ✅ **Agent references resolve** — every `agent` in team steps exists in `agents/`
- ✅ **Skill references resolve** — every skill referenced by an agent exists in `skills/`
- ✅ **Input references** — `from` fields reference valid `step.output` paths
- ✅ **Deployment targets** — platform names are valid, output paths don't conflict
- ✅ **Frontmatter schema** — agent `.md` files have valid YAML frontmatter with required fields
- ⏳ **Conditional execution rules** — `run_if` expressions reference valid step IDs (not yet)

Example:
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

---

## `generate` + `validate` Integration (IMPLEMENTED)

`generate` runs `validate` first and refuses to generate if validation fails:

```bash
assistantkit generate --specs=specs

=== AssistantKit Generator ===
Specs directory: /path/to/specs
Target: local
Output directory: .

Validating specs...
✓ Validation passed (6 agents, 4 skills, 2 commands)

Team: my-team
Loaded: 2 commands, 4 skills, 6 agents

Generated targets:
  - local-kiro: plugins/kiro
  - local-claude: .claude/agents

Done!
```

Add `--skip-validate` flag to bypass:
```bash
assistantkit generate --specs=specs --skip-validate
```

---

## Single-Source Prefix for Kiro (IMPLEMENTED)

The prefix is now defined **once** in the deployment config and applied everywhere during Kiro generation.

### Before (prefix in multiple places)

- Agent frontmatter: `name: cext_pm` ❌
- Team references: `agent: pm` — mismatch!
- Deployment config: `prefix: "cext_"`

### After (prefix in ONE place)

- Agent frontmatter: `name: pm` ✅ (canonical short name)
- Team references: `agent: pm` ✅ (matches!)
- Deployment config: `kiroCli.prefix: "cext_"` ✅ (applied at generation)

### Deployment config format

```json
{
  "targets": [
    {
      "name": "local-kiro",
      "platform": "kiro-cli",
      "output": "plugins/kiro",
      "kiroCli": {
        "prefix": "cext_",
        "pluginDir": "plugins/kiro",
        "format": "json"
      }
    }
  ]
}
```

### Generated output

- Agent JSON filename: `cext_pm.json`
- Agent name field: `"name": "cext_pm"`
- Steering filename: `cext_use-case-analysis.md`
- README examples: `kiro-cli chat --agent cext_pm`

This enables multiple teams (cext_, prd_, rel_) to share a single Kiro agents directory without naming conflicts.

---

## `assistantkit test` (DEFERRED)

Dry-run orchestration with mock agent outputs. Verifies the DAG executes correctly without calling any LLM.

Checks:
- **Execution order** — phases run in correct topological order
- **Conditional branching** — given mock outputs, correct agents are skipped/triggered
- **Short-circuit paths** — e.g., "JAR not in use" skips all downstream agents
- **Output propagation** — downstream agents receive expected inputs from upstream

Takes a test fixture file with mock agent outputs:
```bash
assistantkit test --specs=specs --fixture=tests/trms-compatible.json

Phase 1: pm ✓, security ✓, performance ✓ (parallel)
Phase 2: trms-migration ✓ (inputs: pm, security, performance)
Phase 3: grails6-migration SKIPPED (trms compatible)
Phase 4: coordinator ✓ (inputs: pm, security, performance, trms-migration)

✓ All 4 phases executed correctly
✓ Conditional skip: grails6-migration (trms_compatible=true)
✓ Disposition: (c) TRMS Migration
```

### Fixture File Format

```json
{
  "name": "trms-compatible",
  "description": "JAR is in use, not OOTB, TRMS compatible — should skip Grails 6",
  "mock_outputs": {
    "pm-analysis": {
      "status": "GO",
      "outputs": {"jar_in_use": true, "disposition": "trms"}
    },
    "security-analysis": {
      "status": "WARN",
      "outputs": {"risk_level": "medium"}
    },
    "performance-analysis": {
      "status": "GO",
      "outputs": {"risk_level": "low"}
    },
    "trms-migration-analysis": {
      "status": "GO",
      "outputs": {"trms_compatible": true}
    }
  },
  "expected": {
    "skipped": ["grails6-migration-analysis"],
    "disposition": "trms",
    "final_status": "GO"
  }
}
```

---

# Bug Fix: Kiro Agent Tool Mapping (FIXED ✅)

`convertToKiroAgent()` was not mapping the `tools` field from agent specs to Kiro tool names.

**Fix applied**: Added `Tools []string` to `KiroAgent` and `mapKiroTools()` with mapping:

| Spec Name | Kiro Tool Name |
|-----------|---------------|
| Read | fs_read |
| Write | fs_write |
| Bash | execute_bash |
| Grep | grep |
| Glob | glob |
| WebFetch | web_fetch |
| Code | code |

Unmapped names pass through as-is (e.g., MCP tool names like `Aha`).

---

# Bug Fix: Steering File Prefix (FIXED ✅)

**Problem**: `generateKiroAgents()` did not apply the `prefix` config from the deployment spec to steering filenames.

**Fix applied**: The prefix from `kiroCli.prefix` is now applied to:
- Agent JSON filenames
- Agent name field inside JSON
- Steering filenames
- README usage examples

---

# Enhancement: Kiro CLI Invocation Docs (FIXED ✅)

The generated `README.md` now includes correct Kiro CLI invocation syntax with prefixed agent names:

```bash
# Single agent
kiro-cli chat --agent cext_pm "<your prompt>"

# Full team (coordinator-driven)
kiro-cli chat --agent cext_coordinator "<your prompt>"
```

The generator detects coordinator/orchestrator agents and shows team usage examples.
