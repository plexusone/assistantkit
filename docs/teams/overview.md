# Teams

The `teams` package provides multi-agent orchestration with support for both deterministic and self-directed workflows.

## Workflow Categories

AssistantKit supports two workflow paradigms from [multi-agent-spec](https://github.com/agentplexus/multi-agent-spec):

| Category | Description | Control |
|----------|-------------|---------|
| **Deterministic** | Schema defines execution paths | Orchestrator controls flow |
| **Self-directed** | Agents decide execution paths | Agents control flow |

## Workflow Types

### Deterministic Workflows

| Type | Pattern | Use Case |
|------|---------|----------|
| `chain` | A → B → C | Sequential pipeline |
| `scatter` | A → [B,C,D] → E | Parallel fan-out/fan-in |
| `graph` | DAG | Complex dependencies |

### Self-Directed Workflows

| Type | Pattern | Use Case |
|------|---------|----------|
| `crew` | Lead → Specialists | Manager delegates to experts |
| `swarm` | Shared queue | Self-organizing agents |
| `council` | Peer debate | Consensus voting |

## Packages

| Package | Description |
|---------|-------------|
| `teams` | Re-exports core types for convenience |
| `teams/core` | `SelfDirectedTeam` wrapper, workflow helpers |
| `teams/claude` | Claude Code adapter for self-directed teams |

## Self-Directed Teams

Self-directed workflows use role-based agents that autonomously coordinate work.

### Agent Role Fields

| Field | Purpose |
|-------|---------|
| `role` | Agent's job title (e.g., "Lead Architect") |
| `goal` | What the agent aims to achieve |
| `backstory` | Context for autonomous decisions |
| `delegation` | Delegation permissions |

### SelfDirectedTeam

The `SelfDirectedTeam` wrapper provides helpers for working with self-directed workflows:

```go
import "github.com/agentplexus/assistantkit/teams/core"

// Create from multi-agent-spec definitions
team, agents := core.FromMultiAgentSpec(masTeam, agentDefs)

// Check workflow type
fmt.Println("Type:", team.WorkflowType())      // crew, swarm, or council
fmt.Println("Self-directed:", team.IsSelfDirected()) // true
fmt.Println("Deterministic:", team.IsDeterministic()) // false

// Get crew members (for crew workflow)
lead := team.Lead()           // Lead agent
specialists := team.Specialists() // Specialist agents

// Validate workflow configuration
if err := team.Validate(); err != nil {
    log.Fatal(err)
}
```

### Configuration

```go
// Default configuration
config := core.DefaultSelfDirectedConfig()

// Custom configuration
config := core.SelfDirectedConfig{
    MaxIterations: 10,
    Verbose:      true,
}

team := core.NewSelfDirectedTeam(masTeam, agents, config)
```

## Claude Code Adapter

The `teams/claude` adapter converts self-directed teams to Claude Code format.

### Generated Files

| File | Contents |
|------|----------|
| `{agent}.md` | Agent definition with role, goal, backstory |
| `teammates.json` | List of team member names |

### Example Output

**architect.md:**

```markdown
---
role: Lead Architect
goal: Design system architecture and delegate implementation
model: sonnet
tools: ["Read", "Grep", "Task"]
---

## Backstory

You are an experienced software architect with 15 years of experience
designing scalable systems. You excel at breaking down complex problems
and delegating work to specialists.

## Instructions

Design the system architecture and delegate implementation tasks to
frontend and backend specialists.
```

**teammates.json:**

```json
["frontend", "backend", "security"]
```

### Usage

```go
import "github.com/agentplexus/assistantkit/teams/claude"

adapter := claude.NewAdapter()
files, err := adapter.Convert(selfDirectedTeam)
if err != nil {
    log.Fatal(err)
}

// Write files
for name, content := range files {
    os.WriteFile(filepath.Join(outputDir, name), []byte(content), 0644)
}
```

## Teams Generation

The `generate` package includes team generation for deployment targets:

```go
import "github.com/agentplexus/assistantkit/generate"

result, err := generate.Teams(generate.TeamsOptions{
    SpecsDir: "specs",
    Output:   ".claude/agents",
    Platform: "claude-code",
})

fmt.Println("Generated files:", result.Files)
```

### Supported Platforms

| Platform | Output |
|----------|--------|
| `claude-code` | Agent markdown + teammates.json |

## Workflow Mapping

Different platforms handle self-directed workflows differently:

| Workflow | Claude Code | CrewAI |
|----------|-------------|--------|
| `crew` | `team_mode: team` | `processType: hierarchical` |
| `swarm` | `team_mode: team` | `processType: consensual` |
| `council` | `team_mode: team` | `processType: consensual` |

## Example: Code Review Council

A council workflow for peer code review:

```go
import (
    "github.com/agentplexus/assistantkit/teams/core"
    "github.com/agentplexus/assistantkit/teams/claude"
    mas "github.com/agentplexus/multi-agent-spec/sdk/go"
)

// Define agents with roles
security := mas.NewAgent("security", "Security review").
    WithRole("Security Analyst").
    WithGoal("Find vulnerabilities").
    WithBackstory("Expert in OWASP top 10...")

performance := mas.NewAgent("performance", "Performance review").
    WithRole("Performance Engineer").
    WithGoal("Identify bottlenecks").
    WithBackstory("Expert in profiling...")

maintainability := mas.NewAgent("maintainability", "Code quality").
    WithRole("Code Quality Specialist").
    WithGoal("Ensure maintainability").
    WithBackstory("Expert in clean code...")

// Create council team
team := mas.NewTeam("code-review", "1.0.0").
    WithAgents("security", "performance", "maintainability").
    WithWorkflow(&mas.Workflow{Type: mas.WorkflowCouncil}).
    WithCollaboration(&mas.CollaborationConfig{
        Consensus: &mas.ConsensusRules{
            RequiredAgreement: 0.67,
            MaxRounds:        3,
        },
    })

// Convert to assistantkit format
sdTeam := core.FromMultiAgentSpec(team, []mas.Agent{security, performance, maintainability})

// Generate Claude Code files
adapter := claude.NewAdapter()
files, _ := adapter.Convert(sdTeam)
```

## See Also

- [multi-agent-spec SDK](https://github.com/agentplexus/multi-agent-spec) - Source schema definitions
- [v0.10.0 Release Notes](../releases/v0.10.0.md) - Self-directed teams feature details
