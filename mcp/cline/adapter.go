// Package cline provides an adapter for Cline VS Code extension MCP configuration.
//
// Cline uses a format similar to Claude, with additional fields:
//   - alwaysAllow: tools that don't require user approval
//   - disabled: whether the server is disabled
//
// File location: cline_mcp_settings.json (in VS Code settings)
package cline

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"

	"github.com/plexusone/assistantkit/mcp/core"
)

const (
	// AdapterName is the identifier for this adapter.
	AdapterName = "cline"

	// ConfigFileName is the config file name.
	ConfigFileName = "cline_mcp_settings.json"
)

// Adapter implements core.Adapter for Cline.
type Adapter struct{}

// NewAdapter creates a new Cline adapter.
func NewAdapter() *Adapter {
	return &Adapter{}
}

// Name returns the adapter name.
func (a *Adapter) Name() string {
	return AdapterName
}

// DefaultPaths returns the default config file paths for Cline.
func (a *Adapter) DefaultPaths() []string {
	paths := []string{}

	// VS Code user settings location varies by OS
	if home, err := os.UserHomeDir(); err == nil {
		var vscodePath string
		switch runtime.GOOS {
		case "darwin":
			vscodePath = filepath.Join(home, "Library/Application Support/Code/User/globalStorage/saoudrizwan.claude-dev")
		case "linux":
			vscodePath = filepath.Join(home, ".config/Code/User/globalStorage/saoudrizwan.claude-dev")
		case "windows":
			vscodePath = filepath.Join(home, "AppData/Roaming/Code/User/globalStorage/saoudrizwan.claude-dev")
		}
		if vscodePath != "" {
			paths = append(paths, filepath.Join(vscodePath, ConfigFileName))
		}
	}

	return paths
}

// Parse parses Cline config data into the canonical format.
func (a *Adapter) Parse(data []byte) (*core.Config, error) {
	var clineCfg Config
	if err := json.Unmarshal(data, &clineCfg); err != nil {
		return nil, &core.ParseError{Format: AdapterName, Err: err}
	}
	return a.ToCore(&clineCfg), nil
}

// Marshal converts canonical config to Cline format.
func (a *Adapter) Marshal(cfg *core.Config) ([]byte, error) {
	clineCfg := a.FromCore(cfg)
	return json.MarshalIndent(clineCfg, "", "  ")
}

// ReadFile reads a Cline config file.
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

// WriteFile writes canonical config to a Cline format file.
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

// ToCore converts Cline config to canonical format.
func (a *Adapter) ToCore(clineCfg *Config) *core.Config {
	cfg := core.NewConfig()

	for name, server := range clineCfg.MCPServers {
		coreServer := core.Server{
			Command:     server.Command,
			Args:        server.Args,
			Env:         server.Env,
			URL:         server.URL,
			Headers:     server.Headers,
			AlwaysAllow: server.AlwaysAllow,
		}

		// Handle disabled state
		if server.Disabled {
			enabled := false
			coreServer.Enabled = &enabled
		}

		// Set transport type
		switch server.Type {
		case "stdio":
			coreServer.Transport = core.TransportStdio
		case "http":
			coreServer.Transport = core.TransportHTTP
		case "sse":
			coreServer.Transport = core.TransportSSE
		default:
			if server.Command != "" {
				coreServer.Transport = core.TransportStdio
			} else if server.URL != "" {
				coreServer.Transport = core.TransportHTTP
			}
		}

		cfg.Servers[name] = coreServer
	}

	return cfg
}

// FromCore converts canonical config to Cline format.
func (a *Adapter) FromCore(cfg *core.Config) *Config {
	clineCfg := &Config{
		MCPServers: make(map[string]ServerConfig),
	}

	for name, server := range cfg.Servers {
		clineServer := ServerConfig{
			Command:     server.Command,
			Args:        server.Args,
			Env:         server.Env,
			URL:         server.URL,
			Headers:     server.Headers,
			AlwaysAllow: server.AlwaysAllow,
			Disabled:    !server.IsEnabled(),
		}

		if server.Transport != "" {
			clineServer.Type = server.Transport.String()
		}

		clineCfg.MCPServers[name] = clineServer
	}

	return clineCfg
}

// init registers the adapter with the default registry.
func init() {
	core.Register(NewAdapter())
}
