# Enhancement: `assistantkit validate` and `assistantkit test`

## Problem

`assistantkit` is codegen-only. There's no way to verify that a specs directory is correct before generating plugins. Errors surface at runtime when agents fail to orchestrate, not at build time.

## Proposed Commands

### `assistantkit validate`

Static validation of specs directory. Catches errors before `generate`.

Checks:
- **plugin.json** — required fields present, valid JSON
- **DAG acyclicity** — team workflow `depends_on` references form a valid DAG (no cycles)
- **Agent references resolve** — every `agent_id` in team steps exists in `agents/`
- **Skill references resolve** — every skill referenced by an agent exists in `skills/`
- **Conditional execution rules** — `run_if` expressions reference valid step IDs and output fields
- **Deployment targets** — platform names are valid, output paths don't conflict
- **Frontmatter schema** — agent `.md` files have valid YAML frontmatter with required fields

Example:
```bash
assistantkit validate --specs=specs

✓ plugin.json valid
✓ DAG acyclic (6 steps, 4 phases)
✓ All agent references resolve (6/6)
✓ All skill references resolve (4/4)
✓ Conditional rules reference valid steps
✓ Deployment targets valid (2 targets)
✗ agents/grails6-migration.md: missing 'model' in frontmatter
```

### `assistantkit test`

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

### `assistantkit validate` + `generate` integration

`generate` should run `validate` first and refuse to generate if validation fails:
```bash
assistantkit generate --specs=specs

✗ Validation failed: agents/grails6-migration.md missing 'model' in frontmatter
  Fix the above errors and re-run.
```

Add `--skip-validate` flag to bypass if needed.

## Fixture File Format

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

## Priority

Medium — not blocking, but would have caught the `plugins/kiro/agents/agents/` nesting issue and would make iterating on team specs much faster.

---

# Bug Fix: Kiro Agent Tool Mapping (FIXED)

`convertToKiroAgent()` was not mapping the `tools` field from agent specs to Kiro tool names. The `KiroAgent` struct had no `Tools` field at all.

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

# Bug Fix: Steering File Prefix

`generateKiroAgents()` does not apply the `prefix` config from the deployment spec to steering filenames. Agent JSONs get the prefix (via the agent `name` field in frontmatter), but steering files use the raw skill name.

**Expected**: `cext_use-case-analysis.md`
**Actual**: `use-case-analysis.md`

**Workaround**: Manual rename after generation.

**Fix needed**: Read `prefix` from deployment config and prepend to steering filenames, matching the agent naming convention.

---

# Enhancement: Kiro CLI Invocation Docs

The generated `README.md` should include correct Kiro CLI invocation syntax. Currently it only shows `cp` install commands but no usage examples.

Should generate:
```bash
# Single agent
kiro-cli chat --agent <prefix>_<agent-name> "<prompt>"

# Full team (coordinator-driven)
kiro-cli chat --agent <prefix>_coordinator "<prompt>"
```

The generator knows the agent names and prefix — it should emit the correct `kiro-cli chat --agent` commands in the README.
