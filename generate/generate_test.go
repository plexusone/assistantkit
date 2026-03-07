package generate

import (
	"strings"
	"testing"

	"github.com/plexusone/assistantkit/agents"
	"github.com/plexusone/assistantkit/commands"
	"github.com/plexusone/assistantkit/plugins/core"
	"github.com/plexusone/assistantkit/skills"
)

func TestGeneratePlatformPlugin_RejectsSubdirectoryPaths(t *testing.T) {
	tests := []struct {
		name      string
		outputDir string
		wantErr   bool
		errSubstr string
	}{
		{
			name:      "agents subdirectory rejected",
			outputDir: "plugins/kiro/agents",
			wantErr:   true,
			errSubstr: "should be the plugin root directory",
		},
		{
			name:      "steering subdirectory rejected",
			outputDir: "plugins/kiro/steering",
			wantErr:   true,
			errSubstr: "should be the plugin root directory",
		},
		{
			name:      "skills subdirectory rejected",
			outputDir: "plugins/claude/skills",
			wantErr:   true,
			errSubstr: "should be the plugin root directory",
		},
		{
			name:      "commands subdirectory rejected",
			outputDir: "plugins/claude/commands",
			wantErr:   true,
			errSubstr: "should be the plugin root directory",
		},
		{
			name:      ".claude/agents rejected",
			outputDir: ".claude/agents",
			wantErr:   true,
			errSubstr: "should be the plugin root directory",
		},
		{
			name:      "valid plugin root accepted",
			outputDir: "plugins/kiro",
			wantErr:   false,
		},
		{
			name:      "valid .claude root accepted",
			outputDir: ".claude",
			wantErr:   false,
		},
		{
			name:      "valid nested path accepted",
			outputDir: "output/plugins/claude",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := DeploymentTarget{
				Name:     "test",
				Platform: "claude-code",
				Output:   tt.outputDir,
			}

			// Use temp dir for valid paths to avoid creating directories
			outputDir := t.TempDir() + "/" + tt.outputDir

			err := generatePlatformPlugin(
				target,
				outputDir,
				&PluginSpec{Plugin: core.Plugin{Name: "test", Version: "1.0.0"}},
				[]*commands.Command{},
				[]*skills.Skill{},
				[]*agents.Agent{},
			)

			if tt.wantErr {
				if err == nil {
					t.Errorf("generatePlatformPlugin() error = nil, want error containing %q", tt.errSubstr)
					return
				}
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("generatePlatformPlugin() error = %v, want error containing %q", err, tt.errSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("generatePlatformPlugin() unexpected error = %v", err)
				}
			}
		})
	}
}
