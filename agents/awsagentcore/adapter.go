// Package awsagentcore provides an adapter for generating AWS Bedrock AgentCore CDK deployments.
// This generates Infrastructure-as-Code for deploying multi-agent systems to AWS.
package awsagentcore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	multiagentspec "github.com/plexusone/multi-agent-spec/sdk/go"

	"github.com/plexusone/assistantkit/agents/core"
)

func init() {
	core.Register(&Adapter{})
}

// Adapter converts canonical Agent definitions to AWS AgentCore CDK format.
type Adapter struct{}

// Name returns the adapter identifier.
func (a *Adapter) Name() string {
	return "aws-agentcore"
}

// FileExtension returns the file extension for CDK files.
func (a *Adapter) FileExtension() string {
	return ".ts"
}

// DefaultDir returns the default directory name for CDK output.
func (a *Adapter) DefaultDir() string {
	return "cdk"
}

// Parse is not typically used for CDK output (it's a generator, not a reader).
func (a *Adapter) Parse(data []byte) (*core.Agent, error) {
	return nil, &core.ParseError{Format: "aws-agentcore", Err: fmt.Errorf("parsing CDK output not supported")}
}

// Marshal converts canonical Agent to CDK construct bytes.
func (a *Adapter) Marshal(agent *core.Agent) ([]byte, error) {
	return generateAgentConstruct(agent)
}

// ReadFile is not typically used for CDK output.
func (a *Adapter) ReadFile(path string) (*core.Agent, error) {
	return nil, &core.ReadError{Path: path, Err: fmt.Errorf("reading CDK files not supported")}
}

// WriteFile writes canonical Agent as CDK construct to path.
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

// AgentCoreConfig holds configuration for AgentCore deployment.
type AgentCoreConfig struct {
	Region          string `json:"region"`
	FoundationModel string `json:"foundation_model"`
	LambdaRuntime   string `json:"lambda_runtime"`
	StackName       string `json:"stack_name"`
}

// DefaultAgentCoreConfig returns default configuration.
func DefaultAgentCoreConfig() *AgentCoreConfig {
	return &AgentCoreConfig{
		Region:          "us-east-1",
		FoundationModel: "anthropic.claude-3-sonnet-20240229-v1:0",
		LambdaRuntime:   "python3.11",
		StackName:       "MultiAgentStack",
	}
}

// Model mapping is delegated to multi-agent-spec BedrockModels.

// Tool to Lambda action mapping.
var toolToAction = map[string]string{
	"WebSearch": "web_search",
	"WebFetch":  "web_fetch",
	"Read":      "read_file",
	"Write":     "write_file",
	"Glob":      "glob_files",
	"Grep":      "grep_content",
	"Bash":      "execute_command",
}

func generateAgentConstruct(agent *core.Agent) ([]byte, error) {
	tmpl, err := template.New("agent").Parse(agentConstructTemplate)
	if err != nil {
		return nil, &core.MarshalError{Format: "aws-agentcore", Err: err}
	}

	// Prepare data for template
	data := map[string]interface{}{
		"Name":            agent.Name,
		"NamePascal":      toPascalCase(agent.Name),
		"Description":     escapeString(agent.Description),
		"Instructions":    escapeString(agent.Instructions),
		"FoundationModel": getFoundationModel(agent.Model),
		"Actions":         getActions(agent.Tools),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, &core.MarshalError{Format: "aws-agentcore", Err: err}
	}

	return buf.Bytes(), nil
}

func toPascalCase(s string) string {
	parts := strings.Split(s, "-")
	var result strings.Builder
	for _, part := range parts {
		if len(part) > 0 {
			result.WriteString(strings.ToUpper(part[:1]))
			result.WriteString(part[1:])
		}
	}
	return result.String()
}

func toCamelCase(s string) string {
	parts := strings.Split(s, "-")
	var result strings.Builder
	for i, part := range parts {
		if len(part) > 0 {
			if i == 0 {
				result.WriteString(strings.ToLower(part[:1]))
			} else {
				result.WriteString(strings.ToUpper(part[:1]))
			}
			result.WriteString(part[1:])
		}
	}
	return result.String()
}

func escapeString(s string) string {
	// Escape backticks and newlines for template literals
	s = strings.ReplaceAll(s, "`", "\\`")
	s = strings.ReplaceAll(s, "${", "\\${")
	return s
}

func getFoundationModel(model core.Model) string {
	// Use multi-agent-spec mapping for Bedrock models
	mapped := multiagentspec.MapModelToBedrock(model)
	if mapped == string(model) {
		// Fallback to sonnet if unknown model
		return multiagentspec.MapModelToBedrock(multiagentspec.ModelSonnet)
	}
	return mapped
}

func getActions(tools []string) []string {
	actions := make([]string, 0, len(tools))
	for _, tool := range tools {
		if action, ok := toolToAction[tool]; ok {
			actions = append(actions, action)
		}
	}
	return actions
}

const agentConstructTemplate = `import * as cdk from 'aws-cdk-lib';
import * as bedrock from 'aws-cdk-lib/aws-bedrock';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as iam from 'aws-cdk-lib/aws-iam';
import { Construct } from 'constructs';

export interface {{.NamePascal}}AgentProps {
  readonly foundationModel?: string;
}

export class {{.NamePascal}}Agent extends Construct {
  public readonly agent: bedrock.CfnAgent;
  public readonly agentAlias: bedrock.CfnAgentAlias;

  constructor(scope: Construct, id: string, props?: {{.NamePascal}}AgentProps) {
    super(scope, id);

    const foundationModel = props?.foundationModel ?? '{{.FoundationModel}}';

    // IAM role for the agent
    const agentRole = new iam.Role(this, 'AgentRole', {
      assumedBy: new iam.ServicePrincipal('bedrock.amazonaws.com'),
      managedPolicies: [
        iam.ManagedPolicy.fromAwsManagedPolicyName('AmazonBedrockFullAccess'),
      ],
    });

    // Agent instruction
    const instruction = ` + "`" + `{{.Instructions}}` + "`" + `;

    // Create the Bedrock Agent
    this.agent = new bedrock.CfnAgent(this, 'Agent', {
      agentName: '{{.Name}}',
      description: '{{.Description}}',
      foundationModel: foundationModel,
      instruction: instruction,
      agentResourceRoleArn: agentRole.roleArn,
      idleSessionTtlInSeconds: 600,
      autoPrepare: true,
    });

    // Create agent alias for invocation
    this.agentAlias = new bedrock.CfnAgentAlias(this, 'AgentAlias', {
      agentId: this.agent.attrAgentId,
      agentAliasName: 'live',
    });

    // Output the agent ID
    new cdk.CfnOutput(this, '{{.NamePascal}}AgentId', {
      value: this.agent.attrAgentId,
      description: 'Agent ID for {{.Name}}',
    });
  }
}
`

// GenerateStack creates a full CDK stack with all agents.
func GenerateStack(teamName string, agents []*core.Agent, config *AgentCoreConfig) ([]byte, error) {
	if config == nil {
		config = DefaultAgentCoreConfig()
	}

	tmpl, err := template.New("stack").Parse(stackTemplate)
	if err != nil {
		return nil, &core.MarshalError{Format: "aws-agentcore", Err: err}
	}

	// Prepare agent data
	type agentData struct {
		Name       string
		NamePascal string
		NameCamel  string
	}
	agentsData := make([]agentData, len(agents))
	for i, agent := range agents {
		agentsData[i] = agentData{
			Name:       agent.Name,
			NamePascal: toPascalCase(agent.Name),
			NameCamel:  toCamelCase(agent.Name),
		}
	}

	data := map[string]interface{}{
		"TeamName":      teamName,
		"TeamPascal":    toPascalCase(teamName),
		"StackName":     config.StackName,
		"Agents":        agentsData,
		"Region":        config.Region,
		"DefaultModel":  config.FoundationModel,
		"LambdaRuntime": config.LambdaRuntime,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, &core.MarshalError{Format: "aws-agentcore", Err: err}
	}

	return buf.Bytes(), nil
}

const stackTemplate = `import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
{{range .Agents}}
import { {{.NamePascal}}Agent } from './agents/{{.Name}}';
{{end}}

export interface {{.TeamPascal}}StackProps extends cdk.StackProps {
  readonly foundationModel?: string;
}

export class {{.TeamPascal}}Stack extends cdk.Stack {
{{range .Agents}}  public readonly {{.NameCamel}}Agent: {{.NamePascal}}Agent;
{{end}}

  constructor(scope: Construct, id: string, props?: {{.TeamPascal}}StackProps) {
    super(scope, id, props);

    const foundationModel = props?.foundationModel ?? '{{.DefaultModel}}';
{{range .Agents}}
    // {{.NamePascal}} Agent
    this.{{.NameCamel}}Agent = new {{.NamePascal}}Agent(this, '{{.NamePascal}}', {
      foundationModel,
    });
{{end}}
  }
}
`

// GenerateCDKApp creates the CDK app entry point.
func GenerateCDKApp(teamName string, config *AgentCoreConfig) ([]byte, error) {
	if config == nil {
		config = DefaultAgentCoreConfig()
	}

	tmpl, err := template.New("app").Parse(appTemplate)
	if err != nil {
		return nil, &core.MarshalError{Format: "aws-agentcore", Err: err}
	}

	data := map[string]interface{}{
		"TeamName":   teamName,
		"TeamPascal": toPascalCase(teamName),
		"StackName":  config.StackName,
		"Region":     config.Region,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, &core.MarshalError{Format: "aws-agentcore", Err: err}
	}

	return buf.Bytes(), nil
}

const appTemplate = `#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from 'aws-cdk-lib';
import { {{.TeamPascal}}Stack } from '../lib/{{.TeamName}}-stack';

const app = new cdk.App();

new {{.TeamPascal}}Stack(app, '{{.StackName}}', {
  env: {
    account: process.env.CDK_DEFAULT_ACCOUNT,
    region: process.env.CDK_DEFAULT_REGION ?? '{{.Region}}',
  },
});
`

// GenerateCDKJSON creates the cdk.json configuration file.
func GenerateCDKJSON(teamName string) ([]byte, error) {
	config := map[string]interface{}{
		"app":     fmt.Sprintf("npx ts-node --prefer-ts-exts bin/%s.ts", teamName),
		"watch":   map[string]interface{}{"include": []string{"**"}},
		"context": map[string]interface{}{},
	}
	return json.MarshalIndent(config, "", "  ")
}

// GeneratePackageJSON creates package.json for the CDK project.
func GeneratePackageJSON(teamName string) ([]byte, error) {
	pkg := map[string]interface{}{
		"name":    teamName + "-cdk",
		"version": "1.0.0",
		"scripts": map[string]string{
			"build":   "tsc",
			"watch":   "tsc -w",
			"cdk":     "cdk",
			"deploy":  "cdk deploy",
			"synth":   "cdk synth",
			"destroy": "cdk destroy",
		},
		"devDependencies": map[string]string{
			"@types/node":        "^20.0.0",
			"aws-cdk":            "^2.170.0",
			"ts-node":            "^10.9.0",
			"typescript":         "^5.0.0",
			"source-map-support": "^0.5.21",
		},
		"dependencies": map[string]string{
			"aws-cdk-lib": "^2.170.0",
			"constructs":  "^10.0.0",
		},
	}
	return json.MarshalIndent(pkg, "", "  ")
}

// WriteCDKProject writes a complete CDK project structure.
func WriteCDKProject(teamName string, agents []*core.Agent, outputDir string, config *AgentCoreConfig) error {
	if config == nil {
		config = DefaultAgentCoreConfig()
	}

	// Create directories
	dirs := []string{
		filepath.Join(outputDir, "bin"),
		filepath.Join(outputDir, "lib"),
		filepath.Join(outputDir, "lib", "agents"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, core.DefaultDirMode); err != nil {
			return &core.WriteError{Path: dir, Err: err}
		}
	}

	// Write cdk.json
	cdkJSON, err := GenerateCDKJSON(teamName)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outputDir, "cdk.json"), cdkJSON, core.DefaultFileMode); err != nil {
		return &core.WriteError{Path: "cdk.json", Err: err}
	}

	// Write package.json
	pkgJSON, err := GeneratePackageJSON(teamName)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outputDir, "package.json"), pkgJSON, core.DefaultFileMode); err != nil {
		return &core.WriteError{Path: "package.json", Err: err}
	}

	// Write app entry point
	appTS, err := GenerateCDKApp(teamName, config)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outputDir, "bin", teamName+".ts"), appTS, core.DefaultFileMode); err != nil {
		return &core.WriteError{Path: "bin/app.ts", Err: err}
	}

	// Write stack
	stackTS, err := GenerateStack(teamName, agents, config)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outputDir, "lib", teamName+"-stack.ts"), stackTS, core.DefaultFileMode); err != nil {
		return &core.WriteError{Path: "lib/stack.ts", Err: err}
	}

	// Write individual agent constructs
	for _, agent := range agents {
		agentTS, err := generateAgentConstruct(agent)
		if err != nil {
			return err
		}
		agentPath := filepath.Join(outputDir, "lib", "agents", agent.Name+".ts")
		if err := os.WriteFile(agentPath, agentTS, core.DefaultFileMode); err != nil {
			return &core.WriteError{Path: agentPath, Err: err}
		}
	}

	// Write tsconfig.json
	tsconfig := `{
  "compilerOptions": {
    "target": "ES2020",
    "module": "commonjs",
    "lib": ["ES2020"],
    "declaration": true,
    "strict": true,
    "noImplicitAny": true,
    "strictNullChecks": true,
    "noImplicitThis": true,
    "alwaysStrict": true,
    "noUnusedLocals": false,
    "noUnusedParameters": false,
    "noImplicitReturns": true,
    "noFallthroughCasesInSwitch": false,
    "inlineSourceMap": true,
    "inlineSources": true,
    "experimentalDecorators": true,
    "strictPropertyInitialization": false,
    "outDir": "./dist",
    "rootDir": "."
  },
  "exclude": ["node_modules", "cdk.out"]
}
`
	if err := os.WriteFile(filepath.Join(outputDir, "tsconfig.json"), []byte(tsconfig), core.DefaultFileMode); err != nil {
		return &core.WriteError{Path: "tsconfig.json", Err: err}
	}

	return nil
}
