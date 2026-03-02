// Package agentkit provides an adapter for generating agentkit local server configurations.
// This enables local development with MCP server support, which serves as a stepping stone
// to AWS AgentCore deployment.
package agentkit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	multiagentspec "github.com/plexusone/multi-agent-spec/sdk/go"

	"github.com/plexusone/assistantkit/agents/core"
)

func init() {
	core.Register(&Adapter{})
}

// Adapter converts canonical Agent definitions to agentkit local config format.
type Adapter struct{}

// Name returns the adapter identifier.
func (a *Adapter) Name() string {
	return "agentkit"
}

// FileExtension returns the file extension for agentkit config files.
func (a *Adapter) FileExtension() string {
	return ".json"
}

// DefaultDir returns the default directory name for agentkit configs.
func (a *Adapter) DefaultDir() string {
	return "plugins/agentkit"
}

// Parse converts agentkit config bytes to canonical Agent.
func (a *Adapter) Parse(data []byte) (*core.Agent, error) {
	var cfg AgentConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, &core.ParseError{Format: "agentkit", Err: err}
	}
	return configToAgent(&cfg), nil
}

// Marshal converts canonical Agent to agentkit config bytes.
func (a *Adapter) Marshal(agent *core.Agent) ([]byte, error) {
	cfg := agentToConfig(agent)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, &core.MarshalError{Format: "agentkit", Err: err}
	}
	return append(data, '\n'), nil
}

// ReadFile reads from path and returns canonical Agent.
func (a *Adapter) ReadFile(path string) (*core.Agent, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &core.ReadError{Path: path, Err: err}
	}
	return a.Parse(data)
}

// WriteFile writes canonical Agent to path.
func (a *Adapter) WriteFile(agent *core.Agent, path string) error {
	data, err := a.Marshal(agent)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, core.DefaultDirMode); err != nil {
		return &core.WriteError{Path: path, Err: err}
	}

	if err := os.WriteFile(path, data, core.DefaultFileMode); err != nil {
		return &core.WriteError{Path: path, Err: err}
	}

	return nil
}

// AgentConfig matches agentkit/platforms/local AgentConfig structure.
type AgentConfig struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Instructions string   `json:"instructions"`
	Tools        []string `json:"tools"`
	Model        string   `json:"model,omitempty"`
	MaxTokens    int      `json:"max_tokens,omitempty"`
}

// Config is the full agentkit local configuration.
type Config struct {
	Mode      string        `json:"mode"`
	Workspace string        `json:"workspace"`
	Agents    []AgentConfig `json:"agents"`
	MCP       MCPConfig     `json:"mcp"`
	LLM       LLMConfig     `json:"llm"`
	Timeouts  TimeoutConfig `json:"timeouts"`
}

// MCPConfig configures the MCP server.
type MCPConfig struct {
	Enabled       bool   `json:"enabled"`
	Transport     string `json:"transport"`
	Port          int    `json:"port,omitempty"`
	ServerName    string `json:"server_name,omitempty"`
	ServerVersion string `json:"server_version,omitempty"`
}

// LLMConfig configures the language model.
type LLMConfig struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	//nolint:gosec // G117: Field name for env var template, not a hardcoded secret
	APIKey      string  `json:"api_key,omitempty"`
	BaseURL     string  `json:"base_url,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

// TimeoutConfig defines operation timeouts.
type TimeoutConfig struct {
	AgentInvoke   string `json:"agent_invoke"`
	ShellCommand  string `json:"shell_command"`
	FileRead      string `json:"file_read"`
	ParallelTotal string `json:"parallel_total"`
}

// agentKitModelMapping maps canonical models to agentkit local model strings.
// AgentKit uses full model strings rather than Bedrock ARNs.
var agentKitModelMapping = map[string]string{
	"haiku":  "claude-3-haiku-20240307",
	"sonnet": "claude-3-5-sonnet-20241022",
	"opus":   "claude-3-opus-20240229",
}

// mapToolToAgentKit converts a canonical tool string to AgentKit tool using multi-agent-spec.
func mapToolToAgentKit(tool string) string {
	return multiagentspec.MapToolToAgentKit(multiagentspec.Tool(tool))
}

// mapModelToAgentKit converts a canonical model to AgentKit model string.
func mapModelToAgentKit(model core.Model) string {
	if mapped, ok := agentKitModelMapping[string(model)]; ok {
		return mapped
	}
	return string(model)
}

func agentToConfig(agent *core.Agent) *AgentConfig {
	cfg := &AgentConfig{
		Name:         agent.Name,
		Description:  agent.Description,
		Instructions: agent.Instructions,
	}

	// Map tools using multi-agent-spec mappings
	toolSet := make(map[string]bool)
	for _, tool := range agent.Tools {
		mapped := mapToolToAgentKit(tool)
		if mapped != "" {
			toolSet[mapped] = true
		} else {
			// Keep unknown tools as-is (lowercase)
			toolSet[strings.ToLower(tool)] = true
		}
	}
	for tool := range toolSet {
		cfg.Tools = append(cfg.Tools, tool)
	}

	// Map model
	if agent.Model != "" {
		cfg.Model = mapModelToAgentKit(agent.Model)
	}

	return cfg
}

// reverseAgentKitToolMapping provides reverse mapping from AgentKit tools to canonical.
// Note: Some AgentKit tools map to multiple canonical tools, so we pick a representative.
var reverseAgentKitToolMapping = map[string]string{
	"shell": "Bash", // shell could be WebSearch, WebFetch, Bash, or Task - default to Bash
	"read":  "Read",
	"write": "Write", // write could be Write or Edit - default to Write
	"glob":  "Glob",
	"grep":  "Grep",
}

func configToAgent(cfg *AgentConfig) *core.Agent {
	agent := &core.Agent{
		Name:         cfg.Name,
		Description:  cfg.Description,
		Instructions: cfg.Instructions,
		Model:        core.Model(cfg.Model),
	}

	// Reverse map tools
	for _, tool := range cfg.Tools {
		if mapped, ok := reverseAgentKitToolMapping[tool]; ok {
			agent.Tools = append(agent.Tools, mapped)
		} else {
			agent.Tools = append(agent.Tools, tool)
		}
	}

	return agent
}

// DefaultConfig returns a default agentkit configuration.
func DefaultConfig() *Config {
	return &Config{
		Mode:      "local",
		Workspace: ".",
		Agents:    []AgentConfig{},
		MCP: MCPConfig{
			Enabled:       true,
			Transport:     "stdio",
			ServerName:    "agentkit-local",
			ServerVersion: "1.0.0",
		},
		//nolint:gosec // G101: Environment variable template, not a hardcoded credential
		LLM: LLMConfig{
			Provider:    "anthropic",
			Model:       "claude-3-5-sonnet-20241022",
			APIKey:      "${ANTHROPIC_API_KEY}",
			Temperature: 0.7,
		},
		Timeouts: TimeoutConfig{
			AgentInvoke:   "5m",
			ShellCommand:  "2m",
			FileRead:      "30s",
			ParallelTotal: "10m",
		},
	}
}

// GenerateFullConfig creates a complete agentkit config from multiple agents.
func GenerateFullConfig(agents []*core.Agent) *Config {
	cfg := DefaultConfig()
	for _, agent := range agents {
		cfg.Agents = append(cfg.Agents, *agentToConfig(agent))
	}
	return cfg
}

// WriteFullConfig writes a complete agentkit configuration file.
func WriteFullConfig(agents []*core.Agent, path string) error {
	cfg := GenerateFullConfig(agents)

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return &core.MarshalError{Format: "agentkit", Err: err}
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, core.DefaultDirMode); err != nil {
		return &core.WriteError{Path: path, Err: err}
	}

	if err := os.WriteFile(path, append(data, '\n'), core.DefaultFileMode); err != nil {
		return &core.WriteError{Path: path, Err: err}
	}

	return nil
}
