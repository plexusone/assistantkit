package claude

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"

	"github.com/plexusone/assistantkit/hooks/core"
)

const (
	// AdapterName is the identifier for this adapter.
	AdapterName = "claude"

	// SettingsFileName is the settings file name containing hooks.
	SettingsFileName = "settings.json"

	// SettingsLocalFileName is the local settings file name.
	SettingsLocalFileName = "settings.local.json"

	// ProjectConfigDir is the project config directory.
	ProjectConfigDir = ".claude"

	// ManagedSettingsFileName is the enterprise managed settings file.
	ManagedSettingsFileName = "managed-settings.json"
)

// Adapter implements core.Adapter for Claude Code hooks.
type Adapter struct{}

// NewAdapter creates a new Claude hooks adapter.
func NewAdapter() *Adapter {
	return &Adapter{}
}

// Name returns the adapter name.
func (a *Adapter) Name() string {
	return AdapterName
}

// DefaultPaths returns the default config file paths for Claude hooks.
func (a *Adapter) DefaultPaths() []string {
	paths := []string{
		filepath.Join(ProjectConfigDir, SettingsFileName),
		filepath.Join(ProjectConfigDir, SettingsLocalFileName),
	}

	// User config
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ProjectConfigDir, SettingsFileName))
	}

	// Enterprise managed config
	switch runtime.GOOS {
	case "darwin":
		paths = append(paths, filepath.Join("/Library/Application Support/ClaudeCode", ManagedSettingsFileName))
	case "linux":
		paths = append(paths, filepath.Join("/etc/claude-code", ManagedSettingsFileName))
	case "windows":
		paths = append(paths, filepath.Join("C:\\Program Files\\ClaudeCode", ManagedSettingsFileName))
	}

	return paths
}

// SupportedEvents returns the events supported by Claude.
func (a *Adapter) SupportedEvents() []core.Event {
	return []core.Event{
		core.BeforeFileRead, core.AfterFileRead,
		core.BeforeFileWrite, core.AfterFileWrite,
		core.BeforeCommand, core.AfterCommand,
		core.BeforeMCP, core.AfterMCP,
		core.BeforePrompt,
		core.OnStop, core.OnSessionStart, core.OnSessionEnd,
		core.OnPermission, core.OnNotification,
		core.BeforeCompact, core.OnSubagentStop,
	}
}

// Parse parses Claude hooks config data into the canonical format.
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
	return json.MarshalIndent(claudeCfg, "", "  ")
}

// ReadFile reads a Claude hooks config file.
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

// ToCore converts Claude hooks config to canonical format.
func (a *Adapter) ToCore(claudeCfg *Config) *core.Config {
	cfg := core.NewConfig()
	cfg.DisableAllHooks = claudeCfg.DisableAllHooks
	cfg.AllowManagedHooksOnly = claudeCfg.AllowManagedHooksOnly

	for claudeEvent, entries := range claudeCfg.Hooks {
		for _, entry := range entries {
			// Determine canonical event based on Claude event and matcher
			canonicalEvent := a.claudeToCanonicalEvent(claudeEvent, entry.Matcher)

			// Convert hooks
			var coreHooks []core.Hook
			for _, h := range entry.Hooks {
				coreHook := core.Hook{
					Command: h.Command,
					Prompt:  h.Prompt,
					Timeout: h.Timeout,
				}
				if h.Type == "command" {
					coreHook.Type = core.HookTypeCommand
				} else if h.Type == "prompt" {
					coreHook.Type = core.HookTypePrompt
				}
				coreHooks = append(coreHooks, coreHook)
			}

			// Add to canonical config
			cfg.Hooks[canonicalEvent] = append(cfg.Hooks[canonicalEvent], core.HookEntry{
				Matcher: entry.Matcher,
				Hooks:   coreHooks,
			})
		}
	}

	return cfg
}

// FromCore converts canonical config to Claude format.
func (a *Adapter) FromCore(cfg *core.Config) *Config {
	claudeCfg := NewConfig()
	claudeCfg.DisableAllHooks = cfg.DisableAllHooks
	claudeCfg.AllowManagedHooksOnly = cfg.AllowManagedHooksOnly

	for event, entries := range cfg.Hooks {
		claudeEvent, matcher := a.canonicalToClaudeEvent(event)
		if claudeEvent == "" {
			continue // Event not supported by Claude
		}

		for _, entry := range entries {
			// Use entry matcher if provided, otherwise use default for event
			m := entry.Matcher
			if m == "" {
				m = matcher
			}

			var claudeHooks []Hook
			for _, h := range entry.Hooks {
				claudeHook := Hook{
					Command: h.Command,
					Prompt:  h.Prompt,
					Timeout: h.Timeout,
				}
				if h.Type == core.HookTypeCommand {
					claudeHook.Type = "command"
				} else if h.Type == core.HookTypePrompt {
					claudeHook.Type = "prompt"
				} else if h.Command != "" {
					claudeHook.Type = "command"
				} else if h.Prompt != "" {
					claudeHook.Type = "prompt"
				}
				claudeHooks = append(claudeHooks, claudeHook)
			}

			claudeCfg.Hooks[claudeEvent] = append(claudeCfg.Hooks[claudeEvent], HookEntry{
				Matcher: m,
				Hooks:   claudeHooks,
			})
		}
	}

	return claudeCfg
}

// claudeToCanonicalEvent converts a Claude event to canonical event.
func (a *Adapter) claudeToCanonicalEvent(claudeEvent ClaudeEvent, matcher string) core.Event {
	// Check direct mapping first
	if event, ok := reverseEventMapping[claudeEvent]; ok {
		return event
	}

	// Handle PreToolUse/PostToolUse based on matcher
	switch claudeEvent {
	case PreToolUse:
		if event, ok := matcherToCanonicalEventBefore[matcher]; ok {
			return event
		}
		// Default to BeforeMCP for unknown matchers (likely MCP tools)
		return core.BeforeMCP
	case PostToolUse:
		if event, ok := matcherToCanonicalEventAfter[matcher]; ok {
			return event
		}
		return core.AfterMCP
	}

	return ""
}

// canonicalToClaudeEvent converts a canonical event to Claude event and matcher.
func (a *Adapter) canonicalToClaudeEvent(event core.Event) (ClaudeEvent, string) {
	// Check if this event is supported by Claude
	if !event.GetToolSupport().Claude {
		return "", ""
	}

	// Get matcher if applicable
	matcher := canonicalEventToMatcher[event]

	// Get Claude event
	if claudeEvent, ok := eventMapping[event]; ok {
		return claudeEvent, matcher
	}

	return "", ""
}

// ReadProjectConfig reads the project-level .claude/settings.json hooks.
func ReadProjectConfig() (*core.Config, error) {
	adapter := NewAdapter()
	return adapter.ReadFile(filepath.Join(ProjectConfigDir, SettingsFileName))
}

// ReadUserConfig reads the user-level ~/.claude/settings.json hooks.
func ReadUserConfig() (*core.Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	adapter := NewAdapter()
	return adapter.ReadFile(filepath.Join(home, ProjectConfigDir, SettingsFileName))
}

// init registers the adapter with the default registry.
func init() {
	core.Register(NewAdapter())
}
