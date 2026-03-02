package cursor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"

	"github.com/plexusone/assistantkit/hooks/core"
)

const (
	// AdapterName is the identifier for this adapter.
	AdapterName = "cursor"

	// ConfigFileName is the hooks config file name.
	ConfigFileName = "hooks.json"

	// ProjectConfigDir is the project config directory.
	ProjectConfigDir = ".cursor"
)

// Adapter implements core.Adapter for Cursor hooks.
type Adapter struct{}

// NewAdapter creates a new Cursor hooks adapter.
func NewAdapter() *Adapter {
	return &Adapter{}
}

// Name returns the adapter name.
func (a *Adapter) Name() string {
	return AdapterName
}

// DefaultPaths returns the default config file paths for Cursor hooks.
func (a *Adapter) DefaultPaths() []string {
	paths := []string{
		filepath.Join(ProjectConfigDir, ConfigFileName),
	}

	// User config
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ProjectConfigDir, ConfigFileName))
	}

	// Enterprise config
	switch runtime.GOOS {
	case "darwin":
		paths = append(paths, filepath.Join("/Library/Application Support/Cursor", ConfigFileName))
	case "linux":
		paths = append(paths, filepath.Join("/etc/cursor", ConfigFileName))
	case "windows":
		paths = append(paths, filepath.Join("C:\\ProgramData\\Cursor", ConfigFileName))
	}

	return paths
}

// SupportedEvents returns the events supported by Cursor.
func (a *Adapter) SupportedEvents() []core.Event {
	return []core.Event{
		core.BeforeFileRead, core.AfterFileWrite,
		core.BeforeCommand, core.AfterCommand,
		core.BeforeMCP, core.AfterMCP,
		core.BeforePrompt, core.OnStop,
		core.AfterResponse, core.AfterThought,
		core.BeforeTabRead, core.AfterTabEdit,
	}
}

// Parse parses Cursor hooks config data into the canonical format.
func (a *Adapter) Parse(data []byte) (*core.Config, error) {
	var cursorCfg Config
	if err := json.Unmarshal(data, &cursorCfg); err != nil {
		return nil, &core.ParseError{Format: AdapterName, Err: err}
	}
	return a.ToCore(&cursorCfg), nil
}

// Marshal converts canonical config to Cursor format.
func (a *Adapter) Marshal(cfg *core.Config) ([]byte, error) {
	cursorCfg := a.FromCore(cfg)
	return json.MarshalIndent(cursorCfg, "", "  ")
}

// ReadFile reads a Cursor hooks config file.
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

// ToCore converts Cursor hooks config to canonical format.
func (a *Adapter) ToCore(cursorCfg *Config) *core.Config {
	cfg := core.NewConfig()
	cfg.Version = cursorCfg.Version

	for cursorEvent, hooks := range cursorCfg.Hooks {
		canonicalEvent, ok := reverseEventMapping[cursorEvent]
		if !ok {
			continue
		}

		var coreHooks []core.Hook
		for _, h := range hooks {
			coreHooks = append(coreHooks, core.Hook{
				Type:    core.HookTypeCommand,
				Command: h.Command,
			})
		}

		cfg.Hooks[canonicalEvent] = append(cfg.Hooks[canonicalEvent], core.HookEntry{
			Hooks: coreHooks,
		})
	}

	return cfg
}

// FromCore converts canonical config to Cursor format.
func (a *Adapter) FromCore(cfg *core.Config) *Config {
	cursorCfg := NewConfig()
	if cfg.Version > 0 {
		cursorCfg.Version = cfg.Version
	}

	for event, entries := range cfg.Hooks {
		cursorEvent, ok := eventMapping[event]
		if !ok {
			continue // Event not supported by Cursor
		}

		for _, entry := range entries {
			for _, h := range entry.Hooks {
				// Cursor only supports command hooks
				if h.Command != "" {
					cursorCfg.Hooks[cursorEvent] = append(cursorCfg.Hooks[cursorEvent], Hook{
						Command: h.Command,
					})
				}
			}
		}
	}

	return cursorCfg
}

// ProjectConfigPath returns the project hooks config path.
func ProjectConfigPath() string {
	return filepath.Join(ProjectConfigDir, ConfigFileName)
}

// ReadProjectConfig reads the project-level .cursor/hooks.json.
func ReadProjectConfig() (*core.Config, error) {
	adapter := NewAdapter()
	return adapter.ReadFile(ProjectConfigPath())
}

// WriteProjectConfig writes to the project-level .cursor/hooks.json.
func WriteProjectConfig(cfg *core.Config) error {
	path := ProjectConfigPath()
	// Ensure directory exists
	if err := os.MkdirAll(ProjectConfigDir, 0755); err != nil {
		return err
	}
	adapter := NewAdapter()
	return adapter.WriteFile(cfg, path)
}

// ReadUserConfig reads the user-level ~/.cursor/hooks.json.
func ReadUserConfig() (*core.Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	adapter := NewAdapter()
	return adapter.ReadFile(filepath.Join(home, ProjectConfigDir, ConfigFileName))
}

// init registers the adapter with the default registry.
func init() {
	core.Register(NewAdapter())
}
