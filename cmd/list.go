package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed Agent Skills",
	Long: `List all Agent Skills currently installed on the local agent.

Displays name, version, and other metadata for installed skills.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
