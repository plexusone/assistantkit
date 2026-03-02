// Package bundle provides a unified type for AI assistant plugin bundles.
// A bundle combines all components (plugin, skills, commands, hooks, agents,
// MCP servers, and context) into a single structure that can be generated
// for any supported tool.
//
// Example usage:
//
//	package main
//
//	import (
//	    "log"
//	    "github.com/plexusone/assistantkit/bundle"
//	)
//
//	func main() {
//	    b := bundle.New("agentcall", "0.1.0", "Voice calling for AI assistants")
//	    b.Plugin.Author = "plexusone"
//	    b.Plugin.Repository = "https://github.com/plexusone/agentcall"
//
//	    // Add MCP server
//	    b.AddMCPServer("agentcall", bundle.MCPServer{
//	        Command: "./agentcall",
//	        Env: map[string]string{"NGROK_AUTHTOKEN": "${NGROK_AUTHTOKEN}"},
//	    })
//
//	    // Add skill
//	    skill := bundle.NewSkill("phone-input", "Voice calling via phone")
//	    skill.Instructions = "Use initiate_call to start a call..."
//	    b.AddSkill(skill)
//
//	    // Generate for Claude Code
//	    if err := b.Generate("claude", "."); err != nil {
//	        log.Fatal(err)
//	    }
//
//	    // Or generate for all supported tools
//	    if err := b.GenerateAll("./plugins"); err != nil {
//	        log.Fatal(err)
//	    }
//	}
package bundle

import (
	agentscore "github.com/plexusone/assistantkit/agents/core"
	commandscore "github.com/plexusone/assistantkit/commands/core"
	contextcore "github.com/plexusone/assistantkit/context/core"
	hookscore "github.com/plexusone/assistantkit/hooks/core"
	mcpcore "github.com/plexusone/assistantkit/mcp/core"
	pluginscore "github.com/plexusone/assistantkit/plugins/core"
	skillscore "github.com/plexusone/assistantkit/skills/core"
)

// SupportedTools lists all tools that have bundle generation support.
var SupportedTools = []string{
	"claude",
	"kiro",
	"gemini",
	"cursor",
	"codex",
}

// Bundle represents a complete plugin bundle with all components.
type Bundle struct {
	// Plugin is the core plugin/extension metadata.
	Plugin *pluginscore.Plugin

	// Skills are the skill definitions.
	Skills []*skillscore.Skill

	// Commands are the command/prompt definitions.
	Commands []*commandscore.Command

	// Hooks is the lifecycle hooks configuration.
	Hooks *hookscore.Config

	// Agents are the agent/subagent definitions.
	Agents []*agentscore.Agent

	// Context is the project context (CLAUDE.md, .cursorrules, etc.).
	Context *contextcore.Context

	// MCP is the MCP server configuration.
	MCP *mcpcore.Config
}

// New creates a new Bundle with the given name, version, and description.
func New(name, version, description string) *Bundle {
	return &Bundle{
		Plugin:   pluginscore.NewPlugin(name, version, description),
		Hooks:    hookscore.NewConfig(),
		MCP:      mcpcore.NewConfig(),
		Skills:   make([]*skillscore.Skill, 0),
		Commands: make([]*commandscore.Command, 0),
		Agents:   make([]*agentscore.Agent, 0),
	}
}

// AddSkill adds a skill to the bundle.
func (b *Bundle) AddSkill(skill *skillscore.Skill) {
	b.Skills = append(b.Skills, skill)
}

// AddCommand adds a command to the bundle.
func (b *Bundle) AddCommand(cmd *commandscore.Command) {
	b.Commands = append(b.Commands, cmd)
}

// AddAgent adds an agent to the bundle.
func (b *Bundle) AddAgent(agent *agentscore.Agent) {
	b.Agents = append(b.Agents, agent)
}

// SetHooks sets the hooks configuration.
func (b *Bundle) SetHooks(cfg *hookscore.Config) {
	b.Hooks = cfg
}

// SetContext sets the project context.
func (b *Bundle) SetContext(ctx *contextcore.Context) {
	b.Context = ctx
}

// AddMCPServer adds an MCP server configuration.
// This is a convenience method that adds the server to both the MCP config
// and the Plugin's MCPServers field.
func (b *Bundle) AddMCPServer(name string, server MCPServer) {
	// Add to MCP config
	b.MCP.Servers[name] = mcpcore.Server{
		Command:   server.Command,
		Args:      server.Args,
		Env:       server.Env,
		Transport: mcpcore.TransportStdio,
	}

	// Also add to plugin for tools that read from plugin manifest
	b.Plugin.AddMCPServer(name, pluginscore.MCPServer{
		Command: server.Command,
		Args:    server.Args,
		Env:     server.Env,
	})
}

// MCPServer represents an MCP server configuration.
type MCPServer struct {
	Command string
	Args    []string
	Env     map[string]string
}

// NewSkill creates a new skill.
func NewSkill(name, description string) *skillscore.Skill {
	return skillscore.NewSkill(name, description)
}

// NewCommand creates a new command.
func NewCommand(name, description string) *commandscore.Command {
	return commandscore.NewCommand(name, description)
}

// NewAgent creates a new agent.
func NewAgent(name, description string) *agentscore.Agent {
	return agentscore.NewAgent(name, description)
}

// NewContext creates a new context.
func NewContext(name string) *contextcore.Context {
	return contextcore.NewContext(name)
}

// NewHooksConfig creates a new hooks configuration.
func NewHooksConfig() *hookscore.Config {
	return hookscore.NewConfig()
}

// Re-export common hook types for convenience.
type (
	// Hook represents a single hook action.
	Hook = hookscore.Hook
	// HookEntry represents a hook entry with optional matcher.
	HookEntry = hookscore.HookEntry
	// Event represents a hook event type.
	Event = hookscore.Event
)

// Re-export common hook events.
const (
	EventOnStop         = hookscore.OnStop
	EventOnNotification = hookscore.OnNotification
	EventOnSubagentStop = hookscore.OnSubagentStop
	EventBeforeCommand  = hookscore.BeforeCommand
	EventAfterCommand   = hookscore.AfterCommand
	EventBeforeFileRead = hookscore.BeforeFileRead
	EventAfterFileRead  = hookscore.AfterFileRead
)

// Re-export command argument type.
type Argument = commandscore.Argument

// Re-export core types for convenience.
type (
	// Skill represents a skill definition.
	Skill = skillscore.Skill
	// Command represents a command definition.
	Command = commandscore.Command
	// Agent represents an agent definition.
	Agent = agentscore.Agent
	// Config represents hooks configuration (alias for convenience).
	Config = hookscore.Config
	// Context represents project context.
	Context = contextcore.Context
)
