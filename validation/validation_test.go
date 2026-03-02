package validation_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/plexusone/assistantkit/validation"
	_ "github.com/plexusone/assistantkit/validation/claude" // Register Claude adapter
	_ "github.com/plexusone/assistantkit/validation/codex"  // Register Codex adapter
	_ "github.com/plexusone/assistantkit/validation/gemini" // Register Gemini adapter
)

// testAreas returns sample validation areas for testing
func testAreas() []*validation.ValidationArea {
	return []*validation.ValidationArea{
		{
			Name:            "qa",
			Description:     "Quality Assurance validation for release readiness.",
			SignOffCriteria: "All tests pass, code is properly linted and formatted",
			Dependencies:    []string{"go", "golangci-lint"},
			Checks: []validation.Check{
				{Name: "build", Description: "Verify project compiles", Command: "go build ./...", Required: true},
				{Name: "tests", Description: "Run all tests", Command: "go test -v ./...", Required: true},
				{Name: "lint", Description: "Check linting", Command: "golangci-lint run", Required: true},
			},
			Model:        "haiku",
			Tools:        []string{"Read", "Grep", "Glob", "Bash"},
			Instructions: "You are a Quality Assurance specialist responsible for validating software quality.",
		},
		{
			Name:            "documentation",
			Description:     "Documentation validation for release readiness.",
			SignOffCriteria: "README exists, release notes created, CHANGELOG up to date",
			Dependencies:    []string{"schangelog"},
			Checks: []validation.Check{
				{Name: "readme", Description: "README.md exists", FilePattern: "README.md", Required: true},
				{Name: "changelog", Description: "CHANGELOG.md exists", FilePattern: "CHANGELOG.md", Required: true},
			},
			Model:        "haiku",
			Tools:        []string{"Read", "Glob", "Write"},
			Instructions: "You are a Documentation specialist responsible for ensuring release documentation is complete.",
		},
		{
			Name:            "security",
			Description:     "Security validation for release readiness.",
			SignOffCriteria: "LICENSE exists, no vulnerabilities, no hardcoded secrets",
			Dependencies:    []string{"govulncheck"},
			Checks: []validation.Check{
				{Name: "license", Description: "LICENSE file exists", FilePattern: "LICENSE*", Required: true},
				{Name: "vulns", Description: "No vulnerabilities", Command: "govulncheck ./...", Required: true},
			},
			Model:        "haiku",
			Tools:        []string{"Read", "Grep", "Glob", "Bash"},
			Instructions: "You are a Security specialist responsible for ensuring release security.",
		},
	}
}

func TestAdapterRegistry(t *testing.T) {
	names := validation.AdapterNames()
	if len(names) != 3 {
		t.Errorf("Expected 3 adapters, got %d: %v", len(names), names)
	}

	expectedAdapters := []string{"claude", "codex", "gemini"}
	for _, expected := range expectedAdapters {
		found := false
		for _, name := range names {
			if name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected adapter %q not found in registry", expected)
		}
	}
}

func TestClaudeAdapter(t *testing.T) {
	adapter, ok := validation.GetAdapter("claude")
	if !ok {
		t.Fatal("Claude adapter not registered")
	}

	if adapter.Name() != "claude" {
		t.Errorf("Expected adapter name 'claude', got %q", adapter.Name())
	}

	if adapter.FileExtension() != ".md" {
		t.Errorf("Expected file extension '.md', got %q", adapter.FileExtension())
	}

	if adapter.DefaultDir() != "agents" {
		t.Errorf("Expected default dir 'agents', got %q", adapter.DefaultDir())
	}
}

func TestGeminiAdapter(t *testing.T) {
	adapter, ok := validation.GetAdapter("gemini")
	if !ok {
		t.Fatal("Gemini adapter not registered")
	}

	if adapter.Name() != "gemini" {
		t.Errorf("Expected adapter name 'gemini', got %q", adapter.Name())
	}

	if adapter.FileExtension() != ".toml" {
		t.Errorf("Expected file extension '.toml', got %q", adapter.FileExtension())
	}

	if adapter.DefaultDir() != "commands" {
		t.Errorf("Expected default dir 'commands', got %q", adapter.DefaultDir())
	}
}

func TestCodexAdapter(t *testing.T) {
	adapter, ok := validation.GetAdapter("codex")
	if !ok {
		t.Fatal("Codex adapter not registered")
	}

	if adapter.Name() != "codex" {
		t.Errorf("Expected adapter name 'codex', got %q", adapter.Name())
	}

	if adapter.FileExtension() != ".md" {
		t.Errorf("Expected file extension '.md', got %q", adapter.FileExtension())
	}

	if adapter.DefaultDir() != "prompts" {
		t.Errorf("Expected default dir 'prompts', got %q", adapter.DefaultDir())
	}
}

func TestMarshalClaudeAdapter(t *testing.T) {
	area := &validation.ValidationArea{
		Name:            "test",
		Description:     "Test validation area",
		SignOffCriteria: "All tests pass",
		Dependencies:    []string{"go", "golangci-lint"},
		Checks: []validation.Check{
			{Name: "build", Command: "go build ./...", Required: true},
			{Name: "test", Command: "go test -v ./...", Required: true},
		},
		Model:        "haiku",
		Tools:        []string{"Read", "Bash"},
		Instructions: "You are a test validator.",
	}

	adapter, _ := validation.GetAdapter("claude")
	data, err := adapter.Marshal(area)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	content := string(data)

	// Check frontmatter
	if !strings.Contains(content, "name: test-validator") {
		t.Error("Expected name in frontmatter")
	}
	if !strings.Contains(content, "model: haiku") {
		t.Error("Expected model in frontmatter")
	}
	if !strings.Contains(content, "tools: Read, Bash") {
		t.Error("Expected tools in frontmatter")
	}

	// Check content
	if !strings.Contains(content, "# Test Validator") {
		t.Error("Expected title")
	}
	if !strings.Contains(content, "## Sign-Off Criteria") {
		t.Error("Expected sign-off criteria section")
	}
	if !strings.Contains(content, "## Validation Checks") {
		t.Error("Expected validation checks section")
	}
	if !strings.Contains(content, "## Instructions") {
		t.Error("Expected instructions section")
	}
	if !strings.Contains(content, "GO") {
		t.Error("Expected GO in reporting format")
	}
}

func TestMarshalGeminiAdapter(t *testing.T) {
	area := &validation.ValidationArea{
		Name:            "test",
		Description:     "Test validation area",
		SignOffCriteria: "All tests pass",
		Dependencies:    []string{"go", "golangci-lint"},
		Checks: []validation.Check{
			{Name: "build", Command: "go build ./...", Required: true},
			{Name: "test", Command: "go test -v ./...", Required: true},
		},
		Instructions: "You are a test validator.",
	}

	adapter, _ := validation.GetAdapter("gemini")
	data, err := adapter.Marshal(area)
	if err != nil {
		t.Fatalf("Failed to marshal Gemini: %v", err)
	}

	content := string(data)

	// Check TOML structure
	if !strings.Contains(content, "[command]") {
		t.Error("Expected [command] section")
	}
	if !strings.Contains(content, `name = "test-validator"`) {
		t.Error("Expected name in command section")
	}
	if !strings.Contains(content, "[[arguments]]") {
		t.Error("Expected [[arguments]] section")
	}
	if !strings.Contains(content, "[content]") {
		t.Error("Expected [content] section")
	}
	if !strings.Contains(content, "# Test Validator") {
		t.Error("Expected title in content")
	}
	if !strings.Contains(content, "GO") {
		t.Error("Expected GO status in reporting format")
	}
}

func TestMarshalCodexAdapter(t *testing.T) {
	area := &validation.ValidationArea{
		Name:            "test",
		Description:     "Test validation area",
		SignOffCriteria: "All tests pass",
		Dependencies:    []string{"go", "golangci-lint"},
		Checks: []validation.Check{
			{Name: "build", Command: "go build ./...", Required: true},
			{Name: "test", Command: "go test -v ./...", Required: true},
		},
		Model:        "gpt-4",
		Instructions: "You are a test validator.",
	}

	adapter, _ := validation.GetAdapter("codex")
	data, err := adapter.Marshal(area)
	if err != nil {
		t.Fatalf("Failed to marshal Codex: %v", err)
	}

	content := string(data)

	// Check Markdown with YAML frontmatter
	if !strings.Contains(content, "---") {
		t.Error("Expected YAML frontmatter delimiter")
	}
	if !strings.Contains(content, "name: test-validator") {
		t.Error("Expected name in frontmatter")
	}
	if !strings.Contains(content, "tags:") {
		t.Error("Expected tags in frontmatter")
	}
	if !strings.Contains(content, "model: gpt-4") {
		t.Error("Expected model in frontmatter")
	}
	if !strings.Contains(content, "# Test Validator") {
		t.Error("Expected title")
	}
	if !strings.Contains(content, "## Sign-Off Criteria") {
		t.Error("Expected sign-off criteria section")
	}
	if !strings.Contains(content, "## Validation Checks") {
		t.Error("Expected validation checks section")
	}
	if !strings.Contains(content, "GO") {
		t.Error("Expected GO status in reporting format")
	}
}

func TestWriteAreasToDir(t *testing.T) {
	areas := testAreas()

	// Create temp directory for output
	tmpDir, err := os.MkdirTemp("", "validation-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test each adapter
	adapters := []struct {
		name string
		ext  string
	}{
		{"claude", ".md"},
		{"gemini", ".toml"},
		{"codex", ".md"},
	}

	for _, adapterInfo := range adapters {
		t.Run(adapterInfo.name, func(t *testing.T) {
			outputDir := filepath.Join(tmpDir, adapterInfo.name)
			err = validation.WriteAreasToDir(areas, outputDir, adapterInfo.name)
			if err != nil {
				t.Fatalf("Failed to write %s files: %v", adapterInfo.name, err)
			}

			// Verify output files exist
			for _, area := range areas {
				expectedFile := filepath.Join(outputDir, area.Name+adapterInfo.ext)
				if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
					t.Errorf("Expected %s file not created: %s", adapterInfo.name, expectedFile)
				} else {
					t.Logf("Created %s: %s", adapterInfo.name, expectedFile)
				}
			}
		})
	}
}

func TestReadWriteCanonicalFile(t *testing.T) {
	area := testAreas()[0] // Use QA area

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "validation-canonical-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write canonical file
	path := filepath.Join(tmpDir, "qa.json")
	err = validation.WriteCanonicalFile(area, path)
	if err != nil {
		t.Fatalf("Failed to write canonical file: %v", err)
	}

	// Read it back
	readArea, err := validation.ReadCanonicalFile(path)
	if err != nil {
		t.Fatalf("Failed to read canonical file: %v", err)
	}

	// Verify fields match
	if readArea.Name != area.Name {
		t.Errorf("Name mismatch: got %q, want %q", readArea.Name, area.Name)
	}
	if readArea.Description != area.Description {
		t.Errorf("Description mismatch: got %q, want %q", readArea.Description, area.Description)
	}
	if len(readArea.Checks) != len(area.Checks) {
		t.Errorf("Checks count mismatch: got %d, want %d", len(readArea.Checks), len(area.Checks))
	}
}

func TestReadCanonicalDir(t *testing.T) {
	areas := testAreas()

	// Create temp directory with JSON files
	tmpDir, err := os.MkdirTemp("", "validation-dir-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write all areas
	for _, area := range areas {
		path := filepath.Join(tmpDir, area.Name+".json")
		err = validation.WriteCanonicalFile(area, path)
		if err != nil {
			t.Fatalf("Failed to write %s: %v", area.Name, err)
		}
	}

	// Read directory
	readAreas, err := validation.ReadCanonicalDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read canonical dir: %v", err)
	}

	if len(readAreas) != len(areas) {
		t.Errorf("Area count mismatch: got %d, want %d", len(readAreas), len(areas))
	}

	// Check all expected areas are present
	areaNames := make(map[string]bool)
	for _, area := range readAreas {
		areaNames[area.Name] = true
	}

	for _, area := range areas {
		if !areaNames[area.Name] {
			t.Errorf("Expected area %q not found", area.Name)
		}
	}
}
