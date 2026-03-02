package cursor

import "github.com/plexusone/assistantkit/mcp/claude"

// Config is an alias for Claude's config since the format is identical.
type Config = claude.Config

// ServerConfig is an alias for Claude's server config.
type ServerConfig = claude.ServerConfig

// NewConfig creates a new Cursor config.
func NewConfig() *Config {
	return claude.NewConfig()
}
