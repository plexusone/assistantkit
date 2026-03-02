// Package codex provides the OpenAI Codex CLI prompt adapter.
package codex

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/plexusone/assistantkit/commands/core"
)

func init() {
	core.Register(&Adapter{})
}

// Adapter converts between canonical Command and OpenAI Codex prompt format.
type Adapter struct{}

// Name returns the adapter identifier.
func (a *Adapter) Name() string {
	return "codex"
}

// FileExtension returns the file extension for Codex prompts.
func (a *Adapter) FileExtension() string {
	return ".md"
}

// DefaultDir returns the default directory name for Codex prompts.
func (a *Adapter) DefaultDir() string {
	return "prompts"
}

// Parse converts Codex prompt Markdown bytes to canonical Command.
func (a *Adapter) Parse(data []byte) (*core.Command, error) {
	frontmatter, body := parseFrontmatter(data)

	cmd := &core.Command{
		Description:  frontmatter["description"],
		Instructions: strings.TrimSpace(body),
	}

	// Parse argument-hint if present (e.g., "VERSION=<semver>")
	if hint, ok := frontmatter["argument-hint"]; ok {
		args := parseArgumentHint(hint)
		cmd.Arguments = args
	}

	return cmd, nil
}

// Marshal converts canonical Command to Codex prompt Markdown bytes.
func (a *Adapter) Marshal(cmd *core.Command) ([]byte, error) {
	var buf bytes.Buffer

	// Write YAML frontmatter
	buf.WriteString("---\n")
	buf.WriteString(fmt.Sprintf("description: %s\n", cmd.Description))

	// Write argument-hint if there are arguments
	if len(cmd.Arguments) > 0 {
		hints := make([]string, 0, len(cmd.Arguments))
		for _, arg := range cmd.Arguments {
			hint := arg.Hint
			if hint == "" {
				hint = fmt.Sprintf("<%s>", arg.Type)
			}
			hints = append(hints, fmt.Sprintf("%s=%s", strings.ToUpper(arg.Name), hint))
		}
		buf.WriteString(fmt.Sprintf("argument-hint: %s\n", strings.Join(hints, " ")))
	}

	buf.WriteString("---\n\n")

	// Write main instructions
	if cmd.Instructions != "" {
		buf.WriteString(cmd.Instructions)
		buf.WriteString("\n\n")
	} else {
		buf.WriteString(cmd.Description)
		buf.WriteString("\n\n")
	}

	// Write arguments section if present
	if len(cmd.Arguments) > 0 {
		buf.WriteString("Arguments:\n")
		for _, arg := range cmd.Arguments {
			desc := arg.Description
			if desc == "" {
				desc = arg.Hint
			}
			buf.WriteString(fmt.Sprintf("- $%s: %s\n", strings.ToUpper(arg.Name), desc))
		}
		buf.WriteString("\n")
	}

	// Write process section if present
	if len(cmd.Process) > 0 {
		buf.WriteString("Process:\n")
		for i, step := range cmd.Process {
			buf.WriteString(fmt.Sprintf("%d. %s\n", i+1, step))
		}
		buf.WriteString("\n")
	}

	// Write dependencies section if present
	if len(cmd.Dependencies) > 0 {
		buf.WriteString("Dependencies:\n")
		for _, dep := range cmd.Dependencies {
			buf.WriteString(fmt.Sprintf("- %s\n", dep))
		}
	}

	return buf.Bytes(), nil
}

// ReadFile reads a Codex prompt Markdown file and returns canonical Command.
func (a *Adapter) ReadFile(path string) (*core.Command, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &core.ReadError{Path: path, Err: err}
	}

	cmd, err := a.Parse(data)
	if err != nil {
		if pe, ok := err.(*core.ParseError); ok {
			pe.Path = path
		}
		return nil, err
	}

	// Infer name from filename if not set
	if cmd.Name == "" {
		base := filepath.Base(path)
		cmd.Name = strings.TrimSuffix(base, filepath.Ext(base))
	}

	return cmd, nil
}

// WriteFile writes canonical Command to a Codex prompt Markdown file.
func (a *Adapter) WriteFile(cmd *core.Command, path string) error {
	data, err := a.Marshal(cmd)
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

// parseFrontmatter extracts YAML frontmatter and body from Markdown.
func parseFrontmatter(data []byte) (map[string]string, string) {
	content := string(data)
	frontmatter := make(map[string]string)

	if !strings.HasPrefix(content, "---") {
		return frontmatter, content
	}

	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return frontmatter, content
	}

	// Parse simple YAML key: value pairs
	lines := strings.Split(strings.TrimSpace(parts[1]), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, ":")
		if idx > 0 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])
			// Remove quotes if present
			value = strings.Trim(value, "\"'")
			frontmatter[key] = value
		}
	}

	return frontmatter, strings.TrimSpace(parts[2])
}

// parseArgumentHint parses Codex argument-hint format (e.g., "VERSION=<semver> FILE=<path>").
func parseArgumentHint(hint string) []core.Argument {
	var args []core.Argument

	parts := strings.Fields(hint)
	for _, part := range parts {
		idx := strings.Index(part, "=")
		if idx > 0 {
			name := strings.ToLower(part[:idx])
			typeHint := part[idx+1:]
			// Remove angle brackets if present
			typeHint = strings.Trim(typeHint, "<>[]")

			args = append(args, core.Argument{
				Name:     name,
				Type:     "string",
				Required: !strings.HasPrefix(part[idx+1:], "["),
				Hint:     typeHint,
			})
		}
	}

	return args
}
