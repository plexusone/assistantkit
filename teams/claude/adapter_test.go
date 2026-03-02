package claude

import (
	"strings"
	"testing"

	multiagentspec "github.com/plexusone/multi-agent-spec/sdk/go"

	"github.com/plexusone/assistantkit/teams/core"
)

func TestAdapter_Name(t *testing.T) {
	adapter := &Adapter{}
	if adapter.Name() != "claude" {
		t.Errorf("Name() = %q, want %q", adapter.Name(), "claude")
	}
}

func TestAdapter_FileExtension(t *testing.T) {
	adapter := &Adapter{}
	if adapter.FileExtension() != ".md" {
		t.Errorf("FileExtension() = %q, want %q", adapter.FileExtension(), ".md")
	}
}

func TestAdapter_Marshal_Basic(t *testing.T) {
	adapter := &Adapter{}

	team := &core.Team{
		Name:        "test-team",
		Description: "A test team",
		Process:     core.ProcessParallel,
		Agents:      []string{"agent-a", "agent-b"},
	}

	data, err := adapter.Marshal(team)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	content := string(data)

	// Check header
	if !strings.Contains(content, "# Test Team Team") {
		t.Error("Missing team header")
	}

	// Check description
	if !strings.Contains(content, "A test team") {
		t.Error("Missing description")
	}

	// Check workflow type
	if !strings.Contains(content, "**Type:** `graph`") {
		t.Error("Missing workflow type")
	}

	// Check agent roster
	if !strings.Contains(content, "| `agent-a` |") {
		t.Error("Missing agent-a in roster")
	}
	if !strings.Contains(content, "| `agent-b` |") {
		t.Error("Missing agent-b in roster")
	}
}

func TestAdapter_MarshalWithSpec_Crew(t *testing.T) {
	adapter := &Adapter{}

	team := &core.Team{
		Name:    "crew-team",
		Process: core.ProcessHierarchical,
		Manager: "lead-agent",
		Agents:  []string{"lead-agent", "specialist-a", "specialist-b"},
	}

	specTeam := &multiagentspec.Team{
		Name:    "crew-team",
		Version: "1.0.0",
		Agents:  []string{"lead-agent", "specialist-a", "specialist-b"},
		Workflow: &multiagentspec.Workflow{
			Type: multiagentspec.WorkflowCrew,
		},
		Collaboration: &multiagentspec.CollaborationConfig{
			Lead:        "lead-agent",
			Specialists: []string{"specialist-a", "specialist-b"},
		},
		PlanApproval: true,
	}

	data, err := adapter.MarshalWithSpec(team, specTeam)
	if err != nil {
		t.Fatalf("MarshalWithSpec() error = %v", err)
	}

	content := string(data)

	// Check workflow type
	if !strings.Contains(content, "**Type:** `crew`") {
		t.Error("Missing crew workflow type")
	}

	// Check category
	if !strings.Contains(content, "**Category:** self-directed") {
		t.Error("Missing self-directed category")
	}

	// Check lead agent in roster
	if !strings.Contains(content, "| `lead-agent` | Lead |") {
		t.Error("Missing lead agent with Lead role")
	}

	// Check crew workflow section
	if !strings.Contains(content, "### Crew Workflow") {
		t.Error("Missing crew workflow section")
	}

	// Check lead agent mention
	if !strings.Contains(content, "lead-agent") {
		t.Error("Missing lead agent in instructions")
	}

	// Check plan approval
	if !strings.Contains(content, "Plan Approval Required") {
		t.Error("Missing plan approval note")
	}
}

func TestAdapter_MarshalWithSpec_Swarm(t *testing.T) {
	adapter := &Adapter{}

	team := &core.Team{
		Name:   "swarm-team",
		Agents: []string{"agent-1", "agent-2", "agent-3"},
	}

	specTeam := &multiagentspec.Team{
		Name:   "swarm-team",
		Agents: []string{"agent-1", "agent-2", "agent-3"},
		Workflow: &multiagentspec.Workflow{
			Type: multiagentspec.WorkflowSwarm,
		},
		Collaboration: &multiagentspec.CollaborationConfig{
			TaskQueue: true,
		},
		SelfClaim: true,
	}

	data, err := adapter.MarshalWithSpec(team, specTeam)
	if err != nil {
		t.Fatalf("MarshalWithSpec() error = %v", err)
	}

	content := string(data)

	// Check workflow type
	if !strings.Contains(content, "**Type:** `swarm`") {
		t.Error("Missing swarm workflow type")
	}

	// Check swarm workflow section
	if !strings.Contains(content, "### Swarm Workflow") {
		t.Error("Missing swarm workflow section")
	}

	// Check task queue
	if !strings.Contains(content, "Task Queue") {
		t.Error("Missing task queue mention")
	}

	// Check self-claim
	if !strings.Contains(content, "Self-Claim") {
		t.Error("Missing self-claim mention")
	}
}

func TestAdapter_MarshalWithSpec_Council(t *testing.T) {
	adapter := &Adapter{}

	team := &core.Team{
		Name:   "council-team",
		Agents: []string{"reviewer-a", "reviewer-b", "reviewer-c"},
	}

	specTeam := &multiagentspec.Team{
		Name:   "council-team",
		Agents: []string{"reviewer-a", "reviewer-b", "reviewer-c"},
		Workflow: &multiagentspec.Workflow{
			Type: multiagentspec.WorkflowCouncil,
		},
		Collaboration: &multiagentspec.CollaborationConfig{
			Consensus: &multiagentspec.ConsensusRules{
				RequiredAgreement: 0.66,
				MaxRounds:         3,
				TieBreaker:        "reviewer-a",
			},
			Channels: []multiagentspec.Channel{
				{Name: "findings", Type: multiagentspec.ChannelBroadcast, Participants: []string{"*"}},
			},
		},
	}

	data, err := adapter.MarshalWithSpec(team, specTeam)
	if err != nil {
		t.Fatalf("MarshalWithSpec() error = %v", err)
	}

	content := string(data)

	// Check workflow type
	if !strings.Contains(content, "**Type:** `council`") {
		t.Error("Missing council workflow type")
	}

	// Check council workflow section
	if !strings.Contains(content, "### Council Workflow") {
		t.Error("Missing council workflow section")
	}

	// Check consensus rules
	if !strings.Contains(content, "66%") {
		t.Error("Missing required agreement percentage")
	}
	if !strings.Contains(content, "3") {
		t.Error("Missing max rounds")
	}
	if !strings.Contains(content, "reviewer-a") {
		t.Error("Missing tie breaker")
	}

	// Check channels
	if !strings.Contains(content, "findings") {
		t.Error("Missing findings channel")
	}
}

func TestAdapter_Registration(t *testing.T) {
	adapter, ok := core.GetAdapter("claude")
	if !ok {
		t.Fatal("Claude adapter not registered")
	}

	if adapter.Name() != "claude" {
		t.Errorf("Registered adapter name = %q, want %q", adapter.Name(), "claude")
	}
}

func TestToTitle(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"test-team", "Test Team"},
		{"code-review-council", "Code Review Council"},
		{"simple", "Simple"},
		{"", ""},
	}

	for _, tt := range tests {
		got := toTitle(tt.input)
		if got != tt.want {
			t.Errorf("toTitle(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
