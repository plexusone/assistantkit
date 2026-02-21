package core

import (
	multiagentspec "github.com/agentplexus/multi-agent-spec/sdk/go"
)

// SelfDirectedTeam wraps a multi-agent-spec Team with assistantkit-specific functionality.
type SelfDirectedTeam struct {
	// Spec is the underlying multi-agent-spec team definition.
	Spec *multiagentspec.Team

	// Agents is the list of agents with their full definitions.
	Agents []*multiagentspec.Agent
}

// NewSelfDirectedTeam creates a new SelfDirectedTeam from a multi-agent-spec Team.
func NewSelfDirectedTeam(spec *multiagentspec.Team) *SelfDirectedTeam {
	return &SelfDirectedTeam{
		Spec:   spec,
		Agents: make([]*multiagentspec.Agent, 0),
	}
}

// AddAgent adds an agent to the team.
func (t *SelfDirectedTeam) AddAgent(agent *multiagentspec.Agent) *SelfDirectedTeam {
	t.Agents = append(t.Agents, agent)
	return t
}

// WorkflowType returns the workflow type of this team.
func (t *SelfDirectedTeam) WorkflowType() multiagentspec.WorkflowType {
	if t.Spec == nil || t.Spec.Workflow == nil {
		return multiagentspec.WorkflowGraph
	}
	return t.Spec.Workflow.Type
}

// IsSelfDirected returns true if this team uses a self-directed workflow.
func (t *SelfDirectedTeam) IsSelfDirected() bool {
	return t.Spec != nil && t.Spec.IsSelfDirected()
}

// IsDeterministic returns true if this team uses a deterministic workflow.
func (t *SelfDirectedTeam) IsDeterministic() bool {
	return t.Spec != nil && t.Spec.IsDeterministic()
}

// Lead returns the lead agent name for crew workflows.
func (t *SelfDirectedTeam) Lead() string {
	if t.Spec == nil {
		return ""
	}
	return t.Spec.EffectiveLead()
}

// GetAgent returns the agent with the given name, or nil if not found.
func (t *SelfDirectedTeam) GetAgent(name string) *multiagentspec.Agent {
	for _, agent := range t.Agents {
		if agent.Name == name {
			return agent
		}
	}
	return nil
}

// Validate validates the team configuration.
func (t *SelfDirectedTeam) Validate() error {
	if t.Spec == nil {
		return &ValidationError{Field: "spec", Message: "team spec is required"}
	}
	return t.Spec.Validate()
}

// ToAssistantKitTeam converts a SelfDirectedTeam to an assistantkit Team.
// This allows using assistantkit's orchestration generation with self-directed teams.
func (t *SelfDirectedTeam) ToAssistantKitTeam() *Team {
	if t.Spec == nil {
		return nil
	}

	// Map workflow type to process
	process := ProcessParallel // default for self-directed
	switch t.WorkflowType() {
	case multiagentspec.WorkflowChain:
		process = ProcessSequential
	case multiagentspec.WorkflowScatter, multiagentspec.WorkflowGraph:
		process = ProcessParallel
	case multiagentspec.WorkflowCrew:
		process = ProcessHierarchical
	case multiagentspec.WorkflowSwarm, multiagentspec.WorkflowCouncil:
		process = ProcessParallel // Self-organizing
	}

	team := &Team{
		Name:        t.Spec.Name,
		Description: t.Spec.Description,
		Process:     process,
		Manager:     t.Lead(),
		Agents:      t.Spec.Agents,
		Version:     t.Spec.Version,
	}

	// Convert workflow steps to tasks if present
	if t.Spec.Workflow != nil && len(t.Spec.Workflow.Steps) > 0 {
		for _, step := range t.Spec.Workflow.Steps {
			task := Task{
				Name:      step.Name,
				Agent:     step.Agent,
				DependsOn: step.DependsOn,
			}
			team.Tasks = append(team.Tasks, task)
		}
	}

	return team
}

// FromMultiAgentSpec creates a SelfDirectedTeam from multi-agent-spec types.
func FromMultiAgentSpec(team *multiagentspec.Team, agents []*multiagentspec.Agent) *SelfDirectedTeam {
	sdt := NewSelfDirectedTeam(team)
	for _, agent := range agents {
		sdt.AddAgent(agent)
	}
	return sdt
}

// SelfDirectedConfig holds configuration for generating self-directed team output.
type SelfDirectedConfig struct {
	// TeamMode specifies the Claude Code team mode ("subagent" or "team").
	TeamMode string

	// TeammateMode specifies how teammates are displayed ("in-process", "tmux", "auto").
	TeammateMode string

	// EnableTeams enables Claude Code experimental agent teams.
	EnableTeams bool

	// OutputDir is the directory for generated files.
	OutputDir string

	// AgentSpecsPath is the path to agent spec files.
	AgentSpecsPath string
}

// DefaultSelfDirectedConfig returns the default configuration.
func DefaultSelfDirectedConfig() *SelfDirectedConfig {
	return &SelfDirectedConfig{
		TeamMode:     "team",
		TeammateMode: "auto",
		EnableTeams:  true,
	}
}
