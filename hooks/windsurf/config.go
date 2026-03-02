// Package windsurf provides an adapter for Windsurf (Codeium) hooks configuration.
//
// Windsurf hooks are configured in hooks.json files:
//   - Workspace: .windsurf/hooks.json
//   - User: ~/.codeium/windsurf/hooks.json
//   - System: /Library/Application Support/Windsurf/hooks.json (macOS)
//
// Windsurf hook events:
//   - pre_read_code: Before file reads (can block)
//   - post_read_code: After successful reads
//   - pre_write_code: Before code modifications (can block)
//   - post_write_code: After code changes
//   - pre_run_command: Before terminal execution (can block)
//   - post_run_command: After command completion
//   - pre_mcp_tool_use: Before MCP invocation (can block)
//   - post_mcp_tool_use: After MCP success
//   - pre_user_prompt: Before prompt processing (can block)
package windsurf

import "github.com/plexusone/assistantkit/hooks/core"

// WindsurfEvent represents Windsurf-specific hook event names.
type WindsurfEvent string

const (
	PreReadCode    WindsurfEvent = "pre_read_code"
	PostReadCode   WindsurfEvent = "post_read_code"
	PreWriteCode   WindsurfEvent = "pre_write_code"
	PostWriteCode  WindsurfEvent = "post_write_code"
	PreRunCommand  WindsurfEvent = "pre_run_command"
	PostRunCommand WindsurfEvent = "post_run_command"
	PreMCPToolUse  WindsurfEvent = "pre_mcp_tool_use"
	PostMCPToolUse WindsurfEvent = "post_mcp_tool_use"
	PreUserPrompt  WindsurfEvent = "pre_user_prompt"
)

// Config represents Windsurf's hooks.json configuration.
type Config struct {
	// Hooks maps event names to hook definitions.
	Hooks map[WindsurfEvent][]Hook `json:"hooks"`
}

// Hook represents a single Windsurf hook definition.
type Hook struct {
	// Command is the shell command to execute.
	Command string `json:"command"`

	// ShowOutput displays hook output in the Cascade UI.
	ShowOutput bool `json:"show_output,omitempty"`

	// WorkingDirectory is the execution directory (defaults to workspace root).
	WorkingDirectory string `json:"working_directory,omitempty"`
}

// NewConfig creates a new empty Windsurf hooks config.
func NewConfig() *Config {
	return &Config{
		Hooks: make(map[WindsurfEvent][]Hook),
	}
}

// eventMapping maps canonical events to Windsurf events.
var eventMapping = map[core.Event]WindsurfEvent{
	core.BeforeFileRead:  PreReadCode,
	core.AfterFileRead:   PostReadCode,
	core.BeforeFileWrite: PreWriteCode,
	core.AfterFileWrite:  PostWriteCode,
	core.BeforeCommand:   PreRunCommand,
	core.AfterCommand:    PostRunCommand,
	core.BeforeMCP:       PreMCPToolUse,
	core.AfterMCP:        PostMCPToolUse,
	core.BeforePrompt:    PreUserPrompt,
}

// reverseEventMapping maps Windsurf events back to canonical events.
var reverseEventMapping = map[WindsurfEvent]core.Event{
	PreReadCode:    core.BeforeFileRead,
	PostReadCode:   core.AfterFileRead,
	PreWriteCode:   core.BeforeFileWrite,
	PostWriteCode:  core.AfterFileWrite,
	PreRunCommand:  core.BeforeCommand,
	PostRunCommand: core.AfterCommand,
	PreMCPToolUse:  core.BeforeMCP,
	PostMCPToolUse: core.AfterMCP,
	PreUserPrompt:  core.BeforePrompt,
}
