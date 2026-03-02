package cursor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/plexusone/assistantkit/hooks/core"
)

func TestNewAdapter(t *testing.T) {
	adapter := NewAdapter()
	if adapter == nil {
		t.Fatal("NewAdapter returned nil")
	}
}

func TestAdapterName(t *testing.T) {
	adapter := NewAdapter()
	if adapter.Name() != "cursor" {
		t.Errorf("Expected name 'cursor', got %q", adapter.Name())
	}
}

func TestAdapterDefaultPaths(t *testing.T) {
	adapter := NewAdapter()
	paths := adapter.DefaultPaths()
	if len(paths) < 1 {
		t.Errorf("Expected at least 1 default path, got %d", len(paths))
	}
	// Check project path is present
	if paths[0] != filepath.Join(ProjectConfigDir, ConfigFileName) {
		t.Errorf("First path should be project config, got %q", paths[0])
	}
}

func TestAdapterSupportedEvents(t *testing.T) {
	adapter := NewAdapter()
	events := adapter.SupportedEvents()
	if len(events) < 8 {
		t.Errorf("Expected at least 8 supported events, got %d", len(events))
	}

	// Check key events are present
	eventSet := make(map[core.Event]bool)
	for _, e := range events {
		eventSet[e] = true
	}

	requiredEvents := []core.Event{
		core.BeforeFileRead, core.AfterFileWrite,
		core.BeforeCommand, core.AfterCommand,
		core.BeforeMCP, core.AfterMCP,
		core.AfterResponse, core.AfterThought,
	}
	for _, e := range requiredEvents {
		if !eventSet[e] {
			t.Errorf("Expected event %q in supported events", e)
		}
	}
}

func TestAdapterParse(t *testing.T) {
	adapter := NewAdapter()

	tests := []struct {
		name      string
		json      string
		wantHooks int
		wantError bool
	}{
		{
			name: "valid beforeShellExecution hook",
			json: `{
				"version": 1,
				"hooks": {
					"beforeShellExecution": [
						{"command": "echo before"}
					]
				}
			}`,
			wantHooks: 1,
			wantError: false,
		},
		{
			name: "valid afterFileEdit hook",
			json: `{
				"version": 1,
				"hooks": {
					"afterFileEdit": [
						{"command": "echo after edit"}
					]
				}
			}`,
			wantHooks: 1,
			wantError: false,
		},
		{
			name: "multiple events",
			json: `{
				"version": 1,
				"hooks": {
					"beforeShellExecution": [
						{"command": "echo 1"}
					],
					"afterAgentResponse": [
						{"command": "echo 2"}
					]
				}
			}`,
			wantHooks: 2,
			wantError: false,
		},
		{
			name: "multiple hooks per event",
			json: `{
				"version": 1,
				"hooks": {
					"beforeShellExecution": [
						{"command": "echo 1"},
						{"command": "echo 2"},
						{"command": "echo 3"}
					]
				}
			}`,
			wantHooks: 3,
			wantError: false,
		},
		{
			name:      "invalid json",
			json:      `{invalid`,
			wantError: true,
		},
		{
			name:      "empty config",
			json:      `{"version": 1, "hooks": {}}`,
			wantHooks: 0,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := adapter.Parse([]byte(tt.json))
			if (err != nil) != tt.wantError {
				t.Errorf("Parse() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && cfg.HookCount() != tt.wantHooks {
				t.Errorf("Parse() got %d hooks, want %d", cfg.HookCount(), tt.wantHooks)
			}
		})
	}
}

func TestAdapterMarshal(t *testing.T) {
	adapter := NewAdapter()

	cfg := core.NewConfig()
	cfg.AddHook(core.BeforeCommand, core.NewCommandHook("echo before"))
	cfg.AddHook(core.AfterResponse, core.NewCommandHook("echo response"))

	data, err := adapter.Marshal(cfg)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	if len(data) == 0 {
		t.Error("Marshal() returned empty data")
	}

	// Parse back and verify
	parsed, err := adapter.Parse(data)
	if err != nil {
		t.Fatalf("Failed to parse marshaled data: %v", err)
	}

	if parsed.HookCount() != 2 {
		t.Errorf("Round-trip got %d hooks, want 2", parsed.HookCount())
	}
}

func TestAdapterRoundTrip(t *testing.T) {
	adapter := NewAdapter()

	original := `{
		"version": 1,
		"hooks": {
			"beforeShellExecution": [
				{"command": "echo before shell"}
			],
			"afterShellExecution": [
				{"command": "echo after shell"}
			],
			"beforeMCPExecution": [
				{"command": "echo before mcp"}
			],
			"afterFileEdit": [
				{"command": "echo after edit"}
			],
			"stop": [
				{"command": "echo stop"}
			]
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

	// Verify hook count matches
	if cfg.HookCount() != cfg2.HookCount() {
		t.Errorf("Hook count mismatch after round-trip: %d vs %d",
			cfg.HookCount(), cfg2.HookCount())
	}
}

func TestAdapterReadWriteFile(t *testing.T) {
	adapter := NewAdapter()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "cursor-hooks-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create config
	cfg := core.NewConfig()
	cfg.AddHook(core.BeforeCommand, core.NewCommandHook("echo test"))
	cfg.Version = 2

	// Write file
	filePath := filepath.Join(tmpDir, "hooks.json")
	if err := adapter.WriteFile(cfg, filePath); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Read file
	readCfg, err := adapter.ReadFile(filePath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	// Verify
	if readCfg.HookCount() != 1 {
		t.Errorf("ReadFile() got %d hooks, want 1", readCfg.HookCount())
	}
	if readCfg.Version != 2 {
		t.Errorf("Version should be 2, got %d", readCfg.Version)
	}
}

func TestAdapterReadFileNotFound(t *testing.T) {
	adapter := NewAdapter()

	_, err := adapter.ReadFile("/nonexistent/path/hooks.json")
	if err == nil {
		t.Error("ReadFile() should return error for nonexistent file")
	}

	// Check it's a ParseError
	if _, ok := err.(*core.ParseError); !ok {
		t.Errorf("Expected ParseError, got %T", err)
	}
}

func TestAdapterToCoreEventMapping(t *testing.T) {
	adapter := NewAdapter()

	tests := []struct {
		cursorEvent CursorEvent
		wantEvent   core.Event
	}{
		{BeforeShellExecution, core.BeforeCommand},
		{AfterShellExecution, core.AfterCommand},
		{BeforeMCPExecution, core.BeforeMCP},
		{AfterMCPExecution, core.AfterMCP},
		{BeforeReadFile, core.BeforeFileRead},
		{AfterFileEdit, core.AfterFileWrite},
		{BeforeSubmitPrompt, core.BeforePrompt},
		{AfterAgentResponse, core.AfterResponse},
		{AfterAgentThought, core.AfterThought},
		{Stop, core.OnStop},
		{BeforeTabFileRead, core.BeforeTabRead},
		{AfterTabFileEdit, core.AfterTabEdit},
	}

	for _, tt := range tests {
		t.Run(string(tt.cursorEvent), func(t *testing.T) {
			json := `{"version": 1, "hooks": {"` + string(tt.cursorEvent) + `": [{"command": "echo test"}]}}`
			cfg, err := adapter.Parse([]byte(json))
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			hooks := cfg.GetAllHooksForEvent(tt.wantEvent)
			if len(hooks) != 1 {
				t.Errorf("Expected 1 hook for event %q, got %d", tt.wantEvent, len(hooks))
			}
		})
	}
}

func TestAdapterFromCoreEventMapping(t *testing.T) {
	adapter := NewAdapter()

	tests := []struct {
		event       core.Event
		wantCursor  CursorEvent
		shouldExist bool
	}{
		{core.BeforeCommand, BeforeShellExecution, true},
		{core.AfterCommand, AfterShellExecution, true},
		{core.BeforeMCP, BeforeMCPExecution, true},
		{core.AfterMCP, AfterMCPExecution, true},
		{core.BeforeFileRead, BeforeReadFile, true},
		{core.AfterFileWrite, AfterFileEdit, true},
		{core.OnStop, Stop, true},
		{core.AfterResponse, AfterAgentResponse, true},
		{core.AfterThought, AfterAgentThought, true},
		// Claude-only events should not map
		{core.OnSessionStart, "", false},
		{core.OnPermission, "", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.event), func(t *testing.T) {
			cfg := core.NewConfig()
			cfg.AddHook(tt.event, core.NewCommandHook("echo test"))

			cursorCfg := adapter.FromCore(cfg)

			if tt.shouldExist {
				hooks := cursorCfg.Hooks[tt.wantCursor]
				if len(hooks) != 1 {
					t.Errorf("Expected 1 hook for cursor event %q, got %d", tt.wantCursor, len(hooks))
				}
			} else {
				// Should have no hooks since the event is unsupported
				totalHooks := 0
				for _, h := range cursorCfg.Hooks {
					totalHooks += len(h)
				}
				if totalHooks != 0 {
					t.Errorf("Expected 0 hooks for unsupported event, got %d", totalHooks)
				}
			}
		})
	}
}

func TestAdapterPromptHooksIgnored(t *testing.T) {
	adapter := NewAdapter()

	// Cursor doesn't support prompt hooks, only command hooks
	cfg := core.NewConfig()
	cfg.AddHook(core.BeforeCommand, core.NewPromptHook("Is this safe?"))

	cursorCfg := adapter.FromCore(cfg)

	// Should have no hooks since prompt hooks are not supported
	totalHooks := 0
	for _, h := range cursorCfg.Hooks {
		totalHooks += len(h)
	}
	if totalHooks != 0 {
		t.Errorf("Expected 0 hooks (prompt hooks not supported), got %d", totalHooks)
	}
}

func TestAdapterVersionPreserved(t *testing.T) {
	adapter := NewAdapter()

	cfg := core.NewConfig()
	cfg.Version = 5
	cfg.AddHook(core.BeforeCommand, core.NewCommandHook("echo test"))

	cursorCfg := adapter.FromCore(cfg)
	if cursorCfg.Version != 5 {
		t.Errorf("Expected version 5, got %d", cursorCfg.Version)
	}

	// Default version should be 1
	cfg2 := core.NewConfig()
	cfg2.AddHook(core.BeforeCommand, core.NewCommandHook("echo test"))
	cursorCfg2 := adapter.FromCore(cfg2)
	if cursorCfg2.Version != 1 {
		t.Errorf("Expected default version 1, got %d", cursorCfg2.Version)
	}
}

func TestAdapterUnknownEventIgnored(t *testing.T) {
	adapter := NewAdapter()

	// JSON with unknown event
	json := `{
		"version": 1,
		"hooks": {
			"unknownEvent": [
				{"command": "echo unknown"}
			],
			"beforeShellExecution": [
				{"command": "echo known"}
			]
		}
	}`

	cfg, err := adapter.Parse([]byte(json))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Should only have the known event's hook
	if cfg.HookCount() != 1 {
		t.Errorf("Expected 1 hook (unknown events ignored), got %d", cfg.HookCount())
	}
}

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()
	if cfg == nil {
		t.Fatal("NewConfig returned nil")
	}
	if cfg.Hooks == nil {
		t.Error("Hooks map should be initialized")
	}
	if cfg.Version != 1 {
		t.Errorf("Default version should be 1, got %d", cfg.Version)
	}
}

func TestProjectConfigPath(t *testing.T) {
	path := ProjectConfigPath()
	expected := filepath.Join(ProjectConfigDir, ConfigFileName)
	if path != expected {
		t.Errorf("ProjectConfigPath() = %q, want %q", path, expected)
	}
}

func TestReadProjectConfigNotFound(t *testing.T) {
	_, err := ReadProjectConfig()
	if err == nil {
		t.Log("ReadProjectConfig() didn't return error, file may exist")
	}
}

func TestReadUserConfigNotFound(t *testing.T) {
	_, err := ReadUserConfig()
	if err == nil {
		t.Log("ReadUserConfig() didn't return error, file may exist")
	}
}

func TestWriteProjectConfig(t *testing.T) {
	// Create a temp directory to simulate the project directory
	origDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	tmpDir, err := os.MkdirTemp("", "cursor-project-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			panic(err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		panic(err)
	}

	cfg := core.NewConfig()
	cfg.AddHook(core.BeforeCommand, core.NewCommandHook("echo test"))

	err = WriteProjectConfig(cfg)
	if err != nil {
		t.Fatalf("WriteProjectConfig() error = %v", err)
	}

	// Verify file was created
	readCfg, err := ReadProjectConfig()
	if err != nil {
		t.Fatalf("ReadProjectConfig() error = %v", err)
	}
	if readCfg.HookCount() != 1 {
		t.Errorf("Expected 1 hook, got %d", readCfg.HookCount())
	}
}

func TestAdapterFromCoreSkipsUnsupportedEvents(t *testing.T) {
	adapter := NewAdapter()

	cfg := core.NewConfig()
	// Add Claude-only event
	cfg.AddHook(core.OnSessionStart, core.NewCommandHook("echo session"))
	// Add Cursor-supported event
	cfg.AddHook(core.BeforeCommand, core.NewCommandHook("echo command"))

	cursorCfg := adapter.FromCore(cfg)

	// Should only have BeforeCommand, not OnSessionStart
	totalHooks := 0
	for _, hooks := range cursorCfg.Hooks {
		totalHooks += len(hooks)
	}
	if totalHooks != 1 {
		t.Errorf("Expected 1 hook (unsupported events filtered), got %d", totalHooks)
	}
}

func TestAdapterToCoreSkipsUnknownEvents(t *testing.T) {
	adapter := NewAdapter()

	// Config with unknown event
	cursorCfg := &Config{
		Version: 1,
		Hooks: map[CursorEvent][]Hook{
			"unknownEvent":       {{Command: "echo unknown"}},
			BeforeShellExecution: {{Command: "echo known"}},
		},
	}

	cfg := adapter.ToCore(cursorCfg)

	// Should only have 1 hook (unknown event skipped)
	if cfg.HookCount() != 1 {
		t.Errorf("Expected 1 hook, got %d", cfg.HookCount())
	}
}
