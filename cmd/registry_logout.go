package cmd

import (
	"fmt"

	"github.com/andrewhowdencom/skr/pkg/auth"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout [server]",
	Short: "Log out from an OCI registry",
	Long: `Log out from an OCI registry.

This removes the stored credentials for the specified registry.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		server := "ghcr.io" // Default
		if len(args) > 0 {
			server = args[0]
		}

		fmt.Printf("Logging out from %s...\n", server)
		if err := auth.Logout(server); err != nil {
			return err
		}

		fmt.Println("Logout Succeeded")
		return nil
	},
}

func init() {
	registryCmd.AddCommand(logoutCmd)
}
