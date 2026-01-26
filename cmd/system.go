package cmd

import (
	"github.com/spf13/cobra"
)

var systemCmd = &cobra.Command{
	Use:   "system",
	Short: "Manage system-wide resources",
	Long:  `Manage system-wide resources such as storage and configuration.`,
}

func init() {
	rootCmd.AddCommand(systemCmd)
}
