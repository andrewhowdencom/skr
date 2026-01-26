package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull an Agent Skill from a registry",
	Long: `Pull an Agent Skill artifact from an OCI registry.

This downloads the skill image but does not install it. Useful for inspecting
content or preparing for offline installation.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
}
