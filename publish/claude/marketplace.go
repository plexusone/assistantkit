// Package claude provides a publisher for the Claude Code official marketplace.
package claude

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/plexusone/assistantkit/publish/core"
	"github.com/plexusone/assistantkit/publish/github"
)

const (
	// MarketplaceOwner is the GitHub org that owns the official marketplace.
	MarketplaceOwner = "anthropics"

	// MarketplaceRepo is the repository name.
	MarketplaceRepo = "claude-plugins-official"

	// ExternalPluginsPath is the directory for third-party plugins.
	ExternalPluginsPath = "external_plugins"

	// DefaultBaseBranch is the default branch to target.
	DefaultBaseBranch = "main"
)

// RequiredFiles lists files that must exist in a plugin.
var RequiredFiles = []string{
	".claude-plugin/plugin.json",
	"README.md",
}

// Publisher submits plugins to the Claude Code official marketplace.
type Publisher struct {
	client *github.Client
	config core.MarketplaceConfig
}

// NewPublisher creates a new Claude marketplace publisher.
func NewPublisher(token string) *Publisher {
	return &Publisher{
		client: github.NewClient(token),
		config: core.MarketplaceConfig{
			Owner:         MarketplaceOwner,
			Repo:          MarketplaceRepo,
			BaseBranch:    DefaultBaseBranch,
			PluginPath:    ExternalPluginsPath,
			RequiredFiles: RequiredFiles,
		},
	}
}

// Name returns the marketplace identifier.
func (p *Publisher) Name() string {
	return "claude"
}

// Validate checks if the plugin directory has all required files.
func (p *Publisher) Validate(pluginDir string) error {
	var missing []string

	for _, file := range p.config.RequiredFiles {
		path := filepath.Join(pluginDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			missing = append(missing, file)
		}
	}

	if len(missing) > 0 {
		return &core.ValidationError{
			PluginDir: pluginDir,
			Missing:   missing,
		}
	}

	return nil
}

// Publish submits the plugin to the Claude Code marketplace.
func (p *Publisher) Publish(ctx context.Context, opts core.PublishOptions) (*core.PublishResult, error) {
	// Validate first
	if err := p.Validate(opts.PluginDir); err != nil {
		return nil, err
	}

	p.client.SetDryRun(opts.DryRun)

	// Get authenticated user if fork owner not specified
	forkOwner := opts.ForkOwner
	if forkOwner == "" {
		user, err := p.client.GetAuthenticatedUser(ctx)
		if err != nil {
			return nil, err
		}
		forkOwner = user
	}

	// Ensure fork exists
	if opts.Verbose {
		fmt.Printf("Ensuring fork of %s/%s exists for %s...\n",
			p.config.Owner, p.config.Repo, forkOwner)
	}
	forkOwner, forkRepo, err := p.client.EnsureFork(ctx, p.config.Owner, p.config.Repo, forkOwner)
	if err != nil {
		return nil, err
	}

	// Get base branch SHA
	baseBranch := p.config.BaseBranch
	baseSHA, err := p.client.GetBranchSHA(ctx, p.config.Owner, p.config.Repo, baseBranch)
	if err != nil {
		return nil, err
	}

	// Create branch name
	branch := opts.Branch
	if branch == "" {
		branch = fmt.Sprintf("add-%s", opts.PluginName)
	}

	// Create branch in fork
	if opts.Verbose {
		fmt.Printf("Creating branch %s...\n", branch)
	}
	if err := p.client.CreateBranch(ctx, forkOwner, forkRepo, branch, baseSHA); err != nil {
		return nil, err
	}

	// Read local plugin files
	destPath := filepath.Join(p.config.PluginPath, opts.PluginName)
	files, err := github.ReadLocalFiles(opts.PluginDir, destPath)
	if err != nil {
		return nil, err
	}

	if opts.Verbose {
		fmt.Printf("Adding %d files to %s...\n", len(files), destPath)
		for _, f := range files {
			fmt.Printf("  %s\n", f.Path)
		}
	}

	// Create commit
	commitMsg := fmt.Sprintf("Add %s plugin", opts.PluginName)
	if opts.Verbose {
		fmt.Printf("Creating commit: %s\n", commitMsg)
	}
	_, err = p.client.CreateCommit(ctx, forkOwner, forkRepo, branch, commitMsg, files)
	if err != nil {
		return nil, err
	}

	// Create PR title and body
	title := opts.Title
	if title == "" {
		title = fmt.Sprintf("Add %s plugin", opts.PluginName)
	}

	body := opts.Body
	if body == "" {
		body = generatePRBody(opts.PluginName, opts.PluginDir)
	}

	// Create PR
	if opts.Verbose {
		fmt.Printf("Creating PR: %s\n", title)
	}
	pr, err := p.client.CreatePR(ctx, p.config.Owner, p.config.Repo, forkOwner, branch, baseBranch, title, body)
	if err != nil {
		return nil, err
	}

	// Build file list
	var fileNames []string
	for _, f := range files {
		fileNames = append(fileNames, f.Path)
	}

	status := "PR created successfully"
	if opts.DryRun {
		status = "Dry run completed - no PR created"
	}

	return &core.PublishResult{
		PRURL:      pr.GetHTMLURL(),
		PRNumber:   pr.GetNumber(),
		Branch:     branch,
		ForkURL:    fmt.Sprintf("https://github.com/%s/%s", forkOwner, forkRepo),
		Status:     status,
		FilesAdded: fileNames,
	}, nil
}

// generatePRBody creates a default PR description.
func generatePRBody(pluginName, pluginDir string) string {
	// Try to read README for description
	readmePath := filepath.Join(pluginDir, "README.md")
	readme, err := os.ReadFile(readmePath)

	var description string
	if err == nil && len(readme) > 0 {
		// Extract first paragraph after title
		description = extractDescription(string(readme))
	}

	body := fmt.Sprintf(`## Summary

Adding the **%s** plugin to the Claude Code marketplace.

`, pluginName)

	if description != "" {
		body += fmt.Sprintf("### Description\n\n%s\n\n", description)
	}

	body += "### Checklist\n\n"
	body += "- [ ] Plugin has `.claude-plugin/plugin.json`\n"
	body += "- [ ] Plugin has `README.md` with documentation\n"
	body += "- [ ] All commands/skills/agents are documented\n"
	body += "- [ ] No security issues or sensitive data\n"
	body += "- [ ] Tested locally with Claude Code\n"
	body += "\n---\n\n"
	body += "*Submitted via [aiassistkit](https://github.com/plexusone/assistantkit) publish tool*\n"

	return body
}

// extractDescription extracts the first paragraph after a markdown title.
func extractDescription(readme string) string {
	lines := splitLines(readme)
	var description []string
	inDescription := false

	for _, line := range lines {
		// Skip title lines
		if len(line) > 0 && line[0] == '#' {
			if inDescription {
				break // Stop at next heading
			}
			inDescription = true
			continue
		}

		// Skip empty lines before description starts
		if !inDescription && line == "" {
			continue
		}

		// Start collecting description
		if inDescription {
			if line == "" && len(description) > 0 {
				break // End of first paragraph
			}
			if line != "" {
				description = append(description, line)
			}
		}
	}

	if len(description) > 3 {
		description = description[:3]
		description = append(description, "...")
	}

	return joinLines(description)
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func joinLines(lines []string) string {
	result := ""
	for i, line := range lines {
		if i > 0 {
			result += " "
		}
		result += line
	}
	return result
}
