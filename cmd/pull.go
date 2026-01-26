package cmd

import (
	"fmt"

	"github.com/andrewhowdencom/skr/pkg/registry"
	"github.com/andrewhowdencom/skr/pkg/store"
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull [ref]",
	Short: "Pull an Agent Skill from a registry",
	Long: `Pull an Agent Skill artifact from an OCI registry.

This downloads the skill image but does not install it. Useful for inspecting
content or preparing for offline installation.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ref := args[0]
		ctx := cmd.Context()

		st, err := store.New("")
		if err != nil {
			return fmt.Errorf("failed to initialize store: %w", err)
		}

		fmt.Printf("Pulling %s...\n", ref)
		if err := registry.Pull(ctx, st, ref); err != nil {
			return err
		}

		fmt.Printf("Successfully pulled %s\n", ref)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
}
