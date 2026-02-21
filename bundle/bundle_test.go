package bundle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewBundle(t *testing.T) {
	b := New("test-plugin", "1.0.0", "A test plugin")

	if b.Plugin == nil {
		t.Fatal("expected Plugin to be initialized")
	}
	if b.Plugin.Name != "test-plugin" {
		t.Errorf("expected name 'test-plugin', got '%s'", b.Plugin.Name)
	}
	if b.Plugin.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", b.Plugin.Version)
	}
	if b.Plugin.Description != "A test plugin" {
		t.Errorf("expected description 'A test plugin', got '%s'", b.Plugin.Description)
	}
	if b.Hooks == nil {
		t.Fatal("expected Hooks to be initialized")
	}
	if b.MCP == nil {
		t.Fatal("expected MCP to be initialized")
	}
}

func TestAddSkill(t *testing.T) {
	b := New("test", "1.0.0", "test")

	skill := NewSkill("phone-input", "Voice calling via phone")
	skill.Instructions = "Test instructions"
	b.AddSkill(skill)

	if len(b.Skills) != 1 {
		t.Errorf("expected 1 skill, got %d", len(b.Skills))
	}
	if b.Skills[0].Name != "phone-input" {
		t.Errorf("expected skill name 'phone-input', got '%s'", b.Skills[0].Name)
	}
}

func TestAddCommand(t *testing.T) {
	b := New("test", "1.0.0", "test")

	cmd := NewCommand("call", "Initiate a phone call")
	cmd.Instructions = "Test instructions"
	b.AddCommand(cmd)

	if len(b.Commands) != 1 {
		t.Errorf("expected 1 command, got %d", len(b.Commands))
	}
	if b.Commands[0].Name != "call" {
		t.Errorf("expected command name 'call', got '%s'", b.Commands[0].Name)
	}
}

func TestAddAgent(t *testing.T) {
	b := New("test", "1.0.0", "test")

	agent := NewAgent("voice-agent", "Handles voice calls")
	agent.Instructions = "Test instructions"
	b.AddAgent(agent)

	if len(b.Agents) != 1 {
		t.Errorf("expected 1 agent, got %d", len(b.Agents))
	}
	if b.Agents[0].Name != "voice-agent" {
		t.Errorf("expected agent name 'voice-agent', got '%s'", b.Agents[0].Name)
	}
}

func TestAddMCPServer(t *testing.T) {
	b := New("test", "1.0.0", "test")

	b.AddMCPServer("agentcall", MCPServer{
		Command: "./agentcall",
		Args:    []string{"--port", "8080"},
		Env:     map[string]string{"DEBUG": "true"},
	})

	// Check MCP config
	if len(b.MCP.Servers) != 1 {
		t.Errorf("expected 1 MCP server, got %d", len(b.MCP.Servers))
	}
	server, ok := b.MCP.Servers["agentcall"]
	if !ok {
		t.Fatal("expected MCP server 'agentcall' to exist")
	}
	if server.Command != "./agentcall" {
		t.Errorf("expected command './agentcall', got '%s'", server.Command)
	}

	// Check Plugin MCPServers
	if len(b.Plugin.MCPServers) != 1 {
		t.Errorf("expected 1 plugin MCP server, got %d", len(b.Plugin.MCPServers))
	}
}

func TestGenerateClaude(t *testing.T) {
	b := New("agentcall", "0.1.0", "Voice calling for AI assistants")
	b.Plugin.Author = "agentplexus"

	// Add MCP server
	//nolint:gosec // G101: Environment variable template, not a hardcoded credential
	b.AddMCPServer("agentcall", MCPServer{
		Command: "./agentcall",
		Env:     map[string]string{"NGROK_AUTHTOKEN": "${NGROK_AUTHTOKEN}"},
	})

	// Add skill
	skill := NewSkill("phone-input", "Voice calling via phone")
	skill.Instructions = "Use initiate_call to start a call..."
	skill.AddTrigger("call")
	skill.AddTrigger("phone")
	b.AddSkill(skill)

	// Add command
	cmd := NewCommand("call", "Initiate a phone call to the user")
	cmd.Instructions = "Initiate a phone call..."
	b.AddCommand(cmd)

	// Create temp dir
	tmpDir, err := os.MkdirTemp("", "bundle-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Generate
	if err := b.Generate("claude", tmpDir); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check plugin.json exists
	pluginPath := filepath.Join(tmpDir, ".claude-plugin", "plugin.json")
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		t.Error("expected plugin.json to be created")
	}

	// Check skills directory exists
	skillsDir := filepath.Join(tmpDir, "skills")
	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		t.Error("expected skills directory to be created")
	}

	// Check skill file exists
	skillFile := filepath.Join(skillsDir, "phone-input", "SKILL.md")
	if _, err := os.Stat(skillFile); os.IsNotExist(err) {
		t.Error("expected SKILL.md to be created")
	}

	// Check commands directory exists
	commandsDir := filepath.Join(tmpDir, "commands")
	if _, err := os.Stat(commandsDir); os.IsNotExist(err) {
		t.Error("expected commands directory to be created")
	}

	// Check command file exists
	cmdFile := filepath.Join(commandsDir, "call.md")
	if _, err := os.Stat(cmdFile); os.IsNotExist(err) {
		t.Error("expected call.md to be created")
	}
}

func TestGenerateKiro(t *testing.T) {
	b := New("agentcall", "0.1.0", "Voice calling for AI assistants")

	// Add MCP server
	//nolint:gosec // G101: Environment variable template, not a hardcoded credential
	b.AddMCPServer("agentcall", MCPServer{
		Command: "./agentcall",
		Env:     map[string]string{"NGROK_AUTHTOKEN": "${NGROK_AUTHTOKEN}"},
	})

	// Add agent
	agent := NewAgent("voice-caller", "Handles voice calling")
	agent.Instructions = "You are a voice calling agent..."
	b.AddAgent(agent)

	// Create temp dir
	tmpDir, err := os.MkdirTemp("", "bundle-test-kiro-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Generate
	if err := b.Generate("kiro", tmpDir); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check MCP config exists
	mcpPath := filepath.Join(tmpDir, ".kiro", "settings", "mcp.json")
	if _, err := os.Stat(mcpPath); os.IsNotExist(err) {
		t.Error("expected mcp.json to be created")
	}

	// Check agents directory exists
	agentsDir := filepath.Join(tmpDir, ".kiro", "agents")
	if _, err := os.Stat(agentsDir); os.IsNotExist(err) {
		t.Error("expected agents directory to be created")
	}

	// Check agent file exists
	agentFile := filepath.Join(agentsDir, "voice-caller.json")
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		t.Error("expected voice-caller.json to be created")
	}
}

func TestToolConfig(t *testing.T) {
	// Verify all supported tools have configs
	for _, tool := range SupportedTools {
		if _, ok := DefaultToolConfigs[tool]; !ok {
			t.Errorf("missing tool config for %s", tool)
		}
	}
}
