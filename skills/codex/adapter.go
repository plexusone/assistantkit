// Package codex provides the OpenAI Codex skill adapter.
package codex

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/plexusone/assistantkit/skills/core"
)

func init() {
	core.Register(&Adapter{})
}

// Adapter converts between canonical Skill and OpenAI Codex skill format.
type Adapter struct{}

// Name returns the adapter identifier.
func (a *Adapter) Name() string {
	return "codex"
}

// SkillFileName returns the skill definition filename.
func (a *Adapter) SkillFileName() string {
	return "SKILL.md"
}

// DefaultDir returns the default directory name for Codex skills.
func (a *Adapter) DefaultDir() string {
	return "skills"
}

// Parse converts Codex SKILL.md bytes to canonical Skill.
func (a *Adapter) Parse(data []byte) (*core.Skill, error) {
	frontmatter, body := parseFrontmatter(data)

	skill := &core.Skill{
		Name:         frontmatter["name"],
		Description:  frontmatter["description"],
		Instructions: strings.TrimSpace(body),
	}

	return skill, nil
}

// Marshal converts canonical Skill to Codex SKILL.md bytes.
func (a *Adapter) Marshal(skill *core.Skill) ([]byte, error) {
	var buf bytes.Buffer

	// Write YAML frontmatter (Codex requires only name and description)
	buf.WriteString("---\n")
	buf.WriteString(fmt.Sprintf("name: %s\n", skill.Name))
	buf.WriteString(fmt.Sprintf("description: %s\n", skill.Description))
	buf.WriteString("---\n\n")

	// Write instructions (Codex puts the main content after frontmatter)
	if skill.Instructions != "" {
		buf.WriteString(skill.Instructions)
		buf.WriteString("\n")
	} else {
		buf.WriteString(skill.Description)
		buf.WriteString("\n")
	}

	return buf.Bytes(), nil
}

// ReadFile reads a Codex SKILL.md file and returns canonical Skill.
func (a *Adapter) ReadFile(path string) (*core.Skill, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &core.ReadError{Path: path, Err: err}
	}

	skill, err := a.Parse(data)
	if err != nil {
		if pe, ok := err.(*core.ParseError); ok {
			pe.Path = path
		}
		return nil, err
	}

	// Infer name from directory if not set
	if skill.Name == "" {
		skill.Name = filepath.Base(filepath.Dir(path))
	}

	return skill, nil
}

// WriteFile writes canonical Skill to a Codex SKILL.md file.
func (a *Adapter) WriteFile(skill *core.Skill, path string) error {
	data, err := a.Marshal(skill)
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

// WriteSkillDir writes the complete skill directory structure.
func (a *Adapter) WriteSkillDir(skill *core.Skill, baseDir string) error {
	// Create skill directory: skills/<skill-name>/
	skillDir := filepath.Join(baseDir, skill.Name)
	if err := os.MkdirAll(skillDir, core.DefaultDirMode); err != nil {
		return &core.WriteError{Path: skillDir, Err: err}
	}

	// Write SKILL.md
	skillPath := filepath.Join(skillDir, a.SkillFileName())
	if err := a.WriteFile(skill, skillPath); err != nil {
		return err
	}

	// Create optional directories based on Codex convention
	if len(skill.Scripts) > 0 {
		scriptsDir := filepath.Join(skillDir, "scripts")
		if err := os.MkdirAll(scriptsDir, core.DefaultDirMode); err != nil {
			return &core.WriteError{Path: scriptsDir, Err: err}
		}
	}

	if len(skill.References) > 0 {
		refsDir := filepath.Join(skillDir, "references")
		if err := os.MkdirAll(refsDir, core.DefaultDirMode); err != nil {
			return &core.WriteError{Path: refsDir, Err: err}
		}
	}

	if len(skill.Assets) > 0 {
		assetsDir := filepath.Join(skillDir, "assets")
		if err := os.MkdirAll(assetsDir, core.DefaultDirMode); err != nil {
			return &core.WriteError{Path: assetsDir, Err: err}
		}
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
