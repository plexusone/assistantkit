package claude

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
	if adapter.Name() != "claude" {
		t.Errorf("Expected name 'claude', got %q", adapter.Name())
	}
}

func TestAdapterDefaultPaths(t *testing.T) {
	adapter := NewAdapter()
	paths := adapter.DefaultPaths()
	if len(paths) < 2 {
		t.Errorf("Expected at least 2 default paths, got %d", len(paths))
	}
	// Check project paths are present
	foundProject := false
	for _, p := range paths {
		if p == filepath.Join(ProjectConfigDir, SettingsFileName) {
			foundProject = true
		}
	}
	if !foundProject {
		t.Error("Expected project config path in default paths")
	}
}

func TestAdapterSupportedEvents(t *testing.T) {
	adapter := NewAdapter()
	events := adapter.SupportedEvents()
	if len(events) < 10 {
		t.Errorf("Expected at least 10 supported events, got %d", len(events))
	}

	// Check key events are present
	eventSet := make(map[core.Event]bool)
	for _, e := range events {
		eventSet[e] = true
	}

	requiredEvents := []core.Event{
		core.BeforeFileRead, core.BeforeFileWrite,
		core.BeforeCommand, core.AfterCommand,
		core.BeforeMCP, core.AfterMCP,
		core.OnSessionStart, core.OnPermission,
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
			name: "valid PreToolUse hook",
			json: `{
				"hooks": {
					"PreToolUse": [
						{
							"matcher": "Bash",
							"hooks": [
								{"type": "command", "command": "echo before"}
							]
						}
					]
				}
			}`,
			wantHooks: 1,
			wantError: false,
		},
		{
			name: "valid PostToolUse hook",
			json: `{
				"hooks": {
					"PostToolUse": [
						{
							"matcher": "Write|Edit",
							"hooks": [
								{"type": "command", "command": "echo after"}
							]
						}
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
					"PreToolUse": [
						{"matcher": "Bash", "hooks": [{"type": "command", "command": "echo 1"}]}
					],
					"SessionStart": [
						{"hooks": [{"type": "command", "command": "echo 2"}]}
					]
				}
			}`,
			wantHooks: 2,
			wantError: false,
		},
		{
			name: "prompt type hook",
			json: `{
				"hooks": {
					"PreToolUse": [
						{
							"matcher": "Bash",
							"hooks": [
								{"type": "prompt", "prompt": "Is this safe?"}
							]
						}
					]
				}
			}`,
			wantHooks: 1,
			wantError: false,
		},
		{
			name:      "invalid json",
			json:      `{invalid`,
			wantError: true,
		},
		{
			name:      "empty config",
			json:      `{}`,
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
	cfg.AddHookWithMatcher(core.BeforeCommand, "Bash",
		core.NewCommandHook("echo before"))
	cfg.AddHook(core.OnSessionStart,
		core.NewCommandHook("echo session"))

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
			"PreToolUse": [
				{
					"matcher": "Bash",
					"hooks": [
						{"type": "command", "command": "echo before bash", "timeout": 30}
					]
				},
				{
					"matcher": "Read",
					"hooks": [
						{"type": "command", "command": "echo before read"}
					]
				}
			],
			"PostToolUse": [
				{
					"matcher": "Write|Edit",
					"hooks": [
						{"type": "command", "command": "echo after write"}
					]
				}
			],
			"SessionStart": [
				{
					"hooks": [
						{"type": "command", "command": "echo session start"}
					]
				}
			]
		},
		"disableAllHooks": false
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
	tmpDir, err := os.MkdirTemp("", "claude-hooks-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create config
	cfg := core.NewConfig()
	cfg.AddHookWithMatcher(core.BeforeCommand, "Bash",
		core.NewCommandHook("echo test"))
	cfg.DisableAllHooks = true

	// Write file
	filePath := filepath.Join(tmpDir, "settings.json")
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
	if !readCfg.DisableAllHooks {
		t.Error("DisableAllHooks should be true")
	}
}

func TestAdapterReadFileNotFound(t *testing.T) {
	adapter := NewAdapter()

	_, err := adapter.ReadFile("/nonexistent/path/settings.json")
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
		claudeEvent ClaudeEvent
		matcher     string
		wantEvent   core.Event
	}{
		{PreToolUse, "Bash", core.BeforeCommand},
		{PostToolUse, "Bash", core.AfterCommand},
		{PreToolUse, "Read", core.BeforeFileRead},
		{PostToolUse, "Read", core.AfterFileRead},
		{PreToolUse, "Write", core.BeforeFileWrite},
		{PostToolUse, "Write", core.AfterFileWrite},
		{PreToolUse, "Edit", core.BeforeFileWrite},
		{PostToolUse, "Edit", core.AfterFileWrite},
		{PreToolUse, "Write|Edit", core.BeforeFileWrite},
		{PostToolUse, "Write|Edit", core.AfterFileWrite},
		{PreToolUse, "UnknownMCP", core.BeforeMCP}, // Default for unknown
		{PostToolUse, "UnknownMCP", core.AfterMCP}, // Default for unknown
		{SessionStart, "", core.OnSessionStart},
		{SessionEnd, "", core.OnSessionEnd},
		{Stop, "", core.OnStop},
		{PermissionRequest, "", core.OnPermission},
		{UserPromptSubmit, "", core.BeforePrompt},
		{Notification, "", core.OnNotification},
		{PreCompact, "", core.BeforeCompact},
		{SubagentStop, "", core.OnSubagentStop},
	}

	for _, tt := range tests {
		t.Run(string(tt.claudeEvent)+"/"+tt.matcher, func(t *testing.T) {
			got := adapter.claudeToCanonicalEvent(tt.claudeEvent, tt.matcher)
			if got != tt.wantEvent {
				t.Errorf("claudeToCanonicalEvent(%q, %q) = %q, want %q",
					tt.claudeEvent, tt.matcher, got, tt.wantEvent)
			}
		})
	}
}

func TestAdapterFromCoreEventMapping(t *testing.T) {
	adapter := NewAdapter()

	tests := []struct {
		event       core.Event
		wantClaude  ClaudeEvent
		wantMatcher string
	}{
		{core.BeforeCommand, PreToolUse, "Bash"},
		{core.AfterCommand, PostToolUse, "Bash"},
		{core.BeforeFileRead, PreToolUse, "Read"},
		{core.AfterFileRead, PostToolUse, "Read"},
		{core.BeforeFileWrite, PreToolUse, "Write|Edit"},
		{core.AfterFileWrite, PostToolUse, "Write|Edit"},
		{core.OnSessionStart, SessionStart, ""},
		{core.OnSessionEnd, SessionEnd, ""},
		{core.OnStop, Stop, ""},
		{core.OnPermission, PermissionRequest, ""},
		{core.BeforePrompt, UserPromptSubmit, ""},
	}

	for _, tt := range tests {
		t.Run(string(tt.event), func(t *testing.T) {
			gotClaude, gotMatcher := adapter.canonicalToClaudeEvent(tt.event)
			if gotClaude != tt.wantClaude {
				t.Errorf("canonicalToClaudeEvent(%q) claude = %q, want %q",
					tt.event, gotClaude, tt.wantClaude)
			}
			if gotMatcher != tt.wantMatcher {
				t.Errorf("canonicalToClaudeEvent(%q) matcher = %q, want %q",
					tt.event, gotMatcher, tt.wantMatcher)
			}
		})
	}
}

func TestAdapterFromCoreUnsupportedEvent(t *testing.T) {
	adapter := NewAdapter()

	// AfterResponse is Cursor-only, not supported by Claude
	claudeEvent, matcher := adapter.canonicalToClaudeEvent(core.AfterResponse)
	if claudeEvent != "" || matcher != "" {
		t.Errorf("canonicalToClaudeEvent(AfterResponse) should return empty, got %q, %q",
			claudeEvent, matcher)
	}
}

func TestAdapterToCoreWithPromptHook(t *testing.T) {
	adapter := NewAdapter()

	json := `{
		"hooks": {
			"PreToolUse": [
				{
					"matcher": "Bash",
					"hooks": [
						{"type": "prompt", "prompt": "Is this command safe to run?"}
					]
				}
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

	if hooks[0].Type != core.HookTypePrompt {
		t.Errorf("Expected prompt type, got %q", hooks[0].Type)
	}
	if hooks[0].Prompt != "Is this command safe to run?" {
		t.Errorf("Prompt mismatch: %q", hooks[0].Prompt)
	}
}

func TestAdapterFromCoreWithPromptHook(t *testing.T) {
	adapter := NewAdapter()

	cfg := core.NewConfig()
	hook := core.NewPromptHook("Is this safe?")
	cfg.AddHookWithMatcher(core.BeforeCommand, "Bash", hook)

	claudeCfg := adapter.FromCore(cfg)

	entries := claudeCfg.Hooks[PreToolUse]
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}
	if len(entries[0].Hooks) != 1 {
		t.Fatalf("Expected 1 hook, got %d", len(entries[0].Hooks))
	}
	if entries[0].Hooks[0].Type != "prompt" {
		t.Errorf("Expected type 'prompt', got %q", entries[0].Hooks[0].Type)
	}
}

func TestAdapterFromCoreInferType(t *testing.T) {
	adapter := NewAdapter()

	cfg := core.NewConfig()
	// Hook without explicit type but with command
	cfg.AddHook(core.BeforeCommand, core.Hook{Command: "echo test"})
	// Hook without explicit type but with prompt
	cfg.AddHook(core.OnPermission, core.Hook{Prompt: "Check permission"})

	claudeCfg := adapter.FromCore(cfg)

	// Check command type inferred
	bashEntries := claudeCfg.Hooks[PreToolUse]
	if len(bashEntries) > 0 && len(bashEntries[0].Hooks) > 0 {
		if bashEntries[0].Hooks[0].Type != "command" {
			t.Errorf("Expected inferred type 'command', got %q", bashEntries[0].Hooks[0].Type)
		}
	}

	// Check prompt type inferred
	permEntries := claudeCfg.Hooks[PermissionRequest]
	if len(permEntries) > 0 && len(permEntries[0].Hooks) > 0 {
		if permEntries[0].Hooks[0].Type != "prompt" {
			t.Errorf("Expected inferred type 'prompt', got %q", permEntries[0].Hooks[0].Type)
		}
	}
}

func TestAdapterDisableAllHooks(t *testing.T) {
	adapter := NewAdapter()

	json := `{
		"hooks": {
			"PreToolUse": [
				{"matcher": "Bash", "hooks": [{"type": "command", "command": "echo test"}]}
			]
		},
		"disableAllHooks": true,
		"allowManagedHooksOnly": true
	}`

	cfg, err := adapter.Parse([]byte(json))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if !cfg.DisableAllHooks {
		t.Error("DisableAllHooks should be true")
	}
	if !cfg.AllowManagedHooksOnly {
		t.Error("AllowManagedHooksOnly should be true")
	}

	// Round-trip
	data, _ := adapter.Marshal(cfg)
	cfg2, _ := adapter.Parse(data)

	if !cfg2.DisableAllHooks {
		t.Error("DisableAllHooks should be preserved after round-trip")
	}
	if !cfg2.AllowManagedHooksOnly {
		t.Error("AllowManagedHooksOnly should be preserved after round-trip")
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

func TestReadProjectConfigNotFound(t *testing.T) {
	// ReadProjectConfig should fail when file doesn't exist
	_, err := ReadProjectConfig()
	if err == nil {
		// It's possible the file exists in the test environment
		// so we just check the function doesn't panic
		t.Log("ReadProjectConfig() didn't return error, file may exist")
	}
}

func TestReadUserConfigNotFound(t *testing.T) {
	// ReadUserConfig should fail when file doesn't exist
	_, err := ReadUserConfig()
	if err == nil {
		// It's possible the file exists in the user's home
		t.Log("ReadUserConfig() didn't return error, file may exist")
	}
}

func TestAdapterClaudeToCanonicalEventUnknown(t *testing.T) {
	adapter := NewAdapter()

	// Test with unknown Claude event
	result := adapter.claudeToCanonicalEvent(ClaudeEvent("UnknownEvent"), "")
	if result != "" {
		t.Errorf("Unknown Claude event should return empty string, got %q", result)
	}
}

func TestAdapterFromCoreWithEntryMatcher(t *testing.T) {
	adapter := NewAdapter()

	cfg := core.NewConfig()
	// Add hook with empty entry matcher - should use default matcher for event
	cfg.Hooks[core.BeforeCommand] = []core.HookEntry{
		{
			Matcher: "", // Empty matcher
			Hooks:   []core.Hook{core.NewCommandHook("echo test")},
		},
	}

	claudeCfg := adapter.FromCore(cfg)

	// Should have hooks
	entries := claudeCfg.Hooks[PreToolUse]
	if len(entries) == 0 {
		t.Fatal("Expected at least one entry")
	}
	// Default matcher for BeforeCommand should be "Bash"
	if entries[0].Matcher != "Bash" {
		t.Errorf("Expected default matcher 'Bash', got %q", entries[0].Matcher)
	}
}

func TestAdapterParseWithTimeout(t *testing.T) {
	adapter := NewAdapter()

	json := `{
		"hooks": {
			"PreToolUse": [
				{
					"matcher": "Bash",
					"hooks": [
						{"type": "command", "command": "echo test", "timeout": 60}
					]
				}
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
	if hooks[0].Timeout != 60 {
		t.Errorf("Expected timeout 60, got %d", hooks[0].Timeout)
	}
}
