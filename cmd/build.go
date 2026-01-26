package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build an Agent Skill artifact",
	Long: `Build an Agent Skill artifact from a local directory.

This command packages the skill definition and assets into an OCI-compatible
image format, ready for distribution.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
