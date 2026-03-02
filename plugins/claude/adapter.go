// Package claude provides the Claude Code plugin adapter.
package claude

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/plexusone/assistantkit/plugins/core"
)

func init() {
	core.Register(&Adapter{})
}

// Adapter converts between canonical Plugin and Claude Code plugin format.
type Adapter struct{}

// Name returns the adapter identifier.
func (a *Adapter) Name() string {
	return "claude"
}

// DefaultPaths returns default file paths for Claude plugin manifest.
func (a *Adapter) DefaultPaths() []string {
	return []string{
		".claude-plugin/plugin.json",
	}
}

// Parse converts Claude plugin.json bytes to canonical Plugin.
func (a *Adapter) Parse(data []byte) (*core.Plugin, error) {
	var cp ClaudePlugin
	if err := json.Unmarshal(data, &cp); err != nil {
		return nil, &core.ParseError{Format: "claude", Err: err}
	}
	return cp.ToCanonical(), nil
}

// Marshal converts canonical Plugin to Claude plugin.json bytes.
func (a *Adapter) Marshal(plugin *core.Plugin) ([]byte, error) {
	cp := FromCanonical(plugin)
	data, err := json.MarshalIndent(cp, "", "  ")
	if err != nil {
		return nil, &core.MarshalError{Format: "claude", Err: err}
	}
	return append(data, '\n'), nil
}

// ReadFile reads a Claude plugin.json file and returns canonical Plugin.
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

// WriteFile writes canonical Plugin to a Claude plugin.json file.
func (a *Adapter) WriteFile(plugin *core.Plugin, path string) error {
	data, err := a.Marshal(plugin)
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

// WritePlugin writes the complete Claude plugin structure to the given directory.
func (a *Adapter) WritePlugin(plugin *core.Plugin, dir string) error {
	// Create .claude-plugin directory
	pluginDir := filepath.Join(dir, ".claude-plugin")
	if err := os.MkdirAll(pluginDir, core.DefaultDirMode); err != nil {
		return &core.WriteError{Path: pluginDir, Err: err}
	}

	// Write plugin.json
	manifestPath := filepath.Join(pluginDir, "plugin.json")
	if err := a.WriteFile(plugin, manifestPath); err != nil {
		return err
	}

	// Create component directories if specified
	if plugin.Commands != "" {
		commandsDir := filepath.Join(dir, "commands")
		if err := os.MkdirAll(commandsDir, core.DefaultDirMode); err != nil {
			return &core.WriteError{Path: commandsDir, Err: err}
		}
	}

	if plugin.Skills != "" {
		skillsDir := filepath.Join(dir, "skills")
		if err := os.MkdirAll(skillsDir, core.DefaultDirMode); err != nil {
			return &core.WriteError{Path: skillsDir, Err: err}
		}
	}

	if plugin.Agents != "" {
		agentsDir := filepath.Join(dir, "agents")
		if err := os.MkdirAll(agentsDir, core.DefaultDirMode); err != nil {
			return &core.WriteError{Path: agentsDir, Err: err}
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
