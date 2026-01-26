package cmd

import (
	"fmt"
	"os/exec"
	"strings"

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

		// Detect Git Remote for source annotation
		annotations := make(map[string]string)
		if sourceURL, err := getGitRemoteURL(); err == nil && sourceURL != "" {
			annotations["org.opencontainers.image.source"] = sourceURL
			fmt.Printf("Detected git source: %s\n", sourceURL)
		}

		if err := st.Build(ctx, s.Path, buildTag, annotations); err != nil {
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

func getGitRemoteURL() (string, error) {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	url := strings.TrimSpace(string(out))
	// Convert SSH URL to HTTPS if needed? GitHub packages supports both but HTTPS is standard for source linking?
	// Actually GHCR works with repo URL.
	// git@github.com:user/repo.git -> https://github.com/user/repo
	if strings.HasPrefix(url, "git@github.com:") {
		url = strings.Replace(url, "git@github.com:", "https://github.com/", 1)
		url = strings.TrimSuffix(url, ".git")
	} else if strings.HasPrefix(url, "https://github.com/") {
		url = strings.TrimSuffix(url, ".git")
	}

	return url, nil
}
