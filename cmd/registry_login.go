package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to an OCI registry",
	Long: `Log in to an OCI registry using credentials.

This establishes an authenticated session for pushing and pulling private skills.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not yet implemented")
	},
}

func init() {
	registryCmd.AddCommand(loginCmd)
}
