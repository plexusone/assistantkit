// Package gemini provides the Gemini CLI validation area adapter.
// It converts ValidationArea definitions to Gemini CLI command format.
package gemini

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/plexusone/assistantkit/validation/core"
)

func init() {
	core.Register(&Adapter{})
}

// Adapter converts between canonical ValidationArea and Gemini CLI command format.
type Adapter struct{}

// Name returns the adapter identifier.
func (a *Adapter) Name() string {
	return "gemini"
}

// FileExtension returns the file extension for Gemini commands.
func (a *Adapter) FileExtension() string {
	return ".toml"
}

// DefaultDir returns the default directory name for Gemini commands.
func (a *Adapter) DefaultDir() string {
	return "commands"
}

// Parse converts Gemini command TOML bytes to canonical ValidationArea.
func (a *Adapter) Parse(data []byte) (*core.ValidationArea, error) {
	// Simple TOML parsing for command files
	content := string(data)
	area := &core.ValidationArea{}

	lines := strings.Split(content, "\n")
	var section string
	var contentBuilder strings.Builder

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Section headers
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.Trim(line, "[]")
			continue
		}

		// Key-value pairs
		if idx := strings.Index(line, "="); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])
			value = strings.Trim(value, "\"'")

			switch section {
			case "command":
				switch key {
				case "name":
					// Remove -validator suffix if present
					area.Name = strings.TrimSuffix(value, "-validator")
				case "description":
					area.Description = value
				}
			case "content":
				switch key {
				case "text":
					contentBuilder.WriteString(value)
				}
			}
		}

		// Multi-line content
		if section == "content" && strings.HasPrefix(line, "text = '''") {
			// Start of multi-line string - handled separately
			continue
		}
	}

	if contentBuilder.Len() > 0 {
		area.Instructions = contentBuilder.String()
	}

	return area, nil
}

// Marshal converts canonical ValidationArea to Gemini command TOML bytes.
func (a *Adapter) Marshal(area *core.ValidationArea) ([]byte, error) {
	var buf bytes.Buffer

	// Generate command name from validation area name
	commandName := area.Name + "-validator"

	// Write command section
	buf.WriteString("[command]\n")
	buf.WriteString(fmt.Sprintf("name = %q\n", commandName))
	buf.WriteString(fmt.Sprintf("description = %q\n", fmt.Sprintf("%s validation for release readiness. %s",
		strings.Title(area.Name), area.Description)))
	buf.WriteString("\n")

	// Write arguments section for target directory
	buf.WriteString("[[arguments]]\n")
	buf.WriteString("name = \"target\"\n")
	buf.WriteString("description = \"Target directory to validate\"\n")
	buf.WriteString("required = false\n")
	buf.WriteString("default = \".\"\n")
	buf.WriteString("\n")

	// Write content section with the prompt
	buf.WriteString("[content]\n")
	buf.WriteString("text = '''\n")

	// Build the validation prompt
	buf.WriteString(fmt.Sprintf("# %s Validator\n\n", strings.Title(strings.ReplaceAll(area.Name, "-", " "))))
	buf.WriteString(fmt.Sprintf("%s\n\n", area.Description))

	// Sign-off criteria
	if area.SignOffCriteria != "" {
		buf.WriteString("## Sign-Off Criteria\n\n")
		buf.WriteString(fmt.Sprintf("%s\n\n", area.SignOffCriteria))
	}

	// Validation checks
	if len(area.Checks) > 0 {
		buf.WriteString("## Validation Checks\n\n")
		for _, check := range area.Checks {
			required := "optional"
			if check.Required {
				required = "required"
			}
			buf.WriteString(fmt.Sprintf("- **%s** (%s)", check.Name, required))
			if check.Description != "" {
				buf.WriteString(fmt.Sprintf(": %s", check.Description))
			}
			buf.WriteString("\n")
			if check.Command != "" {
				buf.WriteString(fmt.Sprintf("  Command: `%s`\n", check.Command))
			}
			if check.Pattern != "" {
				buf.WriteString(fmt.Sprintf("  Pattern: `%s`\n", check.Pattern))
			}
			if check.FilePattern != "" {
				buf.WriteString(fmt.Sprintf("  Files: `%s`\n", check.FilePattern))
			}
		}
		buf.WriteString("\n")
	}

	// Dependencies
	if len(area.Dependencies) > 0 {
		buf.WriteString("## Dependencies\n\n")
		buf.WriteString("Required CLI tools:\n")
		for _, dep := range area.Dependencies {
			buf.WriteString(fmt.Sprintf("- %s\n", dep))
		}
		buf.WriteString("\n")
	}

	// Instructions
	if area.Instructions != "" {
		buf.WriteString("## Instructions\n\n")
		buf.WriteString(area.Instructions)
		buf.WriteString("\n\n")
	}

	// Go/No-Go reporting format
	buf.WriteString("## Reporting Format\n\n")
	buf.WriteString("Report results in Go/No-Go format:\n\n")
	buf.WriteString(fmt.Sprintf("- GO: Check passed\n"))
	buf.WriteString(fmt.Sprintf("- NO-GO: Check failed (blocking)\n"))
	buf.WriteString(fmt.Sprintf("- WARN: Check failed (non-blocking)\n"))
	buf.WriteString(fmt.Sprintf("- SKIP: Check skipped\n\n"))
	buf.WriteString(fmt.Sprintf("Final status: %s VALIDATION: GO or NO-GO\n", strings.ToUpper(area.Name)))

	buf.WriteString("\n'''\n")

	return buf.Bytes(), nil
}

// ReadFile reads a Gemini command TOML file and returns canonical ValidationArea.
func (a *Adapter) ReadFile(path string) (*core.ValidationArea, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &core.ReadError{Path: path, Err: err}
	}

	area, err := a.Parse(data)
	if err != nil {
		if pe, ok := err.(*core.ParseError); ok {
			pe.Path = path
		}
		return nil, err
	}

	// Infer name from filename if not set
	if area.Name == "" {
		base := filepath.Base(path)
		name := strings.TrimSuffix(base, filepath.Ext(base))
		name = strings.TrimSuffix(name, "-validator")
		area.Name = name
	}

	return area, nil
}

// WriteFile writes canonical ValidationArea to a Gemini command TOML file.
func (a *Adapter) WriteFile(area *core.ValidationArea, path string) error {
	data, err := a.Marshal(area)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, core.DefaultDirMode); err != nil {
		return &core.WriteError{Path: path, Err: err}
	}

	if err := os.WriteFile(path, data, core.DefaultFileMode); err != nil {
		return &core.WriteError{Path: path, Err: err}
	}

	return nil
}
