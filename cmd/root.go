package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "skr",
	Short: "Skill Registry (skr) - Manage Agent Skills via OCI Registries",
	Long: `skr is a CLI tool for managing Agent Skills, enabling you to build, push, pull,
and install skills using standard OCI (Open Container Initiative) registries.

It simplifies the distribution of AI agent capabilities, treating them as versioned
artifacts similar to container images.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Help()
		return fmt.Errorf("subcommand required")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
