// Package kiro provides the Kiro CLI skill adapter for steering files.
package kiro

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/plexusone/assistantkit/skills/core"
)

const (
	// AdapterName is the identifier for this adapter.
	AdapterName = "kiro"

	// SteeringDir is the default steering directory name.
	SteeringDir = "steering"
)

func init() {
	core.Register(&Adapter{})
}

// Adapter converts between canonical Skill and Kiro CLI steering file format.
type Adapter struct{}

// Name returns the adapter identifier.
func (a *Adapter) Name() string {
	return AdapterName
}

// SkillFileName returns the skill definition filename.
// For Kiro, steering files are named <skill-name>.md directly.
func (a *Adapter) SkillFileName() string {
	return ".md" // Used as suffix
}

// DefaultDir returns the default directory name for Kiro steering files.
func (a *Adapter) DefaultDir() string {
	return SteeringDir
}

// Parse converts Kiro steering file bytes to canonical Skill.
func (a *Adapter) Parse(data []byte) (*core.Skill, error) {
	content := string(data)
	lines := strings.SplitN(content, "\n", 2)

	skill := &core.Skill{}

	// Extract name from first line (# Title)
	if len(lines) > 0 && strings.HasPrefix(lines[0], "# ") {
		title := strings.TrimPrefix(lines[0], "# ")
		skill.Name = toKebabCase(title)
		skill.Description = title
	}

	// Rest is instructions
	if len(lines) > 1 {
		skill.Instructions = strings.TrimSpace(lines[1])
	}

	return skill, nil
}

// Marshal converts canonical Skill to Kiro steering file bytes.
func (a *Adapter) Marshal(skill *core.Skill) ([]byte, error) {
	var buf bytes.Buffer

	// Write title from name (convert kebab-case to Title Case)
	title := toTitleCase(skill.Name)
	buf.WriteString(fmt.Sprintf("# %s\n\n", title))

	// Write description if different from title
	if skill.Description != "" && skill.Description != title {
		buf.WriteString(fmt.Sprintf("%s\n\n", skill.Description))
	}

	// Write instructions directly (they contain the markdown content)
	if skill.Instructions != "" {
		buf.WriteString(skill.Instructions)
		if !strings.HasSuffix(skill.Instructions, "\n") {
			buf.WriteString("\n")
		}
	}

	return buf.Bytes(), nil
}

// ReadFile reads a Kiro steering file and returns canonical Skill.
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

	// Infer name from filename if not set
	if skill.Name == "" {
		base := filepath.Base(path)
		skill.Name = strings.TrimSuffix(base, filepath.Ext(base))
	}

	return skill, nil
}

// WriteFile writes canonical Skill to a Kiro steering file.
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

// WriteSkillDir writes the skill as a steering file.
// For Kiro, skills are flat files in the steering directory, not subdirectories.
func (a *Adapter) WriteSkillDir(skill *core.Skill, baseDir string) error {
	// Ensure directory exists
	if err := os.MkdirAll(baseDir, core.DefaultDirMode); err != nil {
		return &core.WriteError{Path: baseDir, Err: err}
	}

	// Write steering file: steering/<skill-name>.md
	steeringPath := filepath.Join(baseDir, skill.Name+".md")
	return a.WriteFile(skill, steeringPath)
}

// toKebabCase converts "Title Case" or "Title-Case" to "title-case".
func toKebabCase(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	return s
}

// toTitleCase converts "kebab-case" to "Title Case".
func toTitleCase(s string) string {
	words := strings.Split(s, "-")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}
