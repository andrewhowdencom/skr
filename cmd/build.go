package cmd

import (
	"fmt"

	"github.com/andrewhowdencom/skr/pkg/skill"
	"github.com/andrewhowdencom/skr/pkg/store"
	"github.com/spf13/cobra"
)

var buildTag string

var buildCmd = &cobra.Command{
	Use:   "build [path]",
	Short: "Build an Agent Skill artifact",
	Long: `Build an Agent Skill artifact from a local directory.

This command packages the skill definition and assets into an OCI-compatible
image format, ready for distribution.

If [path] is not provided, defaults to the current directory.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		s, err := skill.Load(path)
		if err != nil {
			return fmt.Errorf("failed to validate skill: %w", err)
		}

		if buildTag == "" {
			buildTag = fmt.Sprintf("%s:latest", s.Name)
			fmt.Printf("No tag provided. Defaulting to: %s\n", buildTag)
		}

		ctx := cmd.Context()
		st, err := store.New("")
		if err != nil {
			return fmt.Errorf("failed to initialize store: %w", err)
		}

		if err := st.Build(ctx, s.Path, buildTag); err != nil {
			return fmt.Errorf("failed to build artifact: %w", err)
		}

		fmt.Printf("Successfully built artifact for skill '%s'\n", s.Name)
		fmt.Printf("Tagged as: %s\n", buildTag)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().StringVarP(&buildTag, "tag", "t", "", "Tag for the built artifact (e.g., registry.com/skill:v1)")
}
