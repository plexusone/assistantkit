package gemini

import (
	"github.com/plexusone/assistantkit/plugins/core"
)

// GeminiExtension represents the Gemini CLI gemini-extension.json format.
// See: https://geminicli.com/docs/extensions/
type GeminiExtension struct {
	Name    string `json:"name"`
	Version string `json:"version"`

	// Optional metadata
	Description string `json:"description,omitempty"`
	Author      string `json:"author,omitempty"`
	License     string `json:"license,omitempty"`
	Repository  string `json:"repository,omitempty"`
	Homepage    string `json:"homepage,omitempty"`

	// MCP server configurations
	MCPServers map[string]GeminiMCPServer `json:"mcpServers,omitempty"`

	// Context file name (defaults to GEMINI.md)
	ContextFileName string `json:"contextFileName,omitempty"`

	// Tools to exclude from the model
	ExcludeTools []string `json:"excludeTools,omitempty"`

	// User-configurable settings
	Settings map[string]GeminiSetting `json:"settings,omitempty"`
}

// GeminiMCPServer represents an MCP server in Gemini format.
type GeminiMCPServer struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Cwd     string            `json:"cwd,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// GeminiSetting represents a user-configurable setting.
type GeminiSetting struct {
	Type        string `json:"type"` // e.g., "string", "boolean"
	Description string `json:"description,omitempty"`
	Default     string `json:"default,omitempty"`
	EnvVar      string `json:"envVar,omitempty"` // Environment variable name
}

// ToCanonical converts GeminiExtension to canonical Plugin.
func (ge *GeminiExtension) ToCanonical() *core.Plugin {
	plugin := &core.Plugin{
		Name:        ge.Name,
		Version:     ge.Version,
		Description: ge.Description,
		Author:      ge.Author,
		License:     ge.License,
		Repository:  ge.Repository,
		Homepage:    ge.Homepage,
	}

	// Convert MCP servers
	if len(ge.MCPServers) > 0 {
		plugin.MCPServers = make(map[string]core.MCPServer)
		for name, server := range ge.MCPServers {
			plugin.MCPServers[name] = core.MCPServer{
				Command: server.Command,
				Args:    server.Args,
				Cwd:     server.Cwd,
				Env:     server.Env,
			}
		}
	}

	return plugin
}

// FromCanonical creates a GeminiExtension from canonical Plugin.
func FromCanonical(p *core.Plugin) *GeminiExtension {
	ge := &GeminiExtension{
		Name:        p.Name,
		Version:     p.Version,
		Description: p.Description,
		Author:      p.Author,
		License:     p.License,
		Repository:  p.Repository,
		Homepage:    p.Homepage,
	}

	// Set context file name if context is provided
	if p.Context != "" {
		ge.ContextFileName = "GEMINI.md"
	}

	// Convert MCP servers
	if len(p.MCPServers) > 0 {
		ge.MCPServers = make(map[string]GeminiMCPServer)
		for name, server := range p.MCPServers {
			ge.MCPServers[name] = GeminiMCPServer{
				Command: server.Command,
				Args:    server.Args,
				Cwd:     server.Cwd,
				Env:     server.Env,
			}
		}
	}

	return ge
}
