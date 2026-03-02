package kiro

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/plexusone/assistantkit/mcp/core"
)

func TestNewAdapter(t *testing.T) {
	adapter := NewAdapter()
	if adapter == nil {
		t.Fatal("NewAdapter returned nil")
	}
}

func TestAdapterName(t *testing.T) {
	adapter := NewAdapter()
	if adapter.Name() != "kiro" {
		t.Errorf("Expected name 'kiro', got %q", adapter.Name())
	}
}

func TestAdapterDefaultPaths(t *testing.T) {
	adapter := NewAdapter()
	paths := adapter.DefaultPaths()
	if len(paths) < 1 {
		t.Errorf("Expected at least 1 default path, got %d", len(paths))
	}

	// Check workspace path is present
	expectedPath := filepath.Join(ProjectConfigDir, SettingsDir, ConfigFileName)
	if paths[0] != expectedPath {
		t.Errorf("Expected first path %q, got %q", expectedPath, paths[0])
	}
}

func TestAdapterParseStdio(t *testing.T) {
	adapter := NewAdapter()

	json := `{
		"mcpServers": {
			"github": {
				"command": "npx",
				"args": ["-y", "@modelcontextprotocol/server-github"],
				"env": {
					"GITHUB_TOKEN": "${GITHUB_TOKEN}"
				}
			}
		}
	}`

	cfg, err := adapter.Parse([]byte(json))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	server, ok := cfg.GetServer("github")
	if !ok {
		t.Fatal("Expected 'github' server")
	}

	if server.Command != "npx" {
		t.Errorf("Expected command 'npx', got %q", server.Command)
	}
	if len(server.Args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(server.Args))
	}
	if server.Transport != core.TransportStdio {
		t.Errorf("Expected stdio transport, got %q", server.Transport)
	}
}

func TestAdapterParseRemote(t *testing.T) {
	adapter := NewAdapter()

	json := `{
		"mcpServers": {
			"remote-api": {
				"url": "https://api.example.com/mcp",
				"headers": {
					"Authorization": "Bearer ${API_TOKEN}"
				}
			}
		}
	}`

	cfg, err := adapter.Parse([]byte(json))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	server, ok := cfg.GetServer("remote-api")
	if !ok {
		t.Fatal("Expected 'remote-api' server")
	}

	if server.URL != "https://api.example.com/mcp" {
		t.Errorf("Expected URL, got %q", server.URL)
	}
	if server.Headers["Authorization"] != "Bearer ${API_TOKEN}" {
		t.Errorf("Expected Authorization header, got %q", server.Headers["Authorization"])
	}
	if server.Transport != core.TransportHTTP {
		t.Errorf("Expected HTTP transport, got %q", server.Transport)
	}
}

func TestAdapterParseDisabled(t *testing.T) {
	adapter := NewAdapter()

	json := `{
		"mcpServers": {
			"disabled-server": {
				"command": "test",
				"disabled": true
			}
		}
	}`

	cfg, err := adapter.Parse([]byte(json))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	server, ok := cfg.GetServer("disabled-server")
	if !ok {
		t.Fatal("Expected 'disabled-server' server")
	}

	if server.IsEnabled() {
		t.Error("Expected server to be disabled")
	}
}

func TestAdapterMarshal(t *testing.T) {
	adapter := NewAdapter()

	cfg := core.NewConfig()
	cfg.AddServer("test-server", core.Server{
		Command:   "test-cmd",
		Args:      []string{"arg1", "arg2"},
		Env:       map[string]string{"KEY": "value"},
		Transport: core.TransportStdio,
	})

	data, err := adapter.Marshal(cfg)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Parse back and verify
	parsed, err := adapter.Parse(data)
	if err != nil {
		t.Fatalf("Failed to parse marshaled data: %v", err)
	}

	server, ok := parsed.GetServer("test-server")
	if !ok {
		t.Fatal("Expected 'test-server' in parsed config")
	}
	if server.Command != "test-cmd" {
		t.Errorf("Expected command 'test-cmd', got %q", server.Command)
	}
}

func TestAdapterMarshalDisabled(t *testing.T) {
	adapter := NewAdapter()

	cfg := core.NewConfig()
	enabled := false
	cfg.AddServer("disabled-server", core.Server{
		Command: "test",
		Enabled: &enabled,
	})

	data, err := adapter.Marshal(cfg)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Parse back
	parsed, err := adapter.Parse(data)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	server, _ := parsed.GetServer("disabled-server")
	if server.IsEnabled() {
		t.Error("Expected server to remain disabled after round-trip")
	}
}

func TestAdapterRoundTrip(t *testing.T) {
	adapter := NewAdapter()

	original := `{
		"mcpServers": {
			"stdio-server": {
				"command": "npx",
				"args": ["-y", "package"],
				"env": {"TOKEN": "secret"}
			},
			"http-server": {
				"url": "https://example.com/mcp",
				"headers": {"Auth": "Bearer token"}
			}
		}
	}`

	// Parse
	cfg, err := adapter.Parse([]byte(original))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Marshal
	data, err := adapter.Marshal(cfg)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Parse again
	cfg2, err := adapter.Parse(data)
	if err != nil {
		t.Fatalf("Second Parse() error = %v", err)
	}

	// Verify both servers exist
	if len(cfg2.Servers) != 2 {
		t.Errorf("Expected 2 servers, got %d", len(cfg2.Servers))
	}

	stdio, ok := cfg2.GetServer("stdio-server")
	if !ok {
		t.Fatal("Expected stdio-server")
	}
	if stdio.Command != "npx" {
		t.Errorf("Expected command 'npx', got %q", stdio.Command)
	}

	http, ok := cfg2.GetServer("http-server")
	if !ok {
		t.Fatal("Expected http-server")
	}
	if http.URL != "https://example.com/mcp" {
		t.Errorf("Expected URL, got %q", http.URL)
	}
}

func TestAdapterReadWriteFile(t *testing.T) {
	adapter := NewAdapter()

	// Create temp directory
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, ".kiro", "settings", "mcp.json")

	// Create config
	cfg := core.NewConfig()
	cfg.AddServer("test", core.Server{
		Command: "test-cmd",
		Args:    []string{"arg1"},
	})

	// Write file
	if err := adapter.WriteFile(cfg, filePath); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("Expected file to exist")
	}

	// Read file
	readCfg, err := adapter.ReadFile(filePath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	// Verify
	server, ok := readCfg.GetServer("test")
	if !ok {
		t.Fatal("Expected 'test' server")
	}
	if server.Command != "test-cmd" {
		t.Errorf("Expected command 'test-cmd', got %q", server.Command)
	}
}

func TestAdapterReadFileNotFound(t *testing.T) {
	adapter := NewAdapter()

	_, err := adapter.ReadFile("/nonexistent/path/mcp.json")
	if err == nil {
		t.Error("ReadFile() should return error for nonexistent file")
	}

	// Check it's a ParseError
	if _, ok := err.(*core.ParseError); !ok {
		t.Errorf("Expected ParseError, got %T", err)
	}
}

func TestAdapterParseInvalidJSON(t *testing.T) {
	adapter := NewAdapter()

	_, err := adapter.Parse([]byte("invalid json"))
	if err == nil {
		t.Error("Parse() should return error for invalid JSON")
	}

	if _, ok := err.(*core.ParseError); !ok {
		t.Errorf("Expected ParseError, got %T", err)
	}
}

func TestAdapterParseEmptyConfig(t *testing.T) {
	adapter := NewAdapter()

	cfg, err := adapter.Parse([]byte(`{}`))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(cfg.Servers) != 0 {
		t.Errorf("Expected 0 servers, got %d", len(cfg.Servers))
	}
}

func TestAdapterRegistered(t *testing.T) {
	adapter, ok := core.GetAdapter("kiro")
	if !ok {
		t.Fatal("kiro adapter should be registered")
	}
	if adapter.Name() != "kiro" {
		t.Errorf("Expected name 'kiro', got %q", adapter.Name())
	}
}

func TestUserConfigPath(t *testing.T) {
	path, err := UserConfigPath()
	if err != nil {
		t.Fatalf("UserConfigPath() error = %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".kiro", "settings", "mcp.json")
	if path != expected {
		t.Errorf("Expected %q, got %q", expected, path)
	}
}

func TestWorkspaceConfigPath(t *testing.T) {
	path := WorkspaceConfigPath("/project")
	expected := filepath.Join("/project", ".kiro", "settings", "mcp.json")
	if path != expected {
		t.Errorf("Expected %q, got %q", expected, path)
	}
}
