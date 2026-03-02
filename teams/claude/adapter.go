// Package claude provides the Claude Code team adapter for self-directed workflows.
package claude

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	multiagentspec "github.com/plexusone/multi-agent-spec/sdk/go"

	"github.com/plexusone/assistantkit/teams/core"
)

func init() {
	core.Register(&Adapter{})
}

// Adapter converts team definitions to Claude Code team format.
type Adapter struct{}

// Name returns the adapter identifier.
func (a *Adapter) Name() string {
	return "claude"
}

// FileExtension returns the file extension for Claude team files.
func (a *Adapter) FileExtension() string {
	return ".md"
}

// Parse converts Claude team Markdown bytes to canonical Team.
// Note: Claude Code team files are typically generated, not parsed.
func (a *Adapter) Parse(data []byte) (*core.Team, error) {
	return nil, fmt.Errorf("claude team adapter does not support parsing")
}

// Marshal converts canonical Team to Claude team Markdown bytes.
func (a *Adapter) Marshal(team *core.Team) ([]byte, error) {
	return a.MarshalWithSpec(team, nil)
}

// MarshalWithSpec converts a Team with multi-agent-spec Team to Claude team Markdown.
func (a *Adapter) MarshalWithSpec(team *core.Team, specTeam *multiagentspec.Team) ([]byte, error) {
	var buf bytes.Buffer

	// Write header
	buf.WriteString(fmt.Sprintf("# %s Team\n\n", toTitle(team.Name)))
	if team.Description != "" {
		buf.WriteString(fmt.Sprintf("%s\n\n", team.Description))
	}

	// Determine workflow type
	workflowType := "graph" // default
	if specTeam != nil && specTeam.Workflow != nil {
		workflowType = string(specTeam.Workflow.Type)
	}

	// Write workflow info
	buf.WriteString("## Workflow Configuration\n\n")
	buf.WriteString(fmt.Sprintf("**Type:** `%s`\n", workflowType))

	// Add category info
	if specTeam != nil && specTeam.Workflow != nil {
		category := specTeam.Workflow.Type.Category()
		buf.WriteString(fmt.Sprintf("**Category:** %s\n", category))
	}

	// Add self-directed workflow specific configuration
	if specTeam != nil && specTeam.IsSelfDirected() {
		buf.WriteString("\n### Self-Directed Configuration\n\n")
		a.writeSelfDirectedConfig(&buf, specTeam)
	}

	buf.WriteString("\n---\n\n")

	// Write agent roster
	buf.WriteString("## Team Roster\n\n")
	buf.WriteString("| Agent | Role | Description |\n")
	buf.WriteString("|-------|------|-------------|\n")
	for _, agentName := range team.Agents {
		role := "Specialist"
		if specTeam != nil && specTeam.EffectiveLead() == agentName {
			role = "Lead"
		}
		buf.WriteString(fmt.Sprintf("| `%s` | %s | - |\n", agentName, role))
	}
	buf.WriteString("\n")

	// Write workflow-specific instructions
	if specTeam != nil && specTeam.IsSelfDirected() {
		a.writeWorkflowInstructions(&buf, specTeam)
	} else {
		// Fallback to deterministic workflow instructions
		a.writeDeterministicInstructions(&buf, team)
	}

	return buf.Bytes(), nil
}

// writeSelfDirectedConfig writes the collaboration configuration section.
func (a *Adapter) writeSelfDirectedConfig(buf *bytes.Buffer, team *multiagentspec.Team) {
	if team.Collaboration == nil {
		return
	}

	collab := team.Collaboration

	// Lead agent
	if collab.Lead != "" {
		buf.WriteString(fmt.Sprintf("**Lead Agent:** `%s`\n", collab.Lead))
	}

	// Specialists
	if len(collab.Specialists) > 0 {
		buf.WriteString(fmt.Sprintf("**Specialists:** %s\n", strings.Join(collab.Specialists, ", ")))
	}

	// Task queue
	if collab.TaskQueue {
		buf.WriteString("**Task Queue:** Enabled (agents can self-claim tasks)\n")
	}

	// Consensus rules
	if collab.Consensus != nil {
		buf.WriteString("\n**Consensus Rules:**\n")
		buf.WriteString(fmt.Sprintf("- Required Agreement: %.0f%%\n", collab.Consensus.EffectiveRequiredAgreement()*100))
		buf.WriteString(fmt.Sprintf("- Max Rounds: %d\n", collab.Consensus.EffectiveMaxRounds()))
		if collab.Consensus.TieBreaker != "" {
			buf.WriteString(fmt.Sprintf("- Tie Breaker: %s\n", collab.Consensus.TieBreaker))
		}
	}

	// Channels
	if len(collab.Channels) > 0 {
		buf.WriteString("\n**Communication Channels:**\n\n")
		buf.WriteString("| Channel | Type | Participants |\n")
		buf.WriteString("|---------|------|-------------|\n")
		for _, ch := range collab.Channels {
			participants := strings.Join(ch.Participants, ", ")
			if len(ch.Participants) == 0 || (len(ch.Participants) == 1 && ch.Participants[0] == "*") {
				participants = "all"
			}
			buf.WriteString(fmt.Sprintf("| %s | %s | %s |\n", ch.Name, ch.Type, participants))
		}
		buf.WriteString("\n")
	}

	// Team-level flags
	if team.SelfClaim {
		buf.WriteString("**Self-Claim:** Agents can claim tasks from shared queue\n")
	}
	if team.PlanApproval {
		buf.WriteString("**Plan Approval:** Required before implementation\n")
	}
}

// writeWorkflowInstructions writes workflow-specific instructions.
func (a *Adapter) writeWorkflowInstructions(buf *bytes.Buffer, team *multiagentspec.Team) {
	if team.Workflow == nil {
		return
	}

	buf.WriteString("## Orchestration Instructions\n\n")

	switch team.Workflow.Type {
	case multiagentspec.WorkflowCrew:
		a.writeCrewInstructions(buf, team)
	case multiagentspec.WorkflowSwarm:
		a.writeSwarmInstructions(buf, team)
	case multiagentspec.WorkflowCouncil:
		a.writeCouncilInstructions(buf, team)
	default:
		// Deterministic workflows
		buf.WriteString("This team uses a deterministic workflow. ")
		buf.WriteString("Tasks are executed according to the step definitions.\n\n")
	}
}

// writeCrewInstructions writes crew workflow instructions.
func (a *Adapter) writeCrewInstructions(buf *bytes.Buffer, team *multiagentspec.Team) {
	lead := team.EffectiveLead()

	buf.WriteString("### Crew Workflow\n\n")
	buf.WriteString(fmt.Sprintf("The **%s** agent leads this crew and delegates tasks to specialists.\n\n", lead))

	buf.WriteString("**Lead Agent Responsibilities:**\n\n")
	buf.WriteString("1. Analyze incoming requests and break them into subtasks\n")
	buf.WriteString("2. Delegate subtasks to appropriate specialist agents\n")
	buf.WriteString("3. Collect and synthesize results from specialists\n")
	buf.WriteString("4. Make final decisions and report outcomes\n\n")

	buf.WriteString("**Specialist Agent Responsibilities:**\n\n")
	buf.WriteString("1. Wait for task delegation from the lead agent\n")
	buf.WriteString("2. Execute assigned tasks within your area of expertise\n")
	buf.WriteString("3. Report findings back to the lead agent\n")
	buf.WriteString("4. Respond to follow-up questions from the lead\n\n")

	buf.WriteString("**Delegation Protocol:**\n\n")
	buf.WriteString("```\n")
	buf.WriteString("Lead → Specialist: delegate_work message with task details\n")
	buf.WriteString("Specialist → Lead: share_finding message with results\n")
	buf.WriteString("Lead → Specialist: ask_question for clarification (if needed)\n")
	buf.WriteString("```\n\n")

	if team.PlanApproval {
		buf.WriteString("**Plan Approval Required:** The lead must submit a plan for approval before delegating work.\n\n")
	}
}

// writeSwarmInstructions writes swarm workflow instructions.
func (a *Adapter) writeSwarmInstructions(buf *bytes.Buffer, _ *multiagentspec.Team) {
	buf.WriteString("### Swarm Workflow\n\n")
	buf.WriteString("This is a self-organizing swarm where agents claim tasks from a shared queue.\n\n")

	buf.WriteString("**Agent Responsibilities:**\n\n")
	buf.WriteString("1. Monitor the shared task queue for available work\n")
	buf.WriteString("2. Claim tasks that match your capabilities\n")
	buf.WriteString("3. Execute claimed tasks and report completion\n")
	buf.WriteString("4. Share relevant findings with other agents\n\n")

	buf.WriteString("**Task Queue Protocol:**\n\n")
	buf.WriteString("```\n")
	buf.WriteString("Agent: task_claimed message when claiming a task\n")
	buf.WriteString("Agent: task_completed message when done\n")
	buf.WriteString("Agent: share_finding to broadcast discoveries\n")
	buf.WriteString("```\n\n")

	buf.WriteString("**Coordination Rules:**\n\n")
	buf.WriteString("- First agent to claim a task owns it\n")
	buf.WriteString("- If stuck, share findings and release the task\n")
	buf.WriteString("- Broadcast important discoveries to all agents\n\n")
}

// writeCouncilInstructions writes council workflow instructions.
func (a *Adapter) writeCouncilInstructions(buf *bytes.Buffer, team *multiagentspec.Team) {
	buf.WriteString("### Council Workflow\n\n")
	buf.WriteString("This is a peer council where agents debate and reach consensus.\n\n")

	buf.WriteString("**Agent Responsibilities:**\n\n")
	buf.WriteString("1. Analyze the problem from your area of expertise\n")
	buf.WriteString("2. Share your findings and recommendations\n")
	buf.WriteString("3. Review and challenge findings from other agents\n")
	buf.WriteString("4. Vote on final recommendations\n\n")

	buf.WriteString("**Debate Protocol:**\n\n")
	buf.WriteString("```\n")
	buf.WriteString("Agent: share_finding to present analysis\n")
	buf.WriteString("Agent: challenge to dispute another's finding\n")
	buf.WriteString("Agent: ask_question for clarification\n")
	buf.WriteString("Agent: vote to cast final vote\n")
	buf.WriteString("```\n\n")

	if team.Collaboration != nil && team.Collaboration.Consensus != nil {
		consensus := team.Collaboration.Consensus
		buf.WriteString("**Consensus Rules:**\n\n")
		buf.WriteString(fmt.Sprintf("- Requires %.0f%% agreement to pass\n", consensus.EffectiveRequiredAgreement()*100))
		buf.WriteString(fmt.Sprintf("- Maximum %d debate rounds\n", consensus.EffectiveMaxRounds()))
		if consensus.TieBreaker != "" {
			buf.WriteString(fmt.Sprintf("- Tie breaker: %s\n", consensus.TieBreaker))
		}
		buf.WriteString("\n")
	}

	buf.WriteString("**Voting Process:**\n\n")
	buf.WriteString("1. All agents share their findings\n")
	buf.WriteString("2. Challenge/debate phase (max rounds apply)\n")
	buf.WriteString("3. Final vote on recommendations\n")
	buf.WriteString("4. Consensus reached or tie-breaker decides\n\n")
}

// writeDeterministicInstructions writes instructions for deterministic workflows.
func (a *Adapter) writeDeterministicInstructions(buf *bytes.Buffer, team *core.Team) {
	buf.WriteString("## Orchestration Instructions\n\n")

	// Get parallel groups
	groups, err := team.ParallelGroups()
	if err != nil {
		buf.WriteString("Execute tasks in dependency order.\n\n")
		return
	}

	for i, group := range groups {
		if len(group) == 0 {
			continue
		}

		if len(group) > 1 {
			buf.WriteString(fmt.Sprintf("### Parallel Group %d\n\n", i+1))
			buf.WriteString("These tasks can run concurrently:\n\n")
		} else {
			buf.WriteString(fmt.Sprintf("### Step %d\n\n", i+1))
		}

		for _, task := range group {
			buf.WriteString(fmt.Sprintf("- **%s** (agent: `%s`)\n", task.Name, task.Agent))
			if task.Description != "" {
				buf.WriteString(fmt.Sprintf("  %s\n", task.Description))
			}
		}
		buf.WriteString("\n")
	}
}

// ReadFile reads a Claude team file and returns canonical Team.
// Note: Claude Code team files are typically generated, not parsed.
func (a *Adapter) ReadFile(path string) (*core.Team, error) {
	return nil, fmt.Errorf("claude team adapter does not support reading")
}

// WriteFile writes canonical Team to a Claude team Markdown file.
func (a *Adapter) WriteFile(team *core.Team, path string) error {
	data, err := a.Marshal(team)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, core.DefaultDirMode); err != nil {
		return &core.WriteError{Path: path, Err: err}
	}

	if err := os.WriteFile(path, data, core.DefaultFileMode); err != nil {
		return &core.WriteError{Path: path, Err: err}
	}

	return nil
}

// WriteFileWithSpec writes a Team with multi-agent-spec Team to a Claude team file.
func (a *Adapter) WriteFileWithSpec(team *core.Team, specTeam *multiagentspec.Team, path string) error {
	data, err := a.MarshalWithSpec(team, specTeam)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, core.DefaultDirMode); err != nil {
		return &core.WriteError{Path: path, Err: err}
	}

	if err := os.WriteFile(path, data, core.DefaultFileMode); err != nil {
		return &core.WriteError{Path: path, Err: err}
	}

	return nil
}

// GenerateTeamDir generates a complete Claude Code team directory.
// This includes the team orchestration file and agent files with role information.
func (a *Adapter) GenerateTeamDir(
	team *core.Team,
	specTeam *multiagentspec.Team,
	agents []*multiagentspec.Agent,
	outputDir string,
) error {
	// Create output directory
	if err := os.MkdirAll(outputDir, core.DefaultDirMode); err != nil {
		return &core.WriteError{Path: outputDir, Err: err}
	}

	// Write team orchestration file
	teamPath := filepath.Join(outputDir, "team.md")
	if err := a.WriteFileWithSpec(team, specTeam, teamPath); err != nil {
		return fmt.Errorf("writing team file: %w", err)
	}

	// Write agent files with role information
	agentsDir := filepath.Join(outputDir, "agents")
	if err := os.MkdirAll(agentsDir, core.DefaultDirMode); err != nil {
		return &core.WriteError{Path: agentsDir, Err: err}
	}

	for _, agt := range agents {
		data := marshalAgentWithRole(agt, specTeam)
		path := filepath.Join(agentsDir, agt.Name+".md")
		if err := os.WriteFile(path, data, core.DefaultFileMode); err != nil {
			return &core.WriteError{Path: path, Err: err}
		}
	}

	return nil
}

// marshalAgentWithRole converts an agent to Markdown with role information.
func marshalAgentWithRole(agent *multiagentspec.Agent, _ *multiagentspec.Team) []byte {
	var buf bytes.Buffer

	// Write YAML frontmatter
	buf.WriteString("---\n")
	buf.WriteString(fmt.Sprintf("name: %s\n", agent.Name))
	if agent.Description != "" {
		buf.WriteString(fmt.Sprintf("description: %s\n", agent.Description))
	}

	if agent.Model != "" {
		buf.WriteString(fmt.Sprintf("model: %s\n", string(agent.Model)))
	}

	if len(agent.Tools) > 0 {
		buf.WriteString(fmt.Sprintf("tools: [%s]\n", strings.Join(agent.Tools, ", ")))
	}

	// Add role-based fields for self-directed workflows
	if agent.Role != "" {
		buf.WriteString(fmt.Sprintf("role: %s\n", agent.Role))
	}
	if agent.Goal != "" {
		buf.WriteString(fmt.Sprintf("goal: %s\n", agent.Goal))
	}

	// Add delegation config
	if agent.Delegation != nil && agent.Delegation.AllowDelegation {
		buf.WriteString("delegation:\n")
		buf.WriteString("  allow_delegation: true\n")
		if len(agent.Delegation.CanDelegateTo) > 0 {
			buf.WriteString(fmt.Sprintf("  can_delegate_to: [%s]\n", strings.Join(agent.Delegation.CanDelegateTo, ", ")))
		}
	}

	buf.WriteString("---\n\n")

	// Write role context if present
	if agent.Role != "" {
		buf.WriteString(fmt.Sprintf("## Role: %s\n\n", agent.Role))
	}
	if agent.Goal != "" {
		buf.WriteString(fmt.Sprintf("**Goal:** %s\n\n", agent.Goal))
	}
	if agent.Backstory != "" {
		buf.WriteString(fmt.Sprintf("**Background:**\n%s\n\n", agent.Backstory))
	}

	// Write instructions
	if agent.Instructions != "" {
		buf.WriteString("## Instructions\n\n")
		buf.WriteString(agent.Instructions)
		buf.WriteString("\n")
	}

	return buf.Bytes()
}

// toTitle converts a kebab-case string to Title Case.
func toTitle(s string) string {
	words := strings.Split(s, "-")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}
