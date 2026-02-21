package core

import "testing"

func TestNewPlugin(t *testing.T) {
	plugin := NewPlugin("test-plugin", "1.0.0", "A test plugin")

	if plugin.Name != "test-plugin" {
		t.Errorf("expected Name 'test-plugin', got '%s'", plugin.Name)
	}
	if plugin.Version != "1.0.0" {
		t.Errorf("expected Version '1.0.0', got '%s'", plugin.Version)
	}
	if plugin.Description != "A test plugin" {
		t.Errorf("expected Description 'A test plugin', got '%s'", plugin.Description)
	}
}

func TestPluginAddDependency(t *testing.T) {
	plugin := NewPlugin("test", "1.0.0", "test")

	plugin.AddDependency("git", "git")
	plugin.AddDependency("node", "node")

	if len(plugin.Dependencies) != 2 {
		t.Errorf("expected 2 dependencies, got %d", len(plugin.Dependencies))
	}

	if plugin.Dependencies[0].Name != "git" {
		t.Errorf("expected first dependency 'git', got '%s'", plugin.Dependencies[0].Name)
	}
	if plugin.Dependencies[0].Optional {
		t.Error("expected first dependency to be required")
	}
}

func TestPluginAddOptionalDependency(t *testing.T) {
	plugin := NewPlugin("test", "1.0.0", "test")

	plugin.AddOptionalDependency("eslint", "eslint")

	if len(plugin.Dependencies) != 1 {
		t.Errorf("expected 1 dependency, got %d", len(plugin.Dependencies))
	}

	if !plugin.Dependencies[0].Optional {
		t.Error("expected dependency to be optional")
	}
}

func TestPluginAddMCPServer(t *testing.T) {
	plugin := NewPlugin("test", "1.0.0", "test")

	//nolint:gosec // G101: Environment variable template, not a hardcoded credential
	plugin.AddMCPServer("github", MCPServer{
		Command: "npx",
		Args:    []string{"-y", "@modelcontextprotocol/server-github"},
		Env:     map[string]string{"GITHUB_TOKEN": "${GITHUB_TOKEN}"},
	})

	if len(plugin.MCPServers) != 1 {
		t.Errorf("expected 1 MCP server, got %d", len(plugin.MCPServers))
	}

	server, ok := plugin.MCPServers["github"]
	if !ok {
		t.Error("expected 'github' MCP server to exist")
	}
	if server.Command != "npx" {
		t.Errorf("expected Command 'npx', got '%s'", server.Command)
	}
}
