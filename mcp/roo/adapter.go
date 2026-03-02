// Package roo provides an adapter for Roo Code VS Code extension MCP configuration.
//
// Roo Code uses a format similar to Cline, with config at:
//   - Global: mcp_settings.json in VS Code globalStorage
//   - Workspace: .roo/mcp.json
package roo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"

	"github.com/plexusone/assistantkit/mcp/core"
)

const (
	// AdapterName is the identifier for this adapter.
	AdapterName = "roo"

	// GlobalConfigFileName is the global config file name.
	GlobalConfigFileName = "mcp_settings.json"

	// WorkspaceConfigDir is the workspace config directory.
	WorkspaceConfigDir = ".roo"

	// WorkspaceConfigFileName is the workspace config file name.
	WorkspaceConfigFileName = "mcp.json"
)

// Adapter implements core.Adapter for Roo Code.
type Adapter struct{}

// NewAdapter creates a new Roo Code adapter.
func NewAdapter() *Adapter {
	return &Adapter{}
}

// Name returns the adapter name.
func (a *Adapter) Name() string {
	return AdapterName
}

// DefaultPaths returns the default config file paths for Roo Code.
func (a *Adapter) DefaultPaths() []string {
	paths := []string{
		filepath.Join(WorkspaceConfigDir, WorkspaceConfigFileName),
	}

	// VS Code globalStorage location varies by OS
	if home, err := os.UserHomeDir(); err == nil {
		var vscodePath string
		switch runtime.GOOS {
		case "darwin":
			vscodePath = filepath.Join(home, "Library/Application Support/Code/User/globalStorage/rooveterinaryinc.roo-cline")
		case "linux":
			vscodePath = filepath.Join(home, ".config/Code/User/globalStorage/rooveterinaryinc.roo-cline")
		case "windows":
			vscodePath = filepath.Join(home, "AppData/Roaming/Code/User/globalStorage/rooveterinaryinc.roo-cline")
		}
		if vscodePath != "" {
			paths = append(paths, filepath.Join(vscodePath, GlobalConfigFileName))
		}
	}

	return paths
}

// Parse parses Roo Code config data into the canonical format.
func (a *Adapter) Parse(data []byte) (*core.Config, error) {
	var rooCfg Config
	if err := json.Unmarshal(data, &rooCfg); err != nil {
		return nil, &core.ParseError{Format: AdapterName, Err: err}
	}
	return a.ToCore(&rooCfg), nil
}

// Marshal converts canonical config to Roo Code format.
func (a *Adapter) Marshal(cfg *core.Config) ([]byte, error) {
	rooCfg := a.FromCore(cfg)
	return json.MarshalIndent(rooCfg, "", "  ")
}

// ReadFile reads a Roo Code config file.
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

// WriteFile writes canonical config to a Roo Code format file.
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

// ToCore converts Roo Code config to canonical format.
func (a *Adapter) ToCore(rooCfg *Config) *core.Config {
	cfg := core.NewConfig()

	for name, server := range rooCfg.MCPServers {
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

// FromCore converts canonical config to Roo Code format.
func (a *Adapter) FromCore(cfg *core.Config) *Config {
	rooCfg := &Config{
		MCPServers: make(map[string]ServerConfig),
	}

	for name, server := range cfg.Servers {
		rooServer := ServerConfig{
			Command:     server.Command,
			Args:        server.Args,
			Env:         server.Env,
			URL:         server.URL,
			Headers:     server.Headers,
			AlwaysAllow: server.AlwaysAllow,
			Disabled:    !server.IsEnabled(),
		}

		if server.Transport != "" {
			rooServer.Type = server.Transport.String()
		}

		rooCfg.MCPServers[name] = rooServer
	}

	return rooCfg
}

// init registers the adapter with the default registry.
func init() {
	core.Register(NewAdapter())
}
