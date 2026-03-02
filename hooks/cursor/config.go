// Package cursor provides an adapter for Cursor IDE hooks configuration.
//
// Cursor hooks are configured in hooks.json files:
//   - Project: .cursor/hooks.json
//   - User: ~/.cursor/hooks.json
//   - Enterprise: /Library/Application Support/Cursor/hooks.json (macOS)
//
// Cursor hook events:
//   - beforeShellExecution: Gates shell commands
//   - afterShellExecution: Audits executed commands
//   - beforeMCPExecution: Controls MCP tool usage
//   - afterMCPExecution: Monitors MCP results
//   - beforeReadFile: Controls file access
//   - afterFileEdit: Processes file modifications
//   - beforeSubmitPrompt: Validates prompts
//   - afterAgentResponse: Observes assistant messages
//   - afterAgentThought: Tracks reasoning blocks
//   - stop: Handles agent loop termination
//   - beforeTabFileRead: Controls Tab file access
//   - afterTabFileEdit: Processes Tab edits
package cursor

import "github.com/plexusone/assistantkit/hooks/core"

// CursorEvent represents Cursor-specific hook event names.
type CursorEvent string

const (
	BeforeShellExecution CursorEvent = "beforeShellExecution"
	AfterShellExecution  CursorEvent = "afterShellExecution"
	BeforeMCPExecution   CursorEvent = "beforeMCPExecution"
	AfterMCPExecution    CursorEvent = "afterMCPExecution"
	BeforeReadFile       CursorEvent = "beforeReadFile"
	AfterFileEdit        CursorEvent = "afterFileEdit"
	BeforeSubmitPrompt   CursorEvent = "beforeSubmitPrompt"
	AfterAgentResponse   CursorEvent = "afterAgentResponse"
	AfterAgentThought    CursorEvent = "afterAgentThought"
	Stop                 CursorEvent = "stop"
	BeforeTabFileRead    CursorEvent = "beforeTabFileRead"
	AfterTabFileEdit     CursorEvent = "afterTabFileEdit"
)

// Config represents Cursor's hooks.json configuration.
type Config struct {
	// Version is the configuration version.
	Version int `json:"version"`

	// Hooks maps event names to hook definitions.
	Hooks map[CursorEvent][]Hook `json:"hooks"`
}

// Hook represents a single Cursor hook definition.
type Hook struct {
	// Command is the shell command to execute.
	Command string `json:"command"`
}

// NewConfig creates a new empty Cursor hooks config.
func NewConfig() *Config {
	return &Config{
		Version: 1,
		Hooks:   make(map[CursorEvent][]Hook),
	}
}

// eventMapping maps canonical events to Cursor events.
var eventMapping = map[core.Event]CursorEvent{
	core.BeforeFileRead: BeforeReadFile,
	core.AfterFileWrite: AfterFileEdit,
	core.BeforeCommand:  BeforeShellExecution,
	core.AfterCommand:   AfterShellExecution,
	core.BeforeMCP:      BeforeMCPExecution,
	core.AfterMCP:       AfterMCPExecution,
	core.BeforePrompt:   BeforeSubmitPrompt,
	core.OnStop:         Stop,
	core.AfterResponse:  AfterAgentResponse,
	core.AfterThought:   AfterAgentThought,
	core.BeforeTabRead:  BeforeTabFileRead,
	core.AfterTabEdit:   AfterTabFileEdit,
}

// reverseEventMapping maps Cursor events back to canonical events.
var reverseEventMapping = map[CursorEvent]core.Event{
	BeforeShellExecution: core.BeforeCommand,
	AfterShellExecution:  core.AfterCommand,
	BeforeMCPExecution:   core.BeforeMCP,
	AfterMCPExecution:    core.AfterMCP,
	BeforeReadFile:       core.BeforeFileRead,
	AfterFileEdit:        core.AfterFileWrite,
	BeforeSubmitPrompt:   core.BeforePrompt,
	AfterAgentResponse:   core.AfterResponse,
	AfterAgentThought:    core.AfterThought,
	Stop:                 core.OnStop,
	BeforeTabFileRead:    core.BeforeTabRead,
	AfterTabFileEdit:     core.AfterTabEdit,
}
