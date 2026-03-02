package windsurf

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"

	"github.com/plexusone/assistantkit/hooks/core"
)

const (
	// AdapterName is the identifier for this adapter.
	AdapterName = "windsurf"

	// ConfigFileName is the hooks config file name.
	ConfigFileName = "hooks.json"

	// UserConfigDir is the user config directory relative to home.
	UserConfigDir = ".codeium/windsurf"

	// WorkspaceConfigDir is the workspace config directory.
	WorkspaceConfigDir = ".windsurf"
)

// Adapter implements core.Adapter for Windsurf hooks.
type Adapter struct{}

// NewAdapter creates a new Windsurf hooks adapter.
func NewAdapter() *Adapter {
	return &Adapter{}
}

// Name returns the adapter name.
func (a *Adapter) Name() string {
	return AdapterName
}

// DefaultPaths returns the default config file paths for Windsurf hooks.
func (a *Adapter) DefaultPaths() []string {
	paths := []string{
		filepath.Join(WorkspaceConfigDir, ConfigFileName),
	}

	// User config
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, UserConfigDir, ConfigFileName))
	}

	// System config
	switch runtime.GOOS {
	case "darwin":
		paths = append(paths, filepath.Join("/Library/Application Support/Windsurf", ConfigFileName))
	case "linux":
		paths = append(paths, filepath.Join("/etc/windsurf", ConfigFileName))
	case "windows":
		paths = append(paths, filepath.Join("C:\\ProgramData\\Windsurf", ConfigFileName))
	}

	return paths
}

// SupportedEvents returns the events supported by Windsurf.
func (a *Adapter) SupportedEvents() []core.Event {
	return []core.Event{
		core.BeforeFileRead, core.AfterFileRead,
		core.BeforeFileWrite, core.AfterFileWrite,
		core.BeforeCommand, core.AfterCommand,
		core.BeforeMCP, core.AfterMCP,
		core.BeforePrompt,
	}
}

// Parse parses Windsurf hooks config data into the canonical format.
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

// ReadFile reads a Windsurf hooks config file.
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

// ToCore converts Windsurf hooks config to canonical format.
func (a *Adapter) ToCore(windsurfCfg *Config) *core.Config {
	cfg := core.NewConfig()

	for windsurfEvent, hooks := range windsurfCfg.Hooks {
		canonicalEvent, ok := reverseEventMapping[windsurfEvent]
		if !ok {
			continue
		}

		var coreHooks []core.Hook
		for _, h := range hooks {
			coreHooks = append(coreHooks, core.Hook{
				Type:       core.HookTypeCommand,
				Command:    h.Command,
				ShowOutput: h.ShowOutput,
				WorkingDir: h.WorkingDirectory,
			})
		}

		cfg.Hooks[canonicalEvent] = append(cfg.Hooks[canonicalEvent], core.HookEntry{
			Hooks: coreHooks,
		})
	}

	return cfg
}

// FromCore converts canonical config to Windsurf format.
func (a *Adapter) FromCore(cfg *core.Config) *Config {
	windsurfCfg := NewConfig()

	for event, entries := range cfg.Hooks {
		windsurfEvent, ok := eventMapping[event]
		if !ok {
			continue // Event not supported by Windsurf
		}

		for _, entry := range entries {
			for _, h := range entry.Hooks {
				// Windsurf only supports command hooks
				if h.Command != "" {
					windsurfCfg.Hooks[windsurfEvent] = append(windsurfCfg.Hooks[windsurfEvent], Hook{
						Command:          h.Command,
						ShowOutput:       h.ShowOutput,
						WorkingDirectory: h.WorkingDir,
					})
				}
			}
		}
	}

	return windsurfCfg
}

// WorkspaceConfigPath returns the workspace hooks config path.
func WorkspaceConfigPath() string {
	return filepath.Join(WorkspaceConfigDir, ConfigFileName)
}

// ReadWorkspaceConfig reads the workspace .windsurf/hooks.json.
func ReadWorkspaceConfig() (*core.Config, error) {
	adapter := NewAdapter()
	return adapter.ReadFile(WorkspaceConfigPath())
}

// WriteWorkspaceConfig writes to the workspace .windsurf/hooks.json.
func WriteWorkspaceConfig(cfg *core.Config) error {
	path := WorkspaceConfigPath()
	// Ensure directory exists
	if err := os.MkdirAll(WorkspaceConfigDir, 0755); err != nil {
		return err
	}
	adapter := NewAdapter()
	return adapter.WriteFile(cfg, path)
}

// UserConfigPath returns the user hooks config path.
func UserConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, UserConfigDir, ConfigFileName), nil
}

// ReadUserConfig reads the user-level ~/.codeium/windsurf/hooks.json.
func ReadUserConfig() (*core.Config, error) {
	path, err := UserConfigPath()
	if err != nil {
		return nil, err
	}
	adapter := NewAdapter()
	return adapter.ReadFile(path)
}

// init registers the adapter with the default registry.
func init() {
	core.Register(NewAdapter())
}
