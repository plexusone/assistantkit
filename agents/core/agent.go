// Package core provides the canonical agent definition types.
// Agent definitions use the multi-agent-spec types as the canonical form,
// which maps losslessly to Claude Code, Kiro CLI, and OpenAI Codex.
package core

import (
	multiagentspec "github.com/plexusone/multi-agent-spec/sdk/go"
)

// Agent is an alias for multiagentspec.Agent.
// This is the canonical agent definition type used across all platforms.
type Agent = multiagentspec.Agent

// Task is an alias for multiagentspec.Task.
type Task = multiagentspec.Task

// Model is an alias for multiagentspec.Model.
type Model = multiagentspec.Model

// Model constants from multiagentspec.
const (
	ModelHaiku  = multiagentspec.ModelHaiku
	ModelSonnet = multiagentspec.ModelSonnet
	ModelOpus   = multiagentspec.ModelOpus
)

// NewAgent creates a new Agent with the given name and description.
func NewAgent(name, description string) *Agent {
	return multiagentspec.NewAgent(name, description)
}
