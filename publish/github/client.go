// Package github provides GitHub API utilities for marketplace submissions.
// This package wraps github.com/grokify/gogithub for marketplace-specific operations.
package github

import (
	"context"

	"github.com/google/go-github/v84/github"
	"github.com/grokify/gogithub/auth"
	"github.com/grokify/gogithub/pr"
	"github.com/grokify/gogithub/repo"
)

// Client wraps the GitHub API client with marketplace-specific operations.
type Client struct {
	gh     *github.Client
	dryRun bool
}

// NewClient creates a new GitHub client with the given token.
func NewClient(token string) *Client {
	gh := auth.NewGitHubClient(context.Background(), token)
	return &Client{gh: gh}
}

// SetDryRun enables or disables dry run mode.
func (c *Client) SetDryRun(dryRun bool) {
	c.dryRun = dryRun
}

// GetAuthenticatedUser returns the authenticated user's login.
func (c *Client) GetAuthenticatedUser(ctx context.Context) (string, error) {
	return auth.GetAuthenticatedUser(ctx, c.gh)
}

// EnsureFork ensures a fork exists for the given repository.
// Returns the fork owner and repo name.
func (c *Client) EnsureFork(ctx context.Context, upstreamOwner, upstreamRepo, forkOwner string) (string, string, error) {
	if c.dryRun {
		return forkOwner, upstreamRepo, nil
	}
	return repo.EnsureFork(ctx, c.gh, upstreamOwner, upstreamRepo, forkOwner)
}

// GetDefaultBranch returns the default branch of a repository.
func (c *Client) GetDefaultBranch(ctx context.Context, owner, repoName string) (string, error) {
	return repo.GetDefaultBranch(ctx, c.gh, owner, repoName)
}

// GetBranchSHA returns the SHA of the given branch.
func (c *Client) GetBranchSHA(ctx context.Context, owner, repoName, branch string) (string, error) {
	return repo.GetBranchSHA(ctx, c.gh, owner, repoName, branch)
}

// CreateBranch creates a new branch from the given base SHA.
func (c *Client) CreateBranch(ctx context.Context, owner, repoName, branch, baseSHA string) error {
	if c.dryRun {
		return nil
	}
	return repo.CreateBranch(ctx, c.gh, owner, repoName, branch, baseSHA)
}

// FileContent represents a file to be committed.
type FileContent = repo.FileContent

// CreateCommit creates a commit with the given files.
func (c *Client) CreateCommit(ctx context.Context, owner, repoName, branch, message string, files []FileContent) (string, error) {
	if c.dryRun {
		return "dry-run-sha", nil
	}
	return repo.CreateCommit(ctx, c.gh, owner, repoName, branch, message, files)
}

// CreatePR creates a pull request.
func (c *Client) CreatePR(ctx context.Context, upstreamOwner, upstreamRepo, forkOwner, branch, baseBranch, title, body string) (*github.PullRequest, error) {
	if c.dryRun {
		return &github.PullRequest{
			HTMLURL: github.Ptr("https://github.com/" + upstreamOwner + "/" + upstreamRepo + "/pull/0"),
			Number:  github.Ptr(0),
			State:   github.Ptr("dry-run"),
		}, nil
	}
	return pr.CreatePR(ctx, c.gh, upstreamOwner, upstreamRepo, forkOwner, branch, baseBranch, title, body)
}

// ReadLocalFiles reads all files from a local directory recursively.
func ReadLocalFiles(dir, prefix string) ([]FileContent, error) {
	return repo.ReadLocalFiles(dir, prefix)
}
