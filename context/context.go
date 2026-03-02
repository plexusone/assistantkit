// Package context provides a tool-agnostic system for managing project context
// that can be converted to various AI assistant formats (CLAUDE.md, .cursorrules, etc.).
//
// The context package uses a JSON-based canonical format (CONTEXT.json) that can be
// validated against a JSON Schema and converted to tool-specific formats.
//
// # Usage
//
// Create a CONTEXT.json file in your project root:
//
//	{
//	  "$schema": "https://github.com/plexusone/assistantkit/context/schema/project-context.schema.json",
//	  "name": "my-project",
//	  "description": "A brief description",
//	  "language": "go",
//	  "commands": {
//	    "build": "go build ./...",
//	    "test": "go test ./..."
//	  }
//	}
//
// Then convert to tool-specific formats:
//
//	import (
//	    "github.com/plexusone/assistantkit/context"
//	    _ "github.com/plexusone/assistantkit/context/claude" // Register converter
//	)
//
//	ctx, _ := context.ReadFile("CONTEXT.json")
//	context.WriteFile(ctx, "claude", "CLAUDE.md")
//
// # Supported Formats
//
//   - claude: CLAUDE.md for Claude Code
//   - (future) cursor: .cursorrules for Cursor IDE
//   - (future) copilot: .github/copilot-instructions.md for GitHub Copilot
package context

import (
	"github.com/plexusone/assistantkit/context/core"
)

// Re-export core types for convenience.
type (
	// Context is the canonical project context.
	Context = core.Context

	// Architecture describes the project architecture.
	Architecture = core.Architecture

	// Diagram represents an architecture diagram.
	Diagram = core.Diagram

	// Package describes a package or module.
	Package = core.Package

	// Dependencies describes project dependencies.
	Dependencies = core.Dependencies

	// Dependency represents a single dependency.
	Dependency = core.Dependency

	// Testing describes the testing strategy.
	Testing = core.Testing

	// Files describes important files.
	Files = core.Files

	// Note represents an additional note.
	Note = core.Note

	// Related represents a related project or resource.
	Related = core.Related

	// Converter is the interface for format converters.
	Converter = core.Converter

	// ParseError represents a parsing error.
	ParseError = core.ParseError

	// WriteError represents a write error.
	WriteError = core.WriteError

	// ConversionError represents a conversion error.
	ConversionError = core.ConversionError
)

// Re-export core errors.
var (
	ErrEmptyContext      = core.ErrEmptyContext
	ErrMissingName       = core.ErrMissingName
	ErrUnsupportedFormat = core.ErrUnsupportedFormat
)

// NewContext creates a new empty Context with the given name.
func NewContext(name string) *Context {
	return core.NewContext(name)
}

// ReadFile reads a Context from a JSON file.
func ReadFile(path string) (*Context, error) {
	return core.ReadFile(path)
}

// Parse parses JSON data into a Context.
func Parse(data []byte) (*Context, error) {
	return core.Parse(data)
}

// Convert converts a context to a specific format.
func Convert(ctx *Context, format string) ([]byte, error) {
	return core.ConvertTo(ctx, format)
}

// WriteFile writes a context to a file in a specific format.
func WriteFile(ctx *Context, format, path string) error {
	return core.DefaultRegistry.WriteFile(ctx, format, path)
}

// GenerateAll generates all supported formats in the given directory.
func GenerateAll(ctx *Context, dir string) error {
	return core.DefaultRegistry.GenerateAll(ctx, dir)
}

// RegisterConverter registers a converter with the default registry.
func RegisterConverter(converter Converter) {
	core.RegisterConverter(converter)
}

// GetConverter returns a converter by name.
func GetConverter(name string) (Converter, bool) {
	return core.GetConverter(name)
}

// ConverterNames returns the names of all registered converters.
func ConverterNames() []string {
	return core.DefaultRegistry.Names()
}
