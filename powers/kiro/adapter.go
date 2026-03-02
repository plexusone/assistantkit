// Package kiro provides the Kiro IDE power adapter.
//
// Kiro Powers are capability packages that bundle MCP servers, steering files,
// and hooks. They use dynamic keyword-based activation to load context efficiently.
//
// Power structure:
//
//	power-name/
//	├── POWER.md           # Frontmatter + onboarding + steering
//	├── mcp.json           # MCP server configuration (optional)
//	└── steering/          # Workflow-specific guidance (optional)
//	    ├── workflow-a.md
//	    └── workflow-b.md
package kiro

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/plexusone/assistantkit/powers/core"
)

const (
	// AdapterName is the identifier for this adapter.
	AdapterName = "kiro"

	// PowerFileName is the main power definition file.
	PowerFileName = "POWER.md"

	// MCPFileName is the MCP server configuration file.
	MCPFileName = "mcp.json"

	// SteeringDir is the directory for steering files.
	SteeringDir = "steering"
)

func init() {
	core.Register(&Adapter{})
}

// Adapter converts between canonical Power and Kiro IDE power format.
type Adapter struct{}

// Name returns the adapter identifier.
func (a *Adapter) Name() string {
	return AdapterName
}

// GeneratePowerDir creates a complete Kiro power directory structure.
func (a *Adapter) GeneratePowerDir(power *core.Power, outputDir string) ([]string, error) {
	if err := power.Validate(); err != nil {
		return nil, err
	}

	var createdFiles []string

	// Create output directory
	if err := os.MkdirAll(outputDir, core.DefaultDirMode); err != nil {
		return nil, &core.GenerateError{Format: AdapterName, Path: outputDir, Message: "failed to create directory", Err: err}
	}

	// Generate POWER.md
	powerMDPath := filepath.Join(outputDir, PowerFileName)
	powerMD := a.generatePowerMD(power)
	if err := os.WriteFile(powerMDPath, []byte(powerMD), core.DefaultFileMode); err != nil {
		return nil, &core.GenerateError{Format: AdapterName, Path: powerMDPath, Message: "failed to write POWER.md", Err: err}
	}
	createdFiles = append(createdFiles, powerMDPath)

	// Generate mcp.json if there are MCP servers
	if len(power.MCPServers) > 0 {
		mcpPath := filepath.Join(outputDir, MCPFileName)
		mcpConfig := a.generateMCPConfig(power)
		mcpJSON, err := json.MarshalIndent(mcpConfig, "", "  ")
		if err != nil {
			return nil, &core.GenerateError{Format: AdapterName, Path: mcpPath, Message: "failed to marshal MCP config", Err: err}
		}
		if err := os.WriteFile(mcpPath, mcpJSON, core.DefaultFileMode); err != nil {
			return nil, &core.GenerateError{Format: AdapterName, Path: mcpPath, Message: "failed to write mcp.json", Err: err}
		}
		createdFiles = append(createdFiles, mcpPath)
	}

	// Generate steering files
	if len(power.SteeringFiles) > 0 {
		steeringPath := filepath.Join(outputDir, SteeringDir)
		if err := os.MkdirAll(steeringPath, core.DefaultDirMode); err != nil {
			return nil, &core.GenerateError{Format: AdapterName, Path: steeringPath, Message: "failed to create steering directory", Err: err}
		}

		for name, sf := range power.SteeringFiles {
			// Determine file path
			var filePath string
			if sf.Path != "" {
				filePath = filepath.Join(outputDir, sf.Path)
			} else {
				filePath = filepath.Join(steeringPath, name+".md")
			}

			// Ensure parent directory exists
			if err := os.MkdirAll(filepath.Dir(filePath), core.DefaultDirMode); err != nil {
				return nil, &core.GenerateError{Format: AdapterName, Path: filePath, Message: "failed to create steering file directory", Err: err}
			}

			// Write content
			content := sf.Content
			if content == "" {
				content = fmt.Sprintf("# %s\n\n%s\n", name, sf.Description)
			}
			if err := os.WriteFile(filePath, []byte(content), core.DefaultFileMode); err != nil {
				return nil, &core.GenerateError{Format: AdapterName, Path: filePath, Message: "failed to write steering file", Err: err}
			}
			createdFiles = append(createdFiles, filePath)
		}
	}

	return createdFiles, nil
}

// ParsePowerDir reads a Kiro power directory and returns a canonical Power.
func (a *Adapter) ParsePowerDir(dir string) (*core.Power, error) {
	powerMDPath := filepath.Join(dir, PowerFileName)
	data, err := os.ReadFile(powerMDPath)
	if err != nil {
		return nil, &core.ParseError{Format: AdapterName, Path: powerMDPath, Err: err}
	}

	power, err := a.parsePowerMD(string(data))
	if err != nil {
		return nil, &core.ParseError{Format: AdapterName, Path: powerMDPath, Err: err}
	}

	// Parse mcp.json if it exists
	mcpPath := filepath.Join(dir, MCPFileName)
	if mcpData, err := os.ReadFile(mcpPath); err == nil {
		if err := a.parseMCPConfig(power, mcpData); err != nil {
			return nil, &core.ParseError{Format: AdapterName, Path: mcpPath, Err: err}
		}
	}

	return power, nil
}

// generatePowerMD generates the POWER.md content.
func (a *Adapter) generatePowerMD(power *core.Power) string {
	var sb strings.Builder

	// Frontmatter
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("name: %q\n", power.Name))
	if power.DisplayName != "" {
		sb.WriteString(fmt.Sprintf("displayName: %q\n", power.DisplayName))
	}
	if power.Description != "" {
		sb.WriteString(fmt.Sprintf("description: %q\n", power.Description))
	}
	if power.Version != "" {
		sb.WriteString(fmt.Sprintf("version: %q\n", power.Version))
	}
	if len(power.Keywords) > 0 {
		sb.WriteString("keywords:\n")
		for _, kw := range power.Keywords {
			sb.WriteString(fmt.Sprintf("  - %q\n", kw))
		}
	}
	sb.WriteString("---\n\n")

	// Title
	title := power.DisplayName
	if title == "" {
		title = power.Name
	}
	sb.WriteString(fmt.Sprintf("# %s\n\n", title))

	// Description
	if power.Description != "" {
		sb.WriteString(power.Description + "\n\n")
	}

	// Onboarding section
	if power.Onboarding != "" {
		sb.WriteString("## Onboarding\n\n")
		sb.WriteString(power.Onboarding + "\n\n")
	}

	// MCP Tools section
	if len(power.MCPServers) > 0 {
		sb.WriteString("## Available Tools\n\n")
		sb.WriteString("This power provides the following MCP servers:\n\n")
		for name, server := range power.MCPServers {
			sb.WriteString(fmt.Sprintf("### %s\n\n", name))
			if server.Description != "" {
				sb.WriteString(server.Description + "\n\n")
			}
			if server.Command != "" {
				sb.WriteString(fmt.Sprintf("**Command:** `%s`\n\n", server.Command))
			}
		}
	}

	// Steering files section
	if len(power.SteeringFiles) > 0 {
		sb.WriteString("## Workflows\n\n")
		sb.WriteString("This power includes steering for the following workflows:\n\n")
		for name, sf := range power.SteeringFiles {
			sb.WriteString(fmt.Sprintf("- **%s**", name))
			if sf.Description != "" {
				sb.WriteString(fmt.Sprintf(": %s", sf.Description))
			}
			if len(sf.Keywords) > 0 {
				sb.WriteString(fmt.Sprintf(" (triggers: %s)", strings.Join(sf.Keywords, ", ")))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// Main instructions
	if power.Instructions != "" {
		sb.WriteString("## Instructions\n\n")
		sb.WriteString(power.Instructions + "\n")
	}

	return sb.String()
}

// MCPConfig represents the mcp.json structure.
type MCPConfig struct {
	MCPServers map[string]MCPServerConfig `json:"mcpServers"`
}

// MCPServerConfig represents an MCP server in mcp.json.
type MCPServerConfig struct {
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	URL     string            `json:"url,omitempty"`
}

// generateMCPConfig creates the mcp.json structure.
func (a *Adapter) generateMCPConfig(power *core.Power) *MCPConfig {
	config := &MCPConfig{
		MCPServers: make(map[string]MCPServerConfig),
	}

	for name, server := range power.MCPServers {
		config.MCPServers[name] = MCPServerConfig{
			Command: server.Command,
			Args:    server.Args,
			Env:     server.Env,
			URL:     server.URL,
		}
	}

	return config
}

// parsePowerMD parses POWER.md content into a Power.
func (a *Adapter) parsePowerMD(content string) (*core.Power, error) {
	power := &core.Power{
		MCPServers:    make(map[string]core.MCPServer),
		SteeringFiles: make(map[string]core.SteeringFile),
	}

	// Simple frontmatter parsing
	if !strings.HasPrefix(content, "---") {
		return nil, fmt.Errorf("POWER.md must start with YAML frontmatter")
	}

	// Find end of frontmatter
	endIdx := strings.Index(content[3:], "---")
	if endIdx == -1 {
		return nil, fmt.Errorf("POWER.md frontmatter not closed")
	}

	frontmatter := content[3 : endIdx+3]
	body := content[endIdx+6:]

	// Parse frontmatter line by line
	lines := strings.Split(frontmatter, "\n")
	var currentKey string
	var inKeywords bool

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "- ") && inKeywords {
			// Keyword item
			kw := strings.Trim(strings.TrimPrefix(line, "- "), "\"")
			power.Keywords = append(power.Keywords, kw)
			continue
		}

		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			currentKey = strings.TrimSpace(parts[0])
			value := strings.Trim(strings.TrimSpace(parts[1]), "\"")

			inKeywords = currentKey == "keywords"

			switch currentKey {
			case "name":
				power.Name = value
			case "displayName":
				power.DisplayName = value
			case "description":
				power.Description = value
			case "version":
				power.Version = value
			}
		}
	}

	// Store body as instructions
	power.Instructions = strings.TrimSpace(body)

	return power, nil
}

// parseMCPConfig parses mcp.json and adds servers to the power.
func (a *Adapter) parseMCPConfig(power *core.Power, data []byte) error {
	var config MCPConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	for name, server := range config.MCPServers {
		power.MCPServers[name] = core.MCPServer{
			Command: server.Command,
			Args:    server.Args,
			Env:     server.Env,
			URL:     server.URL,
		}
	}

	return nil
}

// UserPowersPath returns the path to the user's powers directory.
func UserPowersPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".kiro", "powers"), nil
}

// InstallPower installs a power to the user's powers directory.
func InstallPower(power *core.Power) error {
	powersDir, err := UserPowersPath()
	if err != nil {
		return err
	}

	outputDir := filepath.Join(powersDir, power.Name)
	adapter := &Adapter{}
	_, err = adapter.GeneratePowerDir(power, outputDir)
	return err
}
