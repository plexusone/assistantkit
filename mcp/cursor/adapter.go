// Package cursor provides an adapter for Cursor IDE MCP configuration.
//
// Cursor uses the same format as Claude Desktop, with config files at:
//   - Global: ~/.cursor/mcp.json
//   - Project: .cursor/mcp.json
package cursor

import (
	"os"
	"path/filepath"

	"github.com/plexusone/assistantkit/mcp/claude"
	"github.com/plexusone/assistantkit/mcp/core"
)

const (
	// AdapterName is the identifier for this adapter.
	AdapterName = "cursor"

	// GlobalConfigDir is the global config directory name.
	GlobalConfigDir = ".cursor"

	// ConfigFileName is the config file name.
	ConfigFileName = "mcp.json"
)

// Adapter implements core.Adapter for Cursor IDE.
type Adapter struct {
	claudeAdapter *claude.Adapter
}

// NewAdapter creates a new Cursor adapter.
func NewAdapter() *Adapter {
	return &Adapter{
		claudeAdapter: claude.NewAdapter(),
	}
}

// Name returns the adapter name.
func (a *Adapter) Name() string {
	return AdapterName
}

// DefaultPaths returns the default config file paths for Cursor.
func (a *Adapter) DefaultPaths() []string {
	paths := []string{}

	// Global config
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, GlobalConfigDir, ConfigFileName))
	}

	// Project config
	paths = append(paths, filepath.Join(".cursor", ConfigFileName))

	return paths
}

// Parse parses Cursor config data into the canonical format.
func (a *Adapter) Parse(data []byte) (*core.Config, error) {
	cfg, err := a.claudeAdapter.Parse(data)
	if err != nil {
		if pe, ok := err.(*core.ParseError); ok {
			pe.Format = AdapterName
		}
		return nil, err
	}
	return cfg, nil
}

// Marshal converts canonical config to Cursor format.
func (a *Adapter) Marshal(cfg *core.Config) ([]byte, error) {
	return a.claudeAdapter.Marshal(cfg)
}

// ReadFile reads a Cursor config file.
func (a *Adapter) ReadFile(path string) (*core.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &core.ParseError{Format: AdapterName, Path: path, Err: err}
	}
	return a.Parse(data)
}

// WriteFile writes canonical config to a Cursor format file.
func (a *Adapter) WriteFile(cfg *core.Config, path string) error {
	data, err := a.Marshal(cfg)
	if err != nil {
		return &core.WriteError{Format: AdapterName, Path: path, Err: err}
	}
	if err := os.WriteFile(path, data, core.DefaultFileMode); err != nil {
		return &core.WriteError{Format: AdapterName, Path: path, Err: err}
	}
	return nil
}

// GlobalConfigPath returns the path to the global Cursor config.
func GlobalConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, GlobalConfigDir, ConfigFileName), nil
}

// ReadGlobalConfig reads the global ~/.cursor/mcp.json file.
func ReadGlobalConfig() (*core.Config, error) {
	path, err := GlobalConfigPath()
	if err != nil {
		return nil, err
	}
	adapter := NewAdapter()
	return adapter.ReadFile(path)
}

// WriteGlobalConfig writes to the global ~/.cursor/mcp.json file.
func WriteGlobalConfig(cfg *core.Config) error {
	path, err := GlobalConfigPath()
	if err != nil {
		return err
	}
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	adapter := NewAdapter()
	return adapter.WriteFile(cfg, path)
}

// init registers the adapter with the default registry.
func init() {
	core.Register(NewAdapter())
}
