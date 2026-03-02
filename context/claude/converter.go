// Package claude provides a converter for generating CLAUDE.md files
// from the canonical project context format.
package claude

import (
	"fmt"
	"strings"

	"github.com/plexusone/assistantkit/context/core"
)

const (
	// ConverterName is the identifier for this converter.
	ConverterName = "claude"

	// OutputFile is the default output file name.
	OutputFile = "CLAUDE.md"
)

// Converter implements core.Converter for Claude Code CLAUDE.md files.
type Converter struct {
	core.BaseConverter
}

// NewConverter creates a new Claude converter.
func NewConverter() *Converter {
	return &Converter{
		BaseConverter: core.NewBaseConverter(ConverterName, OutputFile),
	}
}

// Convert converts the context to CLAUDE.md format.
func (c *Converter) Convert(ctx *core.Context) ([]byte, error) {
	if ctx == nil {
		return nil, &core.ConversionError{Format: ConverterName, Err: core.ErrEmptyContext}
	}
	if ctx.Name == "" {
		return nil, &core.ConversionError{Format: ConverterName, Err: core.ErrMissingName}
	}

	var b strings.Builder

	// Header
	b.WriteString(fmt.Sprintf("# %s\n\n", ctx.Name))

	// Description
	if ctx.Description != "" {
		b.WriteString(fmt.Sprintf("%s\n\n", ctx.Description))
	}

	// Version and Language
	if ctx.Version != "" || ctx.Language != "" {
		if ctx.Version != "" && ctx.Language != "" {
			b.WriteString(fmt.Sprintf("**Version:** %s | **Language:** %s\n\n", ctx.Version, ctx.Language))
		} else if ctx.Version != "" {
			b.WriteString(fmt.Sprintf("**Version:** %s\n\n", ctx.Version))
		} else {
			b.WriteString(fmt.Sprintf("**Language:** %s\n\n", ctx.Language))
		}
	}

	// Architecture
	if ctx.Architecture != nil {
		b.WriteString("## Architecture\n\n")
		if ctx.Architecture.Pattern != "" {
			b.WriteString(fmt.Sprintf("**Pattern:** %s\n\n", ctx.Architecture.Pattern))
		}
		if ctx.Architecture.Summary != "" {
			b.WriteString(fmt.Sprintf("%s\n\n", ctx.Architecture.Summary))
		}
		for _, diagram := range ctx.Architecture.Diagrams {
			if diagram.Title != "" {
				b.WriteString(fmt.Sprintf("### %s\n\n", diagram.Title))
			}
			if diagram.Type == "mermaid" {
				b.WriteString("```mermaid\n")
			} else {
				b.WriteString("```\n")
			}
			b.WriteString(diagram.Content)
			b.WriteString("\n```\n\n")
		}
	}

	// Packages
	if len(ctx.Packages) > 0 {
		b.WriteString("## Packages\n\n")
		b.WriteString("| Package | Purpose |\n")
		b.WriteString("|---------|----------|\n")
		for _, pkg := range ctx.Packages {
			b.WriteString(fmt.Sprintf("| `%s` | %s |\n", pkg.Path, pkg.Purpose))
		}
		b.WriteString("\n")
	}

	// Commands
	if len(ctx.Commands) > 0 {
		b.WriteString("## Commands\n\n")
		b.WriteString("```bash\n")
		// Order matters for readability - common commands first
		orderedKeys := []string{"build", "test", "lint", "format", "run"}
		written := make(map[string]bool)
		for _, key := range orderedKeys {
			if cmd, ok := ctx.Commands[key]; ok {
				b.WriteString(fmt.Sprintf("# %s\n%s\n\n", key, cmd))
				written[key] = true
			}
		}
		// Then any additional commands
		for key, cmd := range ctx.Commands {
			if !written[key] {
				b.WriteString(fmt.Sprintf("# %s\n%s\n\n", key, cmd))
			}
		}
		b.WriteString("```\n\n")
	}

	// Conventions
	if len(ctx.Conventions) > 0 {
		b.WriteString("## Conventions\n\n")
		for _, conv := range ctx.Conventions {
			b.WriteString(fmt.Sprintf("- %s\n", conv))
		}
		b.WriteString("\n")
	}

	// Dependencies
	if ctx.Dependencies != nil {
		if len(ctx.Dependencies.Runtime) > 0 || len(ctx.Dependencies.Development) > 0 {
			b.WriteString("## Dependencies\n\n")
			if len(ctx.Dependencies.Runtime) > 0 {
				b.WriteString("### Runtime\n\n")
				for _, dep := range ctx.Dependencies.Runtime {
					if dep.Purpose != "" {
						b.WriteString(fmt.Sprintf("- **%s** - %s\n", dep.Name, dep.Purpose))
					} else {
						b.WriteString(fmt.Sprintf("- %s\n", dep.Name))
					}
				}
				b.WriteString("\n")
			}
			if len(ctx.Dependencies.Development) > 0 {
				b.WriteString("### Development\n\n")
				for _, dep := range ctx.Dependencies.Development {
					if dep.Purpose != "" {
						b.WriteString(fmt.Sprintf("- **%s** - %s\n", dep.Name, dep.Purpose))
					} else {
						b.WriteString(fmt.Sprintf("- %s\n", dep.Name))
					}
				}
				b.WriteString("\n")
			}
		}
	}

	// Testing
	if ctx.Testing != nil {
		b.WriteString("## Testing\n\n")
		if ctx.Testing.Framework != "" {
			b.WriteString(fmt.Sprintf("**Framework:** %s\n\n", ctx.Testing.Framework))
		}
		if ctx.Testing.Coverage != "" {
			b.WriteString(fmt.Sprintf("**Coverage:** %s\n\n", ctx.Testing.Coverage))
		}
		if len(ctx.Testing.Patterns) > 0 {
			b.WriteString("**Patterns:**\n")
			for _, pattern := range ctx.Testing.Patterns {
				b.WriteString(fmt.Sprintf("- %s\n", pattern))
			}
			b.WriteString("\n")
		}
	}

	// Files
	if ctx.Files != nil {
		hasContent := len(ctx.Files.EntryPoints) > 0 || len(ctx.Files.Config) > 0
		if hasContent {
			b.WriteString("## Key Files\n\n")
			if len(ctx.Files.EntryPoints) > 0 {
				b.WriteString("**Entry Points:**\n")
				for _, f := range ctx.Files.EntryPoints {
					b.WriteString(fmt.Sprintf("- `%s`\n", f))
				}
				b.WriteString("\n")
			}
			if len(ctx.Files.Config) > 0 {
				b.WriteString("**Configuration:**\n")
				for _, f := range ctx.Files.Config {
					b.WriteString(fmt.Sprintf("- `%s`\n", f))
				}
				b.WriteString("\n")
			}
		}
	}

	// Notes
	if len(ctx.Notes) > 0 {
		b.WriteString("## Notes\n\n")
		for _, note := range ctx.Notes {
			severity := note.GetSeverity()
			prefix := ""
			switch severity {
			case "warning":
				prefix = "**Warning:** "
			case "critical":
				prefix = "**CRITICAL:** "
			}
			if note.Title != "" {
				b.WriteString(fmt.Sprintf("### %s\n\n%s%s\n\n", note.Title, prefix, note.Content))
			} else {
				b.WriteString(fmt.Sprintf("- %s%s\n", prefix, note.Content))
			}
		}
		b.WriteString("\n")
	}

	// Related
	if len(ctx.Related) > 0 {
		b.WriteString("## Related\n\n")
		for _, rel := range ctx.Related {
			if rel.URL != "" {
				b.WriteString(fmt.Sprintf("- [%s](%s)", rel.Name, rel.URL))
			} else {
				b.WriteString(fmt.Sprintf("- %s", rel.Name))
			}
			if rel.Description != "" {
				b.WriteString(fmt.Sprintf(" - %s", rel.Description))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Footer
	b.WriteString("---\n")
	b.WriteString("*Generated from CONTEXT.json*\n")

	return []byte(b.String()), nil
}

// WriteFile writes the converted context to a file.
func (c *Converter) WriteFile(ctx *core.Context, path string) error {
	data, err := c.Convert(ctx)
	if err != nil {
		return err
	}
	return c.WriteFileWithData(data, path)
}

// init registers the converter with the default registry.
func init() {
	core.RegisterConverter(NewConverter())
}
