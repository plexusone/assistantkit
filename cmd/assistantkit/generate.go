package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/plexusone/assistantkit/generate"
	"github.com/spf13/cobra"
)

var (
	genSpecsDir     string
	genTarget       string
	genOutputDir    string
	genSkipValidate bool
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate platform-specific plugins from canonical specs",
	Long: `Generate platform-specific plugins from a unified specs directory.

This command reads from a specs directory and generates complete plugins for
each platform defined in the deployment file.

The specs directory should contain:
  - plugin.json: Plugin metadata (name, version, keywords, mcpServers)
  - commands/: Command definitions (*.md or *.json)
  - skills/: Skill definitions (*.md or *.json)
  - agents/: Agent definitions (*.md with YAML frontmatter)
  - deployments/: Deployment definitions (*.json)

Each deployment target receives a complete plugin:
  - claude/claude-code: .claude-plugin/, commands/, skills/, agents/
  - kiro/kiro-cli: POWER.md + mcp.json or agents/*.json
  - gemini/gemini-cli: gemini-extension.json, commands/, agents/

Deployment output paths:
  The output path in your deployment target should be the plugin ROOT directory.
  The generator creates subdirectories (agents/, steering/, etc.) automatically.

  For distribution:  "output": "plugins/kiro"    -> plugins/kiro/agents/*.json
  For local install: "output": "~/.kiro"         -> ~/.kiro/agents/*.json

  Do NOT include subdirectories in output paths (e.g., "plugins/kiro/agents").

Example:
  assistantkit generate
  assistantkit generate --specs=specs --target=local --output=.`,
	RunE: runGenerate,
}

var (
	specDir    string
	outputDir  string
	platforms  []string
	configFile string
)

var generatePluginsCmd = &cobra.Command{
	Use:   "plugins",
	Short: "Generate plugins for all configured platforms",
	Long: `Generate platform-specific plugins from a canonical spec directory.

The spec directory should contain:
  - plugin.json: Plugin metadata
  - commands/: Command definitions (*.json)
  - skills/: Skill definitions (*.json)
  - agents/: Agent definitions (*.json)

Example:
  assistantkit generate plugins --spec=plugins/spec --output=plugins --platforms=claude,kiro`,
	RunE: runGeneratePlugins,
}

var (
	deploymentSpecDir string
	deploymentFile    string
)

var generateDeploymentCmd = &cobra.Command{
	Use:   "deployment",
	Short: "Generate deployment artifacts from multi-agent-spec definitions",
	Long: `Generate platform-specific deployment artifacts from multi-agent-spec format.

The specs directory should contain:
  - agents/: Agent definitions (*.md with YAML frontmatter)
  - teams/: Team definitions (*.json)
  - deployments/: Deployment definitions (*.json)

Each target in the deployment file specifies a platform and output directory.

Supported platforms:
  - claude-code: Claude Code agent markdown files
  - kiro-cli: Kiro CLI agent JSON files
  - gemini-cli: Gemini CLI agent TOML files

Example:
  assistantkit generate deployment --specs=specs --deployment=specs/deployments/my-team.json`,
	RunE: runGenerateDeployment,
}

var (
	agentsSpecDir   string
	agentsTarget    string
	agentsOutputDir string
)

var generateAgentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "Generate agents from specs directory (simplified)",
	Long: `Generate platform-specific agents from a specs directory.

This is a simplified command that reads from specs/agents/*.md and uses
specs/deployments/<target>.json to determine output locations.

The specs directory should contain:
  - agents/: Agent definitions (*.md with YAML frontmatter)
  - deployments/: Deployment definitions (*.json, defaults to local.json)

Example:
  assistantkit generate agents
  assistantkit generate agents --specs=specs --target=local --output=.`,
	RunE: runGenerateAgents,
}

var (
	allSpecsDir  string
	allTarget    string
	allOutputDir string
	allPlatforms []string
)

var generateAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Generate all plugin artifacts from a unified specs directory",
	Long: `Generate all platform-specific artifacts from a unified specs directory.

This command combines 'generate plugins' and 'generate agents' into a single
operation. It reads all specs from a single directory and generates complete
plugin packages for each platform.

The specs directory should contain:
  - plugin.json: Plugin metadata
  - commands/: Command definitions (*.md or *.json)
  - skills/: Skill definitions (*.md or *.json)
  - agents/: Agent definitions (*.md with YAML frontmatter)
  - deployments/: Deployment definitions (*.json)

Example:
  assistantkit generate all --specs=specs --target=local
  assistantkit generate all --specs=specs --target=local --output=. --platforms=claude,kiro,gemini`,
	RunE: runGenerateAll,
}

func init() {
	generateCmd.AddCommand(generatePluginsCmd)
	generateCmd.AddCommand(generateDeploymentCmd)
	generateCmd.AddCommand(generateAgentsCmd)
	generateCmd.AddCommand(generateAllCmd)

	// Main generate command flags
	generateCmd.Flags().StringVar(&genSpecsDir, "specs", "specs", "Path to unified specs directory")
	generateCmd.Flags().StringVar(&genTarget, "target", "local", "Deployment target (looks for specs/deployments/<target>.json)")
	generateCmd.Flags().StringVar(&genOutputDir, "output", ".", "Output base directory for relative paths")
	generateCmd.Flags().BoolVar(&genSkipValidate, "skip-validate", false, "Skip validation before generation")

	generatePluginsCmd.Flags().StringVar(&specDir, "spec", "plugins/spec", "Path to canonical spec directory")
	generatePluginsCmd.Flags().StringVar(&outputDir, "output", "plugins", "Output directory for generated plugins")
	generatePluginsCmd.Flags().StringSliceVar(&platforms, "platforms", []string{"claude", "kiro"}, "Platforms to generate (claude,kiro,gemini)")
	generatePluginsCmd.Flags().StringVar(&configFile, "config", "", "Config file (default: assistantkit.yaml if exists)")

	generateDeploymentCmd.Flags().StringVar(&deploymentSpecDir, "specs", "specs", "Path to multi-agent-spec directory")
	generateDeploymentCmd.Flags().StringVar(&deploymentFile, "deployment", "", "Path to deployment definition file (required)")
	_ = generateDeploymentCmd.MarkFlagRequired("deployment")

	generateAgentsCmd.Flags().StringVar(&agentsSpecDir, "specs", "specs", "Path to specs directory")
	generateAgentsCmd.Flags().StringVar(&agentsTarget, "target", "local", "Deployment target (looks for specs/deployments/<target>.json)")
	generateAgentsCmd.Flags().StringVar(&agentsOutputDir, "output", ".", "Output base directory (repo root)")

	generateAllCmd.Flags().StringVar(&allSpecsDir, "specs", "specs", "Path to unified specs directory")
	generateAllCmd.Flags().StringVar(&allTarget, "target", "local", "Deployment target (looks for specs/deployments/<target>.json)")
	generateAllCmd.Flags().StringVar(&allOutputDir, "output", ".", "Output base directory (repo root)")
	generateAllCmd.Flags().StringSliceVar(&allPlatforms, "platforms", []string{"claude", "kiro", "gemini"}, "Platforms to generate")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// Resolve paths
	absSpecsDir, err := filepath.Abs(genSpecsDir)
	if err != nil {
		return fmt.Errorf("resolving specs dir: %w", err)
	}

	absOutputDir, err := filepath.Abs(genOutputDir)
	if err != nil {
		return fmt.Errorf("resolving output dir: %w", err)
	}

	// Validate specs directory exists
	if _, err := os.Stat(absSpecsDir); os.IsNotExist(err) {
		return fmt.Errorf("specs directory not found: %s", absSpecsDir)
	}

	// Print header
	fmt.Println("=== AssistantKit Generator ===")
	fmt.Printf("Specs directory: %s\n", absSpecsDir)
	fmt.Printf("Target: %s\n", genTarget)
	fmt.Printf("Output directory: %s\n", absOutputDir)
	fmt.Println()

	// Run validation unless skipped
	if !genSkipValidate {
		fmt.Println("Validating specs...")
		validateResult := generate.Validate(absSpecsDir)

		if !validateResult.IsValid() {
			fmt.Println()
			fmt.Println("Validation failed:")
			for _, e := range validateResult.Errors {
				fmt.Printf("  ✗ %s: %s\n", e.File, e.Message)
			}
			fmt.Println()
			fmt.Println("Fix the above errors and re-run, or use --skip-validate to bypass.")
			return fmt.Errorf("validation failed with %d error(s)", len(validateResult.Errors))
		}
		fmt.Printf("✓ Validation passed (%d agents, %d skills, %d commands)\n\n",
			validateResult.Stats.Agents, validateResult.Stats.Skills, validateResult.Stats.Commands)
	}

	// Generate using the unified Generate function
	result, err := generate.Generate(absSpecsDir, genTarget, absOutputDir)
	if err != nil {
		return fmt.Errorf("generating: %w", err)
	}

	// Print results
	if result.TeamName != "" {
		fmt.Printf("Team: %s\n", result.TeamName)
	}
	fmt.Printf("Loaded: %d commands, %d skills, %d agents\n\n", result.CommandCount, result.SkillCount, result.AgentCount)

	fmt.Println("Generated targets:")
	for _, target := range result.TargetsGenerated {
		dir := result.GeneratedDirs[target]
		fmt.Printf("  - %s: %s\n", target, dir)
	}

	// Print install instructions for generated platforms
	printInstallInstructions(result.GeneratedDirs)

	fmt.Println("\nDone!")
	return nil
}

func runGenerateDeployment(cmd *cobra.Command, args []string) error {
	fmt.Println("Note: 'generate deployment' is deprecated. Use 'generate --specs=... --target=...' instead.")
	fmt.Println()
	// Resolve paths
	absSpecsDir, err := filepath.Abs(deploymentSpecDir)
	if err != nil {
		return fmt.Errorf("resolving specs dir: %w", err)
	}

	absDeploymentFile, err := filepath.Abs(deploymentFile)
	if err != nil {
		return fmt.Errorf("resolving deployment file: %w", err)
	}

	// Validate paths exist
	if _, err := os.Stat(absSpecsDir); os.IsNotExist(err) {
		return fmt.Errorf("specs directory not found: %s", absSpecsDir)
	}
	if _, err := os.Stat(absDeploymentFile); os.IsNotExist(err) {
		return fmt.Errorf("deployment file not found: %s", absDeploymentFile)
	}

	// Print header
	fmt.Println("=== AssistantKit Deployment Generator ===")
	fmt.Printf("Specs directory: %s\n", absSpecsDir)
	fmt.Printf("Deployment file: %s\n", absDeploymentFile)
	fmt.Println()

	// Generate deployment
	result, err := generate.Deployment(absSpecsDir, absDeploymentFile)
	if err != nil {
		return fmt.Errorf("generating deployment: %w", err)
	}

	// Print results
	fmt.Printf("Team: %s\n", result.TeamName)
	fmt.Printf("Loaded: %d agents\n\n", result.AgentCount)

	fmt.Println("Generated targets:")
	for _, target := range result.TargetsGenerated {
		dir := result.GeneratedDirs[target]
		fmt.Printf("  - %s: %s\n", target, dir)
	}

	fmt.Println("\nDone!")
	return nil
}

func runGeneratePlugins(cmd *cobra.Command, args []string) error {
	fmt.Println("Note: 'generate plugins' is deprecated. Use 'generate --specs=... --target=...' instead.")
	fmt.Println()

	// Resolve paths
	absSpecDir, err := filepath.Abs(specDir)
	if err != nil {
		return fmt.Errorf("resolving spec dir: %w", err)
	}

	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return fmt.Errorf("resolving output dir: %w", err)
	}

	// Validate spec directory exists
	if _, err := os.Stat(absSpecDir); os.IsNotExist(err) {
		return fmt.Errorf("spec directory not found: %s", absSpecDir)
	}

	// Print header
	fmt.Println("=== AssistantKit Plugin Generator ===")
	fmt.Printf("Spec directory: %s\n", absSpecDir)
	fmt.Printf("Output directory: %s\n", absOutputDir)
	fmt.Printf("Platforms: %s\n", strings.Join(platforms, ", "))
	fmt.Println()

	// Generate plugins
	result, err := generate.Plugins(absSpecDir, absOutputDir, platforms)
	if err != nil {
		return fmt.Errorf("generating plugins: %w", err)
	}

	// Print results
	fmt.Printf("Loaded: %d commands, %d skills, %d agents\n\n",
		result.CommandCount, result.SkillCount, result.AgentCount)

	for platform, dir := range result.GeneratedDirs {
		fmt.Printf("Generated %s: %s\n", platform, dir)
	}

	fmt.Println("\nDone!")
	return nil
}

func runGenerateAgents(cmd *cobra.Command, args []string) error {
	fmt.Println("Note: 'generate agents' is deprecated. Use 'generate --specs=... --target=...' instead.")
	fmt.Println()

	// Resolve paths
	absSpecsDir, err := filepath.Abs(agentsSpecDir)
	if err != nil {
		return fmt.Errorf("resolving specs dir: %w", err)
	}

	absOutputDir, err := filepath.Abs(agentsOutputDir)
	if err != nil {
		return fmt.Errorf("resolving output dir: %w", err)
	}

	// Validate specs directory exists
	if _, err := os.Stat(absSpecsDir); os.IsNotExist(err) {
		return fmt.Errorf("specs directory not found: %s", absSpecsDir)
	}

	// Print header
	fmt.Println("=== AssistantKit Agent Generator ===")
	fmt.Printf("Specs directory: %s\n", absSpecsDir)
	fmt.Printf("Target: %s\n", agentsTarget)
	fmt.Printf("Output directory: %s\n", absOutputDir)
	fmt.Println()

	// Generate agents
	result, err := generate.Agents(absSpecsDir, agentsTarget, absOutputDir)
	if err != nil {
		return fmt.Errorf("generating agents: %w", err)
	}

	// Print results
	fmt.Printf("Team: %s\n", result.TeamName)
	fmt.Printf("Loaded: %d agents\n\n", result.AgentCount)

	fmt.Println("Generated targets:")
	for _, target := range result.TargetsGenerated {
		dir := result.GeneratedDirs[target]
		fmt.Printf("  - %s: %s\n", target, dir)
	}

	fmt.Println("\nDone!")
	return nil
}

func runGenerateAll(cmd *cobra.Command, args []string) error {
	fmt.Println("Note: 'generate all' is deprecated. Use 'generate --specs=... --target=...' instead.")
	fmt.Println()

	// Resolve paths
	absSpecsDir, err := filepath.Abs(allSpecsDir)
	if err != nil {
		return fmt.Errorf("resolving specs dir: %w", err)
	}

	absOutputDir, err := filepath.Abs(allOutputDir)
	if err != nil {
		return fmt.Errorf("resolving output dir: %w", err)
	}

	// Validate specs directory exists
	if _, err := os.Stat(absSpecsDir); os.IsNotExist(err) {
		return fmt.Errorf("specs directory not found: %s", absSpecsDir)
	}

	// Print header
	fmt.Println("=== AssistantKit Unified Generator ===")
	fmt.Printf("Specs directory: %s\n", absSpecsDir)
	fmt.Printf("Output directory: %s\n", absOutputDir)
	fmt.Printf("Target: %s\n", allTarget)
	fmt.Printf("Platforms: %s\n", strings.Join(allPlatforms, ", "))
	fmt.Println()

	// Step 1: Generate plugins (commands, skills, plugin manifest)
	pluginsOutputDir := filepath.Join(absOutputDir, "plugins")
	fmt.Println("1. Generating plugins (commands, skills, manifest)...")

	pluginResult, err := generate.Plugins(absSpecsDir, pluginsOutputDir, allPlatforms)
	if err != nil {
		return fmt.Errorf("generating plugins: %w", err)
	}

	fmt.Printf("   Loaded: %d commands, %d skills\n", pluginResult.CommandCount, pluginResult.SkillCount)
	for platform, dir := range pluginResult.GeneratedDirs {
		fmt.Printf("   Generated %s: %s\n", platform, dir)
	}
	fmt.Println()

	// Step 2: Generate agents from deployment target
	fmt.Println("2. Generating agents from deployment target...")

	agentResult, err := generate.Agents(absSpecsDir, allTarget, absOutputDir)
	if err != nil {
		return fmt.Errorf("generating agents: %w", err)
	}

	fmt.Printf("   Team: %s\n", agentResult.TeamName)
	fmt.Printf("   Loaded: %d agents\n", agentResult.AgentCount)
	for _, target := range agentResult.TargetsGenerated {
		dir := agentResult.GeneratedDirs[target]
		fmt.Printf("   Generated %s: %s\n", target, dir)
	}

	fmt.Println("\nDone!")
	return nil
}

// printInstallInstructions prints platform-specific install commands based on generated directories.
func printInstallInstructions(generatedDirs map[string]string) {
	var kiroDir, claudeDir string

	for target, dir := range generatedDirs {
		targetLower := strings.ToLower(target)
		if strings.Contains(targetLower, "kiro") {
			kiroDir = dir
		} else if strings.Contains(targetLower, "claude") {
			claudeDir = dir
		}
	}

	if kiroDir == "" && claudeDir == "" {
		return
	}

	fmt.Println("\nTo install:")

	if kiroDir != "" {
		// Check if already installed to ~/.kiro
		if !strings.Contains(kiroDir, ".kiro") {
			fmt.Println("\n  Kiro CLI:")
			fmt.Printf("    cp %s/agents/*.json ~/.kiro/agents/\n", kiroDir)
			// Check if steering directory exists in the output
			steeringDir := filepath.Join(kiroDir, "steering")
			if _, err := os.Stat(steeringDir); err == nil {
				fmt.Printf("    cp %s/steering/*.md ~/.kiro/steering/\n", kiroDir)
			}
		} else {
			fmt.Println("\n  Kiro CLI: ✓ Already installed to ~/.kiro")
		}
	}

	if claudeDir != "" {
		// Check if already installed to ~/.claude or .claude
		if !strings.Contains(claudeDir, ".claude") {
			fmt.Println("\n  Claude Code:")
			fmt.Printf("    cp -r %s/.claude-plugin ~/.claude-plugin\n", claudeDir)
			fmt.Printf("    cp %s/agents/*.md .claude/agents/\n", claudeDir)
		} else {
			fmt.Println("\n  Claude Code: ✓ Already installed")
		}
	}

	fmt.Println("\n  Tip: Set output to ~/.kiro or .claude in your deployment file to install directly.")
}
