package cmd

import (
	"fmt"

	"github.com/andrewhowdencom/skr/pkg/store"
	"github.com/spf13/cobra"
)

var systemPruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove unused data",
	Long:  `Remove unused data (blob garbage collection). This command deletes any content in the local store that is not referenced by any tag.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := store.New("")
		if err != nil {
			return fmt.Errorf("failed to initialize store: %w", err)
		}

		ctx := cmd.Context()
		count, size, err := st.Prune(ctx)
		if err != nil {
			return fmt.Errorf("failed to prune system: %w", err)
		}

		fmt.Printf("Deleted %d artifacts\n", count)
		fmt.Printf("Reclaimed %.2f MB\n", float64(size)/1024/1024)

		return nil
	},
}

func init() {
	systemCmd.AddCommand(systemPruneCmd)
}
