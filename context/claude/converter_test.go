package claude

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/plexusone/assistantkit/context/core"
)

func TestNewConverter(t *testing.T) {
	c := NewConverter()

	if c.Name() != ConverterName {
		t.Errorf("expected name '%s', got '%s'", ConverterName, c.Name())
	}
	if c.OutputFileName() != OutputFile {
		t.Errorf("expected output file '%s', got '%s'", OutputFile, c.OutputFileName())
	}
}

func TestConverterConvertBasic(t *testing.T) {
	c := NewConverter()
	ctx := core.NewContext("test-project")
	ctx.Description = "A test project"

	data, err := c.Convert(ctx)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	md := string(data)

	// Check header
	if !strings.Contains(md, "# test-project") {
		t.Error("expected markdown to contain project name header")
	}

	// Check description
	if !strings.Contains(md, "A test project") {
		t.Error("expected markdown to contain description")
	}

	// Check footer
	if !strings.Contains(md, "Generated from CONTEXT.json") {
		t.Error("expected markdown to contain footer")
	}
}

func TestConverterConvertNilContext(t *testing.T) {
	c := NewConverter()

	_, err := c.Convert(nil)
	if err == nil {
		t.Error("expected error for nil context")
	}
}

func TestConverterConvertMissingName(t *testing.T) {
	c := NewConverter()
	ctx := &core.Context{}

	_, err := c.Convert(ctx)
	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestConverterConvertWithVersion(t *testing.T) {
	c := NewConverter()
	ctx := core.NewContext("test")
	ctx.Version = "1.0.0"
	ctx.Language = "go"

	data, err := c.Convert(ctx)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	md := string(data)
	if !strings.Contains(md, "**Version:** 1.0.0") {
		t.Error("expected markdown to contain version")
	}
	if !strings.Contains(md, "**Language:** go") {
		t.Error("expected markdown to contain language")
	}
}

func TestConverterConvertWithArchitecture(t *testing.T) {
	c := NewConverter()
	ctx := core.NewContext("test")
	ctx.Architecture = &core.Architecture{
		Pattern: "adapter",
		Summary: "Uses adapter pattern",
		Diagrams: []core.Diagram{
			{Title: "Overview", Type: "ascii", Content: "A -> B"},
			{Title: "Flow", Type: "mermaid", Content: "graph LR\n  A --> B"},
		},
	}

	data, err := c.Convert(ctx)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	md := string(data)
	if !strings.Contains(md, "## Architecture") {
		t.Error("expected markdown to contain Architecture section")
	}
	if !strings.Contains(md, "**Pattern:** adapter") {
		t.Error("expected markdown to contain pattern")
	}
	if !strings.Contains(md, "```mermaid") {
		t.Error("expected markdown to contain mermaid code block")
	}
}

func TestConverterConvertWithPackages(t *testing.T) {
	c := NewConverter()
	ctx := core.NewContext("test")
	ctx.AddPackage("pkg/core", "Core functionality")
	ctx.AddPackage("pkg/adapter", "Adapter implementations")

	data, err := c.Convert(ctx)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	md := string(data)
	if !strings.Contains(md, "## Packages") {
		t.Error("expected markdown to contain Packages section")
	}
	if !strings.Contains(md, "`pkg/core`") {
		t.Error("expected markdown to contain package path")
	}
	if !strings.Contains(md, "Core functionality") {
		t.Error("expected markdown to contain package purpose")
	}
}

func TestConverterConvertWithCommands(t *testing.T) {
	c := NewConverter()
	ctx := core.NewContext("test")
	ctx.SetCommand("build", "go build ./...")
	ctx.SetCommand("test", "go test ./...")
	ctx.SetCommand("custom", "make custom")

	data, err := c.Convert(ctx)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	md := string(data)
	if !strings.Contains(md, "## Commands") {
		t.Error("expected markdown to contain Commands section")
	}
	if !strings.Contains(md, "go build ./...") {
		t.Error("expected markdown to contain build command")
	}
}

func TestConverterConvertWithConventions(t *testing.T) {
	c := NewConverter()
	ctx := core.NewContext("test")
	ctx.AddConvention("Use gofmt")
	ctx.AddConvention("Follow Go idioms")

	data, err := c.Convert(ctx)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	md := string(data)
	if !strings.Contains(md, "## Conventions") {
		t.Error("expected markdown to contain Conventions section")
	}
	if !strings.Contains(md, "- Use gofmt") {
		t.Error("expected markdown to contain convention")
	}
}

func TestConverterConvertWithDependencies(t *testing.T) {
	c := NewConverter()
	ctx := core.NewContext("test")
	ctx.Dependencies = &core.Dependencies{
		Runtime: []core.Dependency{
			{Name: "go-toml/v2", Purpose: "TOML parsing"},
		},
		Development: []core.Dependency{
			{Name: "testify"},
		},
	}

	data, err := c.Convert(ctx)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	md := string(data)
	if !strings.Contains(md, "## Dependencies") {
		t.Error("expected markdown to contain Dependencies section")
	}
	if !strings.Contains(md, "### Runtime") {
		t.Error("expected markdown to contain Runtime subsection")
	}
	if !strings.Contains(md, "**go-toml/v2**") {
		t.Error("expected markdown to contain dependency with purpose")
	}
}

func TestConverterConvertWithTesting(t *testing.T) {
	c := NewConverter()
	ctx := core.NewContext("test")
	ctx.Testing = &core.Testing{
		Framework: "go test",
		Coverage:  "80%",
		Patterns:  []string{"Table-driven tests", "Subtests"},
	}

	data, err := c.Convert(ctx)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	md := string(data)
	if !strings.Contains(md, "## Testing") {
		t.Error("expected markdown to contain Testing section")
	}
	if !strings.Contains(md, "**Framework:** go test") {
		t.Error("expected markdown to contain framework")
	}
	if !strings.Contains(md, "**Coverage:** 80%") {
		t.Error("expected markdown to contain coverage")
	}
}

func TestConverterConvertWithFiles(t *testing.T) {
	c := NewConverter()
	ctx := core.NewContext("test")
	ctx.Files = &core.Files{
		EntryPoints: []string{"main.go", "cmd/cli/main.go"},
		Config:      []string{"go.mod", "go.sum"},
	}

	data, err := c.Convert(ctx)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	md := string(data)
	if !strings.Contains(md, "## Key Files") {
		t.Error("expected markdown to contain Key Files section")
	}
	if !strings.Contains(md, "`main.go`") {
		t.Error("expected markdown to contain entry point")
	}
}

func TestConverterConvertWithNotes(t *testing.T) {
	c := NewConverter()
	ctx := core.NewContext("test")
	ctx.AddNote("Simple note")
	ctx.AddNoteWithSeverity("Warning Title", "This is a warning", "warning")
	ctx.AddNoteWithSeverity("Critical Issue", "This is critical", "critical")

	data, err := c.Convert(ctx)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	md := string(data)
	if !strings.Contains(md, "## Notes") {
		t.Error("expected markdown to contain Notes section")
	}
	if !strings.Contains(md, "**Warning:**") {
		t.Error("expected markdown to contain warning prefix")
	}
	if !strings.Contains(md, "**CRITICAL:**") {
		t.Error("expected markdown to contain critical prefix")
	}
}

func TestConverterConvertWithRelated(t *testing.T) {
	c := NewConverter()
	ctx := core.NewContext("test")
	ctx.Related = []core.Related{
		{Name: "OmniLLM", URL: "https://github.com/example/omnillm", Description: "LLM abstraction"},
		{Name: "Other Project"},
	}

	data, err := c.Convert(ctx)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	md := string(data)
	if !strings.Contains(md, "## Related") {
		t.Error("expected markdown to contain Related section")
	}
	if !strings.Contains(md, "[OmniLLM](https://github.com/example/omnillm)") {
		t.Error("expected markdown to contain linked related project")
	}
}

func TestConverterWriteFile(t *testing.T) {
	c := NewConverter()
	ctx := core.NewContext("test-project")
	ctx.Description = "A test project"

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "CLAUDE.md")

	if err := c.WriteFile(ctx, path); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Verify file exists and has content
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read written file: %v", err)
	}

	if !strings.Contains(string(data), "# test-project") {
		t.Error("written file should contain project header")
	}
}

func TestConverterWriteFileError(t *testing.T) {
	c := NewConverter()
	ctx := core.NewContext("test")

	err := c.WriteFile(ctx, "/nonexistent/directory/CLAUDE.md")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestConverterRegistered(t *testing.T) {
	converter, ok := core.GetConverter(ConverterName)
	if !ok {
		t.Fatal("claude converter should be registered")
	}
	if converter.Name() != ConverterName {
		t.Errorf("expected name '%s', got '%s'", ConverterName, converter.Name())
	}
}

func TestConverterFullContext(t *testing.T) {
	c := NewConverter()
	ctx := &core.Context{
		Name:        "full-project",
		Description: "A project with all fields",
		Version:     "1.0.0",
		Language:    "go",
		Architecture: &core.Architecture{
			Pattern: "adapter",
			Summary: "Uses adapter pattern",
		},
		Packages: []core.Package{
			{Path: "pkg/core", Purpose: "Core types"},
		},
		Commands: map[string]string{
			"build": "go build ./...",
			"test":  "go test ./...",
		},
		Conventions: []string{"Use gofmt"},
		Dependencies: &core.Dependencies{
			Runtime: []core.Dependency{{Name: "dep1", Purpose: "Purpose1"}},
		},
		Testing: &core.Testing{
			Framework: "go test",
		},
		Files: &core.Files{
			EntryPoints: []string{"main.go"},
		},
		Notes: []core.Note{
			{Content: "A note"},
		},
		Related: []core.Related{
			{Name: "Related1", URL: "https://example.com"},
		},
	}

	data, err := c.Convert(ctx)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	md := string(data)

	// Verify all major sections are present
	sections := []string{
		"# full-project",
		"## Architecture",
		"## Packages",
		"## Commands",
		"## Conventions",
		"## Dependencies",
		"## Testing",
		"## Key Files",
		"## Notes",
		"## Related",
	}

	for _, section := range sections {
		if !strings.Contains(md, section) {
			t.Errorf("expected markdown to contain '%s'", section)
		}
	}
}
