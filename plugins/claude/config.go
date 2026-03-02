package claude

import (
	"github.com/plexusone/assistantkit/plugins/core"
)

// ClaudePlugin represents the Claude Code plugin.json format.
// See: https://docs.anthropic.com/en/docs/claude-code/plugins
type ClaudePlugin struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`

	// Optional metadata
	Author     string `json:"author,omitempty"`
	License    string `json:"license,omitempty"`
	Repository string `json:"repository,omitempty"`
	Homepage   string `json:"homepage,omitempty"`

	// MCP Servers - embedded directly in plugin.json for consolidated config
	MCPServers map[string]MCPServerConfig `json:"mcpServers,omitempty"`

	// Hooks - embedded directly in plugin.json for consolidated config
	Hooks *HooksConfig `json:"hooks,omitempty"`

	// Component paths (relative to plugin root)
	Commands string `json:"commands,omitempty"` // e.g., "./commands/"
	Skills   string `json:"skills,omitempty"`   // e.g., "./skills/"
	Agents   string `json:"agents,omitempty"`   // e.g., "./agents/"
}

// MCPServerConfig represents an MCP server configuration in Claude format.
type MCPServerConfig struct {
	Command  string            `json:"command"`
	Args     []string          `json:"args,omitempty"`
	Env      map[string]string `json:"env,omitempty"`
	Cwd      string            `json:"cwd,omitempty"`
	Disabled bool              `json:"disabled,omitempty"`
}

// HooksConfig represents the hooks configuration embedded in plugin.json.
type HooksConfig struct {
	// Event-based hooks (Claude Code format)
	PreToolUse   []HookEntry `json:"PreToolUse,omitempty"`
	PostToolUse  []HookEntry `json:"PostToolUse,omitempty"`
	Notification []HookEntry `json:"Notification,omitempty"`
	Stop         []HookEntry `json:"Stop,omitempty"`
	SubagentStop []HookEntry `json:"SubagentStop,omitempty"`
}

// HookEntry represents a hook entry with optional matcher.
type HookEntry struct {
	Matcher string `json:"matcher,omitempty"`
	Hooks   []Hook `json:"hooks"`
}

// Hook represents a single hook action.
type Hook struct {
	Type    string `json:"type"`              // "command" or "prompt"
	Command string `json:"command,omitempty"` // For command hooks
	Prompt  string `json:"prompt,omitempty"`  // For prompt hooks
}

// ToCanonical converts ClaudePlugin to canonical Plugin.
func (cp *ClaudePlugin) ToCanonical() *core.Plugin {
	p := &core.Plugin{
		Name:        cp.Name,
		Version:     cp.Version,
		Description: cp.Description,
		Author:      cp.Author,
		License:     cp.License,
		Repository:  cp.Repository,
		Homepage:    cp.Homepage,
		Commands:    cp.Commands,
		Skills:      cp.Skills,
		Agents:      cp.Agents,
	}

	// Convert MCP servers
	if len(cp.MCPServers) > 0 {
		p.MCPServers = make(map[string]core.MCPServer)
		for name, server := range cp.MCPServers {
			p.MCPServers[name] = core.MCPServer{
				Command: server.Command,
				Args:    server.Args,
				Env:     server.Env,
				Cwd:     server.Cwd,
			}
		}
	}

	return p
}

// FromCanonical creates a ClaudePlugin from canonical Plugin.
func FromCanonical(p *core.Plugin) *ClaudePlugin {
	cp := &ClaudePlugin{
		Name:        p.Name,
		Version:     p.Version,
		Description: p.Description,
		Author:      p.Author,
		License:     p.License,
		Repository:  p.Repository,
		Homepage:    p.Homepage,
	}

	// Set default paths if components are specified
	if p.Commands != "" {
		cp.Commands = "./commands/"
	}
	if p.Skills != "" {
		cp.Skills = "./skills/"
	}
	if p.Agents != "" {
		cp.Agents = "./agents/"
	}

	// Convert MCP servers
	if len(p.MCPServers) > 0 {
		cp.MCPServers = make(map[string]MCPServerConfig)
		for name, server := range p.MCPServers {
			cp.MCPServers[name] = MCPServerConfig{
				Command: server.Command,
				Args:    server.Args,
				Env:     server.Env,
				Cwd:     server.Cwd,
			}
		}
	}

	return cp
}

// SetHooks sets the hooks configuration on the plugin.
func (cp *ClaudePlugin) SetHooks(hooks *HooksConfig) {
	cp.Hooks = hooks
}
