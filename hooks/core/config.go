package core

import (
	"encoding/json"
	"io/fs"
	"os"
)

// DefaultFileMode is the default permission mode for configuration files.
// This can be used by adapters or overridden with WriteFileWithMode.
const DefaultFileMode fs.FileMode = 0600

// Config represents the canonical hooks configuration that can be
// converted to/from various AI assistant formats.
type Config struct {
	// Version is the configuration version (used by Cursor).
	Version int `json:"version,omitempty"`

	// Hooks maps events to their hook entries.
	Hooks map[Event][]HookEntry `json:"hooks"`

	// DisableAllHooks disables all hooks when true (Claude-specific).
	DisableAllHooks bool `json:"disableAllHooks,omitempty"`

	// AllowManagedHooksOnly restricts to enterprise-managed hooks only (Claude-specific).
	AllowManagedHooksOnly bool `json:"allowManagedHooksOnly,omitempty"`
}

// NewConfig creates a new empty hooks Config.
func NewConfig() *Config {
	return &Config{
		Hooks: make(map[Event][]HookEntry),
	}
}

// AddHook adds a hook for the specified event.
func (c *Config) AddHook(event Event, hook Hook) {
	c.AddHookWithMatcher(event, "", hook)
}

// AddHookWithMatcher adds a hook with a matcher pattern for the specified event.
func (c *Config) AddHookWithMatcher(event Event, matcher string, hook Hook) {
	if c.Hooks == nil {
		c.Hooks = make(map[Event][]HookEntry)
	}

	// Find existing entry with same matcher
	entries := c.Hooks[event]
	for i, entry := range entries {
		if entry.Matcher == matcher {
			entries[i].Hooks = append(entries[i].Hooks, hook)
			c.Hooks[event] = entries
			return
		}
	}

	// Create new entry
	c.Hooks[event] = append(entries, HookEntry{
		Matcher: matcher,
		Hooks:   []Hook{hook},
	})
}

// GetHooks returns all hook entries for an event.
func (c *Config) GetHooks(event Event) []HookEntry {
	return c.Hooks[event]
}

// GetAllHooksForEvent returns a flat list of all hooks for an event.
func (c *Config) GetAllHooksForEvent(event Event) []Hook {
	var hooks []Hook
	for _, entry := range c.Hooks[event] {
		hooks = append(hooks, entry.Hooks...)
	}
	return hooks
}

// RemoveHooks removes all hooks for an event.
func (c *Config) RemoveHooks(event Event) {
	delete(c.Hooks, event)
}

// Events returns all events that have hooks configured.
func (c *Config) Events() []Event {
	events := make([]Event, 0, len(c.Hooks))
	for event := range c.Hooks {
		events = append(events, event)
	}
	return events
}

// HasHooks returns true if any hooks are configured.
func (c *Config) HasHooks() bool {
	return len(c.Hooks) > 0
}

// HookCount returns the total number of hooks across all events.
func (c *Config) HookCount() int {
	count := 0
	for _, entries := range c.Hooks {
		for _, entry := range entries {
			count += len(entry.Hooks)
		}
	}
	return count
}

// Merge combines another config into this one.
// Hooks from the other config are appended to existing hooks.
func (c *Config) Merge(other *Config) {
	if other == nil {
		return
	}
	for event, entries := range other.Hooks {
		c.Hooks[event] = append(c.Hooks[event], entries...)
	}
	// Take the more restrictive settings
	if other.DisableAllHooks {
		c.DisableAllHooks = true
	}
	if other.AllowManagedHooksOnly {
		c.AllowManagedHooksOnly = true
	}
}

// FilterByTool returns a new config with only hooks supported by the specified tool.
func (c *Config) FilterByTool(tool string) *Config {
	filtered := NewConfig()
	filtered.Version = c.Version
	filtered.DisableAllHooks = c.DisableAllHooks
	filtered.AllowManagedHooksOnly = c.AllowManagedHooksOnly

	for event, entries := range c.Hooks {
		support := event.GetToolSupport()
		var supported bool
		switch tool {
		case "claude":
			supported = support.Claude
		case "cursor":
			supported = support.Cursor
		case "windsurf":
			supported = support.Windsurf
		}
		if supported {
			filtered.Hooks[event] = entries
		}
	}
	return filtered
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	for event, entries := range c.Hooks {
		for i, entry := range entries {
			for j, hook := range entry.Hooks {
				if err := hook.Validate(); err != nil {
					return &HookValidationError{
						Event:      event,
						EntryIndex: i,
						HookIndex:  j,
						Err:        err,
					}
				}
			}
		}
	}
	return nil
}

// MarshalJSON implements json.Marshaler.
func (c *Config) MarshalJSON() ([]byte, error) {
	type Alias Config
	return json.Marshal((*Alias)(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *Config) UnmarshalJSON(data []byte) error {
	type Alias Config
	aux := (*Alias)(c)
	return json.Unmarshal(data, aux)
}

// WriteFile writes the config to a file in JSON format using DefaultFileMode.
func (c *Config) WriteFile(path string) error {
	return c.WriteFileWithMode(path, DefaultFileMode)
}

// WriteFileWithMode writes the config to a file in JSON format with the specified permission mode.
func (c *Config) WriteFileWithMode(path string, mode fs.FileMode) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, mode)
}

// ReadFile reads a config from a JSON file.
func ReadFile(path string) (*Config, error) {
	//nolint:gosec // G703: Path is user-provided config file location
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
