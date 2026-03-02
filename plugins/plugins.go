// Package plugins provides adapters for AI assistant plugin/extension manifests.
//
// Supported tools:
//   - Claude Code: .claude-plugin/plugin.json
//   - Gemini CLI: gemini-extension.json
//
// Example usage:
//
//	package main
//
//	import (
//	    "github.com/plexusone/assistantkit/plugins"
//	    "github.com/plexusone/assistantkit/plugins/core"
//	)
//
//	func main() {
//	    // Create a new plugin
//	    plugin := plugins.NewPlugin("my-plugin", "1.0.0", "My awesome plugin")
//	    plugin.AddDependency("git", "git")
//	    plugin.Commands = "commands"
//	    plugin.Skills = "skills"
//
//	    // Write to Claude format
//	    claudeAdapter, _ := plugins.GetAdapter("claude")
//	    claudeAdapter.WritePlugin(plugin, "./plugins/claude")
//
//	    // Write to Gemini format
//	    geminiAdapter, _ := plugins.GetAdapter("gemini")
//	    geminiAdapter.WritePlugin(plugin, "./plugins/gemini")
//	}
package plugins

import (
	"github.com/plexusone/assistantkit/plugins/core"

	// Import adapters for side-effect registration
	_ "github.com/plexusone/assistantkit/plugins/claude"
	_ "github.com/plexusone/assistantkit/plugins/gemini"
)

// Re-export core types for convenience
type (
	Plugin     = core.Plugin
	Dependency = core.Dependency
	MCPServer  = core.MCPServer
	Adapter    = core.Adapter
)

// Re-export core functions
var (
	NewPlugin          = core.NewPlugin
	GetAdapter         = core.GetAdapter
	AdapterNames       = core.AdapterNames
	Convert            = core.Convert
	ReadCanonicalFile  = core.ReadCanonicalFile
	WriteCanonicalFile = core.WriteCanonicalFile
)

// Re-export error types
type (
	ParseError      = core.ParseError
	MarshalError    = core.MarshalError
	ReadError       = core.ReadError
	WriteError      = core.WriteError
	ValidationError = core.ValidationError
)
