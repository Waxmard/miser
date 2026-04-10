package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var genDocsCmd = &cobra.Command{
	Use:    "gen-docs [output-dir]",
	Short:  "Generate markdown documentation for all commands",
	Hidden: true,
	Args:   cobra.MaximumNArgs(1),
	RunE:   runGenDocs,
}

func init() {
	rootCmd.AddCommand(genDocsCmd)
}

func runGenDocs(_ *cobra.Command, args []string) error {
	outputDir := "docs/commands"
	if len(args) > 0 {
		outputDir = args[0]
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	rootCmd.DisableAutoGenTag = true
	if err := doc.GenMarkdownTree(rootCmd, outputDir); err != nil {
		return fmt.Errorf("generate docs: %w", err)
	}

	fmt.Printf("Docs written to %s/\n", outputDir)
	return nil
}
