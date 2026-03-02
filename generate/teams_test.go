package generate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	multiagentspec "github.com/plexusone/multi-agent-spec/sdk/go"
)

func TestTeams_CrewWorkflow(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "teams-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create specs directory structure
	specsDir := filepath.Join(tmpDir, "specs")
	teamsDir := filepath.Join(specsDir, "teams")
	agentsDir := filepath.Join(specsDir, "agents")

	if err := os.MkdirAll(teamsDir, 0755); err != nil {
		t.Fatalf("Failed to create teams dir: %v", err)
	}
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents dir: %v", err)
	}

	// Create team definition
	team := multiagentspec.Team{
		Name:    "test-crew",
		Version: "1.0.0",
		Agents:  []string{"lead", "specialist-a"},
		Workflow: &multiagentspec.Workflow{
			Type: multiagentspec.WorkflowCrew,
		},
		Collaboration: &multiagentspec.CollaborationConfig{
			Lead:        "lead",
			Specialists: []string{"specialist-a"},
		},
	}

	teamData, _ := json.MarshalIndent(team, "", "  ")
	if err := os.WriteFile(filepath.Join(teamsDir, "test-crew.json"), teamData, 0644); err != nil {
		t.Fatalf("Failed to write team file: %v", err)
	}

	// Create agent definitions
	leadAgent := `---
name: lead
description: The lead agent
model: sonnet
tools: [Read, Write]
role: Team Lead
goal: Coordinate the team
---

# Instructions

Lead the team.
`
	if err := os.WriteFile(filepath.Join(agentsDir, "lead.md"), []byte(leadAgent), 0644); err != nil {
		t.Fatalf("Failed to write lead agent: %v", err)
	}

	specialistAgent := `---
name: specialist-a
description: A specialist agent
model: haiku
tools: [Read]
role: Specialist
goal: Do specialized work
---

# Instructions

Do specialized work.
`
	if err := os.WriteFile(filepath.Join(agentsDir, "specialist-a.md"), []byte(specialistAgent), 0644); err != nil {
		t.Fatalf("Failed to write specialist agent: %v", err)
	}

	// Run team generation
	outputDir := filepath.Join(tmpDir, "output")
	result, err := Teams(specsDir, "test-crew", "claude-code", outputDir)
	if err != nil {
		t.Fatalf("Teams() error = %v", err)
	}

	// Verify result
	if result.TeamName != "test-crew" {
		t.Errorf("TeamName = %q, want %q", result.TeamName, "test-crew")
	}
	if result.WorkflowType != "crew" {
		t.Errorf("WorkflowType = %q, want %q", result.WorkflowType, "crew")
	}
	if result.AgentCount != 2 {
		t.Errorf("AgentCount = %d, want %d", result.AgentCount, 2)
	}

	// Verify generated files exist
	teamFile := filepath.Join(outputDir, "team.md")
	if _, err := os.Stat(teamFile); os.IsNotExist(err) {
		t.Error("team.md not generated")
	}

	settingsFile := filepath.Join(outputDir, "settings.json")
	if _, err := os.Stat(settingsFile); os.IsNotExist(err) {
		t.Error("settings.json not generated")
	}

	// Verify settings content
	settingsData, err := os.ReadFile(settingsFile)
	if err != nil {
		t.Fatalf("Failed to read settings: %v", err)
	}

	var settings ClaudeTeamSettings
	if err := json.Unmarshal(settingsData, &settings); err != nil {
		t.Fatalf("Failed to parse settings: %v", err)
	}

	if settings.TeamMode != "team" {
		t.Errorf("TeamMode = %q, want %q", settings.TeamMode, "team")
	}
	if settings.TeammateMode != "in-process" {
		t.Errorf("TeammateMode = %q, want %q", settings.TeammateMode, "in-process")
	}
	if !settings.EnableTeams {
		t.Error("EnableTeams should be true")
	}
}

func TestTeams_SwarmWorkflow(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "teams-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create specs directory structure
	specsDir := filepath.Join(tmpDir, "specs")
	teamsDir := filepath.Join(specsDir, "teams")
	agentsDir := filepath.Join(specsDir, "agents")

	if err := os.MkdirAll(teamsDir, 0755); err != nil {
		t.Fatalf("Failed to create teams dir: %v", err)
	}
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents dir: %v", err)
	}

	// Create team definition
	team := multiagentspec.Team{
		Name:    "test-swarm",
		Version: "1.0.0",
		Agents:  []string{"worker-1", "worker-2"},
		Workflow: &multiagentspec.Workflow{
			Type: multiagentspec.WorkflowSwarm,
		},
		Collaboration: &multiagentspec.CollaborationConfig{
			TaskQueue: true,
		},
		SelfClaim: true,
	}

	teamData, _ := json.MarshalIndent(team, "", "  ")
	if err := os.WriteFile(filepath.Join(teamsDir, "test-swarm.json"), teamData, 0644); err != nil {
		t.Fatalf("Failed to write team file: %v", err)
	}

	// Create agent definitions
	worker1 := `---
name: worker-1
description: Worker 1
model: haiku
---

# Instructions

Claim and complete tasks.
`
	if err := os.WriteFile(filepath.Join(agentsDir, "worker-1.md"), []byte(worker1), 0644); err != nil {
		t.Fatalf("Failed to write worker-1 agent: %v", err)
	}

	worker2 := `---
name: worker-2
description: Worker 2
model: haiku
---

# Instructions

Claim and complete tasks.
`
	if err := os.WriteFile(filepath.Join(agentsDir, "worker-2.md"), []byte(worker2), 0644); err != nil {
		t.Fatalf("Failed to write worker-2 agent: %v", err)
	}

	// Run team generation
	outputDir := filepath.Join(tmpDir, "output")
	result, err := Teams(specsDir, "test-swarm", "claude-code", outputDir)
	if err != nil {
		t.Fatalf("Teams() error = %v", err)
	}

	// Verify result
	if result.WorkflowType != "swarm" {
		t.Errorf("WorkflowType = %q, want %q", result.WorkflowType, "swarm")
	}

	// Verify team.md contains swarm instructions
	teamData2, err := os.ReadFile(filepath.Join(outputDir, "team.md"))
	if err != nil {
		t.Fatalf("Failed to read team.md: %v", err)
	}

	content := string(teamData2)
	if !containsSubstring(content, "Swarm Workflow") {
		t.Error("team.md should contain Swarm Workflow section")
	}
}

func TestTeams_CouncilWorkflow(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "teams-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create specs directory structure
	specsDir := filepath.Join(tmpDir, "specs")
	teamsDir := filepath.Join(specsDir, "teams")
	agentsDir := filepath.Join(specsDir, "agents")

	if err := os.MkdirAll(teamsDir, 0755); err != nil {
		t.Fatalf("Failed to create teams dir: %v", err)
	}
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents dir: %v", err)
	}

	// Create team definition with consensus rules
	team := multiagentspec.Team{
		Name:    "test-council",
		Version: "1.0.0",
		Agents:  []string{"reviewer-a", "reviewer-b"},
		Workflow: &multiagentspec.Workflow{
			Type: multiagentspec.WorkflowCouncil,
		},
		Collaboration: &multiagentspec.CollaborationConfig{
			Consensus: &multiagentspec.ConsensusRules{
				RequiredAgreement: 0.66,
				MaxRounds:         3,
				TieBreaker:        "reviewer-a",
			},
		},
	}

	teamData, _ := json.MarshalIndent(team, "", "  ")
	if err := os.WriteFile(filepath.Join(teamsDir, "test-council.json"), teamData, 0644); err != nil {
		t.Fatalf("Failed to write team file: %v", err)
	}

	// Create agent definitions
	reviewer := `---
name: reviewer-a
description: Reviewer A
model: sonnet
role: Security Reviewer
---

# Instructions

Review for security issues.
`
	if err := os.WriteFile(filepath.Join(agentsDir, "reviewer-a.md"), []byte(reviewer), 0644); err != nil {
		t.Fatalf("Failed to write reviewer-a agent: %v", err)
	}
	if err := os.WriteFile(filepath.Join(agentsDir, "reviewer-b.md"), []byte(reviewer), 0644); err != nil {
		t.Fatalf("Failed to write reviewer-b agent: %v", err)
	}

	// Run team generation
	outputDir := filepath.Join(tmpDir, "output")
	result, err := Teams(specsDir, "test-council", "claude-code", outputDir)
	if err != nil {
		t.Fatalf("Teams() error = %v", err)
	}

	// Verify result
	if result.WorkflowType != "council" {
		t.Errorf("WorkflowType = %q, want %q", result.WorkflowType, "council")
	}

	// Verify team.md contains council instructions
	teamData2, err := os.ReadFile(filepath.Join(outputDir, "team.md"))
	if err != nil {
		t.Fatalf("Failed to read team.md: %v", err)
	}

	content := string(teamData2)
	if !containsSubstring(content, "Council Workflow") {
		t.Error("team.md should contain Council Workflow section")
	}
	if !containsSubstring(content, "66%") {
		t.Error("team.md should contain consensus percentage")
	}
}

func TestFilterTeamAgents(t *testing.T) {
	allAgents := []*multiagentspec.Agent{
		{Name: "agent-a"},
		{Name: "agent-b"},
		{Name: "agent-c"},
		{Name: "agent-d"},
	}

	teamAgentNames := []string{"agent-a", "agent-c"}

	filtered := filterTeamAgents(allAgents, teamAgentNames)

	if len(filtered) != 2 {
		t.Errorf("len(filtered) = %d, want %d", len(filtered), 2)
	}

	names := make(map[string]bool)
	for _, a := range filtered {
		names[a.Name] = true
	}

	if !names["agent-a"] {
		t.Error("Missing agent-a in filtered")
	}
	if !names["agent-c"] {
		t.Error("Missing agent-c in filtered")
	}
	if names["agent-b"] {
		t.Error("agent-b should not be in filtered")
	}
}

func TestJoinStrings(t *testing.T) {
	tests := []struct {
		strs []string
		sep  string
		want string
	}{
		{[]string{"a", "b", "c"}, ", ", "a, b, c"},
		{[]string{"single"}, ", ", "single"},
		{[]string{}, ", ", ""},
		{[]string{"x", "y"}, "-", "x-y"},
	}

	for _, tt := range tests {
		got := joinStrings(tt.strs, tt.sep)
		if got != tt.want {
			t.Errorf("joinStrings(%v, %q) = %q, want %q", tt.strs, tt.sep, got, tt.want)
		}
	}
}

func containsSubstring(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
