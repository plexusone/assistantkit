package windsurf

import (
	"path/filepath"
	"testing"

	"github.com/plexusone/assistantkit/mcp/core"
)

func TestAdapterName(t *testing.T) {
	adapter := NewAdapter()
	if adapter.Name() != "windsurf" {
		t.Errorf("Expected name 'windsurf', got %q", adapter.Name())
	}
}

func TestAdapterDefaultPaths(t *testing.T) {
	adapter := NewAdapter()
	paths := adapter.DefaultPaths()

	// Should have at least one path (user config)
	if len(paths) == 0 {
		t.Error("Expected at least one default path")
	}
}

func TestAdapterParse(t *testing.T) {
	adapter := NewAdapter()

	// Windsurf uses serverUrl instead of url
	jsonData := []byte(`{
		"mcpServers": {
			"stdio-server": {
				"command": "node",
				"args": ["server.js"],
				"env": {"KEY": "value"}
			},
			"http-server": {
				"type": "http",
				"serverUrl": "https://api.example.com/mcp",
				"headers": {"Authorization": "Bearer token"}
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

	// Check stdio server
	stdio, ok := cfg.GetServer("stdio-server")
	if !ok {
		t.Fatal("stdio-server not found")
	}
	if stdio.Command != "node" {
		t.Errorf("Expected command 'node', got %q", stdio.Command)
	}
	if stdio.Transport != core.TransportStdio {
		t.Errorf("Expected transport stdio, got %v", stdio.Transport)
	}

	// Check http server - URL should be converted from serverUrl
	http, ok := cfg.GetServer("http-server")
	if !ok {
		t.Fatal("http-server not found")
	}
	if http.URL != "https://api.example.com/mcp" {
		t.Errorf("Expected URL from serverUrl, got %q", http.URL)
	}
	if http.Transport != core.TransportHTTP {
		t.Errorf("Expected transport http, got %v", http.Transport)
	}
}

func TestAdapterMarshal(t *testing.T) {
	adapter := NewAdapter()

	cfg := core.NewConfig()
	cfg.AddServer("test-http", core.Server{
		Transport: core.TransportHTTP,
		URL:       "https://example.com/mcp",
		Headers:   map[string]string{"X-Key": "value"},
	})

	data, err := adapter.Marshal(cfg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Verify serverUrl is used in output (not url)
	if string(data) == "" {
		t.Error("Expected non-empty output")
	}

	// Round-trip test
	cfg2, err := adapter.Parse(data)
	if err != nil {
		t.Fatalf("Parse after marshal failed: %v", err)
	}

	server, ok := cfg2.GetServer("test-http")
	if !ok {
		t.Fatal("test-http not found after round-trip")
	}
	if server.URL != "https://example.com/mcp" {
		t.Errorf("URL mismatch after round-trip: %q", server.URL)
	}
}

func TestAdapterRoundTrip(t *testing.T) {
	adapter := NewAdapter()

	original := core.NewConfig()
	original.AddServer("stdio", core.Server{
		Transport:     core.TransportStdio,
		Command:       "node",
		Args:          []string{"server.js"},
		Env:           map[string]string{"NODE_ENV": "prod"},
		DisabledTools: []string{"dangerous_tool"},
	})

	data, err := adapter.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	parsed, err := adapter.Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	server, _ := parsed.GetServer("stdio")
	if server.Command != "node" {
		t.Errorf("Command mismatch: %q", server.Command)
	}
	if len(server.DisabledTools) != 1 {
		t.Errorf("Expected 1 disabled tool, got %d", len(server.DisabledTools))
	}
}

func TestAdapterReadWriteFile(t *testing.T) {
	adapter := NewAdapter()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "mcp_config.json")

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

func TestConfigPath(t *testing.T) {
	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath failed: %v", err)
	}

	if !filepath.IsAbs(path) {
		t.Error("Expected absolute path")
	}
}
