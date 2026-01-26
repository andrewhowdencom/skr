package cmd

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/andrewhowdencom/skr/pkg/git"
	"github.com/andrewhowdencom/skr/pkg/registry"
	"github.com/andrewhowdencom/skr/pkg/store"
	"github.com/spf13/cobra"
)

var batchPublishCmd = &cobra.Command{
	Use:   "publish [path]",
	Short: "Build and push multiple skills",
	Long: `Recursively find skills in a directory and publish them.
If --base is provided, only publishes skills that have changed since that git reference.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		rootDir := "."
		if len(args) > 0 {
			rootDir = args[0]
		}

		baseRef, _ := cmd.Flags().GetString("base")
		registryHost, _ := cmd.Flags().GetString("registry")
		namespace, _ := cmd.Flags().GetString("namespace")

		if registryHost == "" || namespace == "" {
			return fmt.Errorf("--registry and --namespace are required for batch publishing")
		}

		// 1. Find all SKILL.md files
		var skills []string
		err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && d.Name() == "SKILL.md" {
				skills = append(skills, filepath.Dir(path))
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to walk directory: %w", err)
		}

		fmt.Printf("Found %d skills in %s\n", len(skills), rootDir)

		// 2. Filter by changed files if --base is set
		if baseRef != "" {
			fmt.Printf("Checking changes against %s...\n", baseRef)
			changedFiles, err := git.ChangedFiles(baseRef)
			if err != nil {
				return fmt.Errorf("failed to detect changes: %w", err)
			}

			var changedSkills []string
			for _, skillPath := range skills {
				// naive check: is any changed file inside the skill directory?
				// Rel path logic...
				isChanged := false
				for _, cf := range changedFiles {
					// We need absolute or relative paths to match?
					// git diff returns paths relative to repo root.
					// skillPath is relative to CWD (rootDir).
					// This comparison is tricky unless we resolve everything to absolute or repo-relative.
					// Assumption: CWD is repo root for now, or close to it.
					// Better approach: Check if `cf` starts with `skillPath` (cleanly).

					// Normalize paths
					relCf := cf
					// If skillPath is ".", logic breaks.
					// If skillPath is "skills/foo", and cf is "skills/foo/SKILL.md", it matches.

					if strings.HasPrefix(relCf, skillPath) {
						isChanged = true
						break
					}
				}
				if isChanged {
					changedSkills = append(changedSkills, skillPath)
				}
			}
			skills = changedSkills
			fmt.Printf("Identified %d changed skills\n", len(skills))
		}

		if len(skills) == 0 {
			fmt.Println("No skills to publish.")
			return nil
		}

		st, err := store.New("")
		if err != nil {
			return fmt.Errorf("failed to initialize store: %w", err)
		}

		// 3. Build and Push each skill
		var errs []error
		for _, skillPath := range skills {
			skillName := filepath.Base(skillPath)
			// Version strategy: use 'latest' and maybe a git-sha tag if we calculated it?
			// For simplicity in this iteration: uses 'latest' + Git SHA if we can get it.
			// Let's stick to :latest for the first pass, as versioning from SKILL.md requires parsing.

			// FIXME: We probably want a version strategy flag.
			// Let's generate a pseudo-version "hash-{shortSHA}" or just use "latest"
			tag := fmt.Sprintf("%s/%s/%s:latest", registryHost, namespace, skillName)

			fmt.Printf("\nProcessing %s -> %s\n", skillName, tag)

			absPath, _ := filepath.Abs(skillPath)
			// Build
			// Annotations?
			if err := st.Build(ctx, absPath, tag, nil); err != nil {
				fmt.Printf("Check failure: %v\n", err)
				errs = append(errs, fmt.Errorf("build failed for %s: %w", skillName, err))
				continue
			}

			// Push
			if err := registry.Push(ctx, st, tag); err != nil {
				fmt.Printf("Push failure: %v\n", err)
				errs = append(errs, fmt.Errorf("push failed for %s: %w", skillName, err))
				continue
			}
			fmt.Printf("Published %s\n", tag)
		}

		if len(errs) > 0 {
			return fmt.Errorf("encountered %d errors during batch publish", len(errs))
		}

		return nil
	},
}

func init() {
	batchCmd.AddCommand(batchPublishCmd)
	batchPublishCmd.Flags().String("base", "", "Git reference to compare against (e.g. origin/main)")
	batchPublishCmd.Flags().String("registry", "", "Registry host (e.g. ghcr.io)")
	batchPublishCmd.Flags().String("namespace", "", "Registry namespace (e.g. user or org)")
	batchPublishCmd.MarkFlagRequired("registry")
	batchPublishCmd.MarkFlagRequired("namespace")
}
