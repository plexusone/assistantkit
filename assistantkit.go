// Package assistantkit provides a unified interface for managing configuration files
// across multiple AI coding assistants including Claude Code, Cursor, Windsurf,
// VS Code, OpenAI Codex CLI, Cline, and Roo Code.
//
// Assistant Kit supports multiple configuration types:
//
//   - MCP (Model Context Protocol) server configurations
//   - Hooks (automation/lifecycle callbacks)
//   - Settings (permissions, sandbox, general settings) - coming soon
//   - Rules (team rules, coding guidelines) - coming soon
//   - Memory (CLAUDE.md, .cursorrules, etc.) - coming soon
//
// # MCP Configuration
//
// The mcp subpackage provides adapters for reading, writing, and converting
// MCP server configurations between different AI assistant formats.
//
// Example usage:
//
//	import (
//	    "github.com/plexusone/assistantkit/mcp"
//	    "github.com/plexusone/assistantkit/mcp/claude"
//	    "github.com/plexusone/assistantkit/mcp/vscode"
//	)
//
//	// Read Claude config and write to VS Code format
//	cfg, _ := claude.ReadProjectConfig()
//	vscode.WriteWorkspaceConfig(cfg)
//
//	// Or use dynamic conversion
//	data, _ := mcp.Convert(jsonData, "claude", "vscode")
//
// # Hooks Configuration
//
// The hooks subpackage provides adapters for automation/lifecycle callbacks
// that execute at defined stages of the agent loop.
//
// Example usage:
//
//	import (
//	    "github.com/plexusone/assistantkit/hooks"
//	    "github.com/plexusone/assistantkit/hooks/claude"
//	)
//
//	// Create hooks configuration
//	cfg := hooks.NewConfig()
//	cfg.AddHook(hooks.BeforeCommand, hooks.NewCommandHook("echo 'before'"))
//
//	// Write to Claude format
//	claude.WriteProjectConfig(cfg)
//
//	// Or convert between formats
//	data, _ := hooks.Convert(jsonData, "claude", "cursor")
//
// # Related Projects
//
// Assistant Kit is part of the AgentPlexus family of Go modules:
//   - Assistant Kit - AI coding assistant configuration management
//   - OmniVault - Unified secrets management
//   - OmniLLM - Multi-provider LLM abstraction
//   - OmniSerp - Search engine abstraction
//   - OmniObserve - LLM observability abstraction
package assistantkit

// Version is the current version of Assistant Kit.
const Version = "0.7.0"

// ConfigType represents the type of configuration.
type ConfigType string

const (
	// ConfigTypeMCP represents MCP server configuration.
	ConfigTypeMCP ConfigType = "mcp"

	// ConfigTypeHooks represents hooks/automation configuration.
	ConfigTypeHooks ConfigType = "hooks"

	// ConfigTypeSettings represents general settings configuration.
	ConfigTypeSettings ConfigType = "settings"

	// ConfigTypeRules represents team rules configuration.
	ConfigTypeRules ConfigType = "rules"

	// ConfigTypeMemory represents memory/context configuration.
	ConfigTypeMemory ConfigType = "memory"
)

// SupportedConfigTypes returns a list of configuration types that Assistant Kit supports.
func SupportedConfigTypes() []ConfigType {
	return []ConfigType{
		ConfigTypeMCP,
		ConfigTypeHooks,
		ConfigTypeSettings,
		ConfigTypeRules,
		ConfigTypeMemory,
	}
}

// SupportedTools returns a list of AI coding tools that Assistant Kit supports.
func SupportedTools() []string {
	return []string{
		"claude",   // Claude Code / Claude Desktop
		"cursor",   // Cursor IDE
		"windsurf", // Windsurf (Codeium)
		"vscode",   // VS Code / GitHub Copilot
		"codex",    // OpenAI Codex CLI
		"cline",    // Cline VS Code extension
		"roo",      // Roo Code VS Code extension
		"kiro",     // AWS Kiro CLI
	}
}
