package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Create a tag for a local Agent Skill",
	Long: `Tag a local Agent Skill image with a new name and/or tag.

This is mostly used to prepare a built skill for pushing to a specific registry.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(tagCmd)
}
