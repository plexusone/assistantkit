package kiro

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/plexusone/assistantkit/agents/core"
)

func TestAdapter_Name(t *testing.T) {
	adapter := &Adapter{}
	if got := adapter.Name(); got != "kiro" {
		t.Errorf("Name() = %q, want %q", got, "kiro")
	}
}

func TestAdapter_FileExtension(t *testing.T) {
	adapter := &Adapter{}
	if got := adapter.FileExtension(); got != ".json" {
		t.Errorf("FileExtension() = %q, want %q", got, ".json")
	}
}

func TestAdapter_DefaultDir(t *testing.T) {
	adapter := &Adapter{}
	if got := adapter.DefaultDir(); got != "agents" {
		t.Errorf("DefaultDir() = %q, want %q", got, "agents")
	}
}

func TestAdapter_Parse(t *testing.T) {
	adapter := &Adapter{}

	input := `{
  "name": "release-agent",
  "description": "Automates software releases",
  "tools": ["fs_read", "fs_write", "execute_bash"],
  "allowedTools": ["fs_read"],
  "resources": ["file://README.md"],
  "prompt": "You are a release automation specialist.",
  "model": "claude-sonnet-4"
}`

	agent, err := adapter.Parse([]byte(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if agent.Name != "release-agent" {
		t.Errorf("Name = %q, want %q", agent.Name, "release-agent")
	}

	if agent.Description != "Automates software releases" {
		t.Errorf("Description = %q, want %q", agent.Description, "Automates software releases")
	}

	if agent.Model != "sonnet" {
		t.Errorf("Model = %q, want %q", agent.Model, "sonnet")
	}

	if agent.Instructions != "You are a release automation specialist." {
		t.Errorf("Instructions = %q, want %q", agent.Instructions, "You are a release automation specialist.")
	}

	// Check tools mapping
	expectedTools := []string{"Read", "Write", "Bash"}
	if len(agent.Tools) != len(expectedTools) {
		t.Errorf("Tools count = %d, want %d", len(agent.Tools), len(expectedTools))
	}
	for i, tool := range expectedTools {
		if i < len(agent.Tools) && agent.Tools[i] != tool {
			t.Errorf("Tools[%d] = %q, want %q", i, agent.Tools[i], tool)
		}
	}
}

func TestAdapter_Marshal(t *testing.T) {
	adapter := &Adapter{}

	agent := &core.Agent{
		Name:         "test-agent",
		Description:  "A test agent",
		Model:        "sonnet",
		Tools:        []string{"Read", "Write", "Bash", "Grep"},
		Skills:       []string{"version-analysis"},
		Instructions: "You are a helpful assistant.",
	}

	data, err := adapter.Marshal(agent)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	output := string(data)

	// Check key fields
	if !strings.Contains(output, `"name": "test-agent"`) {
		t.Error("Output should contain name field")
	}

	if !strings.Contains(output, `"description": "A test agent"`) {
		t.Error("Output should contain description field")
	}

	if !strings.Contains(output, `"model": "claude-sonnet-4"`) {
		t.Error("Output should contain model field with Kiro model name")
	}

	if !strings.Contains(output, `"prompt": "You are a helpful assistant."`) {
		t.Error("Output should contain prompt field")
	}

	// Check tools mapping
	if !strings.Contains(output, `"fs_read"`) {
		t.Error("Output should contain fs_read tool")
	}
	if !strings.Contains(output, `"execute_bash"`) {
		t.Error("Output should contain execute_bash tool (mapped from Bash)")
	}

	// Check skills mapped to resources
	if !strings.Contains(output, `"resources"`) {
		t.Error("Output should contain resources field")
	}
	if !strings.Contains(output, `"file://.kiro/steering/version-analysis.md"`) {
		t.Error("Output should map skills to steering files")
	}
}

func TestAdapter_RoundTrip(t *testing.T) {
	adapter := &Adapter{}

	original := &core.Agent{
		Name:         "round-trip-agent",
		Description:  "Tests round-trip conversion",
		Model:        "opus",
		Tools:        []string{"Read", "Write"},
		Instructions: "System instructions here.",
	}

	// Marshal to Kiro format
	data, err := adapter.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Parse back to canonical
	parsed, err := adapter.Parse(data)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Verify fields preserved
	if parsed.Name != original.Name {
		t.Errorf("Name = %q, want %q", parsed.Name, original.Name)
	}
	if parsed.Description != original.Description {
		t.Errorf("Description = %q, want %q", parsed.Description, original.Description)
	}
	if parsed.Model != original.Model {
		t.Errorf("Model = %q, want %q", parsed.Model, original.Model)
	}
	if parsed.Instructions != original.Instructions {
		t.Errorf("Instructions = %q, want %q", parsed.Instructions, original.Instructions)
	}
}

func TestAdapter_WriteFile_ReadFile(t *testing.T) {
	adapter := &Adapter{}

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "kiro-agent-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	agent := &core.Agent{
		Name:         "file-test-agent",
		Description:  "Tests file operations",
		Model:        "haiku",
		Tools:        []string{"Read", "Grep", "Glob"},
		Instructions: "You help with file operations.",
	}

	// Write to file
	path := filepath.Join(tmpDir, "file-test-agent.json")
	if err := adapter.WriteFile(agent, path); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("WriteFile() did not create file")
	}

	// Read back
	readAgent, err := adapter.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	// Verify content
	if readAgent.Name != agent.Name {
		t.Errorf("Name = %q, want %q", readAgent.Name, agent.Name)
	}
	if readAgent.Description != agent.Description {
		t.Errorf("Description = %q, want %q", readAgent.Description, agent.Description)
	}
}

func TestModelMapping(t *testing.T) {
	tests := []struct {
		kiroModel      string
		canonicalModel core.Model
	}{
		{"claude-sonnet-4", core.ModelSonnet},
		{"claude-4-sonnet", core.ModelSonnet},
		{"claude-opus-4", core.ModelOpus},
		{"claude-4-opus", core.ModelOpus},
		{"claude-haiku", core.ModelHaiku},
		{"claude-3-haiku", core.ModelHaiku},
		{"unknown-model", core.Model("unknown-model")},
	}

	for _, tt := range tests {
		got := mapKiroModelToCanonical(tt.kiroModel)
		if got != tt.canonicalModel {
			t.Errorf("mapKiroModelToCanonical(%q) = %q, want %q", tt.kiroModel, got, tt.canonicalModel)
		}
	}
}

func TestToolMapping(t *testing.T) {
	kiroTools := []string{"fs_read", "fs_write", "execute_bash", "web_search", "grep"}
	expected := []string{"Read", "Write", "Bash", "WebSearch", "Grep"}

	got := mapKiroToolsToCanonical(kiroTools)

	if len(got) != len(expected) {
		t.Fatalf("Tool count = %d, want %d", len(got), len(expected))
	}

	for i, tool := range expected {
		if got[i] != tool {
			t.Errorf("Tool[%d] = %q, want %q", i, got[i], tool)
		}
	}
}

func TestReverseToolMapping(t *testing.T) {
	canonicalTools := []string{"Read", "Write", "Bash", "WebFetch", "Edit"}
	// Edit maps to fs_write which is deduplicated with Write's fs_write
	expected := []string{"fs_read", "fs_write", "execute_bash", "web_fetch"}

	got := mapCanonicalToolsToKiro(canonicalTools)

	if len(got) != len(expected) {
		t.Fatalf("Tool count = %d, want %d", len(got), len(expected))
	}

	for i, tool := range expected {
		if got[i] != tool {
			t.Errorf("Tool[%d] = %q, want %q", i, got[i], tool)
		}
	}
}
