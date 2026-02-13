package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "skr",
	Short: "Skill Registry (skr) - Manage Agent Skills via OCI Registries",
	Long: `skr is a CLI tool for managing Agent Skills, enabling you to build, push, pull,
and install skills using standard OCI (Open Container Initiative) registries.

It simplifies the distribution of AI agent capabilities, treating them as versioned
artifacts similar to container images.`,

	// RunE removed to allow default Cobra behavior (print help)
	SilenceErrors: true,
	SilenceUsage:  true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
