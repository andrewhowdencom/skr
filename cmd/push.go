package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push an Agent Skill to a registry",
	Long: `Push a built Agent Skill artifact to an OCI registry.

This command uploads the skill image to the configured registry, making it
available for others to pull and install.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
}
