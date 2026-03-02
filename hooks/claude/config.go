// Package claude provides an adapter for Claude Code hooks configuration.
//
// Claude hooks are configured in the "hooks" section of settings.json files:
//   - Project: .claude/settings.json
//   - User: ~/.claude/settings.json
//   - Local: .claude/settings.local.json
//   - Enterprise: /Library/Application Support/ClaudeCode/managed-settings.json
//
// Claude hook events:
//   - PreToolUse: Before tool execution (can block with exit code 2)
//   - PostToolUse: After tool execution
//   - PermissionRequest: When permission dialog is shown
//   - UserPromptSubmit: When user submits a prompt
//   - Stop: When agent stops
//   - SessionStart: At session start
//   - SessionEnd: At session end
//   - Notification: When notifications are sent
//   - PreCompact: Before context compaction
//   - SubagentStop: When subagent stops
package claude

import "github.com/plexusone/assistantkit/hooks/core"

// ClaudeEvent represents Claude-specific hook event names.
type ClaudeEvent string

const (
	PreToolUse        ClaudeEvent = "PreToolUse"
	PostToolUse       ClaudeEvent = "PostToolUse"
	PermissionRequest ClaudeEvent = "PermissionRequest"
	UserPromptSubmit  ClaudeEvent = "UserPromptSubmit"
	Stop              ClaudeEvent = "Stop"
	SessionStart      ClaudeEvent = "SessionStart"
	SessionEnd        ClaudeEvent = "SessionEnd"
	Notification      ClaudeEvent = "Notification"
	PreCompact        ClaudeEvent = "PreCompact"
	SubagentStop      ClaudeEvent = "SubagentStop"
)

// Config represents the hooks section of Claude's settings.json.
type Config struct {
	Hooks                 map[ClaudeEvent][]HookEntry `json:"hooks,omitempty"`
	DisableAllHooks       bool                        `json:"disableAllHooks,omitempty"`
	AllowManagedHooksOnly bool                        `json:"allowManagedHooksOnly,omitempty"`
}

// HookEntry represents a Claude hook entry with matcher and hooks.
type HookEntry struct {
	// Matcher filters which tools trigger this hook.
	// Examples: "Bash", "Write", "Edit", "Read", "Bash|Write"
	Matcher string `json:"matcher,omitempty"`

	// Hooks is the list of hooks to execute.
	Hooks []Hook `json:"hooks"`
}

// Hook represents a single Claude hook definition.
type Hook struct {
	// Type is "command" or "prompt".
	Type string `json:"type"`

	// Command is the shell command to execute (for command type).
	Command string `json:"command,omitempty"`

	// Prompt is the LLM prompt for context-aware decisions (for prompt type).
	Prompt string `json:"prompt,omitempty"`

	// Timeout in seconds for hook execution.
	Timeout int `json:"timeout,omitempty"`
}

// NewConfig creates a new empty Claude hooks config.
func NewConfig() *Config {
	return &Config{
		Hooks: make(map[ClaudeEvent][]HookEntry),
	}
}

// eventMapping maps canonical events to Claude events.
var eventMapping = map[core.Event]ClaudeEvent{
	core.BeforeFileRead:  PreToolUse,  // with matcher "Read"
	core.AfterFileRead:   PostToolUse, // with matcher "Read"
	core.BeforeFileWrite: PreToolUse,  // with matcher "Write|Edit"
	core.AfterFileWrite:  PostToolUse, // with matcher "Write|Edit"
	core.BeforeCommand:   PreToolUse,  // with matcher "Bash"
	core.AfterCommand:    PostToolUse, // with matcher "Bash"
	core.BeforeMCP:       PreToolUse,  // with MCP tool matcher
	core.AfterMCP:        PostToolUse, // with MCP tool matcher
	core.BeforePrompt:    UserPromptSubmit,
	core.OnStop:          Stop,
	core.OnSessionStart:  SessionStart,
	core.OnSessionEnd:    SessionEnd,
	core.OnPermission:    PermissionRequest,
	core.OnNotification:  Notification,
	core.BeforeCompact:   PreCompact,
	core.OnSubagentStop:  SubagentStop,
}

// reverseEventMapping maps Claude events back to canonical events.
// Note: PreToolUse/PostToolUse need matcher context to determine exact canonical event.
var reverseEventMapping = map[ClaudeEvent]core.Event{
	UserPromptSubmit:  core.BeforePrompt,
	Stop:              core.OnStop,
	SessionStart:      core.OnSessionStart,
	SessionEnd:        core.OnSessionEnd,
	PermissionRequest: core.OnPermission,
	Notification:      core.OnNotification,
	PreCompact:        core.BeforeCompact,
	SubagentStop:      core.OnSubagentStop,
}

// matcherToCanonicalEvent maps matchers to canonical events for Pre/PostToolUse.
var matcherToCanonicalEventBefore = map[string]core.Event{
	"Read":       core.BeforeFileRead,
	"Write":      core.BeforeFileWrite,
	"Edit":       core.BeforeFileWrite,
	"Write|Edit": core.BeforeFileWrite,
	"Bash":       core.BeforeCommand,
}

var matcherToCanonicalEventAfter = map[string]core.Event{
	"Read":       core.AfterFileRead,
	"Write":      core.AfterFileWrite,
	"Edit":       core.AfterFileWrite,
	"Write|Edit": core.AfterFileWrite,
	"Bash":       core.AfterCommand,
}

// canonicalEventToMatcher maps canonical events to Claude matchers.
var canonicalEventToMatcher = map[core.Event]string{
	core.BeforeFileRead:  "Read",
	core.AfterFileRead:   "Read",
	core.BeforeFileWrite: "Write|Edit",
	core.AfterFileWrite:  "Write|Edit",
	core.BeforeCommand:   "Bash",
	core.AfterCommand:    "Bash",
}
