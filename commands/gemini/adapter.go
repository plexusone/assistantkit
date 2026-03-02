// Package gemini provides the Gemini CLI command adapter.
package gemini

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/plexusone/assistantkit/commands/core"
)

func init() {
	core.Register(&Adapter{})
}

// Adapter converts between canonical Command and Gemini CLI command format.
type Adapter struct{}

// Name returns the adapter identifier.
func (a *Adapter) Name() string {
	return "gemini"
}

// FileExtension returns the file extension for Gemini commands.
func (a *Adapter) FileExtension() string {
	return ".toml"
}

// DefaultDir returns the default directory name for Gemini commands.
func (a *Adapter) DefaultDir() string {
	return "commands"
}

// GeminiCommand represents a Gemini CLI command in TOML format.
type GeminiCommand struct {
	Command   CommandSection `toml:"command"`
	Arguments []ArgumentToml `toml:"arguments,omitempty"`
	Content   ContentSection `toml:"content"`
	Process   []string       `toml:"process,omitempty"`
	Examples  []ExampleToml  `toml:"examples,omitempty"`
}

// CommandSection contains command metadata.
type CommandSection struct {
	Name        string `toml:"name"`
	Description string `toml:"description"`
}

// ArgumentToml represents an argument in TOML format.
type ArgumentToml struct {
	Name        string `toml:"name"`
	Type        string `toml:"type,omitempty"`
	Required    bool   `toml:"required,omitempty"`
	Default     string `toml:"default,omitempty"`
	Hint        string `toml:"hint,omitempty"`
	Description string `toml:"description,omitempty"`
}

// ContentSection contains the command instructions.
type ContentSection struct {
	Instructions string `toml:"instructions"`
}

// ExampleToml represents an example in TOML format.
type ExampleToml struct {
	Description string `toml:"description,omitempty"`
	Input       string `toml:"input"`
	Output      string `toml:"output,omitempty"`
}

// Parse converts Gemini command TOML bytes to canonical Command.
func (a *Adapter) Parse(data []byte) (*core.Command, error) {
	var gc GeminiCommand
	if err := toml.Unmarshal(data, &gc); err != nil {
		return nil, &core.ParseError{Format: "gemini", Err: err}
	}

	cmd := &core.Command{
		Name:         gc.Command.Name,
		Description:  gc.Command.Description,
		Instructions: gc.Content.Instructions,
		Process:      gc.Process,
	}

	// Convert arguments
	for _, arg := range gc.Arguments {
		cmd.Arguments = append(cmd.Arguments, core.Argument{
			Name:        arg.Name,
			Type:        arg.Type,
			Required:    arg.Required,
			Default:     arg.Default,
			Hint:        arg.Hint,
			Description: arg.Description,
		})
	}

	// Convert examples
	for _, ex := range gc.Examples {
		cmd.Examples = append(cmd.Examples, core.Example{
			Description: ex.Description,
			Input:       ex.Input,
			Output:      ex.Output,
		})
	}

	return cmd, nil
}

// Marshal converts canonical Command to Gemini command TOML bytes.
func (a *Adapter) Marshal(cmd *core.Command) ([]byte, error) {
	gc := GeminiCommand{
		Command: CommandSection{
			Name:        cmd.Name,
			Description: cmd.Description,
		},
		Content: ContentSection{
			Instructions: cmd.Instructions,
		},
		Process: cmd.Process,
	}

	// Convert arguments
	for _, arg := range cmd.Arguments {
		gc.Arguments = append(gc.Arguments, ArgumentToml{
			Name:        arg.Name,
			Type:        arg.Type,
			Required:    arg.Required,
			Default:     arg.Default,
			Hint:        arg.Hint,
			Description: arg.Description,
		})
	}

	// Convert examples
	for _, ex := range cmd.Examples {
		gc.Examples = append(gc.Examples, ExampleToml{
			Description: ex.Description,
			Input:       ex.Input,
			Output:      ex.Output,
		})
	}

	data, err := toml.Marshal(gc)
	if err != nil {
		return nil, &core.MarshalError{Format: "gemini", Err: err}
	}

	return data, nil
}

// ReadFile reads a Gemini command TOML file and returns canonical Command.
func (a *Adapter) ReadFile(path string) (*core.Command, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &core.ReadError{Path: path, Err: err}
	}

	cmd, err := a.Parse(data)
	if err != nil {
		if pe, ok := err.(*core.ParseError); ok {
			pe.Path = path
		}
		return nil, err
	}

	// Infer name from filename if not set
	if cmd.Name == "" {
		base := filepath.Base(path)
		cmd.Name = strings.TrimSuffix(base, filepath.Ext(base))
	}

	return cmd, nil
}

// WriteFile writes canonical Command to a Gemini command TOML file.
func (a *Adapter) WriteFile(cmd *core.Command, path string) error {
	data, err := a.Marshal(cmd)
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
