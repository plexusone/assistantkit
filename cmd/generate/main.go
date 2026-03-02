// Command generate converts CONTEXT.json to tool-specific formats.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/plexusone/assistantkit/context"
	_ "github.com/plexusone/assistantkit/context/claude"
)

func main() {
	input := flag.String("input", "CONTEXT.json", "Input context file")
	output := flag.String("output", "", "Output file (default: format-specific)")
	format := flag.String("format", "claude", "Output format (claude)")
	flag.Parse()

	ctx, err := context.ReadFile(*input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", *input, err)
		os.Exit(1)
	}

	outputPath := *output
	if outputPath == "" {
		converter, ok := context.GetConverter(*format)
		if !ok {
			fmt.Fprintf(os.Stderr, "Unknown format: %s\n", *format)
			os.Exit(1)
		}
		outputPath = converter.OutputFileName()
	}

	if err := context.WriteFile(ctx, *format, outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", outputPath, err)
		os.Exit(1)
	}

	fmt.Printf("Generated %s from %s\n", outputPath, *input)
}
