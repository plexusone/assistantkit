package codex

import (
	"path/filepath"
	"testing"

	"github.com/plexusone/assistantkit/mcp/core"
)

func TestAdapterName(t *testing.T) {
	adapter := NewAdapter()
	if adapter.Name() != "codex" {
		t.Errorf("Expected name 'codex', got %q", adapter.Name())
	}
}

func TestAdapterDefaultPaths(t *testing.T) {
	adapter := NewAdapter()
	paths := adapter.DefaultPaths()

	// Should have user config path
	if len(paths) == 0 {
		t.Error("Expected at least one default path")
	}
}

func TestAdapterParse(t *testing.T) {
	adapter := NewAdapter()

	// Codex uses TOML format
	tomlData := []byte(`
[mcp_servers.stdio-server]
command = "node"
args = ["server.js"]
cwd = "/path/to/dir"

[mcp_servers.stdio-server.env]
KEY = "value"

[mcp_servers.http-server]
url = "https://api.example.com/mcp"
bearer_token_env_var = "API_TOKEN"
enabled_tools = ["tool1", "tool2"]
disabled_tools = ["dangerous"]
startup_timeout_sec = 30
tool_timeout_sec = 60
enabled = true

[mcp_servers.http-server.http_headers]
X-Custom = "header"
`)

	cfg, err := adapter.Parse(tomlData)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(cfg.Servers) != 2 {
		t.Errorf("Expected 2 servers, got %d", len(cfg.Servers))
	}

	// Check stdio server
	stdio, ok := cfg.GetServer("stdio-server")
	if !ok {
		t.Fatal("stdio-server not found")
	}
	if stdio.Command != "node" {
		t.Errorf("Expected command 'node', got %q", stdio.Command)
	}
	if stdio.Cwd != "/path/to/dir" {
		t.Errorf("Expected cwd '/path/to/dir', got %q", stdio.Cwd)
	}
	if stdio.Transport != core.TransportStdio {
		t.Errorf("Expected stdio transport, got %v", stdio.Transport)
	}

	// Check http server
	http, ok := cfg.GetServer("http-server")
	if !ok {
		t.Fatal("http-server not found")
	}
	if http.URL != "https://api.example.com/mcp" {
		t.Errorf("Expected URL, got %q", http.URL)
	}
	if http.BearerTokenEnvVar != "API_TOKEN" {
		t.Errorf("Expected bearer token env var 'API_TOKEN', got %q", http.BearerTokenEnvVar)
	}
	if len(http.EnabledTools) != 2 {
		t.Errorf("Expected 2 enabled tools, got %d", len(http.EnabledTools))
	}
	if len(http.DisabledTools) != 1 {
		t.Errorf("Expected 1 disabled tool, got %d", len(http.DisabledTools))
	}
	if http.StartupTimeoutSec != 30 {
		t.Errorf("Expected startup timeout 30, got %d", http.StartupTimeoutSec)
	}
	if http.ToolTimeoutSec != 60 {
		t.Errorf("Expected tool timeout 60, got %d", http.ToolTimeoutSec)
	}
}

func TestAdapterMarshal(t *testing.T) {
	adapter := NewAdapter()

	enabled := true
	cfg := core.NewConfig()
	cfg.AddServer("test", core.Server{
		Transport:         core.TransportStdio,
		Command:           "npx",
		Args:              []string{"-y", "server"},
		Cwd:               "/work",
		BearerTokenEnvVar: "TOKEN",
		EnabledTools:      []string{"safe"},
		DisabledTools:     []string{"unsafe"},
		StartupTimeoutSec: 15,
		ToolTimeoutSec:    30,
		Enabled:           &enabled,
	})

	data, err := adapter.Marshal(cfg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Round-trip
	cfg2, err := adapter.Parse(data)
	if err != nil {
		t.Fatalf("Parse after marshal failed: %v", err)
	}

	server, ok := cfg2.GetServer("test")
	if !ok {
		t.Fatal("test not found after round-trip")
	}
	if server.Command != "npx" {
		t.Errorf("Command mismatch: %q", server.Command)
	}
	if server.Cwd != "/work" {
		t.Errorf("Cwd mismatch: %q", server.Cwd)
	}
	if server.StartupTimeoutSec != 15 {
		t.Errorf("Startup timeout mismatch: %d", server.StartupTimeoutSec)
	}
}

func TestAdapterTransportInference(t *testing.T) {
	adapter := NewAdapter()

	tests := []struct {
		name     string
		toml     string
		expected core.TransportType
	}{
		{
			name: "infer stdio from command",
			toml: `[mcp_servers.s]
command = "cmd"`,
			expected: core.TransportStdio,
		},
		{
			name: "infer http from url",
			toml: `[mcp_servers.s]
url = "http://example.com"`,
			expected: core.TransportHTTP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := adapter.Parse([]byte(tt.toml))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}
			server, _ := cfg.GetServer("s")
			if server.Transport != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, server.Transport)
			}
		})
	}
}

func TestAdapterReadWriteFile(t *testing.T) {
	adapter := NewAdapter()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ConfigFileName)

	cfg := core.NewConfig()
	cfg.AddServer("file-test", core.Server{
		Command: "echo",
		Args:    []string{"hello"},
	})

	if err := adapter.WriteFile(cfg, path); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	loaded, err := adapter.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if len(loaded.Servers) != 1 {
		t.Errorf("Expected 1 server, got %d", len(loaded.Servers))
	}

	server, _ := loaded.GetServer("file-test")
	if server.Command != "echo" {
		t.Errorf("Expected command 'echo', got %q", server.Command)
	}
}

func TestAdapterReadFileNotFound(t *testing.T) {
	adapter := NewAdapter()

	_, err := adapter.ReadFile("/nonexistent/config.toml")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestAdapterParseInvalid(t *testing.T) {
	adapter := NewAdapter()

	_, err := adapter.Parse([]byte("invalid toml [[["))
	if err == nil {
		t.Error("Expected error for invalid TOML")
	}
}

func TestConfigPath(t *testing.T) {
	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath failed: %v", err)
	}

	if !filepath.IsAbs(path) {
		t.Error("Expected absolute path")
	}
}

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()

	if cfg.MCPServers == nil {
		t.Error("Expected MCPServers to be initialized")
	}
}
