// Package publish provides publishers for submitting plugins to AI assistant marketplaces.
//
// Supported marketplaces:
//   - Claude Code: anthropics/claude-plugins-official
//
// Example usage:
//
//	package main
//
//	import (
//	    "context"
//	    "fmt"
//	    "os"
//
//	    "github.com/plexusone/assistantkit/publish"
//	    "github.com/plexusone/assistantkit/publish/claude"
//	)
//
//	func main() {
//	    token := os.Getenv("GITHUB_TOKEN")
//	    publisher := claude.NewPublisher(token)
//
//	    result, err := publisher.Publish(context.Background(), publish.PublishOptions{
//	        PluginDir:  "./plugins/claude",
//	        PluginName: "my-plugin",
//	    })
//	    if err != nil {
//	        panic(err)
//	    }
//
//	    fmt.Printf("PR created: %s\n", result.PRURL)
//	}
package publish

import (
	"github.com/plexusone/assistantkit/publish/core"

	// Import publishers for side-effect registration
	_ "github.com/plexusone/assistantkit/publish/claude"
)

// Re-export core types for convenience.
type (
	Publisher         = core.Publisher
	PublishOptions    = core.PublishOptions
	PublishResult     = core.PublishResult
	MarketplaceConfig = core.MarketplaceConfig
)

// Re-export error types.
type (
	ValidationError = core.ValidationError
	ForkError       = core.ForkError
	BranchError     = core.BranchError
	CommitError     = core.CommitError
	PRError         = core.PRError
	AuthError       = core.AuthError
)
