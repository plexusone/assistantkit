package cursor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/plexusone/assistantkit/mcp/core"
)

func TestAdapterName(t *testing.T) {
	adapter := NewAdapter()
	if adapter.Name() != "cursor" {
		t.Errorf("Expected name 'cursor', got %q", adapter.Name())
	}
}

func TestAdapterDefaultPaths(t *testing.T) {
	adapter := NewAdapter()
	paths := adapter.DefaultPaths()

	if len(paths) == 0 {
		t.Error("Expected at least one default path")
	}
}

func TestAdapterParse(t *testing.T) {
	adapter := NewAdapter()

	jsonData := []byte(`{
		"mcpServers": {
			"test-server": {
				"command": "node",
				"args": ["server.js"],
				"env": {
					"API_KEY": "test-key"
				}
			}
		}
	}`)

	cfg, err := adapter.Parse(jsonData)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(cfg.Servers) != 1 {
		t.Errorf("Expected 1 server, got %d", len(cfg.Servers))
	}

	server, ok := cfg.GetServer("test-server")
	if !ok {
		t.Fatal("test-server not found")
	}
	if server.Command != "node" {
		t.Errorf("Expected command 'node', got %q", server.Command)
	}
}

func TestAdapterMarshal(t *testing.T) {
	adapter := NewAdapter()

	cfg := core.NewConfig()
	cfg.AddServer("test", core.Server{
		Transport: core.TransportStdio,
		Command:   "npx",
		Args:      []string{"test-server"},
	})

	data, err := adapter.Marshal(cfg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Verify round-trip
	cfg2, err := adapter.Parse(data)
	if err != nil {
		t.Fatalf("Parse after marshal failed: %v", err)
	}

	if len(cfg2.Servers) != 1 {
		t.Errorf("Expected 1 server after round-trip, got %d", len(cfg2.Servers))
	}
}

func TestAdapterReadWriteFile(t *testing.T) {
	adapter := NewAdapter()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "mcp.json")

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
}

func TestAdapterReadFileNotFound(t *testing.T) {
	adapter := NewAdapter()

	_, err := adapter.ReadFile("/nonexistent/path/mcp.json")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestGlobalConfigPath(t *testing.T) {
	path, err := GlobalConfigPath()
	if err != nil {
		t.Fatalf("GlobalConfigPath failed: %v", err)
	}

	if !filepath.IsAbs(path) {
		t.Error("Expected absolute path")
	}
}

func TestWriteGlobalConfig(t *testing.T) {
	// Use temp dir to simulate home
	tmpDir := t.TempDir()
	cursorDir := filepath.Join(tmpDir, ".cursor")

	cfg := core.NewConfig()
	cfg.AddServer("test", core.Server{Command: "test"})

	// Create directory and write
	if err := os.MkdirAll(cursorDir, 0755); err != nil {
		t.Fatalf("Failed to create dir: %v", err)
	}

	adapter := NewAdapter()
	path := filepath.Join(cursorDir, ConfigFileName)
	if err := adapter.WriteFile(cfg, path); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}
}
