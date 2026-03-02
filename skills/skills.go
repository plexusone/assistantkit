// Package skills provides adapters for AI assistant skill definitions.
//
// Supported tools:
//   - Claude Code: skills/<name>/SKILL.md
//   - OpenAI Codex: skills/<name>/SKILL.md
//
// Example usage:
//
//	package main
//
//	import (
//	    "github.com/plexusone/assistantkit/skills"
//	)
//
//	func main() {
//	    // Create a new skill
//	    skill := skills.NewSkill("version-analysis", "Analyze git history for semantic versioning")
//	    skill.Instructions = "Analyze commits since the last tag..."
//	    skill.AddTrigger("version")
//	    skill.AddTrigger("semver")
//	    skill.AddDependency("git")
//
//	    // Write to Claude format
//	    claudeAdapter, _ := skills.GetAdapter("claude")
//	    claudeAdapter.WriteSkillDir(skill, "./skills")
//
//	    // Write to Codex format
//	    codexAdapter, _ := skills.GetAdapter("codex")
//	    codexAdapter.WriteSkillDir(skill, "./skills")
//	}
package skills

import (
	"github.com/plexusone/assistantkit/skills/core"

	// Import adapters for side-effect registration
	_ "github.com/plexusone/assistantkit/skills/claude"
	_ "github.com/plexusone/assistantkit/skills/codex"
	_ "github.com/plexusone/assistantkit/skills/kiro"
)

// Re-export core types for convenience
type (
	Skill   = core.Skill
	Adapter = core.Adapter
)

// Re-export core functions
var (
	NewSkill           = core.NewSkill
	GetAdapter         = core.GetAdapter
	AdapterNames       = core.AdapterNames
	Convert            = core.Convert
	ReadCanonicalFile  = core.ReadCanonicalFile
	WriteCanonicalFile = core.WriteCanonicalFile
	ReadCanonicalDir   = core.ReadCanonicalDir
	WriteSkillsToDir   = core.WriteSkillsToDir
)

// Re-export error types
type (
	ParseError   = core.ParseError
	MarshalError = core.MarshalError
	ReadError    = core.ReadError
	WriteError   = core.WriteError
)
