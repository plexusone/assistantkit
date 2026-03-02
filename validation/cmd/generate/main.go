// Command generate creates validation area files from canonical specs.
// Supports multiple output formats: Claude (agents), Gemini (commands), Codex (prompts).
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/plexusone/assistantkit/validation"
	_ "github.com/plexusone/assistantkit/validation/claude" // Register Claude adapter
	_ "github.com/plexusone/assistantkit/validation/codex"  // Register Codex adapter
	_ "github.com/plexusone/assistantkit/validation/gemini" // Register Gemini adapter
)

func main() {
	var (
		specsDir  = flag.String("specs", "validation/specs", "Directory containing canonical JSON specs")
		outputDir = flag.String("output", "/tmp/validation-agents", "Output directory")
		adapters  = flag.String("adapters", "claude", "Comma-separated list of adapters (claude, gemini, codex, or all)")
		listOnly  = flag.Bool("list", false, "List available adapters and exit")
	)

	flag.Usage = func() {
		//nolint:gosec // G705: CLI usage output to stderr, not web content
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Generate validation area files from canonical specs.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -specs=./specs -output=./output -adapters=claude\n", os.Args[0]) //nolint:gosec // G705: CLI usage
		fmt.Fprintf(os.Stderr, "  %s -specs=./specs -output=./output -adapters=all\n", os.Args[0])    //nolint:gosec // G705: CLI usage
		fmt.Fprintf(os.Stderr, "  %s -list\n", os.Args[0])                                            //nolint:gosec // G705: CLI usage
	}

	flag.Parse()

	// List adapters and exit
	if *listOnly {
		fmt.Println("Available adapters:")
		for _, name := range validation.AdapterNames() {
			adapter, _ := validation.GetAdapter(name)
			fmt.Printf("  - %s (ext: %s, dir: %s)\n", name, adapter.FileExtension(), adapter.DefaultDir())
		}
		return
	}

	// Read canonical specs
	areas, err := validation.ReadCanonicalDir(*specsDir)
	if err != nil {
		log.Fatalf("Failed to read specs from %s: %v", *specsDir, err)
	}

	fmt.Printf("Found %d validation areas\n", len(areas))

	// Determine which adapters to use
	var adapterNames []string
	if *adapters == "all" {
		adapterNames = validation.AdapterNames()
	} else {
		adapterNames = strings.Split(*adapters, ",")
		for i := range adapterNames {
			adapterNames[i] = strings.TrimSpace(adapterNames[i])
		}
	}

	// Generate files for each adapter
	for _, adapterName := range adapterNames {
		adapter, ok := validation.GetAdapter(adapterName)
		if !ok {
			log.Printf("Warning: unknown adapter %q, skipping", adapterName)
			continue
		}

		adapterDir := filepath.Join(*outputDir, adapterName)
		err = validation.WriteAreasToDir(areas, adapterDir, adapterName)
		if err != nil {
			log.Fatalf("Failed to write %s files: %v", adapterName, err)
		}

		fmt.Printf("\nGenerated %s %s in %s:\n", adapterName, adapter.DefaultDir(), adapterDir)
		for _, area := range areas {
			fmt.Printf("  - %s%s\n", area.Name, adapter.FileExtension())
		}
	}
}
