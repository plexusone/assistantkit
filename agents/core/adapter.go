package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	multiagentspec "github.com/plexusone/multi-agent-spec/sdk/go"
)

// DefaultFileMode is the default permission for generated files.
const DefaultFileMode fs.FileMode = 0600

// DefaultDirMode is the default permission for generated directories.
const DefaultDirMode fs.FileMode = 0700

// Adapter converts between canonical Agent definitions and tool-specific formats.
type Adapter interface {
	// Name returns the adapter identifier (e.g., "claude", "gemini").
	Name() string

	// FileExtension returns the file extension for agent files.
	FileExtension() string

	// DefaultDir returns the default directory name for agents.
	DefaultDir() string

	// Parse converts tool-specific bytes to canonical Agent.
	Parse(data []byte) (*Agent, error)

	// Marshal converts canonical Agent to tool-specific bytes.
	Marshal(agent *Agent) ([]byte, error)

	// ReadFile reads from path and returns canonical Agent.
	ReadFile(path string) (*Agent, error)

	// WriteFile writes canonical Agent to path.
	WriteFile(agent *Agent, path string) error
}

// Registry manages adapter registration and lookup.
type Registry struct {
	mu       sync.RWMutex
	adapters map[string]Adapter
}

// NewRegistry creates a new adapter registry.
func NewRegistry() *Registry {
	return &Registry{
		adapters: make(map[string]Adapter),
	}
}

// Register adds an adapter to the registry.
func (r *Registry) Register(adapter Adapter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.adapters[adapter.Name()] = adapter
}

// GetAdapter returns an adapter by name.
func (r *Registry) GetAdapter(name string) (Adapter, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	adapter, ok := r.adapters[name]
	return adapter, ok
}

// AdapterNames returns all registered adapter names sorted alphabetically.
func (r *Registry) AdapterNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.adapters))
	for name := range r.adapters {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// DefaultRegistry is the global adapter registry.
var DefaultRegistry = NewRegistry()

// Register adds an adapter to the default registry.
func Register(adapter Adapter) {
	DefaultRegistry.Register(adapter)
}

// GetAdapter returns an adapter from the default registry.
func GetAdapter(name string) (Adapter, bool) {
	return DefaultRegistry.GetAdapter(name)
}

// AdapterNames returns adapter names from the default registry.
func AdapterNames() []string {
	return DefaultRegistry.AdapterNames()
}

// ReadCanonicalFile reads a canonical agent file (Markdown + YAML frontmatter or JSON).
// The format is auto-detected based on file extension or content.
func ReadCanonicalFile(path string) (*Agent, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &ReadError{Path: path, Err: err}
	}

	// Detect format: if it starts with "---" or has .md extension, use multi-agent-spec loader
	ext := filepath.Ext(path)
	if ext == ".md" || (len(data) >= 3 && string(data[:3]) == "---") {
		agent, err := multiagentspec.ParseAgentMarkdown(data)
		if err != nil {
			return nil, &ParseError{Format: "markdown", Path: path, Err: err}
		}
		// Infer name from filename if not set
		if agent.Name == "" {
			base := filepath.Base(path)
			agent.Name = strings.TrimSuffix(base, filepath.Ext(base))
		}
		return agent, nil
	}

	// Fall back to JSON for .json files or other formats
	var agent Agent
	if err := json.Unmarshal(data, &agent); err != nil {
		return nil, &ParseError{Format: "canonical", Path: path, Err: err}
	}

	return &agent, nil
}

// WriteCanonicalFile writes a canonical agent file in Markdown + YAML frontmatter format.
func WriteCanonicalFile(agent *Agent, path string) error {
	data := MarshalMarkdownAgent(agent)

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, DefaultDirMode); err != nil {
		return &WriteError{Path: path, Err: err}
	}

	if err := os.WriteFile(path, data, DefaultFileMode); err != nil {
		return &WriteError{Path: path, Err: err}
	}

	return nil
}

// WriteCanonicalJSON writes a canonical agent.json file (for validation/schema compatibility).
func WriteCanonicalJSON(agent *Agent, path string) error {
	data, err := json.MarshalIndent(agent, "", "  ")
	if err != nil {
		return &MarshalError{Format: "canonical", Err: err}
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, DefaultDirMode); err != nil {
		return &WriteError{Path: path, Err: err}
	}

	if err := os.WriteFile(path, append(data, '\n'), DefaultFileMode); err != nil {
		return &WriteError{Path: path, Err: err}
	}

	return nil
}

// ReadCanonicalDir reads all agent files (.md or .json) from a directory.
// This delegates to multiagentspec.LoadAgentsFromDir for markdown files.
func ReadCanonicalDir(dir string) ([]*Agent, error) {
	// Try multiagentspec loader first (handles .md files properly)
	agents, err := multiagentspec.LoadAgentsFromDir(dir)
	if err != nil {
		return nil, &ReadError{Path: dir, Err: err}
	}

	// Also load any .json files that multiagentspec loader skips
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, &ReadError{Path: dir, Err: err}
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := filepath.Ext(entry.Name())
		if ext != ".json" {
			continue // .md files already loaded by multiagentspec
		}

		path := filepath.Join(dir, entry.Name())
		agent, err := ReadCanonicalFile(path)
		if err != nil {
			return nil, err
		}
		agents = append(agents, agent)
	}

	return agents, nil
}

// ParseMarkdownAgent parses a Markdown file with YAML frontmatter into an Agent.
// Deprecated: Use multiagentspec.ParseAgentMarkdown directly.
func ParseMarkdownAgent(data []byte, path string) (*Agent, error) {
	agent, err := multiagentspec.ParseAgentMarkdown(data)
	if err != nil {
		return nil, err
	}

	// Infer name from filename if not set
	if agent.Name == "" && path != "" {
		base := filepath.Base(path)
		agent.Name = strings.TrimSuffix(base, filepath.Ext(base))
	}

	return agent, nil
}

// MarshalMarkdownAgent converts an Agent to Markdown + YAML frontmatter bytes.
func MarshalMarkdownAgent(agent *Agent) []byte {
	var buf bytes.Buffer

	// Write YAML frontmatter
	buf.WriteString("---\n")
	buf.WriteString(fmt.Sprintf("name: %s\n", agent.Name))
	buf.WriteString(fmt.Sprintf("description: %s\n", agent.Description))

	if agent.Model != "" {
		buf.WriteString(fmt.Sprintf("model: %s\n", string(agent.Model)))
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

	if len(agent.Requires) > 0 {
		buf.WriteString(fmt.Sprintf("requires: [%s]\n", strings.Join(agent.Requires, ", ")))
	}

	buf.WriteString("---\n\n")

	// Write instructions directly (they already contain markdown formatting)
	if agent.Instructions != "" {
		buf.WriteString(agent.Instructions)
		buf.WriteString("\n")
	}

	return buf.Bytes()
}

// WriteAgentsToDir writes multiple agents to a directory using the specified adapter.
func WriteAgentsToDir(agents []*Agent, dir string, adapterName string) error {
	adapter, ok := GetAdapter(adapterName)
	if !ok {
		return &AdapterError{Name: adapterName}
	}

	if err := os.MkdirAll(dir, DefaultDirMode); err != nil {
		return &WriteError{Path: dir, Err: err}
	}

	for _, agent := range agents {
		filename := agent.Name + adapter.FileExtension()
		path := filepath.Join(dir, filename)
		if err := adapter.WriteFile(agent, path); err != nil {
			return err
		}
	}

	return nil
}
