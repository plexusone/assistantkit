// Package codex provides an adapter for OpenAI Codex CLI MCP configuration.
//
// Codex uses TOML format instead of JSON, with additional features:
//   - bearer_token_env_var for OAuth
//   - enabled_tools / disabled_tools for tool filtering
//   - startup_timeout_sec / tool_timeout_sec for timeouts
//   - enabled flag to disable without deleting
//
// File location: ~/.codex/config.toml
package codex

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
	"github.com/plexusone/assistantkit/mcp/core"
)

const (
	// AdapterName is the identifier for this adapter.
	AdapterName = "codex"

	// ConfigDir is the config directory relative to home.
	ConfigDir = ".codex"

	// ConfigFileName is the config file name.
	ConfigFileName = "config.toml"
)

// Adapter implements core.Adapter for Codex.
type Adapter struct{}

// NewAdapter creates a new Codex adapter.
func NewAdapter() *Adapter {
	return &Adapter{}
}

// Name returns the adapter name.
func (a *Adapter) Name() string {
	return AdapterName
}

// DefaultPaths returns the default config file paths for Codex.
func (a *Adapter) DefaultPaths() []string {
	if home, err := os.UserHomeDir(); err == nil {
		return []string{filepath.Join(home, ConfigDir, ConfigFileName)}
	}
	return []string{}
}

// Parse parses Codex TOML config data into the canonical format.
func (a *Adapter) Parse(data []byte) (*core.Config, error) {
	var codexCfg Config
	if err := toml.Unmarshal(data, &codexCfg); err != nil {
		return nil, &core.ParseError{Format: AdapterName, Err: err}
	}
	return a.ToCore(&codexCfg), nil
}

// Marshal converts canonical config to Codex TOML format.
func (a *Adapter) Marshal(cfg *core.Config) ([]byte, error) {
	codexCfg := a.FromCore(cfg)
	return toml.Marshal(codexCfg)
}

// ReadFile reads a Codex config file.
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

// WriteFile writes canonical config to a Codex TOML format file.
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

// ToCore converts Codex config to canonical format.
func (a *Adapter) ToCore(codexCfg *Config) *core.Config {
	cfg := core.NewConfig()

	for name, server := range codexCfg.MCPServers {
		coreServer := core.Server{
			Command:           server.Command,
			Args:              server.Args,
			Env:               server.Env,
			Cwd:               server.Cwd,
			URL:               server.URL,
			Headers:           server.HTTPHeaders,
			BearerTokenEnvVar: server.BearerTokenEnvVar,
			EnabledTools:      server.EnabledTools,
			DisabledTools:     server.DisabledTools,
			StartupTimeoutSec: server.StartupTimeoutSec,
			ToolTimeoutSec:    server.ToolTimeoutSec,
			Enabled:           server.Enabled,
		}

		// Infer transport
		if server.Command != "" {
			coreServer.Transport = core.TransportStdio
		} else if server.URL != "" {
			coreServer.Transport = core.TransportHTTP
		}

		cfg.Servers[name] = coreServer
	}

	return cfg
}

// FromCore converts canonical config to Codex format.
func (a *Adapter) FromCore(cfg *core.Config) *Config {
	codexCfg := &Config{
		MCPServers: make(map[string]ServerConfig),
	}

	for name, server := range cfg.Servers {
		codexServer := ServerConfig{
			Command:           server.Command,
			Args:              server.Args,
			Env:               server.Env,
			Cwd:               server.Cwd,
			URL:               server.URL,
			HTTPHeaders:       server.Headers,
			BearerTokenEnvVar: server.BearerTokenEnvVar,
			EnabledTools:      server.EnabledTools,
			DisabledTools:     server.DisabledTools,
			StartupTimeoutSec: server.StartupTimeoutSec,
			ToolTimeoutSec:    server.ToolTimeoutSec,
			Enabled:           server.Enabled,
		}

		codexCfg.MCPServers[name] = codexServer
	}

	return codexCfg
}

// ConfigPath returns the default Codex config path.
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ConfigDir, ConfigFileName), nil
}

// ReadConfig reads the Codex config file.
func ReadConfig() (*core.Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}
	adapter := NewAdapter()
	return adapter.ReadFile(path)
}

// WriteConfig writes to the Codex config file.
func WriteConfig(cfg *core.Config) error {
	path, err := ConfigPath()
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
