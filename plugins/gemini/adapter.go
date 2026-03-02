// Package gemini provides the Gemini CLI extension adapter.
package gemini

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/plexusone/assistantkit/plugins/core"
)

func init() {
	core.Register(&Adapter{})
}

// Adapter converts between canonical Plugin and Gemini CLI extension format.
type Adapter struct{}

// Name returns the adapter identifier.
func (a *Adapter) Name() string {
	return "gemini"
}

// DefaultPaths returns default file paths for Gemini extension manifest.
func (a *Adapter) DefaultPaths() []string {
	return []string{
		"gemini-extension.json",
	}
}

// Parse converts Gemini extension JSON bytes to canonical Plugin.
func (a *Adapter) Parse(data []byte) (*core.Plugin, error) {
	var ge GeminiExtension
	if err := json.Unmarshal(data, &ge); err != nil {
		return nil, &core.ParseError{Format: "gemini", Err: err}
	}
	return ge.ToCanonical(), nil
}

// Marshal converts canonical Plugin to Gemini extension JSON bytes.
func (a *Adapter) Marshal(plugin *core.Plugin) ([]byte, error) {
	ge := FromCanonical(plugin)
	data, err := json.MarshalIndent(ge, "", "  ")
	if err != nil {
		return nil, &core.MarshalError{Format: "gemini", Err: err}
	}
	return append(data, '\n'), nil
}

// ReadFile reads a Gemini extension JSON file and returns canonical Plugin.
func (a *Adapter) ReadFile(path string) (*core.Plugin, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &core.ReadError{Path: path, Err: err}
	}

	plugin, err := a.Parse(data)
	if err != nil {
		if pe, ok := err.(*core.ParseError); ok {
			pe.Path = path
		}
		return nil, err
	}

	return plugin, nil
}

// WriteFile writes canonical Plugin to a Gemini extension JSON file.
func (a *Adapter) WriteFile(plugin *core.Plugin, path string) error {
	data, err := a.Marshal(plugin)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, core.DefaultDirMode); err != nil {
			return &core.WriteError{Path: path, Err: err}
		}
	}

	if err := os.WriteFile(path, data, core.DefaultFileMode); err != nil {
		return &core.WriteError{Path: path, Err: err}
	}

	return nil
}

// WritePlugin writes the complete Gemini extension structure to the given directory.
func (a *Adapter) WritePlugin(plugin *core.Plugin, dir string) error {
	// Create extension directory
	if err := os.MkdirAll(dir, core.DefaultDirMode); err != nil {
		return &core.WriteError{Path: dir, Err: err}
	}

	// Write gemini-extension.json
	manifestPath := filepath.Join(dir, "gemini-extension.json")
	if err := a.WriteFile(plugin, manifestPath); err != nil {
		return err
	}

	// Write GEMINI.md context file if context is provided
	if plugin.Context != "" {
		contextPath := filepath.Join(dir, "GEMINI.md")
		if err := os.WriteFile(contextPath, []byte(plugin.Context), core.DefaultFileMode); err != nil {
			return &core.WriteError{Path: contextPath, Err: err}
		}
	}

	// Create component directories if specified
	if plugin.Commands != "" {
		commandsDir := filepath.Join(dir, "commands")
		if err := os.MkdirAll(commandsDir, core.DefaultDirMode); err != nil {
			return &core.WriteError{Path: commandsDir, Err: err}
		}
	}

	if plugin.Hooks != "" {
		hooksDir := filepath.Join(dir, "hooks")
		if err := os.MkdirAll(hooksDir, core.DefaultDirMode); err != nil {
			return &core.WriteError{Path: hooksDir, Err: err}
		}
	}

	return nil
}
