// Package validation provides configuration types for release validation areas.
// Validation areas represent departments or areas of responsibility in the
// release process (e.g., QA, Documentation, Release Management, Security).
//
// Validation areas are platform-agnostic and can be converted to tool-specific
// formats using adapters:
//
//   - Claude Code: Sub-agents (agents/*.md)
//   - Gemini CLI: Commands or prompts (future)
//   - Codex: Prompts (future)
//
// Example usage:
//
//	import (
//	    "github.com/plexusone/assistantkit/validation"
//	    _ "github.com/plexusone/assistantkit/validation/claude" // Register Claude adapter
//	)
//
//	// Read canonical validation area
//	area, err := validation.ReadCanonicalFile("validation/qa.json")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Write as Claude agent
//	adapter, _ := validation.GetAdapter("claude")
//	err = adapter.WriteFile(area, "agents/qa-validator.md")
package validation

import (
	"github.com/plexusone/assistantkit/validation/core"
)

// ValidationArea is the canonical validation area type.
type ValidationArea = core.ValidationArea

// Check is the canonical check type.
type Check = core.Check

// CheckStatus represents the result of a check.
type CheckStatus = core.CheckStatus

// Status constants
const (
	StatusGo   = core.StatusGo
	StatusNoGo = core.StatusNoGo
	StatusWarn = core.StatusWarn
	StatusSkip = core.StatusSkip
)

// Predefined validation areas
var (
	AreaQA            = core.AreaQA
	AreaDocumentation = core.AreaDocumentation
	AreaRelease       = core.AreaRelease
	AreaSecurity      = core.AreaSecurity
)

// NewValidationArea creates a new ValidationArea.
func NewValidationArea(name, description string) *ValidationArea {
	return core.NewValidationArea(name, description)
}

// Adapter is the adapter interface.
type Adapter = core.Adapter

// Register adds an adapter to the default registry.
func Register(adapter Adapter) {
	core.Register(adapter)
}

// GetAdapter returns an adapter by name.
func GetAdapter(name string) (Adapter, bool) {
	return core.GetAdapter(name)
}

// AdapterNames returns all registered adapter names.
func AdapterNames() []string {
	return core.AdapterNames()
}

// ReadCanonicalFile reads a canonical validation-area.json file.
func ReadCanonicalFile(path string) (*ValidationArea, error) {
	return core.ReadCanonicalFile(path)
}

// WriteCanonicalFile writes a canonical validation-area.json file.
func WriteCanonicalFile(area *ValidationArea, path string) error {
	return core.WriteCanonicalFile(area, path)
}

// ReadCanonicalDir reads all validation-area.json files from a directory.
func ReadCanonicalDir(dir string) ([]*ValidationArea, error) {
	return core.ReadCanonicalDir(dir)
}

// WriteAreasToDir writes validation areas to a directory using the specified adapter.
func WriteAreasToDir(areas []*ValidationArea, dir string, adapterName string) error {
	return core.WriteAreasToDir(areas, dir, adapterName)
}
