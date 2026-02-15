package generate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/agentplexus/assistantkit/agents"
	"github.com/agentplexus/assistantkit/commands"
	"github.com/agentplexus/assistantkit/skills"
	multiagentspec "github.com/agentplexus/multi-agent-spec/sdk/go"
)

// ValidateResult contains the results of specs validation.
type ValidateResult struct {
	// Checks contains results for each validation check.
	Checks []CheckResult

	// Errors contains all validation errors.
	Errors []ValidateError

	// Warnings contains non-fatal issues.
	Warnings []ValidateError

	// Stats contains counts of loaded items.
	Stats ValidateStats
}

// CheckResult represents the result of a single validation check.
type CheckResult struct {
	Name    string
	Passed  bool
	Message string
}

// ValidateError represents a validation error or warning.
type ValidateError struct {
	Check   string
	File    string
	Message string
}

// ValidateStats contains counts of loaded items.
type ValidateStats struct {
	Agents      int
	Commands    int
	Skills      int
	Teams       int
	Steps       int
	Phases      int
	Deployments int
	Targets     int
}

// IsValid returns true if there are no errors.
func (r *ValidateResult) IsValid() bool {
	return len(r.Errors) == 0
}

// AddCheck adds a check result.
func (r *ValidateResult) addCheck(name string, passed bool, message string) {
	r.Checks = append(r.Checks, CheckResult{
		Name:    name,
		Passed:  passed,
		Message: message,
	})
}

// AddError adds a validation error.
func (r *ValidateResult) addError(check, file, message string) {
	r.Errors = append(r.Errors, ValidateError{
		Check:   check,
		File:    file,
		Message: message,
	})
}

// AddWarning adds a validation warning.
func (r *ValidateResult) addWarning(check, file, message string) {
	r.Warnings = append(r.Warnings, ValidateError{
		Check:   check,
		File:    file,
		Message: message,
	})
}

// validPlatforms lists recognized platform names.
var validPlatforms = map[string]bool{
	"claude":      true,
	"claude-code": true,
	"kiro":        true,
	"kiro-cli":    true,
	"gemini":      true,
	"gemini-cli":  true,
}

// Validate performs static validation on a specs directory.
func Validate(specsDir string) *ValidateResult {
	result := &ValidateResult{}

	// Check directory exists
	if _, err := os.Stat(specsDir); os.IsNotExist(err) {
		result.addError("directory", specsDir, "specs directory not found")
		return result
	}

	// Validate plugin.json
	validatePluginJSON(specsDir, result)

	// Load and validate agents
	agentMap := validateAgentsDir(specsDir, result)

	// Load and validate skills
	skillMap := validateSkillsDir(specsDir, result)

	// Load and validate commands
	validateCommandsDir(specsDir, result)

	// Validate agent skill references
	validateAgentSkillRefs(agentMap, skillMap, result)

	// Validate teams (DAG, agent refs)
	validateTeamsDir(specsDir, agentMap, result)

	// Validate deployments
	validateDeploymentsDir(specsDir, result)

	return result
}

func validatePluginJSON(specsDir string, result *ValidateResult) {
	pluginPath := filepath.Join(specsDir, "plugin.json")

	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		result.addCheck("plugin.json", true, "not found (optional)")
		return
	}

	data, err := os.ReadFile(pluginPath)
	if err != nil {
		result.addCheck("plugin.json", false, fmt.Sprintf("read error: %v", err))
		result.addError("plugin.json", pluginPath, fmt.Sprintf("cannot read: %v", err))
		return
	}

	var plugin PluginSpec
	if err := json.Unmarshal(data, &plugin); err != nil {
		result.addCheck("plugin.json", false, fmt.Sprintf("invalid JSON: %v", err))
		result.addError("plugin.json", pluginPath, fmt.Sprintf("invalid JSON: %v", err))
		return
	}

	// Check required fields
	var missing []string
	if plugin.Name == "" {
		missing = append(missing, "name")
	}
	if plugin.Version == "" {
		missing = append(missing, "version")
	}

	if len(missing) > 0 {
		result.addCheck("plugin.json", false, fmt.Sprintf("missing: %s", strings.Join(missing, ", ")))
		result.addError("plugin.json", pluginPath, fmt.Sprintf("missing required fields: %s", strings.Join(missing, ", ")))
		return
	}

	result.addCheck("plugin.json", true, fmt.Sprintf("%s v%s", plugin.Name, plugin.Version))
}

func validateAgentsDir(specsDir string, result *ValidateResult) map[string]*agents.Agent {
	agentsDir := filepath.Join(specsDir, "agents")
	agentMap := make(map[string]*agents.Agent)

	if _, err := os.Stat(agentsDir); os.IsNotExist(err) {
		result.addCheck("agents", true, "no agents/ directory (optional)")
		return agentMap
	}

	agts, err := agents.ReadCanonicalDir(agentsDir)
	if err != nil {
		result.addCheck("agents", false, fmt.Sprintf("load error: %v", err))
		result.addError("agents", agentsDir, fmt.Sprintf("cannot load: %v", err))
		return agentMap
	}

	result.Stats.Agents = len(agts)

	// Validate each agent
	var errorCount int
	for _, agt := range agts {
		agentMap[agt.Name] = agt

		// Check required fields
		if agt.Name == "" {
			result.addError("agents", agentsDir, "agent with empty name found")
			errorCount++
		}
		if agt.Model == "" {
			result.addWarning("agents", agt.Name, "missing 'model' in frontmatter")
		}
	}

	if errorCount > 0 {
		result.addCheck("agents", false, fmt.Sprintf("%d agents, %d errors", len(agts), errorCount))
	} else {
		result.addCheck("agents", true, fmt.Sprintf("%d agents", len(agts)))
	}

	return agentMap
}

func validateSkillsDir(specsDir string, result *ValidateResult) map[string]*skills.Skill {
	skillsDir := filepath.Join(specsDir, "skills")
	skillMap := make(map[string]*skills.Skill)

	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		result.addCheck("skills", true, "no skills/ directory (optional)")
		return skillMap
	}

	skls, err := skills.ReadCanonicalDir(skillsDir)
	if err != nil {
		result.addCheck("skills", false, fmt.Sprintf("load error: %v", err))
		result.addError("skills", skillsDir, fmt.Sprintf("cannot load: %v", err))
		return skillMap
	}

	result.Stats.Skills = len(skls)

	for _, skl := range skls {
		skillMap[skl.Name] = skl
	}

	result.addCheck("skills", true, fmt.Sprintf("%d skills", len(skls)))
	return skillMap
}

func validateCommandsDir(specsDir string, result *ValidateResult) {
	commandsDir := filepath.Join(specsDir, "commands")

	if _, err := os.Stat(commandsDir); os.IsNotExist(err) {
		result.addCheck("commands", true, "no commands/ directory (optional)")
		return
	}

	cmds, err := commands.ReadCanonicalDir(commandsDir)
	if err != nil {
		result.addCheck("commands", false, fmt.Sprintf("load error: %v", err))
		result.addError("commands", commandsDir, fmt.Sprintf("cannot load: %v", err))
		return
	}

	result.Stats.Commands = len(cmds)
	result.addCheck("commands", true, fmt.Sprintf("%d commands", len(cmds)))
}

func validateAgentSkillRefs(agentMap map[string]*agents.Agent, skillMap map[string]*skills.Skill, result *ValidateResult) {
	if len(agentMap) == 0 {
		return
	}

	var totalRefs int
	var unresolvedCount int

	for _, agt := range agentMap {
		for _, skillName := range agt.Skills {
			totalRefs++
			if _, ok := skillMap[skillName]; !ok {
				result.addError("skill-refs", agt.Name, fmt.Sprintf("references unknown skill: %s", skillName))
				unresolvedCount++
			}
		}
	}

	if unresolvedCount > 0 {
		result.addCheck("skill-refs", false, fmt.Sprintf("%d/%d unresolved", unresolvedCount, totalRefs))
	} else if totalRefs > 0 {
		result.addCheck("skill-refs", true, fmt.Sprintf("%d references resolve", totalRefs))
	}
}

func validateTeamsDir(specsDir string, agentMap map[string]*agents.Agent, result *ValidateResult) {
	teamsDir := filepath.Join(specsDir, "teams")

	if _, err := os.Stat(teamsDir); os.IsNotExist(err) {
		result.addCheck("teams", true, "no teams/ directory (optional)")
		return
	}

	entries, err := os.ReadDir(teamsDir)
	if err != nil {
		result.addCheck("teams", false, fmt.Sprintf("read error: %v", err))
		result.addError("teams", teamsDir, fmt.Sprintf("cannot read: %v", err))
		return
	}

	var teamCount int
	var totalSteps int
	var totalPhases int
	var errorCount int

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		teamCount++
		path := filepath.Join(teamsDir, entry.Name())

		data, err := os.ReadFile(path)
		if err != nil {
			errorCount++
			result.addError("teams", path, fmt.Sprintf("cannot read: %v", err))
			continue
		}

		var team multiagentspec.Team
		if err := json.Unmarshal(data, &team); err != nil {
			errorCount++
			result.addError("teams", path, fmt.Sprintf("invalid JSON: %v", err))
			continue
		}

		// Validate workflow exists
		if team.Workflow == nil || len(team.Workflow.Steps) == 0 {
			result.addWarning("teams", path, "no workflow steps defined")
			continue
		}

		totalSteps += len(team.Workflow.Steps)

		// Build step map for reference checking
		stepMap := make(map[string]*multiagentspec.Step)
		stepOutputs := make(map[string]map[string]bool) // step -> output names
		for i := range team.Workflow.Steps {
			step := &team.Workflow.Steps[i]
			stepMap[step.Name] = step
			stepOutputs[step.Name] = make(map[string]bool)
			for _, out := range step.Outputs {
				stepOutputs[step.Name][out.Name] = true
			}
		}

		// Validate agent references in steps
		for _, step := range team.Workflow.Steps {
			if step.Agent == "" {
				errorCount++
				result.addError("teams", path, fmt.Sprintf("step '%s' has no agent", step.Name))
				continue
			}
			if _, ok := agentMap[step.Agent]; !ok {
				errorCount++
				result.addError("teams", path, fmt.Sprintf("step '%s' references unknown agent: %s", step.Name, step.Agent))
			}
		}

		// Validate depends_on references
		for _, step := range team.Workflow.Steps {
			for _, dep := range step.DependsOn {
				if _, ok := stepMap[dep]; !ok {
					errorCount++
					result.addError("teams", path, fmt.Sprintf("step '%s' depends on unknown step: %s", step.Name, dep))
				}
			}
		}

		// Validate input 'from' references (format: step_name.output_name)
		for _, step := range team.Workflow.Steps {
			for _, input := range step.Inputs {
				if input.From == "" {
					continue
				}
				parts := strings.SplitN(input.From, ".", 2)
				if len(parts) != 2 {
					errorCount++
					result.addError("teams", path, fmt.Sprintf("step '%s' input '%s' has invalid from reference: %s (expected step.output)", step.Name, input.Name, input.From))
					continue
				}
				srcStep, srcOutput := parts[0], parts[1]
				if _, ok := stepMap[srcStep]; !ok {
					errorCount++
					result.addError("teams", path, fmt.Sprintf("step '%s' input '%s' references unknown step: %s", step.Name, input.Name, srcStep))
				} else if !stepOutputs[srcStep][srcOutput] {
					result.addWarning("teams", path, fmt.Sprintf("step '%s' input '%s' references undeclared output: %s.%s", step.Name, input.Name, srcStep, srcOutput))
				}
			}
		}

		// Check for cycles using topological sort
		phases, err := topologicalSort(team.Workflow.Steps)
		if err != nil {
			errorCount++
			result.addError("teams", path, fmt.Sprintf("DAG error: %v", err))
		} else {
			totalPhases += len(phases)
		}
	}

	result.Stats.Teams = teamCount
	result.Stats.Steps = totalSteps
	result.Stats.Phases = totalPhases

	if errorCount > 0 {
		result.addCheck("teams", false, fmt.Sprintf("%d teams, %d steps, %d errors", teamCount, totalSteps, errorCount))
	} else if teamCount > 0 {
		result.addCheck("teams", true, fmt.Sprintf("%d teams, %d steps, %d phases", teamCount, totalSteps, totalPhases))
	} else {
		result.addCheck("teams", true, "no team files found")
	}
}

// topologicalSort performs a topological sort on workflow steps.
// Returns phases (groups of steps that can run in parallel) or error if cycle detected.
func topologicalSort(steps []multiagentspec.Step) ([][]string, error) {
	// Build dependency graph
	inDegree := make(map[string]int)
	dependents := make(map[string][]string)

	for _, step := range steps {
		if _, exists := inDegree[step.Name]; !exists {
			inDegree[step.Name] = 0
		}
		for _, dep := range step.DependsOn {
			inDegree[step.Name]++
			dependents[dep] = append(dependents[dep], step.Name)
		}
	}

	// Process in phases (BFS by level)
	var phases [][]string
	remaining := len(steps)

	for remaining > 0 {
		// Find all steps with no remaining dependencies
		var phase []string
		for _, step := range steps {
			if inDegree[step.Name] == 0 {
				phase = append(phase, step.Name)
			}
		}

		if len(phase) == 0 {
			return nil, fmt.Errorf("circular dependency detected")
		}

		phases = append(phases, phase)

		// Remove processed steps and update dependents
		for _, name := range phase {
			inDegree[name] = -1 // Mark as processed
			remaining--
			for _, dependent := range dependents[name] {
				inDegree[dependent]--
			}
		}
	}

	return phases, nil
}

func validateDeploymentsDir(specsDir string, result *ValidateResult) {
	deploymentsDir := filepath.Join(specsDir, "deployments")

	if _, err := os.Stat(deploymentsDir); os.IsNotExist(err) {
		result.addCheck("deployments", true, "no deployments/ directory (optional)")
		return
	}

	entries, err := os.ReadDir(deploymentsDir)
	if err != nil {
		result.addCheck("deployments", false, fmt.Sprintf("read error: %v", err))
		result.addError("deployments", deploymentsDir, fmt.Sprintf("cannot read: %v", err))
		return
	}

	var deploymentCount int
	var targetCount int
	var errorCount int
	outputPaths := make(map[string]string) // output -> target name (for conflict detection)

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		deploymentCount++
		path := filepath.Join(deploymentsDir, entry.Name())

		data, err := os.ReadFile(path)
		if err != nil {
			errorCount++
			result.addError("deployments", path, fmt.Sprintf("cannot read: %v", err))
			continue
		}

		var deployment DeploymentSpec
		if err := json.Unmarshal(data, &deployment); err != nil {
			errorCount++
			result.addError("deployments", path, fmt.Sprintf("invalid JSON: %v", err))
			continue
		}

		// Validate targets
		for _, target := range deployment.Targets {
			targetCount++

			// Check platform is valid
			if !validPlatforms[target.Platform] {
				errorCount++
				result.addError("deployments", path, fmt.Sprintf("target '%s' has unknown platform: %s", target.Name, target.Platform))
			}

			// Check for output path conflicts within this deployment
			if existing, ok := outputPaths[target.Output]; ok {
				errorCount++
				result.addError("deployments", path, fmt.Sprintf("output '%s' conflicts with target '%s'", target.Output, existing))
			} else {
				outputPaths[target.Output] = target.Name
			}
		}
	}

	result.Stats.Deployments = deploymentCount
	result.Stats.Targets = targetCount

	if errorCount > 0 {
		result.addCheck("deployments", false, fmt.Sprintf("%d deployments, %d targets, %d errors", deploymentCount, targetCount, errorCount))
	} else {
		result.addCheck("deployments", true, fmt.Sprintf("%d deployments, %d targets", deploymentCount, targetCount))
	}
}
