// Package mcp provides a unified interface for managing MCP (Model Context Protocol)
// server configurations across multiple AI coding assistants.
//
// Supported tools:
//   - Claude Code / Claude Desktop (.mcp.json)
//   - Cursor IDE (~/.cursor/mcp.json)
//   - Windsurf / Codeium (~/.codeium/windsurf/mcp_config.json)
//   - VS Code / GitHub Copilot (.vscode/mcp.json)
//   - OpenAI Codex CLI (~/.codex/config.toml)
//   - Cline VS Code extension (cline_mcp_settings.json)
//   - Roo Code VS Code extension (mcp_settings.json)
//   - AWS Kiro CLI (.kiro/settings/mcp.json)
//
// The package provides:
//   - A canonical Config type that represents MCP configuration
//   - Adapters for reading/writing tool-specific formats
//   - Conversion between different tool formats
//
// Example usage:
//
//	// Read Claude config
//	cfg, err := claude.ReadProjectConfig()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Write to VS Code format
//	err = vscode.WriteWorkspaceConfig(cfg)
//
//	// Or use the adapter registry for dynamic conversion
//	data, err := mcp.Convert(jsonData, "claude", "vscode")
package mcp

import (
	"github.com/plexusone/assistantkit/mcp/core"

	// Import adapters to register them
	_ "github.com/plexusone/assistantkit/mcp/claude"
	_ "github.com/plexusone/assistantkit/mcp/cline"
	_ "github.com/plexusone/assistantkit/mcp/codex"
	_ "github.com/plexusone/assistantkit/mcp/cursor"
	_ "github.com/plexusone/assistantkit/mcp/kiro"
	_ "github.com/plexusone/assistantkit/mcp/roo"
	_ "github.com/plexusone/assistantkit/mcp/vscode"
	_ "github.com/plexusone/assistantkit/mcp/windsurf"
)

// Re-export core types for convenience
type (
	// Config is the canonical MCP configuration.
	Config = core.Config

	// Server is the canonical MCP server configuration.
	Server = core.Server

	// InputVariable is a placeholder for sensitive values.
	InputVariable = core.InputVariable

	// TransportType is the communication protocol.
	TransportType = core.TransportType

	// Adapter is the interface for tool-specific adapters.
	Adapter = core.Adapter
)

// Transport type constants
const (
	TransportStdio = core.TransportStdio
	TransportHTTP  = core.TransportHTTP
	TransportSSE   = core.TransportSSE
)

// NewConfig creates a new empty configuration.
func NewConfig() *Config {
	return core.NewConfig()
}

// GetAdapter returns an adapter by name from the default registry.
// Supported names: "claude", "cursor", "windsurf", "vscode", "codex", "cline", "roo", "kiro"
func GetAdapter(name string) (Adapter, bool) {
	return core.GetAdapter(name)
}

// Convert converts configuration data between formats.
// Example: Convert(data, "claude", "vscode")
func Convert(data []byte, from, to string) ([]byte, error) {
	return core.Convert(data, from, to)
}

// AdapterNames returns the names of all registered adapters.
func AdapterNames() []string {
	return core.DefaultRegistry.Names()
}

// SupportedTools returns a list of supported AI coding tools.
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
