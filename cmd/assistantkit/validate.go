package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/plexusone/assistantkit/generate"
	"github.com/spf13/cobra"
)

var validateSpecsDir string

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate specs directory before generation",
	Long: `Validate a specs directory for correctness before running generate.

Checks performed:
  - plugin.json: Valid JSON with required fields (name, version)
  - agents/: Valid markdown with YAML frontmatter, required fields
  - skills/: Valid skill definitions
  - commands/: Valid command definitions
  - skill-refs: Agent skill references resolve to existing skills
  - deployments/: Valid deployment JSON, known platforms, no output conflicts

Example:
  assistantkit validate
  assistantkit validate --specs=specs`,
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().StringVar(&validateSpecsDir, "specs", "specs", "Path to specs directory")
}

func runValidate(cmd *cobra.Command, args []string) error {
	// Resolve path
	absSpecsDir, err := filepath.Abs(validateSpecsDir)
	if err != nil {
		return fmt.Errorf("resolving specs dir: %w", err)
	}

	// Validate specs directory exists
	if _, err := os.Stat(absSpecsDir); os.IsNotExist(err) {
		return fmt.Errorf("specs directory not found: %s", absSpecsDir)
	}

	// Print header
	fmt.Println("=== AssistantKit Validator ===")
	fmt.Printf("Specs directory: %s\n\n", absSpecsDir)

	// Run validation
	result := generate.Validate(absSpecsDir)

	// Print check results
	for _, check := range result.Checks {
		if check.Passed {
			fmt.Printf("✓ %-14s %s\n", check.Name, check.Message)
		} else {
			fmt.Printf("✗ %-14s %s\n", check.Name, check.Message)
		}
	}
	fmt.Println()

	// Print warnings
	if len(result.Warnings) > 0 {
		fmt.Println("Warnings:")
		for _, w := range result.Warnings {
			fmt.Printf("  ⚠ %s: %s\n", w.File, w.Message)
		}
		fmt.Println()
	}

	// Print errors
	if len(result.Errors) > 0 {
		fmt.Println("Errors:")
		for _, e := range result.Errors {
			fmt.Printf("  ✗ %s: %s\n", e.File, e.Message)
		}
		fmt.Println()
	}

	// Print summary
	if result.IsValid() {
		fmt.Printf("✓ Validation passed (%d agents, %d skills, %d commands)\n",
			result.Stats.Agents, result.Stats.Skills, result.Stats.Commands)
		return nil
	}

	return fmt.Errorf("validation failed with %d error(s)", len(result.Errors))
}
