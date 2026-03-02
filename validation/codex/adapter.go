// Package codex provides the OpenAI Codex validation area adapter.
// It converts ValidationArea definitions to Codex prompt format.
package codex

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

// Adapter converts between canonical ValidationArea and Codex prompt format.
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

// Parse converts Codex prompt Markdown bytes to canonical ValidationArea.
func (a *Adapter) Parse(data []byte) (*core.ValidationArea, error) {
	frontmatter, body := parseFrontmatter(data)

	area := &core.ValidationArea{
		Name:         frontmatter["name"],
		Description:  frontmatter["description"],
		Instructions: strings.TrimSpace(body),
	}

	// Parse tags as tools if present
	if tags, ok := frontmatter["tags"]; ok {
		area.Tools = parseList(tags)
	}

	return area, nil
}

// Marshal converts canonical ValidationArea to Codex prompt Markdown bytes.
func (a *Adapter) Marshal(area *core.ValidationArea) ([]byte, error) {
	var buf bytes.Buffer

	// Generate prompt name from validation area name
	promptName := area.Name + "-validator"

	// Write YAML frontmatter
	buf.WriteString("---\n")
	buf.WriteString(fmt.Sprintf("name: %s\n", promptName))
	buf.WriteString(fmt.Sprintf("description: %s validation for release readiness. %s\n",
		strings.Title(area.Name), area.Description))

	// Add tags for categorization
	buf.WriteString("tags:\n")
	buf.WriteString("  - validation\n")
	buf.WriteString("  - release\n")
	buf.WriteString(fmt.Sprintf("  - %s\n", area.Name))

	// Add model preference if specified
	if area.Model != "" {
		buf.WriteString(fmt.Sprintf("model: %s\n", area.Model))
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

	// Write validation checks
	if len(area.Checks) > 0 {
		buf.WriteString("## Validation Checks\n\n")
		for i, check := range area.Checks {
			required := "Optional"
			icon := "⚠️"
			if check.Required {
				required = "Required"
				icon = "🔴"
			}

			buf.WriteString(fmt.Sprintf("### %d. %s %s (%s)\n\n", i+1, icon, check.Name, required))

			if check.Description != "" {
				buf.WriteString(fmt.Sprintf("%s\n\n", check.Description))
			}

			if check.Command != "" {
				buf.WriteString("**Command:**\n\n")
				buf.WriteString(fmt.Sprintf("```bash\n%s\n```\n\n", check.Command))
			}

			if check.Pattern != "" {
				buf.WriteString("**Pattern to check:**\n\n")
				buf.WriteString(fmt.Sprintf("```\n%s\n```\n\n", check.Pattern))
			}

			if check.FilePattern != "" {
				buf.WriteString(fmt.Sprintf("**Files:** `%s`\n\n", check.FilePattern))
			}
		}
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
		buf.WriteString("\n\n")
	}

	// Add Go/No-Go reporting format
	buf.WriteString("## Reporting Format\n\n")
	buf.WriteString("Report results using the following status indicators:\n\n")
	buf.WriteString("| Status | Meaning |\n")
	buf.WriteString("|--------|----------|\n")
	buf.WriteString("| ✅ GO | Check passed |\n")
	buf.WriteString("| ❌ NO-GO | Check failed (blocking) |\n")
	buf.WriteString("| ⚠️ WARN | Check failed (non-blocking) |\n")
	buf.WriteString("| ⏭️ SKIP | Check skipped |\n\n")

	buf.WriteString("### Final Report Template\n\n")
	buf.WriteString("```\n")
	buf.WriteString(fmt.Sprintf("%s VALIDATION REPORT\n", strings.ToUpper(area.Name)))
	buf.WriteString("========================\n\n")
	buf.WriteString("Checks:\n")
	for _, check := range area.Checks {
		buf.WriteString(fmt.Sprintf("- [ ] %s: [GO/NO-GO/WARN/SKIP]\n", check.Name))
	}
	buf.WriteString("\n")
	buf.WriteString(fmt.Sprintf("FINAL STATUS: %s VALIDATION [GO/NO-GO]\n", strings.ToUpper(area.Name)))
	buf.WriteString("```\n")

	return buf.Bytes(), nil
}

// ReadFile reads a Codex prompt Markdown file and returns canonical ValidationArea.
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

// WriteFile writes canonical ValidationArea to a Codex prompt Markdown file.
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
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "-") {
			continue
		}
		idx := strings.Index(line, ":")
		if idx > 0 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])
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
