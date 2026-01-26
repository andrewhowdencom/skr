package cmd

import (
	"github.com/spf13/cobra"
)

var batchCmd = &cobra.Command{
	Use:   "batch",
	Short: "Perform batch operations on multiple skills",
	Long:  `Perform operations on collections of skills, useful for monorepo workflows.`,
}

func init() {
	rootCmd.AddCommand(batchCmd)
}
