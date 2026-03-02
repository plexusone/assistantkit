// Package windsurf provides an adapter for Windsurf (Codeium) MCP configuration.
//
// Windsurf uses a format similar to Claude, but with one key difference:
// HTTP servers use "serverUrl" instead of "url".
//
// File location: ~/.codeium/windsurf/mcp_config.json
package windsurf

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/plexusone/assistantkit/mcp/core"
)

const (
	// AdapterName is the identifier for this adapter.
	AdapterName = "windsurf"

	// ConfigDir is the config directory relative to home.
	ConfigDir = ".codeium/windsurf"

	// ConfigFileName is the config file name.
	ConfigFileName = "mcp_config.json"
)

// Adapter implements core.Adapter for Windsurf.
type Adapter struct{}

// NewAdapter creates a new Windsurf adapter.
func NewAdapter() *Adapter {
	return &Adapter{}
}

// Name returns the adapter name.
func (a *Adapter) Name() string {
	return AdapterName
}

// DefaultPaths returns the default config file paths for Windsurf.
func (a *Adapter) DefaultPaths() []string {
	if home, err := os.UserHomeDir(); err == nil {
		return []string{filepath.Join(home, ConfigDir, ConfigFileName)}
	}
	return []string{}
}

// Parse parses Windsurf config data into the canonical format.
func (a *Adapter) Parse(data []byte) (*core.Config, error) {
	var windsurfCfg Config
	if err := json.Unmarshal(data, &windsurfCfg); err != nil {
		return nil, &core.ParseError{Format: AdapterName, Err: err}
	}
	return a.ToCore(&windsurfCfg), nil
}

// Marshal converts canonical config to Windsurf format.
func (a *Adapter) Marshal(cfg *core.Config) ([]byte, error) {
	windsurfCfg := a.FromCore(cfg)
	return json.MarshalIndent(windsurfCfg, "", "  ")
}

// ReadFile reads a Windsurf config file.
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

// WriteFile writes canonical config to a Windsurf format file.
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

// ToCore converts Windsurf config to canonical format.
func (a *Adapter) ToCore(windsurfCfg *Config) *core.Config {
	cfg := core.NewConfig()

	for name, server := range windsurfCfg.MCPServers {
		coreServer := core.Server{
			Command:       server.Command,
			Args:          server.Args,
			Env:           server.Env,
			URL:           server.ServerURL, // Note: serverUrl -> URL
			Headers:       server.Headers,
			DisabledTools: server.DisabledTools,
		}

		// Set transport type
		switch server.Type {
		case "stdio":
			coreServer.Transport = core.TransportStdio
		case "http":
			coreServer.Transport = core.TransportHTTP
		default:
			if server.Command != "" {
				coreServer.Transport = core.TransportStdio
			} else if server.ServerURL != "" {
				coreServer.Transport = core.TransportHTTP
			}
		}

		cfg.Servers[name] = coreServer
	}

	return cfg
}

// FromCore converts canonical config to Windsurf format.
func (a *Adapter) FromCore(cfg *core.Config) *Config {
	windsurfCfg := &Config{
		MCPServers: make(map[string]ServerConfig),
	}

	for name, server := range cfg.Servers {
		windsurfServer := ServerConfig{
			Command:       server.Command,
			Args:          server.Args,
			Env:           server.Env,
			ServerURL:     server.URL, // Note: URL -> serverUrl
			Headers:       server.Headers,
			DisabledTools: server.DisabledTools,
		}

		if server.Transport != "" {
			windsurfServer.Type = server.Transport.String()
		}

		windsurfCfg.MCPServers[name] = windsurfServer
	}

	return windsurfCfg
}

// ConfigPath returns the default Windsurf config path.
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ConfigDir, ConfigFileName), nil
}

// ReadConfig reads the Windsurf config file.
func ReadConfig() (*core.Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}
	adapter := NewAdapter()
	return adapter.ReadFile(path)
}

// WriteConfig writes to the Windsurf config file.
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
