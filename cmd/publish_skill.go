package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/andrewhowdencom/skr/pkg/registry"
	"github.com/andrewhowdencom/skr/pkg/skill"
	"github.com/andrewhowdencom/skr/pkg/store"
	"github.com/spf13/cobra"
)

var publishSkillCmd = &cobra.Command{
	Use:   "publish [path]",
	Short: "Build and push a skill artifact",
	Long:  `Build a skill from a directory and immediately push it to a registry.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		srcDir := "."
		if len(args) > 0 {
			srcDir = args[0]
		}

		tag, _ := cmd.Flags().GetString("tag")
		if tag == "" {
			return fmt.Errorf("a tag is required for publishing (e.g. --tag ghcr.io/user/skill:v1)")
		}

		// 1. Build
		absPath, err := filepath.Abs(srcDir)
		if err != nil {
			return fmt.Errorf("failed to resolve absolute path: %w", err)
		}

		// Load skill metadata for annotations
		s, err := skill.Load(srcDir)
		if err != nil {
			return fmt.Errorf("failed to load skill metadata: %w", err)
		}

		st, err := store.New("")
		if err != nil {
			return fmt.Errorf("failed to initialize store: %w", err)
		}

		// Build annotations
		annotations := map[string]string{
			"org.opencontainers.image.created": "true", // Placeholder, actual time added by store
			"org.opencontainers.image.title":   s.Name,
		}
		if s.Description != "" {
			annotations["org.opencontainers.image.description"] = s.Description
		}
		if s.Metadata.Author != "" {
			annotations["com.skr.author"] = s.Metadata.Author
		}
		if s.Metadata.Version != "" {
			annotations["com.skr.version"] = s.Metadata.Version
		}

		fmt.Printf("Building skill from %s...\n", srcDir)
		if err := st.Build(ctx, absPath, tag, annotations); err != nil {
			return fmt.Errorf("failed to build artifact: %w", err)
		}
		fmt.Printf("Successfully built %s\n", tag)

		// 2. Push
		fmt.Printf("Pushing %s...\n", tag)

		if err := registry.Push(ctx, st, tag); err != nil {
			return fmt.Errorf("failed to push artifact: %w", err)
		}

		fmt.Printf("Successfully published %s\n", tag)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(publishSkillCmd)
	publishSkillCmd.Flags().StringP("tag", "t", "", "Tag for the artifact (required)")
	publishSkillCmd.MarkFlagRequired("tag")
}
