// Package claude provides the Claude Code agent adapter.
package claude

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/plexusone/assistantkit/agents/core"
)

func init() {
	core.Register(&Adapter{})
}

// Adapter converts between canonical Agent and Claude Code agent format.
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

// Parse converts Claude agent Markdown bytes to canonical Agent.
func (a *Adapter) Parse(data []byte) (*core.Agent, error) {
	frontmatter, body := parseFrontmatter(data)

	agent := &core.Agent{
		Name:         frontmatter["name"],
		Description:  frontmatter["description"],
		Model:        core.Model(frontmatter["model"]),
		Instructions: strings.TrimSpace(body),
	}

	// Parse tools if present
	if tools, ok := frontmatter["tools"]; ok {
		agent.Tools = parseList(tools)
	}

	// Parse skills if present
	if skills, ok := frontmatter["skills"]; ok {
		agent.Skills = parseList(skills)
	}

	// Parse dependencies if present
	if deps, ok := frontmatter["dependencies"]; ok {
		agent.Dependencies = parseList(deps)
	}

	return agent, nil
}

// Marshal converts canonical Agent to Claude agent Markdown bytes.
func (a *Adapter) Marshal(agent *core.Agent) ([]byte, error) {
	var buf bytes.Buffer

	// Write YAML frontmatter
	buf.WriteString("---\n")
	buf.WriteString(fmt.Sprintf("name: %s\n", agent.Name))
	buf.WriteString(fmt.Sprintf("description: %s\n", agent.Description))

	if agent.Model != "" {
		buf.WriteString(fmt.Sprintf("model: %s\n", agent.Model))
	}

	if len(agent.Tools) > 0 {
		buf.WriteString(fmt.Sprintf("tools: [%s]\n", strings.Join(agent.Tools, ", ")))
	}

	if len(agent.Skills) > 0 {
		buf.WriteString(fmt.Sprintf("skills: [%s]\n", strings.Join(agent.Skills, ", ")))
	}

	if len(agent.Dependencies) > 0 {
		buf.WriteString(fmt.Sprintf("dependencies: [%s]\n", strings.Join(agent.Dependencies, ", ")))
	}

	buf.WriteString("---\n\n")

	// Write instructions directly (they already contain markdown formatting)
	if agent.Instructions != "" {
		buf.WriteString(agent.Instructions)
		buf.WriteString("\n")
	}

	return buf.Bytes(), nil
}

// ReadFile reads a Claude agent Markdown file and returns canonical Agent.
func (a *Adapter) ReadFile(path string) (*core.Agent, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &core.ReadError{Path: path, Err: err}
	}

	agent, err := a.Parse(data)
	if err != nil {
		if pe, ok := err.(*core.ParseError); ok {
			pe.Path = path
		}
		return nil, err
	}

	// Infer name from filename if not set
	if agent.Name == "" {
		base := filepath.Base(path)
		agent.Name = strings.TrimSuffix(base, filepath.Ext(base))
	}

	return agent, nil
}

// WriteFile writes canonical Agent to a Claude agent Markdown file.
func (a *Adapter) WriteFile(agent *core.Agent, path string) error {
	data, err := a.Marshal(agent)
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

// parseList parses a list in either YAML array format [a, b, c] or comma-separated format.
func parseList(s string) []string {
	s = strings.TrimSpace(s)

	// Handle YAML array syntax: [a, b, c]
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		s = s[1 : len(s)-1]
	}

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
