package generate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/agentplexus/assistantkit/agents"
	"github.com/agentplexus/assistantkit/teams"
	"github.com/agentplexus/assistantkit/teams/claude"
	"github.com/agentplexus/assistantkit/teams/core"

	multiagentspec "github.com/agentplexus/multi-agent-spec/sdk/go"
)

// TeamResult contains the results of team generation.
type TeamResult struct {
	// TeamName is the name of the generated team.
	TeamName string

	// WorkflowType is the workflow type of the team.
	WorkflowType string

	// AgentCount is the number of agents in the team.
	AgentCount int

	// GeneratedFiles lists the paths of generated files.
	GeneratedFiles []string

	// OutputDir is the directory where files were generated.
	OutputDir string
}

// Teams generates platform-specific team files from multi-agent-spec definitions.
//
// The specsDir should contain:
//   - agents/: Agent definitions (*.md with YAML frontmatter)
//   - teams/: Team definitions (*.json)
//
// The teamName specifies which team file to use (looks for {teamName}.json).
// The platform specifies the target platform (e.g., "claude-code").
// The outputDir is the directory for generated files.
func Teams(specsDir, teamName, platform, outputDir string) (*TeamResult, error) {
	result := &TeamResult{
		TeamName:       teamName,
		GeneratedFiles: []string{},
		OutputDir:      outputDir,
	}

	// Load the team definition
	teamFile := filepath.Join(specsDir, "teams", teamName+".json")
	specTeam, err := loadMultiAgentSpecTeam(teamFile)
	if err != nil {
		return nil, fmt.Errorf("loading team: %w", err)
	}

	if specTeam.Workflow != nil {
		result.WorkflowType = string(specTeam.Workflow.Type)
	}

	// Load agents
	agentsDir := filepath.Join(specsDir, "agents")
	allAgents, err := loadMultiAgentSpecAgents(agentsDir)
	if err != nil {
		return nil, fmt.Errorf("loading agents: %w", err)
	}

	// Filter to agents in this team
	teamAgents := filterTeamAgents(allAgents, specTeam.Agents)
	result.AgentCount = len(teamAgents)

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("creating output dir: %w", err)
	}

	// Generate based on platform
	switch platform {
	case "claude", "claude-code":
		files, err := generateClaudeTeam(specTeam, teamAgents, outputDir)
		if err != nil {
			return nil, fmt.Errorf("generating claude team: %w", err)
		}
		result.GeneratedFiles = files

	default:
		return nil, fmt.Errorf("unsupported team platform: %s", platform)
	}

	return result, nil
}

// loadMultiAgentSpecTeam loads a team definition from a JSON file.
func loadMultiAgentSpecTeam(path string) (*multiagentspec.Team, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var team multiagentspec.Team
	if err := json.Unmarshal(data, &team); err != nil {
		return nil, err
	}

	return &team, nil
}

// filterTeamAgents returns only the agents that are in the team's agent list.
func filterTeamAgents(allAgents []*agents.Agent, teamAgentNames []string) []*multiagentspec.Agent {
	nameSet := make(map[string]bool)
	for _, name := range teamAgentNames {
		nameSet[name] = true
	}

	var result []*multiagentspec.Agent
	for _, agent := range allAgents {
		if nameSet[agent.Name] {
			result = append(result, agent)
		}
	}
	return result
}

// generateClaudeTeam generates Claude Code team files.
func generateClaudeTeam(specTeam *multiagentspec.Team, agents []*multiagentspec.Agent, outputDir string) ([]string, error) {
	var files []string

	// Convert to assistantkit Team for compatibility
	akTeam := &core.Team{
		Name:        specTeam.Name,
		Description: specTeam.Description,
		Agents:      specTeam.Agents,
		Version:     specTeam.Version,
	}

	// Map workflow type to process
	if specTeam.Workflow != nil {
		switch specTeam.Workflow.Type {
		case multiagentspec.WorkflowChain:
			akTeam.Process = teams.ProcessSequential
		case multiagentspec.WorkflowCrew:
			akTeam.Process = teams.ProcessHierarchical
			akTeam.Manager = specTeam.EffectiveLead()
		default:
			akTeam.Process = teams.ProcessParallel
		}
	}

	// Use Claude team adapter
	adapter := &claude.Adapter{}

	// Generate team orchestration file
	teamPath := filepath.Join(outputDir, "team.md")
	if err := adapter.WriteFileWithSpec(akTeam, specTeam, teamPath); err != nil {
		return nil, fmt.Errorf("writing team file: %w", err)
	}
	files = append(files, teamPath)

	// Generate agent files with role information
	agentsDir := filepath.Join(outputDir, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		return nil, err
	}

	for _, agt := range agents {
		path := filepath.Join(agentsDir, agt.Name+".md")
		if err := writeAgentWithRole(agt, specTeam, path); err != nil {
			return nil, fmt.Errorf("writing agent %s: %w", agt.Name, err)
		}
		files = append(files, path)
	}

	// Generate settings.json for Claude Code team mode
	settingsPath := filepath.Join(outputDir, "settings.json")
	if err := writeClaudeTeamSettings(specTeam, settingsPath); err != nil {
		return nil, fmt.Errorf("writing settings: %w", err)
	}
	files = append(files, settingsPath)

	return files, nil
}

// writeAgentWithRole writes an agent file with role information.
func writeAgentWithRole(agent *multiagentspec.Agent, team *multiagentspec.Team, path string) error {
	var buf stringBuilder

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
		buf.WriteString(fmt.Sprintf("tools: [%s]\n", joinStrings(agent.Tools, ", ")))
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
			buf.WriteString(fmt.Sprintf("  can_delegate_to: [%s]\n", joinStrings(agent.Delegation.CanDelegateTo, ", ")))
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

	// Write team context
	if team != nil && team.IsSelfDirected() {
		buf.WriteString("\n## Team Context\n\n")
		buf.WriteString(fmt.Sprintf("You are part of the **%s** team.\n\n", team.Name))

		// Add workflow-specific context
		if team.Workflow != nil {
			switch team.Workflow.Type {
			case multiagentspec.WorkflowCrew:
				lead := team.EffectiveLead()
				if agent.Name == lead {
					buf.WriteString("**You are the lead agent.** Delegate tasks to specialists and coordinate the team.\n")
				} else {
					buf.WriteString(fmt.Sprintf("**You report to:** %s\n", lead))
				}
			case multiagentspec.WorkflowSwarm:
				buf.WriteString("**Workflow:** Self-organizing swarm. Claim tasks from the shared queue.\n")
			case multiagentspec.WorkflowCouncil:
				buf.WriteString("**Workflow:** Peer council. Debate and reach consensus with other agents.\n")
			}
		}
	}

	return os.WriteFile(path, []byte(buf.String()), 0600)
}

// ClaudeTeamSettings represents Claude Code team settings.
type ClaudeTeamSettings struct {
	TeamMode     string `json:"team_mode"`
	TeammateMode string `json:"teammate_mode,omitempty"`
	EnableTeams  bool   `json:"enable_teams,omitempty"`
}

// writeClaudeTeamSettings writes Claude Code team settings.
func writeClaudeTeamSettings(team *multiagentspec.Team, path string) error {
	settings := ClaudeTeamSettings{
		TeamMode:    "team",
		EnableTeams: true,
	}

	// Use in-process mode for self-directed workflows
	if team.IsSelfDirected() {
		settings.TeammateMode = "in-process"
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, append(data, '\n'), 0600)
}

// joinStrings joins strings with a separator.
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// GenerateWithTeams generates platform-specific plugins including team files.
// This extends Generate() to also process team definitions.
func GenerateWithTeams(specsDir, target, outputDir string) (*GenerateResult, error) {
	// First, run normal generation
	result, err := Generate(specsDir, target, outputDir)
	if err != nil {
		return nil, err
	}

	// Then process teams if present
	teamsDir := filepath.Join(specsDir, "teams")
	if _, err := os.Stat(teamsDir); os.IsNotExist(err) {
		return result, nil // No teams directory
	}

	// Read teams directory
	entries, err := os.ReadDir(teamsDir)
	if err != nil {
		return result, nil // Ignore errors reading teams dir
	}

	// Process each team file
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := filepath.Ext(entry.Name())
		if ext != ".json" {
			continue
		}

		teamName := entry.Name()[:len(entry.Name())-5] // Remove .json

		// Generate team files in a subdirectory
		teamOutputDir := filepath.Join(outputDir, "teams", teamName)

		// Determine platform from target
		platform := "claude-code" // default

		_, err := Teams(specsDir, teamName, platform, teamOutputDir)
		if err != nil {
			// Log warning but continue
			fmt.Printf("  Warning: failed to generate team %s: %v\n", teamName, err)
			continue
		}
	}

	return result, nil
}
