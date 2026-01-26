package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm",
	Short: "Remove an Agent Skill",
	Long: `Remove an installed Agent Skill from the local agent.

This command deletes the skill's files and configuration from the agent's
skill directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
}
