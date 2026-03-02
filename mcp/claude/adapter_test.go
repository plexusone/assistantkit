package claude

import (
	"testing"

	"github.com/plexusone/assistantkit/mcp/core"
)

func TestAdapterName(t *testing.T) {
	adapter := NewAdapter()
	if adapter.Name() != "claude" {
		t.Errorf("Expected name 'claude', got %q", adapter.Name())
	}
}

func TestAdapterParse(t *testing.T) {
	adapter := NewAdapter()

	jsonData := []byte(`{
		"mcpServers": {
			"github": {
				"command": "npx",
				"args": ["-y", "@modelcontextprotocol/server-github"],
				"env": {
					"GITHUB_TOKEN": "test-token"
				}
			},
			"sentry": {
				"type": "http",
				"url": "https://mcp.sentry.dev/mcp",
				"headers": {
					"Authorization": "Bearer token"
				}
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
	github, ok := cfg.GetServer("github")
	if !ok {
		t.Fatal("github server not found")
	}
	if github.Command != "npx" {
		t.Errorf("Expected command 'npx', got %q", github.Command)
	}
	if github.Transport != core.TransportStdio {
		t.Errorf("Expected transport stdio, got %v", github.Transport)
	}

	// Check http server
	sentry, ok := cfg.GetServer("sentry")
	if !ok {
		t.Fatal("sentry server not found")
	}
	if sentry.URL != "https://mcp.sentry.dev/mcp" {
		t.Errorf("Expected URL 'https://mcp.sentry.dev/mcp', got %q", sentry.URL)
	}
	if sentry.Transport != core.TransportHTTP {
		t.Errorf("Expected transport http, got %v", sentry.Transport)
	}
}

func TestAdapterMarshal(t *testing.T) {
	adapter := NewAdapter()

	cfg := core.NewConfig()
	cfg.AddServer("test", core.Server{
		Transport: core.TransportStdio,
		Command:   "npx",
		Args:      []string{"-y", "test-server"},
		Env:       map[string]string{"KEY": "value"},
	})

	data, err := adapter.Marshal(cfg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Verify we can parse it back
	cfg2, err := adapter.Parse(data)
	if err != nil {
		t.Fatalf("Failed to parse marshaled data: %v", err)
	}

	server, ok := cfg2.GetServer("test")
	if !ok {
		t.Fatal("test server not found after round-trip")
	}
	if server.Command != "npx" {
		t.Errorf("Expected command 'npx', got %q", server.Command)
	}
}

func TestAdapterRoundTrip(t *testing.T) {
	adapter := NewAdapter()

	original := core.NewConfig()
	original.AddServer("stdio-server", core.Server{
		Transport: core.TransportStdio,
		Command:   "node",
		Args:      []string{"server.js", "--port", "3000"},
		Env:       map[string]string{"NODE_ENV": "production"},
	})
	original.AddServer("http-server", core.Server{
		Transport: core.TransportHTTP,
		URL:       "https://api.example.com/mcp",
		Headers:   map[string]string{"X-API-Key": "secret"},
	})

	// Marshal
	data, err := adapter.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Parse
	parsed, err := adapter.Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify
	if len(parsed.Servers) != 2 {
		t.Errorf("Expected 2 servers, got %d", len(parsed.Servers))
	}

	stdio, _ := parsed.GetServer("stdio-server")
	if stdio.Command != "node" {
		t.Errorf("stdio server command mismatch")
	}

	http, _ := parsed.GetServer("http-server")
	if http.URL != "https://api.example.com/mcp" {
		t.Errorf("http server URL mismatch")
	}
}
