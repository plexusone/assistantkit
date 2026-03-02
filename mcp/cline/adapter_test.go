package cline

import (
	"path/filepath"
	"testing"

	"github.com/plexusone/assistantkit/mcp/core"
)

func TestAdapterName(t *testing.T) {
	adapter := NewAdapter()
	if adapter.Name() != "cline" {
		t.Errorf("Expected name 'cline', got %q", adapter.Name())
	}
}

func TestAdapterDefaultPaths(t *testing.T) {
	adapter := NewAdapter()
	paths := adapter.DefaultPaths()

	// Paths depend on OS and home dir availability
	if paths == nil {
		t.Error("Expected paths slice, got nil")
	}
}

func TestAdapterParse(t *testing.T) {
	adapter := NewAdapter()

	jsonData := []byte(`{
		"mcpServers": {
			"test-server": {
				"type": "stdio",
				"command": "node",
				"args": ["server.js"],
				"env": {"KEY": "value"},
				"alwaysAllow": ["read_file", "write_file"],
				"disabled": false
			},
			"disabled-server": {
				"command": "disabled",
				"disabled": true
			}
		}
	}`)

	cfg, err := adapter.Parse(jsonData)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(cfg.Servers) != 2 {
		t.Errorf("Expected 2 servers, got %d", len(cfg.Servers))
	}

	// Check enabled server
	server, ok := cfg.GetServer("test-server")
	if !ok {
		t.Fatal("test-server not found")
	}
	if server.Command != "node" {
		t.Errorf("Expected command 'node', got %q", server.Command)
	}
	if len(server.AlwaysAllow) != 2 {
		t.Errorf("Expected 2 always-allow tools, got %d", len(server.AlwaysAllow))
	}
	if !server.IsEnabled() {
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
}

func TestAdapterMarshal(t *testing.T) {
	adapter := NewAdapter()

	enabled := false
	cfg := core.NewConfig()
	cfg.AddServer("test", core.Server{
		Transport:   core.TransportStdio,
		Command:     "npx",
		Args:        []string{"server"},
		AlwaysAllow: []string{"safe_tool"},
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
	if len(server.AlwaysAllow) != 1 {
		t.Errorf("Expected 1 always-allow tool, got %d", len(server.AlwaysAllow))
	}
	if server.IsEnabled() {
		t.Error("Expected server to be disabled after round-trip")
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
			json:     `{"mcpServers": {"s": {"type": "http", "url": "http://example.com"}}}`,
			expected: core.TransportHTTP,
		},
		{
			name:     "infer stdio from command",
			json:     `{"mcpServers": {"s": {"command": "cmd"}}}`,
			expected: core.TransportStdio,
		},
		{
			name:     "infer http from url",
			json:     `{"mcpServers": {"s": {"url": "http://example.com"}}}`,
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
	path := filepath.Join(tmpDir, ConfigFileName)

	cfg := core.NewConfig()
	cfg.AddServer("file-test", core.Server{
		Command:     "echo",
		AlwaysAllow: []string{"tool"},
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
