package vscode

import (
	"path/filepath"
	"testing"

	"github.com/plexusone/assistantkit/mcp/core"
)

func TestAdapterName(t *testing.T) {
	adapter := NewAdapter()
	if adapter.Name() != "vscode" {
		t.Errorf("Expected name 'vscode', got %q", adapter.Name())
	}
}

func TestAdapterDefaultPaths(t *testing.T) {
	adapter := NewAdapter()
	paths := adapter.DefaultPaths()

	if len(paths) == 0 {
		t.Error("Expected at least one default path")
	}
	// First path should be workspace config
	if paths[0] != filepath.Join(WorkspaceConfigDir, ConfigFileName) {
		t.Errorf("Expected workspace config path first, got %q", paths[0])
	}
}

func TestAdapterParse(t *testing.T) {
	adapter := NewAdapter()

	// VS Code uses "servers" not "mcpServers"
	jsonData := []byte(`{
		"inputs": [
			{
				"type": "promptString",
				"id": "api-key",
				"description": "Enter your API key",
				"password": true
			}
		],
		"servers": {
			"test-server": {
				"type": "stdio",
				"command": "node",
				"args": ["server.js"],
				"env": {"KEY": "${input:api-key}"}
			},
			"http-server": {
				"type": "http",
				"url": "https://api.example.com/mcp"
			}
		}
	}`)

	cfg, err := adapter.Parse(jsonData)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Check inputs
	if len(cfg.Inputs) != 1 {
		t.Errorf("Expected 1 input, got %d", len(cfg.Inputs))
	}
	if cfg.Inputs[0].ID != "api-key" {
		t.Errorf("Expected input ID 'api-key', got %q", cfg.Inputs[0].ID)
	}
	if !cfg.Inputs[0].Password {
		t.Error("Expected password to be true")
	}

	// Check servers
	if len(cfg.Servers) != 2 {
		t.Errorf("Expected 2 servers, got %d", len(cfg.Servers))
	}

	stdio, ok := cfg.GetServer("test-server")
	if !ok {
		t.Fatal("test-server not found")
	}
	if stdio.Transport != core.TransportStdio {
		t.Errorf("Expected stdio transport, got %v", stdio.Transport)
	}

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

	cfg := core.NewConfig()
	cfg.AddInput(core.InputVariable{
		Type:        "promptString",
		ID:          "token",
		Description: "API Token",
		Password:    true,
	})
	cfg.AddServer("test", core.Server{
		Transport: core.TransportStdio,
		Command:   "npx",
		Args:      []string{"server"},
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

	if len(cfg2.Inputs) != 1 {
		t.Errorf("Expected 1 input after round-trip, got %d", len(cfg2.Inputs))
	}
	if len(cfg2.Servers) != 1 {
		t.Errorf("Expected 1 server after round-trip, got %d", len(cfg2.Servers))
	}
}

func TestAdapterTypeInference(t *testing.T) {
	adapter := NewAdapter()

	// Test that type is inferred correctly when marshaling
	cfg := core.NewConfig()
	cfg.AddServer("stdio-infer", core.Server{
		Command: "node", // No explicit transport
		Args:    []string{"server.js"},
	})
	cfg.AddServer("http-infer", core.Server{
		URL: "https://example.com/mcp", // No explicit transport
	})

	data, err := adapter.Marshal(cfg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	cfg2, err := adapter.Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	stdio, _ := cfg2.GetServer("stdio-infer")
	if stdio.Transport != core.TransportStdio {
		t.Errorf("Expected inferred stdio transport, got %v", stdio.Transport)
	}

	http, _ := cfg2.GetServer("http-infer")
	if http.Transport != core.TransportHTTP {
		t.Errorf("Expected inferred http transport, got %v", http.Transport)
	}
}

func TestAdapterReadWriteFile(t *testing.T) {
	adapter := NewAdapter()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "mcp.json")

	cfg := core.NewConfig()
	cfg.AddServer("file-test", core.Server{
		Transport: core.TransportStdio,
		Command:   "test",
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

	_, err := adapter.ReadFile("/nonexistent/mcp.json")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestWorkspaceConfigPath(t *testing.T) {
	path := WorkspaceConfigPath()

	expected := filepath.Join(WorkspaceConfigDir, ConfigFileName)
	if path != expected {
		t.Errorf("Expected %q, got %q", expected, path)
	}
}

func TestSSETransport(t *testing.T) {
	adapter := NewAdapter()

	jsonData := []byte(`{
		"servers": {
			"sse-server": {
				"type": "sse",
				"url": "https://api.example.com/sse"
			}
		}
	}`)

	cfg, err := adapter.Parse(jsonData)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	server, ok := cfg.GetServer("sse-server")
	if !ok {
		t.Fatal("sse-server not found")
	}
	if server.Transport != core.TransportSSE {
		t.Errorf("Expected sse transport, got %v", server.Transport)
	}
}
