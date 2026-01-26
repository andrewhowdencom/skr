package cmd

import (
	"fmt"

	"github.com/andrewhowdencom/skr/pkg/registry"
	"github.com/andrewhowdencom/skr/pkg/store"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push [ref]",
	Short: "Push an Agent Skill to a registry",
	Long: `Push a built Agent Skill artifact to an OCI registry.

This command uploads the skill image to the configured registry, making it
available for others to pull and install.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ref := args[0]
		ctx := cmd.Context()

		st, err := store.New("")
		if err != nil {
			return fmt.Errorf("failed to initialize store: %w", err)
		}

		fmt.Printf("Pushing %s...\n", ref)
		if err := registry.Push(ctx, st, ref); err != nil {
			return err
		}

		fmt.Printf("Successfully pushed %s\n", ref)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
}
