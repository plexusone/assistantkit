package claude

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"

	"github.com/plexusone/assistantkit/mcp/core"
)

const (
	// AdapterName is the identifier for this adapter.
	AdapterName = "claude"

	// ProjectConfigFile is the project-level config file name.
	ProjectConfigFile = ".mcp.json"

	// UserConfigFile is the user-level config file name.
	UserConfigFile = ".claude.json"

	// ManagedConfigFile is the enterprise managed config file name.
	ManagedConfigFile = "managed-mcp.json"
)

// Adapter implements core.Adapter for Claude Code / Claude Desktop.
type Adapter struct{}

// NewAdapter creates a new Claude adapter.
func NewAdapter() *Adapter {
	return &Adapter{}
}

// Name returns the adapter name.
func (a *Adapter) Name() string {
	return AdapterName
}

// DefaultPaths returns the default config file paths for Claude.
func (a *Adapter) DefaultPaths() []string {
	paths := []string{ProjectConfigFile}

	// User config
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, UserConfigFile))
	}

	// Enterprise managed config
	switch runtime.GOOS {
	case "darwin":
		paths = append(paths, filepath.Join("/Library/Application Support/ClaudeCode", ManagedConfigFile))
	case "linux":
		paths = append(paths, filepath.Join("/etc/claude-code", ManagedConfigFile))
	case "windows":
		paths = append(paths, filepath.Join("C:\\Program Files\\ClaudeCode", ManagedConfigFile))
	}

	return paths
}

// Parse parses Claude config data into the canonical format.
func (a *Adapter) Parse(data []byte) (*core.Config, error) {
	var claudeCfg Config
	if err := json.Unmarshal(data, &claudeCfg); err != nil {
		return nil, &core.ParseError{Format: AdapterName, Err: err}
	}
	return a.ToCore(&claudeCfg), nil
}

// Marshal converts canonical config to Claude format.
func (a *Adapter) Marshal(cfg *core.Config) ([]byte, error) {
	claudeCfg := a.FromCore(cfg)
	data, err := json.MarshalIndent(claudeCfg, "", "  ")
	if err != nil {
		return nil, err
	}
	return data, nil
}

// ReadFile reads a Claude config file.
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

// WriteFile writes canonical config to a Claude format file.
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

// ToCore converts Claude config to canonical format.
func (a *Adapter) ToCore(claudeCfg *Config) *core.Config {
	cfg := core.NewConfig()

	for name, server := range claudeCfg.MCPServers {
		coreServer := core.Server{
			Command: server.Command,
			Args:    server.Args,
			Env:     server.Env,
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
		default:
			// Infer from fields
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

// FromCore converts canonical config to Claude format.
func (a *Adapter) FromCore(cfg *core.Config) *Config {
	claudeCfg := NewConfig()

	for name, server := range cfg.Servers {
		claudeServer := ServerConfig{
			Command: server.Command,
			Args:    server.Args,
			Env:     server.Env,
			URL:     server.URL,
			Headers: server.Headers,
		}

		// Set type if explicitly specified
		if server.Transport != "" {
			claudeServer.Type = server.Transport.String()
		}

		claudeCfg.MCPServers[name] = claudeServer
	}

	return claudeCfg
}

// ReadProjectConfig reads the project-level .mcp.json file.
func ReadProjectConfig() (*core.Config, error) {
	adapter := NewAdapter()
	return adapter.ReadFile(ProjectConfigFile)
}

// ReadUserConfig reads the user-level ~/.claude.json file.
func ReadUserConfig() (*core.Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	adapter := NewAdapter()
	return adapter.ReadFile(filepath.Join(home, UserConfigFile))
}

// WriteProjectConfig writes to the project-level .mcp.json file.
func WriteProjectConfig(cfg *core.Config) error {
	adapter := NewAdapter()
	return adapter.WriteFile(cfg, ProjectConfigFile)
}

// init registers the adapter with the default registry.
func init() {
	core.Register(NewAdapter())
}
