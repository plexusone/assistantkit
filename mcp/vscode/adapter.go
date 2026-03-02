// Package vscode provides an adapter for VS Code / GitHub Copilot MCP configuration.
//
// VS Code uses a different format than Claude:
//   - Root key is "servers" (not "mcpServers")
//   - Has "inputs" section for secret management
//   - Requires explicit "type" field
//   - Supports "envFile" for loading env files
//
// File locations:
//   - Workspace: .vscode/mcp.json
//   - User: depends on OS and profile
package vscode

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"

	"github.com/plexusone/assistantkit/mcp/core"
)

const (
	// AdapterName is the identifier for this adapter.
	AdapterName = "vscode"

	// WorkspaceConfigDir is the workspace config directory.
	WorkspaceConfigDir = ".vscode"

	// ConfigFileName is the config file name.
	ConfigFileName = "mcp.json"
)

// Adapter implements core.Adapter for VS Code.
type Adapter struct{}

// NewAdapter creates a new VS Code adapter.
func NewAdapter() *Adapter {
	return &Adapter{}
}

// Name returns the adapter name.
func (a *Adapter) Name() string {
	return AdapterName
}

// DefaultPaths returns the default config file paths for VS Code.
func (a *Adapter) DefaultPaths() []string {
	paths := []string{
		filepath.Join(WorkspaceConfigDir, ConfigFileName),
	}

	// User config location varies by OS
	if home, err := os.UserHomeDir(); err == nil {
		var userPath string
		switch runtime.GOOS {
		case "darwin":
			userPath = filepath.Join(home, "Library/Application Support/Code/User")
		case "linux":
			userPath = filepath.Join(home, ".config/Code/User")
		case "windows":
			userPath = filepath.Join(home, "AppData/Roaming/Code/User")
		}
		if userPath != "" {
			paths = append(paths, filepath.Join(userPath, ConfigFileName))
		}
	}

	return paths
}

// Parse parses VS Code config data into the canonical format.
func (a *Adapter) Parse(data []byte) (*core.Config, error) {
	var vscodeCfg Config
	if err := json.Unmarshal(data, &vscodeCfg); err != nil {
		return nil, &core.ParseError{Format: AdapterName, Err: err}
	}
	return a.ToCore(&vscodeCfg), nil
}

// Marshal converts canonical config to VS Code format.
func (a *Adapter) Marshal(cfg *core.Config) ([]byte, error) {
	vscodeCfg := a.FromCore(cfg)
	return json.MarshalIndent(vscodeCfg, "", "  ")
}

// ReadFile reads a VS Code config file.
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

// WriteFile writes canonical config to a VS Code format file.
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

// ToCore converts VS Code config to canonical format.
func (a *Adapter) ToCore(vscodeCfg *Config) *core.Config {
	cfg := core.NewConfig()

	// Convert inputs
	for _, input := range vscodeCfg.Inputs {
		cfg.Inputs = append(cfg.Inputs, core.InputVariable{
			Type:        input.Type,
			ID:          input.ID,
			Description: input.Description,
			Password:    input.Password,
		})
	}

	// Convert servers
	for name, server := range vscodeCfg.Servers {
		coreServer := core.Server{
			Command: server.Command,
			Args:    server.Args,
			Env:     server.Env,
			EnvFile: server.EnvFile,
			URL:     server.URL,
			Headers: server.Headers,
		}

		// Set transport type
		switch server.Type {
		case "stdio":
			coreServer.Transport = core.TransportStdio
		case "http":
			coreServer.Transport = core.TransportHTTP
		case "sse":
			coreServer.Transport = core.TransportSSE
		}

		cfg.Servers[name] = coreServer
	}

	return cfg
}

// FromCore converts canonical config to VS Code format.
func (a *Adapter) FromCore(cfg *core.Config) *Config {
	vscodeCfg := &Config{
		Servers: make(map[string]ServerConfig),
	}

	// Convert inputs
	for _, input := range cfg.Inputs {
		vscodeCfg.Inputs = append(vscodeCfg.Inputs, InputVariable{
			Type:        input.Type,
			ID:          input.ID,
			Description: input.Description,
			Password:    input.Password,
		})
	}

	// Convert servers
	for name, server := range cfg.Servers {
		vscodeServer := ServerConfig{
			Command: server.Command,
			Args:    server.Args,
			Env:     server.Env,
			EnvFile: server.EnvFile,
			URL:     server.URL,
			Headers: server.Headers,
		}

		// VS Code requires explicit type
		transport := server.InferTransport()
		if transport != "" {
			vscodeServer.Type = transport.String()
		} else if server.Command != "" {
			vscodeServer.Type = "stdio"
		} else if server.URL != "" {
			vscodeServer.Type = "http"
		}

		vscodeCfg.Servers[name] = vscodeServer
	}

	return vscodeCfg
}

// WorkspaceConfigPath returns the workspace config path.
func WorkspaceConfigPath() string {
	return filepath.Join(WorkspaceConfigDir, ConfigFileName)
}

// ReadWorkspaceConfig reads the workspace .vscode/mcp.json file.
func ReadWorkspaceConfig() (*core.Config, error) {
	adapter := NewAdapter()
	return adapter.ReadFile(WorkspaceConfigPath())
}

// WriteWorkspaceConfig writes to the workspace .vscode/mcp.json file.
func WriteWorkspaceConfig(cfg *core.Config) error {
	path := WorkspaceConfigPath()
	// Ensure directory exists
	if err := os.MkdirAll(WorkspaceConfigDir, 0755); err != nil {
		return err
	}
	adapter := NewAdapter()
	return adapter.WriteFile(cfg, path)
}

// init registers the adapter with the default registry.
func init() {
	core.Register(NewAdapter())
}
