// Package claude provides the Claude Code validation area adapter.
// It converts ValidationArea definitions to Claude Code sub-agents.
package claude

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

// Adapter converts between canonical ValidationArea and Claude Code agent format.
type Adapter struct{}

// Name returns the adapter identifier.
func (a *Adapter) Name() string {
	return "claude"
}

// FileExtension returns the file extension for Claude agents.
func (a *Adapter) FileExtension() string {
	return ".md"
}

// DefaultDir returns the default directory name for Claude agents.
func (a *Adapter) DefaultDir() string {
	return "agents"
}

// Parse converts Claude agent Markdown bytes to canonical ValidationArea.
func (a *Adapter) Parse(data []byte) (*core.ValidationArea, error) {
	frontmatter, body := parseFrontmatter(data)

	area := &core.ValidationArea{
		Name:         frontmatter["name"],
		Description:  frontmatter["description"],
		Model:        frontmatter["model"],
		Instructions: strings.TrimSpace(body),
	}

	// Parse tools if present
	if tools, ok := frontmatter["tools"]; ok {
		area.Tools = parseList(tools)
	}

	// Parse skills if present
	if skills, ok := frontmatter["skills"]; ok {
		area.Skills = parseList(skills)
	}

	return area, nil
}

// Marshal converts canonical ValidationArea to Claude agent Markdown bytes.
func (a *Adapter) Marshal(area *core.ValidationArea) ([]byte, error) {
	var buf bytes.Buffer

	// Generate agent name from validation area name
	agentName := area.Name + "-validator"

	// Write YAML frontmatter
	buf.WriteString("---\n")
	buf.WriteString(fmt.Sprintf("name: %s\n", agentName))
	buf.WriteString(fmt.Sprintf("description: %s validation agent for release readiness. %s\n",
		strings.Title(area.Name), area.Description))

	// Default to haiku for validation agents (fast, cost-effective)
	model := area.Model
	if model == "" {
		model = "haiku"
	}
	buf.WriteString(fmt.Sprintf("model: %s\n", model))

	// Default tools for validation
	tools := area.Tools
	if len(tools) == 0 {
		tools = []string{"Read", "Grep", "Glob", "Bash"}
	}
	buf.WriteString(fmt.Sprintf("tools: %s\n", strings.Join(tools, ", ")))

	if len(area.Skills) > 0 {
		buf.WriteString(fmt.Sprintf("skills: %s\n", strings.Join(area.Skills, ", ")))
	}

	buf.WriteString("---\n\n")

	// Write title
	title := strings.Title(strings.ReplaceAll(area.Name, "-", " ")) + " Validator"
	buf.WriteString(fmt.Sprintf("# %s\n\n", title))

	// Write description
	buf.WriteString(fmt.Sprintf("%s\n\n", area.Description))

	// Write sign-off criteria if present
	if area.SignOffCriteria != "" {
		buf.WriteString("## Sign-Off Criteria\n\n")
		buf.WriteString(fmt.Sprintf("%s\n\n", area.SignOffCriteria))
	}

	// Write checks section
	if len(area.Checks) > 0 {
		buf.WriteString("## Validation Checks\n\n")
		buf.WriteString("| Check | Required | Command/Pattern |\n")
		buf.WriteString("|-------|----------|----------------|\n")
		for _, check := range area.Checks {
			required := "⚠️ Warning"
			if check.Required {
				required = "🔴 Required"
			}
			cmdOrPattern := check.Command
			if cmdOrPattern == "" {
				cmdOrPattern = check.Pattern
			}
			buf.WriteString(fmt.Sprintf("| %s | %s | `%s` |\n", check.Name, required, cmdOrPattern))
		}
		buf.WriteString("\n")
	}

	// Write dependencies if present
	if len(area.Dependencies) > 0 {
		buf.WriteString("## Dependencies\n\n")
		buf.WriteString("Required CLI tools:\n\n")
		for _, dep := range area.Dependencies {
			buf.WriteString(fmt.Sprintf("- `%s`\n", dep))
		}
		buf.WriteString("\n")
	}

	// Write instructions
	if area.Instructions != "" {
		buf.WriteString("## Instructions\n\n")
		buf.WriteString(area.Instructions)
		buf.WriteString("\n")
	}

	// Add Go/No-Go reporting format
	buf.WriteString("\n## Reporting Format\n\n")
	buf.WriteString("Report results in Go/No-Go format:\n\n")
	buf.WriteString("```\n")
	buf.WriteString(fmt.Sprintf("╔══════════════════════════════════════════════════════════════╗\n"))
	buf.WriteString(fmt.Sprintf("║                    %s VALIDATION                             ║\n", strings.ToUpper(area.Name)))
	buf.WriteString(fmt.Sprintf("╠══════════════════════════════════════════════════════════════╣\n"))
	buf.WriteString(fmt.Sprintf("║ 🟢 GO     Check passed                                       ║\n"))
	buf.WriteString(fmt.Sprintf("║ 🔴 NO-GO  Check failed (blocking)                            ║\n"))
	buf.WriteString(fmt.Sprintf("║ 🟡 WARN   Check failed (non-blocking)                        ║\n"))
	buf.WriteString(fmt.Sprintf("║ ⚪ SKIP   Check skipped                                      ║\n"))
	buf.WriteString(fmt.Sprintf("╠══════════════════════════════════════════════════════════════╣\n"))
	buf.WriteString(fmt.Sprintf("║                    🚀 %s: GO 🚀                              ║\n", strings.ToUpper(area.Name)))
	buf.WriteString(fmt.Sprintf("╚══════════════════════════════════════════════════════════════╝\n"))
	buf.WriteString("```\n")

	return buf.Bytes(), nil
}

// ReadFile reads a Claude agent Markdown file and returns canonical ValidationArea.
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
		// Remove -validator suffix if present
		name = strings.TrimSuffix(name, "-validator")
		area.Name = name
	}

	return area, nil
}

// WriteFile writes canonical ValidationArea to a Claude agent Markdown file.
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

// parseList parses a comma-separated list.
func parseList(s string) []string {
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
