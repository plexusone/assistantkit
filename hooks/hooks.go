// Package hooks provides a unified interface for managing hook configurations
// across multiple AI coding assistants.
//
// Hooks are automation callbacks that execute at defined stages of the agent loop.
// They can observe, block, or modify agent behavior.
//
// Supported tools:
//   - Claude Code (.claude/settings.json)
//   - Cursor IDE (.cursor/hooks.json)
//   - Windsurf / Codeium (.windsurf/hooks.json)
//
// The package provides:
//   - A canonical Config type that represents hook configuration
//   - Adapters for reading/writing tool-specific formats
//   - Conversion between different tool formats
//
// Example usage:
//
//	// Read Claude hooks config
//	cfg, err := claude.ReadProjectConfig()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Write to Cursor format
//	err = cursor.WriteProjectConfig(cfg)
//
//	// Or use the adapter registry for dynamic conversion
//	data, err := hooks.Convert(jsonData, "claude", "cursor")
package hooks

import (
	"github.com/plexusone/assistantkit/hooks/core"

	// Import adapters to register them
	_ "github.com/plexusone/assistantkit/hooks/claude"
	_ "github.com/plexusone/assistantkit/hooks/cursor"
	_ "github.com/plexusone/assistantkit/hooks/windsurf"
)

// Re-export core types for convenience
type (
	// Config is the canonical hooks configuration.
	Config = core.Config

	// Hook is a single hook definition.
	Hook = core.Hook

	// HookEntry is a collection of hooks with optional matcher.
	HookEntry = core.HookEntry

	// Event is a hook event type.
	Event = core.Event

	// HookType is the type of hook execution.
	HookType = core.HookType

	// Adapter is the interface for tool-specific adapters.
	Adapter = core.Adapter
)

// Hook type constants
const (
	HookTypeCommand = core.HookTypeCommand
	HookTypePrompt  = core.HookTypePrompt
)

// Event constants - File operations
const (
	BeforeFileRead  = core.BeforeFileRead
	AfterFileRead   = core.AfterFileRead
	BeforeFileWrite = core.BeforeFileWrite
	AfterFileWrite  = core.AfterFileWrite
)

// Event constants - Command operations
const (
	BeforeCommand = core.BeforeCommand
	AfterCommand  = core.AfterCommand
)

// Event constants - MCP operations
const (
	BeforeMCP = core.BeforeMCP
	AfterMCP  = core.AfterMCP
)

// Event constants - Prompt/Lifecycle
const (
	BeforePrompt   = core.BeforePrompt
	OnStop         = core.OnStop
	OnSessionStart = core.OnSessionStart
	OnSessionEnd   = core.OnSessionEnd
)

// Event constants - Tool-specific
const (
	AfterResponse  = core.AfterResponse  // Cursor
	AfterThought   = core.AfterThought   // Cursor
	OnPermission   = core.OnPermission   // Claude
	OnNotification = core.OnNotification // Claude
	BeforeCompact  = core.BeforeCompact  // Claude
	OnSubagentStop = core.OnSubagentStop // Claude
	BeforeTabRead  = core.BeforeTabRead  // Cursor
	AfterTabEdit   = core.AfterTabEdit   // Cursor
)

// NewConfig creates a new empty configuration.
func NewConfig() *Config {
	return core.NewConfig()
}

// NewCommandHook creates a new command-type hook.
func NewCommandHook(command string) Hook {
	return core.NewCommandHook(command)
}

// NewPromptHook creates a new prompt-type hook (Claude-specific).
func NewPromptHook(prompt string) Hook {
	return core.NewPromptHook(prompt)
}

// GetAdapter returns an adapter by name from the default registry.
// Supported names: "claude", "cursor", "windsurf"
func GetAdapter(name string) (Adapter, bool) {
	return core.GetAdapter(name)
}

// Convert converts configuration data between formats.
// Example: Convert(data, "claude", "cursor")
func Convert(data []byte, from, to string) ([]byte, error) {
	return core.Convert(data, from, to)
}

// AdapterNames returns the names of all registered adapters.
func AdapterNames() []string {
	return core.DefaultRegistry.Names()
}

// SupportedTools returns a list of tools that support hooks.
func SupportedTools() []string {
	return []string{
		"claude",   // Claude Code
		"cursor",   // Cursor IDE
		"windsurf", // Windsurf (Codeium)
	}
}

// AllEvents returns all defined canonical events.
func AllEvents() []Event {
	return core.AllEvents()
}
