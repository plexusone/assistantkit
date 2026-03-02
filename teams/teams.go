// Package teams provides multi-agent team orchestration definitions.
//
// Teams coordinate multiple agents to accomplish complex workflows.
// Each team defines:
//   - Process type (sequential, parallel, hierarchical)
//   - Tasks with agent assignments
//   - Subtasks with Go/No-Go status tracking
//   - Task dependencies for execution ordering
//
// Example usage:
//
//	package main
//
//	import (
//	    "fmt"
//	    "github.com/plexusone/assistantkit/teams"
//	)
//
//	func main() {
//	    // Read a team definition
//	    team, err := teams.ReadTeamFile("release-team.yaml")
//	    if err != nil {
//	        panic(err)
//	    }
//
//	    // Validate the team
//	    if err := team.Validate(); err != nil {
//	        panic(err)
//	    }
//
//	    // Get tasks in dependency order
//	    sorted, _ := team.TopologicalSort()
//	    for _, task := range sorted {
//	        fmt.Printf("Task: %s (agent: %s)\n", task.Name, task.Agent)
//	        for _, st := range task.Subtasks {
//	            fmt.Printf("  - %s: %s\n", st.Name, st.Type())
//	        }
//	    }
//	}
package teams

import (
	"github.com/plexusone/assistantkit/teams/core"

	// Import adapters for side-effect registration
	_ "github.com/plexusone/assistantkit/teams/claude"
)

// Re-export core types for convenience.
type (
	Team    = core.Team
	Task    = core.Task
	Subtask = core.Subtask
	Process = core.Process
	Status  = core.Status
	Adapter = core.Adapter

	// Result types
	TeamResult    = core.TeamResult
	TaskResult    = core.TaskResult
	SubtaskResult = core.SubtaskResult

	// Orchestration
	OrchestrationConfig = core.OrchestrationConfig

	// Self-directed workflow types
	SelfDirectedTeam   = core.SelfDirectedTeam
	SelfDirectedConfig = core.SelfDirectedConfig
)

// Re-export process constants.
const (
	ProcessSequential   = core.ProcessSequential
	ProcessParallel     = core.ProcessParallel
	ProcessHierarchical = core.ProcessHierarchical
)

// Re-export status constants.
const (
	StatusGo      = core.StatusGo
	StatusNoGo    = core.StatusNoGo
	StatusWarn    = core.StatusWarn
	StatusSkip    = core.StatusSkip
	StatusPending = core.StatusPending
	StatusRunning = core.StatusRunning
)

// Re-export core functions.
var (
	NewTeam      = core.NewTeam
	NewTask      = core.NewTask
	NewSubtask   = core.NewSubtask
	GetAdapter   = core.GetAdapter
	Register     = core.Register
	AdapterNames = core.AdapterNames

	// File I/O
	ReadTeamFile  = core.ReadTeamFile
	WriteTeamFile = core.WriteTeamFile
	WriteTeamJSON = core.WriteTeamJSON
	ReadTeamDir   = core.ReadTeamDir
	ParseYAML     = core.ParseYAML
	ParseJSON     = core.ParseJSON

	// Status computation
	ComputeTaskStatus = core.ComputeTaskStatus
	ComputeTeamStatus = core.ComputeTeamStatus

	// Self-directed workflow functions
	NewSelfDirectedTeam       = core.NewSelfDirectedTeam
	FromMultiAgentSpec        = core.FromMultiAgentSpec
	DefaultSelfDirectedConfig = core.DefaultSelfDirectedConfig
)

// Re-export error types.
type (
	ParseError      = core.ParseError
	MarshalError    = core.MarshalError
	ReadError       = core.ReadError
	WriteError      = core.WriteError
	ValidationError = core.ValidationError
)
