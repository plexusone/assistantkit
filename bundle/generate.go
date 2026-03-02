package bundle

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	agentscore "github.com/plexusone/assistantkit/agents/core"
	commandscore "github.com/plexusone/assistantkit/commands/core"
	contextcore "github.com/plexusone/assistantkit/context/core"
	hooksclaude "github.com/plexusone/assistantkit/hooks/claude"
	hookscore "github.com/plexusone/assistantkit/hooks/core"
	mcpcore "github.com/plexusone/assistantkit/mcp/core"
	pluginsclaude "github.com/plexusone/assistantkit/plugins/claude"
	pluginscore "github.com/plexusone/assistantkit/plugins/core"
	skillscore "github.com/plexusone/assistantkit/skills/core"

	// Import adapters for side-effect registration
	_ "github.com/plexusone/assistantkit/agents/claude"
	_ "github.com/plexusone/assistantkit/agents/codex"
	_ "github.com/plexusone/assistantkit/agents/gemini"
	_ "github.com/plexusone/assistantkit/agents/kiro"
	_ "github.com/plexusone/assistantkit/commands/claude"
	_ "github.com/plexusone/assistantkit/commands/codex"
	_ "github.com/plexusone/assistantkit/commands/gemini"
	_ "github.com/plexusone/assistantkit/context/claude"
	_ "github.com/plexusone/assistantkit/hooks/claude"
	_ "github.com/plexusone/assistantkit/hooks/cursor"
	_ "github.com/plexusone/assistantkit/hooks/windsurf"
	_ "github.com/plexusone/assistantkit/mcp/claude"
	_ "github.com/plexusone/assistantkit/mcp/codex"
	_ "github.com/plexusone/assistantkit/mcp/cursor"
	_ "github.com/plexusone/assistantkit/mcp/kiro"
	_ "github.com/plexusone/assistantkit/mcp/vscode"
	_ "github.com/plexusone/assistantkit/plugins/claude"
	_ "github.com/plexusone/assistantkit/plugins/gemini"
	_ "github.com/plexusone/assistantkit/skills/claude"
	_ "github.com/plexusone/assistantkit/skills/codex"
)

// ToolConfig defines the output paths and supported components for a tool.
type ToolConfig struct {
	// PluginDir is the directory for the plugin manifest.
	PluginDir string
	// PluginFile is the plugin manifest filename.
	PluginFile string
	// SkillsDir is the directory for skills.
	SkillsDir string
	// CommandsDir is the directory for commands.
	CommandsDir string
	// HooksDir is the directory for hooks config.
	HooksDir string
	// HooksFile is the hooks config filename.
	HooksFile string
	// AgentsDir is the directory for agents.
	AgentsDir string
	// MCPDir is the directory for MCP config.
	MCPDir string
	// MCPFile is the MCP config filename.
	MCPFile string
	// ContextDir is the directory for context files.
	ContextDir string
	// ContextFile is the context filename.
	ContextFile string
}

// DefaultToolConfigs maps tool names to their configurations.
var DefaultToolConfigs = map[string]ToolConfig{
	"claude": {
		PluginDir:   ".claude-plugin",
		PluginFile:  "plugin.json",
		SkillsDir:   "skills",
		CommandsDir: "commands",
		AgentsDir:   "agents",
		// Note: Hooks and MCP are embedded in plugin.json for Claude (consolidated format)
		// HooksDir and MCPDir are intentionally empty
		ContextDir:  ".",
		ContextFile: "CLAUDE.md",
	},
	"kiro": {
		AgentsDir: ".kiro/agents",
		MCPDir:    ".kiro/settings",
		MCPFile:   "mcp.json",
	},
	"gemini": {
		PluginDir:   ".",
		PluginFile:  "gemini-extension.json",
		CommandsDir: "commands",
		AgentsDir:   "agents",
	},
	"cursor": {
		HooksDir:    ".",
		HooksFile:   ".cursorrules",
		MCPDir:      ".cursor",
		MCPFile:     "mcp.json",
		ContextDir:  ".",
		ContextFile: ".cursorrules",
	},
	"codex": {
		SkillsDir:   "skills",
		CommandsDir: "prompts",
		AgentsDir:   "agents",
		MCPDir:      ".codex",
		MCPFile:     "mcp.json",
		ContextDir:  ".",
		ContextFile: "AGENTS.md",
	},
	"vscode": {
		MCPDir:  ".vscode",
		MCPFile: "mcp.json",
	},
}

// Generate outputs the bundle for a specific tool to the given directory.
func (b *Bundle) Generate(tool, outputDir string) error {
	config, ok := DefaultToolConfigs[tool]
	if !ok {
		return &GenerateError{Tool: tool, Err: fmt.Errorf("unsupported tool")}
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return &GenerateError{Tool: tool, Err: err}
	}

	// Generate plugin manifest
	if err := b.generatePlugin(tool, outputDir, config); err != nil {
		return err
	}

	// Generate skills
	if err := b.generateSkills(tool, outputDir, config); err != nil {
		return err
	}

	// Generate commands
	if err := b.generateCommands(tool, outputDir, config); err != nil {
		return err
	}

	// Generate hooks
	if err := b.generateHooks(tool, outputDir, config); err != nil {
		return err
	}

	// Generate agents
	if err := b.generateAgents(tool, outputDir, config); err != nil {
		return err
	}

	// Generate MCP config
	if err := b.generateMCP(tool, outputDir, config); err != nil {
		return err
	}

	// Generate context
	if err := b.generateContext(tool, outputDir, config); err != nil {
		return err
	}

	return nil
}

// GenerateAll outputs the bundle for all supported tools.
func (b *Bundle) GenerateAll(outputDir string) error {
	for _, tool := range SupportedTools {
		toolDir := filepath.Join(outputDir, tool)
		if err := b.Generate(tool, toolDir); err != nil {
			return err
		}
	}
	return nil
}

// generatePlugin generates the plugin manifest for a tool.
func (b *Bundle) generatePlugin(tool, outputDir string, config ToolConfig) error {
	if config.PluginDir == "" || config.PluginFile == "" {
		return nil // Tool doesn't support plugin manifests
	}

	// Update plugin paths based on what components we have
	if len(b.Skills) > 0 && config.SkillsDir != "" {
		b.Plugin.Skills = config.SkillsDir
	}
	if len(b.Commands) > 0 && config.CommandsDir != "" {
		b.Plugin.Commands = config.CommandsDir
	}
	if len(b.Agents) > 0 && config.AgentsDir != "" {
		b.Plugin.Agents = config.AgentsDir
	}

	pluginPath := filepath.Join(outputDir, config.PluginDir, config.PluginFile)

	// For Claude, use consolidated format with embedded MCP and hooks
	if tool == "claude" {
		return b.generateClaudePlugin(config, pluginPath)
	}

	// For other tools, use standard adapter
	adapter, ok := pluginscore.GetAdapter(tool)
	if !ok {
		return nil // No adapter for this tool
	}

	if b.Hooks != nil && b.Hooks.HasHooks() && config.HooksDir != "" {
		b.Plugin.Hooks = filepath.Join(config.HooksDir, config.HooksFile)
	}

	if err := adapter.WriteFile(b.Plugin, pluginPath); err != nil {
		return &GenerateError{Tool: tool, Component: "plugin", Err: err}
	}

	return nil
}

// generateSkills generates skills for a tool.
func (b *Bundle) generateSkills(tool, outputDir string, config ToolConfig) error {
	if len(b.Skills) == 0 || config.SkillsDir == "" {
		return nil
	}

	adapter, ok := skillscore.GetAdapter(tool)
	if !ok {
		return nil // No adapter for this tool
	}

	skillsDir := filepath.Join(outputDir, config.SkillsDir)
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return &GenerateError{Tool: tool, Component: "skills", Err: err}
	}

	for _, skill := range b.Skills {
		if err := adapter.WriteSkillDir(skill, skillsDir); err != nil {
			return &GenerateError{Tool: tool, Component: "skill:" + skill.Name, Err: err}
		}
	}

	return nil
}

// generateCommands generates commands for a tool.
func (b *Bundle) generateCommands(tool, outputDir string, config ToolConfig) error {
	if len(b.Commands) == 0 || config.CommandsDir == "" {
		return nil
	}

	adapter, ok := commandscore.GetAdapter(tool)
	if !ok {
		return nil // No adapter for this tool
	}

	commandsDir := filepath.Join(outputDir, config.CommandsDir)
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		return &GenerateError{Tool: tool, Component: "commands", Err: err}
	}

	for _, cmd := range b.Commands {
		filename := cmd.Name + adapter.FileExtension()
		cmdPath := filepath.Join(commandsDir, filename)
		if err := adapter.WriteFile(cmd, cmdPath); err != nil {
			return &GenerateError{Tool: tool, Component: "command:" + cmd.Name, Err: err}
		}
	}

	return nil
}

// generateHooks generates hooks configuration for a tool.
func (b *Bundle) generateHooks(tool, outputDir string, config ToolConfig) error {
	if b.Hooks == nil || !b.Hooks.HasHooks() || config.HooksDir == "" {
		return nil
	}

	adapter, ok := hookscore.GetAdapter(tool)
	if !ok {
		return nil // No adapter for this tool
	}

	hooksPath := filepath.Join(outputDir, config.HooksDir, config.HooksFile)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(hooksPath), 0755); err != nil {
		return &GenerateError{Tool: tool, Component: "hooks", Err: err}
	}

	if err := adapter.WriteFile(b.Hooks, hooksPath); err != nil {
		return &GenerateError{Tool: tool, Component: "hooks", Err: err}
	}

	return nil
}

// generateAgents generates agents for a tool.
func (b *Bundle) generateAgents(tool, outputDir string, config ToolConfig) error {
	if len(b.Agents) == 0 || config.AgentsDir == "" {
		return nil
	}

	adapter, ok := agentscore.GetAdapter(tool)
	if !ok {
		return nil // No adapter for this tool
	}

	agentsDir := filepath.Join(outputDir, config.AgentsDir)
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		return &GenerateError{Tool: tool, Component: "agents", Err: err}
	}

	for _, agent := range b.Agents {
		filename := agent.Name + adapter.FileExtension()
		agentPath := filepath.Join(agentsDir, filename)
		if err := adapter.WriteFile(agent, agentPath); err != nil {
			return &GenerateError{Tool: tool, Component: "agent:" + agent.Name, Err: err}
		}
	}

	return nil
}

// generateMCP generates MCP server configuration for a tool.
func (b *Bundle) generateMCP(tool, outputDir string, config ToolConfig) error {
	if b.MCP == nil || len(b.MCP.Servers) == 0 || config.MCPDir == "" {
		return nil
	}

	adapter, ok := mcpcore.GetAdapter(tool)
	if !ok {
		return nil // No adapter for this tool
	}

	mcpPath := filepath.Join(outputDir, config.MCPDir, config.MCPFile)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(mcpPath), 0755); err != nil {
		return &GenerateError{Tool: tool, Component: "mcp", Err: err}
	}

	if err := adapter.WriteFile(b.MCP, mcpPath); err != nil {
		return &GenerateError{Tool: tool, Component: "mcp", Err: err}
	}

	return nil
}

// generateContext generates context file for a tool.
func (b *Bundle) generateContext(tool, outputDir string, config ToolConfig) error {
	if b.Context == nil || config.ContextFile == "" {
		return nil
	}

	converter, ok := contextcore.GetConverter(tool)
	if !ok {
		return nil // No converter for this tool
	}

	contextPath := filepath.Join(outputDir, config.ContextDir, config.ContextFile)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(contextPath), 0755); err != nil {
		return &GenerateError{Tool: tool, Component: "context", Err: err}
	}

	if err := converter.WriteFile(b.Context, contextPath); err != nil {
		return &GenerateError{Tool: tool, Component: "context", Err: err}
	}

	return nil
}

// generateClaudePlugin generates a consolidated plugin.json for Claude Code.
// This format embeds MCP servers and hooks directly in plugin.json instead of
// using separate files, providing a cleaner single-file configuration.
func (b *Bundle) generateClaudePlugin(config ToolConfig, pluginPath string) error {
	// Create Claude plugin from canonical plugin
	claudePlugin := pluginsclaude.FromCanonical(b.Plugin)

	// Override component paths based on actual content
	if len(b.Skills) > 0 && config.SkillsDir != "" {
		claudePlugin.Skills = "./" + config.SkillsDir + "/"
	}
	if len(b.Commands) > 0 && config.CommandsDir != "" {
		claudePlugin.Commands = "./" + config.CommandsDir + "/"
	}
	if len(b.Agents) > 0 && config.AgentsDir != "" {
		claudePlugin.Agents = "./" + config.AgentsDir + "/"
	}

	// Embed MCP servers directly in plugin.json
	if b.MCP != nil && len(b.MCP.Servers) > 0 {
		claudePlugin.MCPServers = make(map[string]pluginsclaude.MCPServerConfig)
		for name, server := range b.MCP.Servers {
			claudePlugin.MCPServers[name] = pluginsclaude.MCPServerConfig{
				Command:  server.Command,
				Args:     server.Args,
				Env:      server.Env,
				Cwd:      server.Cwd,
				Disabled: !server.IsEnabled(),
			}
		}
	}

	// Embed hooks directly in plugin.json
	if b.Hooks != nil && b.Hooks.HasHooks() {
		claudePlugin.Hooks = convertHooksToClaudeFormat(b.Hooks)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(pluginPath), 0755); err != nil {
		return &GenerateError{Tool: "claude", Component: "plugin", Err: err}
	}

	// Write plugin.json
	data, err := json.MarshalIndent(claudePlugin, "", "  ")
	if err != nil {
		return &GenerateError{Tool: "claude", Component: "plugin", Err: err}
	}

	if err := os.WriteFile(pluginPath, data, 0600); err != nil {
		return &GenerateError{Tool: "claude", Component: "plugin", Err: err}
	}

	return nil
}

// convertHooksToClaudeFormat converts canonical hooks config to Claude's embedded format.
func convertHooksToClaudeFormat(hooks *hookscore.Config) *pluginsclaude.HooksConfig {
	// Use the Claude hooks adapter to convert canonical to Claude format
	adapter := hooksclaude.NewAdapter()
	claudeHooks := adapter.FromCore(hooks)

	// Convert the Claude hooks config to the embedded plugin format
	hooksConfig := &pluginsclaude.HooksConfig{}

	for event, entries := range claudeHooks.Hooks {
		var pluginEntries []pluginsclaude.HookEntry
		for _, entry := range entries {
			var pluginHooks []pluginsclaude.Hook
			for _, h := range entry.Hooks {
				pluginHooks = append(pluginHooks, pluginsclaude.Hook{
					Type:    h.Type,
					Command: h.Command,
					Prompt:  h.Prompt,
				})
			}
			pluginEntries = append(pluginEntries, pluginsclaude.HookEntry{
				Matcher: entry.Matcher,
				Hooks:   pluginHooks,
			})
		}

		switch event {
		case hooksclaude.PreToolUse:
			hooksConfig.PreToolUse = pluginEntries
		case hooksclaude.PostToolUse:
			hooksConfig.PostToolUse = pluginEntries
		case hooksclaude.Notification:
			hooksConfig.Notification = pluginEntries
		case hooksclaude.Stop:
			hooksConfig.Stop = pluginEntries
		case hooksclaude.SubagentStop:
			hooksConfig.SubagentStop = pluginEntries
		}
	}

	return hooksConfig
}
