// Package generate provides functions for generating platform-specific plugins
// from canonical JSON specifications.
//
// This package is the core library used by the assistantkit CLI and can be
// used directly by projects that need programmatic plugin generation.
//
// Example usage:
//
//	result, err := generate.Plugins("plugins/spec", "plugins", []string{"claude", "kiro"})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Generated %d commands, %d skills\n", result.CommandCount, result.SkillCount)
package generate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/plexusone/assistantkit/agents"
	"github.com/plexusone/assistantkit/commands"
	"github.com/plexusone/assistantkit/plugins"
	powercore "github.com/plexusone/assistantkit/powers/core"
	"github.com/plexusone/assistantkit/powers/kiro"
	"github.com/plexusone/assistantkit/skills"
)

// Result contains the results of plugin generation.
type Result struct {
	// CommandCount is the number of commands loaded.
	CommandCount int

	// SkillCount is the number of skills loaded.
	SkillCount int

	// AgentCount is the number of agents loaded.
	AgentCount int

	// GeneratedDirs maps platform names to their output directories.
	GeneratedDirs map[string]string
}

// PluginSpec extends the base Plugin with power-specific fields.
type PluginSpec struct {
	plugins.Plugin
	DisplayName string               `json:"displayName,omitempty"`
	Keywords    []string             `json:"keywords,omitempty"`
	MCPServers  map[string]MCPServer `json:"mcpServers,omitempty"`
}

// MCPServer defines an MCP server configuration.
type MCPServer struct {
	Command     string   `json:"command"`
	Args        []string `json:"args,omitempty"`
	Description string   `json:"description,omitempty"`
}

// Plugins generates platform-specific plugins from a canonical spec directory.
//
// The specDir should contain:
//   - plugin.json: Plugin metadata
//   - commands/: Command definitions (*.json)
//   - skills/: Skill definitions (*.json)
//   - agents/: Agent definitions (*.json)
//
// Generated plugins are written to outputDir/<platform>/.
func Plugins(specDir, outputDir string, platforms []string) (*Result, error) {
	result := &Result{
		GeneratedDirs: make(map[string]string),
	}

	// Load canonical specs
	plugin, err := loadPlugin(filepath.Join(specDir, "plugin.json"))
	if err != nil {
		return nil, fmt.Errorf("loading plugin spec: %w", err)
	}

	cmds, err := loadCommands(filepath.Join(specDir, "commands"))
	if err != nil {
		return nil, fmt.Errorf("loading commands: %w", err)
	}
	result.CommandCount = len(cmds)

	skls, err := loadSkills(filepath.Join(specDir, "skills"))
	if err != nil {
		return nil, fmt.Errorf("loading skills: %w", err)
	}
	result.SkillCount = len(skls)

	agts, err := loadAgents(filepath.Join(specDir, "agents"))
	if err != nil {
		return nil, fmt.Errorf("loading agents: %w", err)
	}
	result.AgentCount = len(agts)

	// Generate each platform
	for _, platform := range platforms {
		platformDir := filepath.Join(outputDir, platform)

		switch platform {
		case "claude":
			if err := generateClaude(platformDir, plugin, cmds, skls, agts); err != nil {
				return nil, fmt.Errorf("generating claude: %w", err)
			}
		case "kiro":
			if err := generateKiro(platformDir, plugin, skls, agts); err != nil {
				return nil, fmt.Errorf("generating kiro: %w", err)
			}
		case "gemini":
			if err := generateGemini(platformDir, plugin, cmds); err != nil {
				return nil, fmt.Errorf("generating gemini: %w", err)
			}
		default:
			return nil, fmt.Errorf("unknown platform: %s", platform)
		}

		result.GeneratedDirs[platform] = platformDir
	}

	return result, nil
}

func loadPlugin(path string) (*PluginSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var plugin PluginSpec
	if err := json.Unmarshal(data, &plugin); err != nil {
		return nil, err
	}

	return &plugin, nil
}

func loadCommands(dir string) ([]*commands.Command, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil // Commands are optional
	}

	// Use ReadCanonicalDir which supports both .json and .md files
	return commands.ReadCanonicalDir(dir)
}

func loadSkills(dir string) ([]*skills.Skill, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil // Skills are optional
	}

	// Use ReadCanonicalDir which supports both .json and .md files
	return skills.ReadCanonicalDir(dir)
}

func loadAgents(dir string) ([]*agents.Agent, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil // Agents are optional
	}

	// Use agents.ReadCanonicalDir which supports both .md (multi-agent-spec) and .json files
	return agents.ReadCanonicalDir(dir)
}

func generateClaude(dir string, plugin *PluginSpec, cmds []*commands.Command, skls []*skills.Skill, agts []*agents.Agent) error {
	// Get adapters
	pluginAdapter, ok := plugins.GetAdapter("claude")
	if !ok {
		return fmt.Errorf("claude plugin adapter not found")
	}

	cmdAdapter, ok := commands.GetAdapter("claude")
	if !ok {
		return fmt.Errorf("claude command adapter not found")
	}

	skillAdapter, ok := skills.GetAdapter("claude")
	if !ok {
		return fmt.Errorf("claude skill adapter not found")
	}

	agentAdapter, ok := agents.GetAdapter("claude")
	if !ok {
		return fmt.Errorf("claude agent adapter not found")
	}

	// Write plugin structure
	if err := pluginAdapter.WritePlugin(&plugin.Plugin, dir); err != nil {
		return fmt.Errorf("write plugin: %w", err)
	}

	// Write commands
	if len(cmds) > 0 {
		commandsDir := filepath.Join(dir, "commands")
		if err := os.MkdirAll(commandsDir, 0755); err != nil {
			return err
		}
		for _, cmd := range cmds {
			path := filepath.Join(commandsDir, cmd.Name+".md")
			if err := cmdAdapter.WriteFile(cmd, path); err != nil {
				return fmt.Errorf("write command %s: %w", cmd.Name, err)
			}
		}
	}

	// Write skills
	if len(skls) > 0 {
		skillsDir := filepath.Join(dir, "skills")
		for _, skl := range skls {
			if err := skillAdapter.WriteSkillDir(skl, skillsDir); err != nil {
				return fmt.Errorf("write skill %s: %w", skl.Name, err)
			}
		}
	}

	// Write agents
	if len(agts) > 0 {
		agentsDir := filepath.Join(dir, "agents")
		if err := os.MkdirAll(agentsDir, 0755); err != nil {
			return err
		}
		for _, agt := range agts {
			path := filepath.Join(agentsDir, agt.Name+".md")
			if err := agentAdapter.WriteFile(agt, path); err != nil {
				return fmt.Errorf("write agent %s: %w", agt.Name, err)
			}
		}
	}

	return nil
}

func generateKiro(dir string, plugin *PluginSpec, skls []*skills.Skill, agts []*agents.Agent) error {
	return generateKiroWithConfig(dir, plugin, skls, agts, nil)
}

func generateKiroWithConfig(dir string, plugin *PluginSpec, skls []*skills.Skill, agts []*agents.Agent, cfg *KiroTargetConfig) error {
	// Determine Kiro format based on plugin spec:
	// - If keywords or MCP servers are present, generate a Kiro Power
	// - Otherwise, generate Kiro Agents format
	isPower := len(plugin.Keywords) > 0 || len(plugin.MCPServers) > 0

	if isPower {
		return generateKiroPower(dir, plugin, skls)
	}
	return generateKiroAgentsWithPrefix(dir, plugin, skls, agts, cfg.getPrefix())
}

func (c *KiroTargetConfig) getPrefix() string {
	if c == nil {
		return ""
	}
	return c.Prefix
}

func generateKiroPower(dir string, plugin *PluginSpec, skls []*skills.Skill) error {
	// Create Power from plugin spec
	power := &powercore.Power{
		Name:        plugin.Name,
		DisplayName: plugin.DisplayName,
		Description: plugin.Description,
		Version:     plugin.Version,
		Keywords:    plugin.Keywords,
		Repository:  plugin.Repository,
		Author:      plugin.Author,
		License:     plugin.License,
	}

	// Add MCP servers
	power.MCPServers = make(map[string]powercore.MCPServer)
	for name, srv := range plugin.MCPServers {
		power.MCPServers[name] = powercore.MCPServer{
			Command:     srv.Command,
			Args:        srv.Args,
			Description: srv.Description,
		}
	}

	// Convert skills to steering files
	power.SteeringFiles = make(map[string]powercore.SteeringFile)
	for _, skl := range skls {
		power.SteeringFiles[skl.Name] = powercore.SteeringFile{
			Path:        filepath.Join("steering", skl.Name+".md"),
			Keywords:    skl.Triggers,
			Description: skl.Description,
			Content:     skl.Instructions,
		}
	}

	// Build instructions from plugin context
	power.Instructions = buildPowerInstructions(plugin, skls)

	// Build onboarding instructions
	power.Onboarding = buildOnboarding(plugin)

	// Use Kiro adapter to write the power
	adapter := &kiro.Adapter{}
	if _, err := adapter.GeneratePowerDir(power, dir); err != nil {
		return fmt.Errorf("write power: %w", err)
	}

	return nil
}

func generateKiroAgentsWithPrefix(dir string, plugin *PluginSpec, skls []*skills.Skill, agts []*agents.Agent, prefix string) error {
	// Create output directory
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Write agents as JSON files with prefix applied to names
	if len(agts) > 0 {
		agentsDir := filepath.Join(dir, "agents")
		if err := os.MkdirAll(agentsDir, 0755); err != nil {
			return err
		}
		for _, agt := range agts {
			// Apply prefix to agent name for Kiro (no namespacing)
			prefixedName := prefix + agt.Name
			path := filepath.Join(agentsDir, prefixedName+".json")
			data, err := json.MarshalIndent(convertToKiroAgentWithName(agt, prefixedName), "", "  ")
			if err != nil {
				return fmt.Errorf("marshal agent %s: %w", agt.Name, err)
			}
			if err := os.WriteFile(path, data, 0600); err != nil {
				return fmt.Errorf("write agent %s: %w", agt.Name, err)
			}
		}
	}

	// Write skills as steering files with prefix applied
	if len(skls) > 0 {
		steeringDir := filepath.Join(dir, "steering")
		if err := os.MkdirAll(steeringDir, 0755); err != nil {
			return err
		}
		for _, skl := range skls {
			// Apply prefix to steering filename to match agent naming convention
			prefixedName := prefix + skl.Name
			path := filepath.Join(steeringDir, prefixedName+".md")
			content := buildSteeringContent(skl)
			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				return fmt.Errorf("write steering %s: %w", skl.Name, err)
			}
		}
	}

	// Write README with prefixed names
	readme := buildKiroAgentsReadmeWithPrefix(plugin, agts, skls, prefix)
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte(readme), 0644); err != nil {
		return fmt.Errorf("write README: %w", err)
	}

	return nil
}

// KiroAgent represents a Kiro CLI agent definition.
type KiroAgent struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Prompt      string   `json:"prompt"`
	Model       string   `json:"model,omitempty"`
	Tools       []string `json:"tools,omitempty"`
}

func convertToKiroAgentWithName(agt *agents.Agent, name string) *KiroAgent {
	return &KiroAgent{
		Name:        name,
		Description: agt.Description,
		Prompt:      agt.Instructions,
		Model:       string(agt.Model),
		Tools:       mapKiroTools(agt.Tools),
	}
}

var kiroToolMap = map[string]string{
	"Read":     "fs_read",
	"Write":    "fs_write",
	"Bash":     "execute_bash",
	"Grep":     "grep",
	"Glob":     "glob",
	"WebFetch": "web_fetch",
	"Code":     "code",
}

func mapKiroTools(tools []string) []string {
	if len(tools) == 0 {
		return nil
	}
	var mapped []string
	for _, t := range tools {
		if m, ok := kiroToolMap[t]; ok {
			mapped = append(mapped, m)
		} else {
			mapped = append(mapped, t)
		}
	}
	return mapped
}

func buildSteeringContent(skl *skills.Skill) string {
	var sb stringBuilder
	sb.WriteString("# " + toTitleCase(skl.Name) + "\n\n")
	sb.WriteString(skl.Description + "\n\n")
	if skl.Instructions != "" {
		sb.WriteString(skl.Instructions + "\n")
	}
	return sb.String()
}

func buildKiroAgentsReadmeWithPrefix(plugin *PluginSpec, agts []*agents.Agent, skls []*skills.Skill, prefix string) string {
	var sb stringBuilder

	title := plugin.DisplayName
	if title == "" {
		title = plugin.Name
	}
	sb.WriteString("# " + title + " - Kiro CLI Plugin\n\n")
	sb.WriteString(plugin.Description + "\n\n")

	if len(agts) > 0 {
		sb.WriteString("## Agents\n\n")
		sb.WriteString("| Agent | Description |\n")
		sb.WriteString("|-------|-------------|\n")
		for _, agt := range agts {
			prefixedName := prefix + agt.Name
			sb.WriteString(fmt.Sprintf("| `%s` | %s |\n", prefixedName, agt.Description))
		}
		sb.WriteString("\n")

		// Add CLI usage examples
		sb.WriteString("## Usage\n\n")
		sb.WriteString("Run an agent with the Kiro CLI:\n\n")
		sb.WriteString("```bash\n")

		// Show example with first agent
		firstAgent := prefix + agts[0].Name
		sb.WriteString(fmt.Sprintf("kiro-cli chat --agent %s \"<your prompt>\"\n", firstAgent))
		sb.WriteString("```\n\n")

		// If there's a coordinator agent, show team usage
		var coordinatorName string
		for _, agt := range agts {
			if hasSubstring(agt.Name, "coordinator") || hasSubstring(agt.Name, "orchestrator") {
				coordinatorName = prefix + agt.Name
				break
			}
		}
		if coordinatorName != "" {
			sb.WriteString("Run the full team (coordinator-driven):\n\n")
			sb.WriteString("```bash\n")
			sb.WriteString(fmt.Sprintf("kiro-cli chat --agent %s \"<your prompt>\"\n", coordinatorName))
			sb.WriteString("```\n\n")
		}
	}

	if len(skls) > 0 {
		sb.WriteString("## Steering Files\n\n")
		sb.WriteString("Copy steering files to `.kiro/steering/` for automatic context loading:\n\n")
		sb.WriteString("```bash\n")
		sb.WriteString("mkdir -p .kiro/steering\n")
		sb.WriteString("cp steering/*.md .kiro/steering/\n")
		sb.WriteString("```\n\n")
	}

	// Installation section
	sb.WriteString("## Installation\n\n")
	sb.WriteString("Copy agents to your Kiro agents directory:\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("mkdir -p ~/.kiro/agents\n")
	sb.WriteString("cp agents/*.json ~/.kiro/agents/\n")
	sb.WriteString("```\n")

	return sb.String()
}

// hasSubstring checks if s contains substr (case-insensitive).
func hasSubstring(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func generateGemini(dir string, plugin *PluginSpec, cmds []*commands.Command) error {
	// Get adapters
	pluginAdapter, ok := plugins.GetAdapter("gemini")
	if !ok {
		return fmt.Errorf("gemini plugin adapter not found")
	}

	cmdAdapter, ok := commands.GetAdapter("gemini")
	if !ok {
		return fmt.Errorf("gemini command adapter not found")
	}

	// Write plugin structure
	if err := pluginAdapter.WritePlugin(&plugin.Plugin, dir); err != nil {
		return fmt.Errorf("write plugin: %w", err)
	}

	// Write commands (Gemini uses TOML)
	if len(cmds) > 0 {
		commandsDir := filepath.Join(dir, "commands")
		if err := os.MkdirAll(commandsDir, 0755); err != nil {
			return err
		}
		for _, cmd := range cmds {
			path := filepath.Join(commandsDir, cmd.Name+".toml")
			if err := cmdAdapter.WriteFile(cmd, path); err != nil {
				return fmt.Errorf("write command %s: %w", cmd.Name, err)
			}
		}
	}

	return nil
}

func buildPowerInstructions(plugin *PluginSpec, skls []*skills.Skill) string {
	var sb stringBuilder

	title := plugin.DisplayName
	if title == "" {
		title = plugin.Name
	}
	sb.WriteString("# " + title + " Power\n\n")

	if plugin.Context != "" {
		sb.WriteString(plugin.Context + "\n\n")
	}

	if len(skls) > 0 {
		sb.WriteString("## Workflows\n\n")
		for _, skl := range skls {
			displayName := toTitleCase(skl.Name)
			sb.WriteString("### " + displayName + " Workflow\n")
			sb.WriteString(skl.Description + "\n\n")
		}
	}

	return sb.String()
}

func buildOnboarding(plugin *PluginSpec) string {
	if len(plugin.MCPServers) == 0 {
		return ""
	}

	var sb stringBuilder
	sb.WriteString("## Prerequisites\n\n")

	for name, srv := range plugin.MCPServers {
		sb.WriteString(fmt.Sprintf("### %s\n\n", name))
		if srv.Description != "" {
			sb.WriteString(srv.Description + "\n\n")
		}
		sb.WriteString("Verify the server is available:\n\n")
		sb.WriteString("```bash\n")
		sb.WriteString(fmt.Sprintf("which %s || echo \"%s not found in PATH\"\n", srv.Command, srv.Command))
		sb.WriteString("```\n\n")
	}

	return sb.String()
}

// stringBuilder is a simple string builder helper.
type stringBuilder struct {
	buf []byte
}

func (sb *stringBuilder) WriteString(s string) {
	sb.buf = append(sb.buf, s...)
}

func (sb *stringBuilder) String() string {
	return string(sb.buf)
}

// toTitleCase converts a kebab-case string to Title Case.
func toTitleCase(s string) string {
	words := splitWords(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = string(upper(word[0])) + lower(word[1:])
		}
	}
	return joinWords(words, " ")
}

func splitWords(s string) []string {
	var words []string
	var word []byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '-' || c == '_' || c == ' ' {
			if len(word) > 0 {
				words = append(words, string(word))
				word = nil
			}
		} else {
			word = append(word, c)
		}
	}
	if len(word) > 0 {
		words = append(words, string(word))
	}
	return words
}

func joinWords(words []string, sep string) string {
	if len(words) == 0 {
		return ""
	}
	result := words[0]
	for i := 1; i < len(words); i++ {
		result += sep + words[i]
	}
	return result
}

func upper(c byte) byte {
	if c >= 'a' && c <= 'z' {
		return c - 32
	}
	return c
}

func lower(s string) string {
	b := []byte(s)
	for i := 0; i < len(b); i++ {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] += 32
		}
	}
	return string(b)
}

// DeploymentResult contains the results of deployment generation.
type DeploymentResult struct {
	// AgentCount is the number of agents loaded.
	AgentCount int

	// TeamName is the name of the team being deployed.
	TeamName string

	// TargetsGenerated lists the names of generated targets.
	TargetsGenerated []string

	// GeneratedDirs maps target names to their output directories.
	GeneratedDirs map[string]string
}

// Deployment generates platform-specific output from multi-agent-spec definitions.
//
// The specsDir should contain:
//   - agents/: Agent definitions (*.md with YAML frontmatter)
//   - teams/: Team definitions (*.json)
//   - deployments/: Deployment definitions (*.json)
//
// Each deployment target specifies a platform and output directory.
func Deployment(specsDir string, deploymentFile string) (*DeploymentResult, error) {
	result := &DeploymentResult{
		GeneratedDirs: make(map[string]string),
	}

	// Load agents from multi-agent-spec format
	agentsDir := filepath.Join(specsDir, "agents")
	agts, err := loadMultiAgentSpecAgents(agentsDir)
	if err != nil {
		return nil, fmt.Errorf("loading agents: %w", err)
	}
	result.AgentCount = len(agts)

	// Build agent map by name
	agentMap := make(map[string]*agents.Agent)
	for _, agt := range agts {
		agentMap[agt.Name] = agt
	}

	// Load deployment
	deployment, err := loadDeployment(deploymentFile)
	if err != nil {
		return nil, fmt.Errorf("loading deployment: %w", err)
	}
	result.TeamName = deployment.Team

	// Generate each target
	for _, target := range deployment.Targets {
		outputDir := target.Output
		if !filepath.IsAbs(outputDir) {
			outputDir = filepath.Join(specsDir, "..", outputDir)
		}

		if err := generateDeploymentTarget(target, agts, outputDir); err != nil {
			return nil, fmt.Errorf("generating target %s: %w", target.Name, err)
		}

		result.TargetsGenerated = append(result.TargetsGenerated, target.Name)
		result.GeneratedDirs[target.Name] = outputDir
	}

	return result, nil
}

// loadMultiAgentSpecAgents loads agents from markdown files with YAML frontmatter.
func loadMultiAgentSpecAgents(dir string) ([]*agents.Agent, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var agts []*agents.Agent
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", entry.Name(), err)
		}

		agt, err := agents.ParseMarkdownAgent(data, path)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", entry.Name(), err)
		}

		// Infer name from filename if not set
		if agt.Name == "" {
			base := filepath.Base(path)
			agt.Name = base[:len(base)-len(filepath.Ext(base))]
		}

		agts = append(agts, agt)
	}

	return agts, nil
}

// DeploymentTarget represents a deployment target configuration.
type DeploymentTarget struct {
	Name     string          `json:"name"`
	Platform string          `json:"platform"`
	Mode     string          `json:"mode,omitempty"`
	Priority string          `json:"priority,omitempty"`
	Output   string          `json:"output"`
	Config   json.RawMessage `json:"config,omitempty"`
	// KiroCli contains Kiro CLI-specific configuration.
	KiroCli *KiroTargetConfig `json:"kiroCli,omitempty"`
}

// KiroTargetConfig contains Kiro-specific deployment configuration.
type KiroTargetConfig struct {
	// Prefix is prepended to agent and steering file names (e.g., "cext_").
	Prefix string `json:"prefix,omitempty"`
	// PluginDir is the output directory for the plugin.
	PluginDir string `json:"pluginDir,omitempty"`
	// Format is the output format (json, yaml).
	Format string `json:"format,omitempty"`
}

// ParseKiroConfig extracts Kiro-specific config from a deployment target.
// Checks both the structured kiroCli field and the generic config field.
func (t *DeploymentTarget) ParseKiroConfig() *KiroTargetConfig {
	// Prefer structured kiroCli field
	if t.KiroCli != nil {
		return t.KiroCli
	}
	// Fall back to generic config field
	if len(t.Config) == 0 {
		return &KiroTargetConfig{}
	}
	var cfg KiroTargetConfig
	if err := json.Unmarshal(t.Config, &cfg); err != nil {
		return &KiroTargetConfig{}
	}
	return &cfg
}

// DeploymentSpec represents a deployment definition.
type DeploymentSpec struct {
	Team    string             `json:"team"`
	Targets []DeploymentTarget `json:"targets"`
}

func loadDeployment(path string) (*DeploymentSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var deployment DeploymentSpec
	if err := json.Unmarshal(data, &deployment); err != nil {
		return nil, err
	}

	return &deployment, nil
}

func generateDeploymentTarget(target DeploymentTarget, agts []*agents.Agent, outputDir string) error {
	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output dir: %w", err)
	}

	switch target.Platform {
	case "claude-code":
		return generateClaudeCodeDeployment(agts, outputDir)
	case "kiro-cli":
		return generateKiroCLIDeployment(agts, outputDir)
	case "gemini-cli":
		return generateGeminiCLIDeployment(agts, outputDir)
	default:
		// For unsupported platforms, log a warning but don't fail
		fmt.Printf("  Warning: platform %s not yet supported, skipping target %s\n", target.Platform, target.Name)
		return nil
	}
}

func generateClaudeCodeDeployment(agts []*agents.Agent, outputDir string) error {
	adapter, ok := agents.GetAdapter("claude")
	if !ok {
		return fmt.Errorf("claude adapter not found")
	}

	for _, agt := range agts {
		path := filepath.Join(outputDir, agt.Name+".md")
		if err := adapter.WriteFile(agt, path); err != nil {
			return fmt.Errorf("writing %s: %w", agt.Name, err)
		}
	}

	return nil
}

func generateKiroCLIDeployment(agts []*agents.Agent, outputDir string) error {
	adapter, ok := agents.GetAdapter("kiro")
	if !ok {
		return fmt.Errorf("kiro adapter not found")
	}

	for _, agt := range agts {
		path := filepath.Join(outputDir, agt.Name+".json")
		if err := adapter.WriteFile(agt, path); err != nil {
			return fmt.Errorf("writing %s: %w", agt.Name, err)
		}
	}

	return nil
}

func generateGeminiCLIDeployment(agts []*agents.Agent, outputDir string) error {
	adapter, ok := agents.GetAdapter("gemini")
	if !ok {
		return fmt.Errorf("gemini adapter not found")
	}

	for _, agt := range agts {
		path := filepath.Join(outputDir, agt.Name+".toml")
		if err := adapter.WriteFile(agt, path); err != nil {
			return fmt.Errorf("writing %s: %w", agt.Name, err)
		}
	}

	return nil
}

// AgentsResult contains the results of simplified agent generation.
type AgentsResult struct {
	// AgentCount is the number of agents loaded.
	AgentCount int

	// TeamName is the name of the team being deployed.
	TeamName string

	// TargetsGenerated lists the names of generated targets.
	TargetsGenerated []string

	// GeneratedDirs maps target names to their output directories.
	GeneratedDirs map[string]string
}

// Agents generates platform-specific agents from a specs directory with simplified options.
//
// The specsDir should contain:
//   - agents/: Agent definitions (*.md with YAML frontmatter)
//   - deployments/: Deployment definitions (*.json)
//
// The target parameter specifies which deployment file to use (looks for {target}.json).
// The outputDir is the base directory for resolving relative output paths in the deployment.
func Agents(specsDir, target, outputDir string) (*AgentsResult, error) {
	result := &AgentsResult{
		GeneratedDirs: make(map[string]string),
	}

	// Load agents from multi-agent-spec format
	agentsDir := filepath.Join(specsDir, "agents")
	agts, err := loadMultiAgentSpecAgents(agentsDir)
	if err != nil {
		return nil, fmt.Errorf("loading agents: %w", err)
	}
	result.AgentCount = len(agts)

	// Construct deployment file path
	deploymentFile := filepath.Join(specsDir, "deployments", target+".json")
	if _, err := os.Stat(deploymentFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("deployment file not found: %s", deploymentFile)
	}

	// Load deployment
	deployment, err := loadDeployment(deploymentFile)
	if err != nil {
		return nil, fmt.Errorf("loading deployment: %w", err)
	}
	result.TeamName = deployment.Team

	// Generate each target
	for _, tgt := range deployment.Targets {
		// Resolve output path relative to outputDir (not specsDir)
		targetOutputDir := tgt.Output
		if !filepath.IsAbs(targetOutputDir) {
			targetOutputDir = filepath.Join(outputDir, targetOutputDir)
		}

		if err := generateDeploymentTarget(tgt, agts, targetOutputDir); err != nil {
			return nil, fmt.Errorf("generating target %s: %w", tgt.Name, err)
		}

		result.TargetsGenerated = append(result.TargetsGenerated, tgt.Name)
		result.GeneratedDirs[tgt.Name] = targetOutputDir
	}

	return result, nil
}

// GenerateResult contains the results of unified plugin generation.
type GenerateResult struct {
	// CommandCount is the number of commands loaded.
	CommandCount int

	// SkillCount is the number of skills loaded.
	SkillCount int

	// AgentCount is the number of agents loaded.
	AgentCount int

	// TeamName is the name of the team being deployed.
	TeamName string

	// TargetsGenerated lists the names of generated targets.
	TargetsGenerated []string

	// GeneratedDirs maps target names to their output directories.
	GeneratedDirs map[string]string
}

// Generate generates platform-specific plugins from a unified specs directory.
// Output is driven by the deployment file at specs/deployments/{target}.json.
//
// Each deployment target receives a complete plugin:
//   - agents (from specs/agents/*.md)
//   - commands (from specs/commands/*.md)
//   - skills (from specs/skills/*.md)
//   - plugin manifest (from specs/plugin.json)
//
// The specsDir should contain:
//   - plugin.json: Plugin metadata
//   - commands/: Command definitions (*.md or *.json)
//   - skills/: Skill definitions (*.md or *.json)
//   - agents/: Agent definitions (*.md with YAML frontmatter)
//   - deployments/: Deployment definitions (*.json)
//
// The target parameter specifies which deployment file to use (looks for {target}.json).
// The outputDir is the base directory for resolving relative output paths in the deployment.
func Generate(specsDir, target, outputDir string) (*GenerateResult, error) {
	result := &GenerateResult{
		GeneratedDirs: make(map[string]string),
	}

	// Load plugin metadata
	pluginPath := filepath.Join(specsDir, "plugin.json")
	var plugin *PluginSpec
	if _, err := os.Stat(pluginPath); err == nil {
		plugin, err = loadPlugin(pluginPath)
		if err != nil {
			return nil, fmt.Errorf("loading plugin spec: %w", err)
		}
	} else {
		// Create minimal plugin spec if not present
		plugin = &PluginSpec{}
	}

	// Load commands
	commandsDir := filepath.Join(specsDir, "commands")
	cmds, err := loadCommands(commandsDir)
	if err != nil {
		return nil, fmt.Errorf("loading commands: %w", err)
	}
	result.CommandCount = len(cmds)

	// Load skills
	skillsDir := filepath.Join(specsDir, "skills")
	skls, err := loadSkills(skillsDir)
	if err != nil {
		return nil, fmt.Errorf("loading skills: %w", err)
	}
	result.SkillCount = len(skls)

	// Load agents from multi-agent-spec format (.md files)
	agentsDir := filepath.Join(specsDir, "agents")
	agts, err := loadMultiAgentSpecAgents(agentsDir)
	if err != nil {
		return nil, fmt.Errorf("loading agents: %w", err)
	}
	result.AgentCount = len(agts)

	// Load deployment
	deploymentFile := filepath.Join(specsDir, "deployments", target+".json")
	if _, err := os.Stat(deploymentFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("deployment file not found: %s", deploymentFile)
	}

	deployment, err := loadDeployment(deploymentFile)
	if err != nil {
		return nil, fmt.Errorf("loading deployment: %w", err)
	}
	result.TeamName = deployment.Team

	// Generate each target
	for _, tgt := range deployment.Targets {
		// Resolve output path relative to outputDir
		targetOutputDir := tgt.Output
		if !filepath.IsAbs(targetOutputDir) {
			targetOutputDir = filepath.Join(outputDir, targetOutputDir)
		}

		if err := generatePlatformPlugin(tgt, targetOutputDir, plugin, cmds, skls, agts); err != nil {
			return nil, fmt.Errorf("generating target %s: %w", tgt.Name, err)
		}

		result.TargetsGenerated = append(result.TargetsGenerated, tgt.Name)
		result.GeneratedDirs[tgt.Name] = targetOutputDir
	}

	return result, nil
}

// generatePlatformPlugin generates a complete plugin for a specific platform.
// It combines agents, commands, skills, and plugin manifest into a platform-specific format.
func generatePlatformPlugin(
	target DeploymentTarget,
	outputDir string,
	plugin *PluginSpec,
	cmds []*commands.Command,
	skls []*skills.Skill,
	agts []*agents.Agent,
) error {
	// Validate output path doesn't end with a generated subdirectory name.
	// The generator creates these subdirectories automatically, so specifying them
	// in the output path would result in duplicate nesting (e.g., agents/agents/).
	base := filepath.Base(outputDir)
	switch base {
	case "agents", "steering", "skills", "commands":
		return fmt.Errorf("output path %q should be the plugin root directory, not a subdirectory (use %q instead)", outputDir, filepath.Dir(outputDir))
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output dir: %w", err)
	}

	switch target.Platform {
	case "claude", "claude-code":
		return generateClaude(outputDir, plugin, cmds, skls, agts)
	case "kiro", "kiro-cli":
		cfg := target.ParseKiroConfig()
		return generateKiroWithConfig(outputDir, plugin, skls, agts, cfg)
	case "gemini", "gemini-cli":
		return generateGemini(outputDir, plugin, cmds)
	default:
		// For unsupported platforms, log a warning but don't fail
		fmt.Printf("  Warning: platform %s not fully supported, generating agents only\n", target.Platform)
		return generateDeploymentTargetAgentsOnly(target.Platform, agts, outputDir)
	}
}

// generateDeploymentTargetAgentsOnly generates only agents for unsupported platforms.
func generateDeploymentTargetAgentsOnly(platform string, agts []*agents.Agent, outputDir string) error {
	if len(agts) == 0 {
		return nil
	}

	// Map platform names to adapter names
	adapterName := platform
	switch platform {
	case "claude-code":
		adapterName = "claude"
	case "kiro-cli":
		adapterName = "kiro"
	case "gemini-cli":
		adapterName = "gemini"
	}

	adapter, ok := agents.GetAdapter(adapterName)
	if !ok {
		return fmt.Errorf("%s adapter not found", adapterName)
	}

	for _, agt := range agts {
		ext := adapter.FileExtension()
		path := filepath.Join(outputDir, agt.Name+ext)
		if err := adapter.WriteFile(agt, path); err != nil {
			return fmt.Errorf("writing %s: %w", agt.Name, err)
		}
	}

	return nil
}
