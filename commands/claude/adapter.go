// Package claude provides the Claude Code command adapter.
package claude

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

// Adapter converts between canonical Command and Claude Code command format.
type Adapter struct{}

// Name returns the adapter identifier.
func (a *Adapter) Name() string {
	return "claude"
}

// FileExtension returns the file extension for Claude commands.
func (a *Adapter) FileExtension() string {
	return ".md"
}

// DefaultDir returns the default directory name for Claude commands.
func (a *Adapter) DefaultDir() string {
	return "commands"
}

// Parse converts Claude command Markdown bytes to canonical Command.
func (a *Adapter) Parse(data []byte) (*core.Command, error) {
	frontmatter, body := parseFrontmatter(data)

	cmd := &core.Command{
		Description:  frontmatter["description"],
		Instructions: strings.TrimSpace(body),
	}

	// Extract name from frontmatter or infer from content
	if name, ok := frontmatter["name"]; ok {
		cmd.Name = name
	}

	return cmd, nil
}

// Marshal converts canonical Command to Claude command Markdown bytes.
func (a *Adapter) Marshal(cmd *core.Command) ([]byte, error) {
	var buf bytes.Buffer

	// Write YAML frontmatter
	buf.WriteString("---\n")
	buf.WriteString(fmt.Sprintf("description: %s\n", cmd.Description))
	buf.WriteString("---\n\n")

	// Write title
	title := strings.ReplaceAll(cmd.Name, "-", " ")
	title = strings.Title(title)
	buf.WriteString(fmt.Sprintf("# %s\n\n", title))

	// Write description
	buf.WriteString(fmt.Sprintf("%s\n\n", cmd.Description))

	// Write usage section if there are arguments
	if len(cmd.Arguments) > 0 {
		buf.WriteString("## Usage\n\n")
		buf.WriteString("```\n")
		buf.WriteString(fmt.Sprintf("/%s", cmd.Name))
		for _, arg := range cmd.Arguments {
			if arg.Required {
				buf.WriteString(fmt.Sprintf(" <%s>", arg.Name))
			} else {
				buf.WriteString(fmt.Sprintf(" [%s]", arg.Name))
			}
		}
		buf.WriteString("\n```\n\n")

		// Write arguments section
		buf.WriteString("## Arguments\n\n")
		for _, arg := range cmd.Arguments {
			required := ""
			if arg.Required {
				required = " (required)"
			}
			desc := arg.Description
			if desc == "" {
				desc = arg.Hint
			}
			buf.WriteString(fmt.Sprintf("- **%s**%s: %s\n", arg.Name, required, desc))
		}
		buf.WriteString("\n")
	}

	// Write process section if there are steps
	if len(cmd.Process) > 0 {
		buf.WriteString("## Process\n\n")
		for i, step := range cmd.Process {
			buf.WriteString(fmt.Sprintf("%d. %s\n", i+1, step))
		}
		buf.WriteString("\n")
	}

	// Write dependencies section if there are dependencies
	if len(cmd.Dependencies) > 0 {
		buf.WriteString("## Dependencies\n\n")
		for _, dep := range cmd.Dependencies {
			buf.WriteString(fmt.Sprintf("- `%s`\n", dep))
		}
		buf.WriteString("\n")
	}

	// Write instructions if different from description
	if cmd.Instructions != "" && cmd.Instructions != cmd.Description {
		buf.WriteString("## Instructions\n\n")
		buf.WriteString(cmd.Instructions)
		buf.WriteString("\n")
	}

	return buf.Bytes(), nil
}

// ReadFile reads a Claude command Markdown file and returns canonical Command.
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

// WriteFile writes canonical Command to a Claude command Markdown file.
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
