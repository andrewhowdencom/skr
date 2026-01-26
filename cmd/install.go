package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install an Agent Skill",
	Long: `Install an Agent Skill to the local agent.

This command pulls the skill artifact from a registry (if not local) and
extracts it to the agent's skill directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
