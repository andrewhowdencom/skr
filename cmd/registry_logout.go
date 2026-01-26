package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out from an OCI registry",
	Long: `Log out from an OCI registry.

This removes the stored credentials for the specified registry.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not yet implemented")
	},
}

func init() {
	registryCmd.AddCommand(logoutCmd)
}
