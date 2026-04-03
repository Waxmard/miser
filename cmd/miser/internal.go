package main

import "github.com/spf13/cobra"

var internalCmd = &cobra.Command{
	Use:   "internal",
	Short: "Commands used by Claude Code cron jobs (not for direct use)",
}

var internalWriteCmd = &cobra.Command{
	Use:   "write",
	Short: "Write Claude's analysis results to the database",
}

func init() {
	internalCmd.AddCommand(internalWriteCmd)
	rootCmd.AddCommand(internalCmd)
}
