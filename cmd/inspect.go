package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect an Agent Skill",
	Long: `Inspect the metadata or content of an Agent Skill.

Can be used on local artifacts or remote registry items to see details like
dependencies, version, and manifest.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(inspectCmd)
}
