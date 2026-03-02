// Package kiro provides an adapter for AWS Kiro CLI MCP configuration.
//
// Kiro uses a format similar to Claude, with mcpServers as the root key.
// It supports both local (stdio) and remote (HTTP/SSE) servers.
//
// Key features:
//   - Environment variable substitution using ${ENV_VAR} syntax
//   - disabled field to disable servers without removing them
//   - Remote MCP with headers for authentication
//
// File locations:
//   - Workspace: <project>/.kiro/settings/mcp.json
//   - User: ~/.kiro/settings/mcp.json
package kiro

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/plexusone/assistantkit/mcp/core"
)

const (
	// AdapterName is the identifier for this adapter.
	AdapterName = "kiro"

	// SettingsDir is the settings directory name.
	SettingsDir = "settings"

	// ConfigFileName is the MCP config file name.
	ConfigFileName = "mcp.json"

	// ProjectConfigDir is the project config directory.
	ProjectConfigDir = ".kiro"
)

// Adapter implements core.Adapter for Kiro.
type Adapter struct{}

// NewAdapter creates a new Kiro adapter.
func NewAdapter() *Adapter {
	return &Adapter{}
}

// Name returns the adapter name.
func (a *Adapter) Name() string {
	return AdapterName
}

// DefaultPaths returns the default config file paths for Kiro.
func (a *Adapter) DefaultPaths() []string {
	paths := []string{
		filepath.Join(ProjectConfigDir, SettingsDir, ConfigFileName),
	}

	// User config
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ProjectConfigDir, SettingsDir, ConfigFileName))
	}

	return paths
}

// Parse parses Kiro config data into the canonical format.
func (a *Adapter) Parse(data []byte) (*core.Config, error) {
	var kiroCfg Config
	if err := json.Unmarshal(data, &kiroCfg); err != nil {
		return nil, &core.ParseError{Format: AdapterName, Err: err}
	}
	return a.ToCore(&kiroCfg), nil
}

// Marshal converts canonical config to Kiro format.
func (a *Adapter) Marshal(cfg *core.Config) ([]byte, error) {
	kiroCfg := a.FromCore(cfg)
	return json.MarshalIndent(kiroCfg, "", "  ")
}

// ReadFile reads a Kiro config file.
func (a *Adapter) ReadFile(path string) (*core.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &core.ParseError{Format: AdapterName, Path: path, Err: err}
	}
	cfg, err := a.Parse(data)
	if err != nil {
		if pe, ok := err.(*core.ParseError); ok {
			pe.Path = path
		}
		return nil, err
	}
	return cfg, nil
}

// WriteFile writes canonical config to a Kiro format file.
func (a *Adapter) WriteFile(cfg *core.Config, path string) error {
	data, err := a.Marshal(cfg)
	if err != nil {
		return &core.WriteError{Format: AdapterName, Path: path, Err: err}
	}
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return &core.WriteError{Format: AdapterName, Path: path, Err: err}
	}
	if err := os.WriteFile(path, data, core.DefaultFileMode); err != nil {
		return &core.WriteError{Format: AdapterName, Path: path, Err: err}
	}
	return nil
}

// ToCore converts Kiro config to canonical format.
func (a *Adapter) ToCore(kiroCfg *Config) *core.Config {
	cfg := core.NewConfig()

	for name, server := range kiroCfg.MCPServers {
		coreServer := core.Server{
			Command: server.Command,
			Args:    server.Args,
			Env:     server.Env,
			URL:     server.URL,
			Headers: server.Headers,
		}

		// Convert disabled to enabled
		if server.Disabled {
			enabled := false
			coreServer.Enabled = &enabled
		}

		// Infer transport type
		if server.Command != "" {
			coreServer.Transport = core.TransportStdio
		} else if server.URL != "" {
			coreServer.Transport = core.TransportHTTP
		}

		cfg.Servers[name] = coreServer
	}

	return cfg
}

// FromCore converts canonical config to Kiro format.
func (a *Adapter) FromCore(cfg *core.Config) *Config {
	kiroCfg := &Config{
		MCPServers: make(map[string]ServerConfig),
	}

	for name, server := range cfg.Servers {
		kiroServer := ServerConfig{
			Command: server.Command,
			Args:    server.Args,
			Env:     server.Env,
			URL:     server.URL,
			Headers: server.Headers,
		}

		// Convert enabled to disabled
		if server.Enabled != nil && !*server.Enabled {
			kiroServer.Disabled = true
		}

		kiroCfg.MCPServers[name] = kiroServer
	}

	return kiroCfg
}

// WorkspaceConfigPath returns the workspace config path for a given project root.
func WorkspaceConfigPath(projectRoot string) string {
	return filepath.Join(projectRoot, ProjectConfigDir, SettingsDir, ConfigFileName)
}

// UserConfigPath returns the user-level config path.
func UserConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ProjectConfigDir, SettingsDir, ConfigFileName), nil
}

// ReadWorkspaceConfig reads the workspace-level .kiro/settings/mcp.json.
func ReadWorkspaceConfig() (*core.Config, error) {
	adapter := NewAdapter()
	return adapter.ReadFile(filepath.Join(ProjectConfigDir, SettingsDir, ConfigFileName))
}

// ReadUserConfig reads the user-level ~/.kiro/settings/mcp.json.
func ReadUserConfig() (*core.Config, error) {
	path, err := UserConfigPath()
	if err != nil {
		return nil, err
	}
	adapter := NewAdapter()
	return adapter.ReadFile(path)
}

// WriteWorkspaceConfig writes to the workspace-level .kiro/settings/mcp.json.
func WriteWorkspaceConfig(cfg *core.Config) error {
	adapter := NewAdapter()
	return adapter.WriteFile(cfg, filepath.Join(ProjectConfigDir, SettingsDir, ConfigFileName))
}

// WriteUserConfig writes to the user-level ~/.kiro/settings/mcp.json.
func WriteUserConfig(cfg *core.Config) error {
	path, err := UserConfigPath()
	if err != nil {
		return err
	}
	adapter := NewAdapter()
	return adapter.WriteFile(cfg, path)
}

// init registers the adapter with the default registry.
func init() {
	core.Register(NewAdapter())
}
