package cmd

import (
	"fmt"
	"strings"

	"github.com/andrewhowdencom/skr/pkg/store"
	"github.com/spf13/cobra"
)

var systemListCmd = &cobra.Command{
	Use:   "list",
	Short: "List built/pulled artifacts in Local Registry",
	Long: `List all skill artifacts stored in the local OCI registry.
	
Shows repository, tag, digest, and size.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		st, err := store.New("")
		if err != nil {
			return fmt.Errorf("failed to initialize store: %w", err)
		}

		tags, err := st.List(ctx)
		if err != nil {
			return fmt.Errorf("failed to list skills: %w", err)
		}

		// Header
		fmt.Printf("%-30s %-15s %-15s %-10s\n", "REPOSITORY", "TAG", "IMAGE ID", "SIZE")

		for _, tag := range tags {
			// Resolve to get digest and size
			desc, err := st.Resolve(ctx, tag)
			if err != nil {
				// warn and continue?
				continue
			}

			// For REPOSITORY/TAG splitting, we assume standard "repo:tag" format.
			repo := tag
			version := "<none>"

			if idx := lastIndex(tag, ":"); idx != -1 {
				repo = tag[:idx]
				version = tag[idx+1:]
			}

			// Short digest
			digestVal := desc.Digest.String()
			if len(digestVal) > 12 {
				digestVal = digestVal[7:19] // sha256:1234... -> 1234...
			}

			// Size (human readable-ish)
			size := fmt.Sprintf("%d B", desc.Size)
			if desc.Size > 1024 {
				size = fmt.Sprintf("%.2f KB", float64(desc.Size)/1024)
			}

			fmt.Printf("%-30s %-15s %-15s %-10s\n", repo, version, digestVal, size)
		}

		return nil
	},
}

func init() {
	systemCmd.AddCommand(systemListCmd)
}

func lastIndex(s, sep string) int {
	return strings.LastIndex(s, sep)
}
