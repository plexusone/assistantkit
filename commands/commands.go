// Package commands provides adapters for AI assistant command/prompt definitions.
//
// Supported tools:
//   - Claude Code: commands/*.md (Markdown with YAML frontmatter)
//   - Gemini CLI: commands/*.toml (TOML format)
//   - OpenAI Codex: prompts/*.md (Markdown with YAML frontmatter)
//
// Example usage:
//
//	package main
//
//	import (
//	    "github.com/plexusone/assistantkit/commands"
//	)
//
//	func main() {
//	    // Create a new command
//	    cmd := commands.NewCommand("release", "Execute full release workflow")
//	    cmd.AddRequiredArgument("version", "Semantic version", "v1.2.3")
//	    cmd.AddProcessStep("Run validation checks")
//	    cmd.AddProcessStep("Generate changelog")
//	    cmd.AddProcessStep("Create and push git tag")
//	    cmd.Instructions = "Execute a full release workflow..."
//
//	    // Write to Claude format
//	    claudeAdapter, _ := commands.GetAdapter("claude")
//	    claudeAdapter.WriteFile(cmd, "./commands/release.md")
//
//	    // Write to Gemini format
//	    geminiAdapter, _ := commands.GetAdapter("gemini")
//	    geminiAdapter.WriteFile(cmd, "./commands/release.toml")
//
//	    // Write to Codex format
//	    codexAdapter, _ := commands.GetAdapter("codex")
//	    codexAdapter.WriteFile(cmd, "./prompts/release.md")
//	}
package commands

import (
	"github.com/plexusone/assistantkit/commands/core"

	// Import adapters for side-effect registration
	_ "github.com/plexusone/assistantkit/commands/claude"
	_ "github.com/plexusone/assistantkit/commands/codex"
	_ "github.com/plexusone/assistantkit/commands/gemini"
)

// Re-export core types for convenience
type (
	Command  = core.Command
	Argument = core.Argument
	Example  = core.Example
	Adapter  = core.Adapter
)

// Re-export core functions
var (
	NewCommand         = core.NewCommand
	GetAdapter         = core.GetAdapter
	AdapterNames       = core.AdapterNames
	Convert            = core.Convert
	ReadCanonicalFile  = core.ReadCanonicalFile
	WriteCanonicalFile = core.WriteCanonicalFile
	ReadCanonicalDir   = core.ReadCanonicalDir
	WriteCommandsToDir = core.WriteCommandsToDir
)

// Re-export error types
type (
	ParseError   = core.ParseError
	MarshalError = core.MarshalError
	ReadError    = core.ReadError
	WriteError   = core.WriteError
)
