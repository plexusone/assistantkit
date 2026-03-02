// Package gemini provides the Gemini CLI agent adapter.
package gemini

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/plexusone/assistantkit/agents/core"
)

func init() {
	core.Register(&Adapter{})
}

// Adapter converts between canonical Agent and Gemini CLI agent format.
type Adapter struct{}

// Name returns the adapter identifier.
func (a *Adapter) Name() string {
	return "gemini"
}

// FileExtension returns the file extension for Gemini agents.
func (a *Adapter) FileExtension() string {
	return ".toml"
}

// DefaultDir returns the default directory name for Gemini agents.
func (a *Adapter) DefaultDir() string {
	return "agents"
}

// GeminiAgent represents a Gemini CLI agent in TOML format.
type GeminiAgent struct {
	Agent        AgentSection `toml:"agent"`
	Instructions string       `toml:"instructions,multiline"`
}

// AgentSection contains agent metadata.
type AgentSection struct {
	Name         string   `toml:"name"`
	Description  string   `toml:"description"`
	Model        string   `toml:"model,omitempty"`
	Tools        []string `toml:"tools,omitempty"`
	Skills       []string `toml:"skills,omitempty"`
	Dependencies []string `toml:"dependencies,omitempty"`
}

// Parse converts Gemini agent TOML bytes to canonical Agent.
func (a *Adapter) Parse(data []byte) (*core.Agent, error) {
	var ga GeminiAgent
	if err := toml.Unmarshal(data, &ga); err != nil {
		return nil, &core.ParseError{Format: "gemini", Err: err}
	}

	agent := &core.Agent{
		Name:         ga.Agent.Name,
		Description:  ga.Agent.Description,
		Model:        mapGeminiModelToCanonical(ga.Agent.Model),
		Tools:        ga.Agent.Tools,
		Skills:       ga.Agent.Skills,
		Dependencies: ga.Agent.Dependencies,
		Instructions: ga.Instructions,
	}

	return agent, nil
}

// Marshal converts canonical Agent to Gemini agent TOML bytes.
func (a *Adapter) Marshal(agent *core.Agent) ([]byte, error) {
	ga := GeminiAgent{
		Agent: AgentSection{
			Name:         agent.Name,
			Description:  agent.Description,
			Model:        mapCanonicalModelToGemini(agent.Model),
			Tools:        agent.Tools,
			Skills:       agent.Skills,
			Dependencies: agent.Dependencies,
		},
		Instructions: agent.Instructions,
	}

	data, err := toml.Marshal(ga)
	if err != nil {
		return nil, &core.MarshalError{Format: "gemini", Err: err}
	}

	return data, nil
}

// ReadFile reads a Gemini agent TOML file and returns canonical Agent.
func (a *Adapter) ReadFile(path string) (*core.Agent, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &core.ReadError{Path: path, Err: err}
	}

	agent, err := a.Parse(data)
	if err != nil {
		if pe, ok := err.(*core.ParseError); ok {
			pe.Path = path
		}
		return nil, err
	}

	// Infer name from filename if not set
	if agent.Name == "" {
		base := filepath.Base(path)
		agent.Name = strings.TrimSuffix(base, filepath.Ext(base))
	}

	return agent, nil
}

// WriteFile writes canonical Agent to a Gemini agent TOML file.
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

// mapGeminiModelToCanonical maps Gemini model names to canonical names.
func mapGeminiModelToCanonical(geminiModel string) core.Model {
	switch strings.ToLower(geminiModel) {
	case "gemini-2.0-flash", "flash":
		return core.ModelHaiku
	case "gemini-2.0-pro", "pro":
		return core.ModelSonnet
	case "gemini-2.0-ultra", "ultra":
		return core.ModelOpus
	default:
		return core.Model(geminiModel)
	}
}

// mapCanonicalModelToGemini maps canonical model names to Gemini names.
func mapCanonicalModelToGemini(model core.Model) string {
	switch model {
	case core.ModelHaiku:
		return "gemini-2.0-flash"
	case core.ModelSonnet:
		return "gemini-2.0-pro"
	case core.ModelOpus:
		return "gemini-2.0-ultra"
	default:
		return string(model)
	}
}
