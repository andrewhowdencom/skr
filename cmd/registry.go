package cmd

import (
	"github.com/spf13/cobra"
)

var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Manage registry interactions",
	Long:  `Manage interactions with OCI registries, such as authentication.`,
}

func init() {
	rootCmd.AddCommand(registryCmd)
}
