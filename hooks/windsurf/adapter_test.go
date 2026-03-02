package windsurf

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
	if adapter.Name() != "windsurf" {
		t.Errorf("Expected name 'windsurf', got %q", adapter.Name())
	}
}

func TestAdapterDefaultPaths(t *testing.T) {
	adapter := NewAdapter()
	paths := adapter.DefaultPaths()
	if len(paths) < 1 {
		t.Errorf("Expected at least 1 default path, got %d", len(paths))
	}
	// Check workspace path is present
	if paths[0] != filepath.Join(WorkspaceConfigDir, ConfigFileName) {
		t.Errorf("First path should be workspace config, got %q", paths[0])
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
		core.BeforeFileRead, core.AfterFileRead,
		core.BeforeFileWrite, core.AfterFileWrite,
		core.BeforeCommand, core.AfterCommand,
		core.BeforeMCP, core.AfterMCP,
		core.BeforePrompt,
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
			name: "valid pre_run_command hook",
			json: `{
				"hooks": {
					"pre_run_command": [
						{"command": "echo before"}
					]
				}
			}`,
			wantHooks: 1,
			wantError: false,
		},
		{
			name: "valid post_write_code hook",
			json: `{
				"hooks": {
					"post_write_code": [
						{"command": "echo after write"}
					]
				}
			}`,
			wantHooks: 1,
			wantError: false,
		},
		{
			name: "hook with show_output",
			json: `{
				"hooks": {
					"pre_run_command": [
						{"command": "echo test", "show_output": true}
					]
				}
			}`,
			wantHooks: 1,
			wantError: false,
		},
		{
			name: "hook with working_directory",
			json: `{
				"hooks": {
					"pre_run_command": [
						{"command": "echo test", "working_directory": "/tmp"}
					]
				}
			}`,
			wantHooks: 1,
			wantError: false,
		},
		{
			name: "multiple events",
			json: `{
				"hooks": {
					"pre_run_command": [
						{"command": "echo 1"}
					],
					"post_write_code": [
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
				"hooks": {
					"pre_run_command": [
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
			json:      `{"hooks": {}}`,
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
	cfg.AddHook(core.AfterFileWrite, core.NewCommandHook("echo after"))

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
		"hooks": {
			"pre_run_command": [
				{"command": "echo before command", "show_output": true}
			],
			"post_run_command": [
				{"command": "echo after command"}
			],
			"pre_mcp_tool_use": [
				{"command": "echo before mcp"}
			],
			"post_write_code": [
				{"command": "echo after write", "working_directory": "/tmp"}
			],
			"pre_user_prompt": [
				{"command": "echo before prompt"}
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
	tmpDir, err := os.MkdirTemp("", "windsurf-hooks-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create config with extra options
	cfg := core.NewConfig()
	hook := core.NewCommandHook("echo test")
	hook.ShowOutput = true
	hook.WorkingDir = "/custom/path"
	cfg.AddHook(core.BeforeCommand, hook)

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

	hooks := readCfg.GetAllHooksForEvent(core.BeforeCommand)
	if len(hooks) != 1 {
		t.Fatalf("Expected 1 hook, got %d", len(hooks))
	}
	if !hooks[0].ShowOutput {
		t.Error("ShowOutput should be true")
	}
	if hooks[0].WorkingDir != "/custom/path" {
		t.Errorf("WorkingDir should be '/custom/path', got %q", hooks[0].WorkingDir)
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
		windsurfEvent WindsurfEvent
		wantEvent     core.Event
	}{
		{PreRunCommand, core.BeforeCommand},
		{PostRunCommand, core.AfterCommand},
		{PreMCPToolUse, core.BeforeMCP},
		{PostMCPToolUse, core.AfterMCP},
		{PreReadCode, core.BeforeFileRead},
		{PostReadCode, core.AfterFileRead},
		{PreWriteCode, core.BeforeFileWrite},
		{PostWriteCode, core.AfterFileWrite},
		{PreUserPrompt, core.BeforePrompt},
	}

	for _, tt := range tests {
		t.Run(string(tt.windsurfEvent), func(t *testing.T) {
			json := `{"hooks": {"` + string(tt.windsurfEvent) + `": [{"command": "echo test"}]}}`
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
		event        core.Event
		wantWindsurf WindsurfEvent
		shouldExist  bool
	}{
		{core.BeforeCommand, PreRunCommand, true},
		{core.AfterCommand, PostRunCommand, true},
		{core.BeforeMCP, PreMCPToolUse, true},
		{core.AfterMCP, PostMCPToolUse, true},
		{core.BeforeFileRead, PreReadCode, true},
		{core.AfterFileRead, PostReadCode, true},
		{core.BeforeFileWrite, PreWriteCode, true},
		{core.AfterFileWrite, PostWriteCode, true},
		{core.BeforePrompt, PreUserPrompt, true},
		// Events not supported by Windsurf
		{core.OnStop, "", false},
		{core.OnSessionStart, "", false},
		{core.AfterResponse, "", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.event), func(t *testing.T) {
			cfg := core.NewConfig()
			cfg.AddHook(tt.event, core.NewCommandHook("echo test"))

			windsurfCfg := adapter.FromCore(cfg)

			if tt.shouldExist {
				hooks := windsurfCfg.Hooks[tt.wantWindsurf]
				if len(hooks) != 1 {
					t.Errorf("Expected 1 hook for windsurf event %q, got %d", tt.wantWindsurf, len(hooks))
				}
			} else {
				// Should have no hooks since the event is unsupported
				totalHooks := 0
				for _, h := range windsurfCfg.Hooks {
					totalHooks += len(h)
				}
				if totalHooks != 0 {
					t.Errorf("Expected 0 hooks for unsupported event, got %d", totalHooks)
				}
			}
		})
	}
}

func TestAdapterShowOutputPreserved(t *testing.T) {
	adapter := NewAdapter()

	json := `{
		"hooks": {
			"pre_run_command": [
				{"command": "echo test", "show_output": true}
			]
		}
	}`

	cfg, err := adapter.Parse([]byte(json))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	hooks := cfg.GetAllHooksForEvent(core.BeforeCommand)
	if len(hooks) != 1 {
		t.Fatalf("Expected 1 hook, got %d", len(hooks))
	}
	if !hooks[0].ShowOutput {
		t.Error("ShowOutput should be true")
	}

	// Round-trip
	data, _ := adapter.Marshal(cfg)
	cfg2, _ := adapter.Parse(data)
	hooks2 := cfg2.GetAllHooksForEvent(core.BeforeCommand)
	if !hooks2[0].ShowOutput {
		t.Error("ShowOutput should be preserved after round-trip")
	}
}

func TestAdapterWorkingDirectoryPreserved(t *testing.T) {
	adapter := NewAdapter()

	json := `{
		"hooks": {
			"pre_run_command": [
				{"command": "echo test", "working_directory": "/custom/dir"}
			]
		}
	}`

	cfg, err := adapter.Parse([]byte(json))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	hooks := cfg.GetAllHooksForEvent(core.BeforeCommand)
	if len(hooks) != 1 {
		t.Fatalf("Expected 1 hook, got %d", len(hooks))
	}
	if hooks[0].WorkingDir != "/custom/dir" {
		t.Errorf("WorkingDir should be '/custom/dir', got %q", hooks[0].WorkingDir)
	}

	// Round-trip
	data, _ := adapter.Marshal(cfg)
	cfg2, _ := adapter.Parse(data)
	hooks2 := cfg2.GetAllHooksForEvent(core.BeforeCommand)
	if hooks2[0].WorkingDir != "/custom/dir" {
		t.Error("WorkingDir should be preserved after round-trip")
	}
}

func TestAdapterPromptHooksIgnored(t *testing.T) {
	adapter := NewAdapter()

	// Windsurf doesn't support prompt hooks, only command hooks
	cfg := core.NewConfig()
	cfg.AddHook(core.BeforeCommand, core.NewPromptHook("Is this safe?"))

	windsurfCfg := adapter.FromCore(cfg)

	// Should have no hooks since prompt hooks are not supported
	totalHooks := 0
	for _, h := range windsurfCfg.Hooks {
		totalHooks += len(h)
	}
	if totalHooks != 0 {
		t.Errorf("Expected 0 hooks (prompt hooks not supported), got %d", totalHooks)
	}
}

func TestAdapterUnknownEventIgnored(t *testing.T) {
	adapter := NewAdapter()

	// JSON with unknown event
	json := `{
		"hooks": {
			"unknown_event": [
				{"command": "echo unknown"}
			],
			"pre_run_command": [
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
}

func TestWorkspaceConfigPath(t *testing.T) {
	path := WorkspaceConfigPath()
	expected := filepath.Join(WorkspaceConfigDir, ConfigFileName)
	if path != expected {
		t.Errorf("WorkspaceConfigPath() = %q, want %q", path, expected)
	}
}

func TestUserConfigPath(t *testing.T) {
	path, err := UserConfigPath()
	if err != nil {
		t.Fatalf("UserConfigPath() error = %v", err)
	}
	if path == "" {
		t.Error("UserConfigPath() returned empty string")
	}
	// Should contain the user config dir
	if filepath.Base(filepath.Dir(path)) != "windsurf" {
		t.Errorf("UserConfigPath() should be in windsurf dir, got %q", path)
	}
}

func TestReadWorkspaceConfigNotFound(t *testing.T) {
	_, err := ReadWorkspaceConfig()
	if err == nil {
		t.Log("ReadWorkspaceConfig() didn't return error, file may exist")
	}
}

func TestReadUserConfigNotFound(t *testing.T) {
	_, err := ReadUserConfig()
	if err == nil {
		t.Log("ReadUserConfig() didn't return error, file may exist")
	}
}

func TestWriteWorkspaceConfig(t *testing.T) {
	// Create a temp directory to simulate the workspace directory
	origDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	tmpDir, err := os.MkdirTemp("", "windsurf-workspace-test")
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

	err = WriteWorkspaceConfig(cfg)
	if err != nil {
		t.Fatalf("WriteWorkspaceConfig() error = %v", err)
	}

	// Verify file was created
	readCfg, err := ReadWorkspaceConfig()
	if err != nil {
		t.Fatalf("ReadWorkspaceConfig() error = %v", err)
	}
	if readCfg.HookCount() != 1 {
		t.Errorf("Expected 1 hook, got %d", readCfg.HookCount())
	}
}

func TestAdapterFromCoreSkipsUnsupportedEvents(t *testing.T) {
	adapter := NewAdapter()

	cfg := core.NewConfig()
	// Add Claude-only event (not supported by Windsurf)
	cfg.AddHook(core.OnSessionStart, core.NewCommandHook("echo session"))
	// Add Windsurf-supported event
	cfg.AddHook(core.BeforeCommand, core.NewCommandHook("echo command"))

	windsurfCfg := adapter.FromCore(cfg)

	// Should only have BeforeCommand, not OnSessionStart
	totalHooks := 0
	for _, hooks := range windsurfCfg.Hooks {
		totalHooks += len(hooks)
	}
	if totalHooks != 1 {
		t.Errorf("Expected 1 hook (unsupported events filtered), got %d", totalHooks)
	}
}

func TestAdapterToCoreSkipsUnknownEvents(t *testing.T) {
	adapter := NewAdapter()

	// Config with unknown event
	windsurfCfg := &Config{
		Hooks: map[WindsurfEvent][]Hook{
			"unknown_event": {{Command: "echo unknown"}},
			PreRunCommand:   {{Command: "echo known"}},
		},
	}

	cfg := adapter.ToCore(windsurfCfg)

	// Should only have 1 hook (unknown event skipped)
	if cfg.HookCount() != 1 {
		t.Errorf("Expected 1 hook, got %d", cfg.HookCount())
	}
}

func TestAdapterAllSupportedEvents(t *testing.T) {
	adapter := NewAdapter()

	// Test all Windsurf events
	allEvents := []WindsurfEvent{
		PreReadCode, PostReadCode,
		PreWriteCode, PostWriteCode,
		PreRunCommand, PostRunCommand,
		PreMCPToolUse, PostMCPToolUse,
		PreUserPrompt,
	}

	for _, event := range allEvents {
		t.Run(string(event), func(t *testing.T) {
			windsurfCfg := &Config{
				Hooks: map[WindsurfEvent][]Hook{
					event: {{Command: "echo test"}},
				},
			}

			cfg := adapter.ToCore(windsurfCfg)
			if cfg.HookCount() != 1 {
				t.Errorf("Event %q should be converted, got %d hooks", event, cfg.HookCount())
			}
		})
	}
}
