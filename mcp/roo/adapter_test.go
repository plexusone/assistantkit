package roo

import (
	"path/filepath"
	"testing"

	"github.com/plexusone/assistantkit/mcp/core"
)

func TestAdapterName(t *testing.T) {
	adapter := NewAdapter()
	if adapter.Name() != "roo" {
		t.Errorf("Expected name 'roo', got %q", adapter.Name())
	}
}

func TestAdapterDefaultPaths(t *testing.T) {
	adapter := NewAdapter()
	paths := adapter.DefaultPaths()

	if len(paths) == 0 {
		t.Error("Expected at least one default path")
	}
	// First path should be workspace config
	if paths[0] != filepath.Join(WorkspaceConfigDir, WorkspaceConfigFileName) {
		t.Errorf("Expected workspace config path first, got %q", paths[0])
	}
}

func TestAdapterParse(t *testing.T) {
	adapter := NewAdapter()

	jsonData := []byte(`{
		"mcpServers": {
			"enabled-server": {
				"type": "stdio",
				"command": "node",
				"args": ["server.js"],
				"alwaysAllow": ["read", "write"],
				"disabled": false
			},
			"disabled-server": {
				"command": "disabled",
				"disabled": true
			},
			"http-server": {
				"type": "http",
				"url": "https://api.example.com/mcp",
				"headers": {"X-Key": "value"}
			}
		}
	}`)

	cfg, err := adapter.Parse(jsonData)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(cfg.Servers) != 3 {
		t.Errorf("Expected 3 servers, got %d", len(cfg.Servers))
	}

	// Check enabled server
	enabled, ok := cfg.GetServer("enabled-server")
	if !ok {
		t.Fatal("enabled-server not found")
	}
	if enabled.Command != "node" {
		t.Errorf("Expected command 'node', got %q", enabled.Command)
	}
	if len(enabled.AlwaysAllow) != 2 {
		t.Errorf("Expected 2 always-allow, got %d", len(enabled.AlwaysAllow))
	}
	if !enabled.IsEnabled() {
		t.Error("Expected server to be enabled")
	}

	// Check disabled server
	disabled, ok := cfg.GetServer("disabled-server")
	if !ok {
		t.Fatal("disabled-server not found")
	}
	if disabled.IsEnabled() {
		t.Error("Expected server to be disabled")
	}

	// Check http server
	http, ok := cfg.GetServer("http-server")
	if !ok {
		t.Fatal("http-server not found")
	}
	if http.Transport != core.TransportHTTP {
		t.Errorf("Expected http transport, got %v", http.Transport)
	}
}

func TestAdapterMarshal(t *testing.T) {
	adapter := NewAdapter()

	enabled := false
	cfg := core.NewConfig()
	cfg.AddServer("test", core.Server{
		Transport:   core.TransportStdio,
		Command:     "npx",
		AlwaysAllow: []string{"tool1", "tool2"},
		Enabled:     &enabled,
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
	if len(server.AlwaysAllow) != 2 {
		t.Errorf("Expected 2 always-allow, got %d", len(server.AlwaysAllow))
	}
	if server.IsEnabled() {
		t.Error("Expected server to be disabled")
	}
}

func TestAdapterTransportInference(t *testing.T) {
	adapter := NewAdapter()

	tests := []struct {
		name     string
		json     string
		expected core.TransportType
	}{
		{
			name:     "explicit stdio",
			json:     `{"mcpServers": {"s": {"type": "stdio", "command": "cmd"}}}`,
			expected: core.TransportStdio,
		},
		{
			name:     "explicit http",
			json:     `{"mcpServers": {"s": {"type": "http", "url": "http://test.com"}}}`,
			expected: core.TransportHTTP,
		},
		{
			name:     "explicit sse",
			json:     `{"mcpServers": {"s": {"type": "sse", "url": "http://test.com"}}}`,
			expected: core.TransportSSE,
		},
		{
			name:     "infer stdio",
			json:     `{"mcpServers": {"s": {"command": "cmd"}}}`,
			expected: core.TransportStdio,
		},
		{
			name:     "infer http",
			json:     `{"mcpServers": {"s": {"url": "http://test.com"}}}`,
			expected: core.TransportHTTP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := adapter.Parse([]byte(tt.json))
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
	path := filepath.Join(tmpDir, WorkspaceConfigFileName)

	cfg := core.NewConfig()
	cfg.AddServer("file-test", core.Server{
		Command: "echo",
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
}

func TestAdapterReadFileNotFound(t *testing.T) {
	adapter := NewAdapter()

	_, err := adapter.ReadFile("/nonexistent/path.json")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()

	if cfg.MCPServers == nil {
		t.Error("Expected MCPServers to be initialized")
	}
}
