package cmd

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/andrewhowdencom/skr/pkg/git"
	"github.com/andrewhowdencom/skr/pkg/registry"
	"github.com/andrewhowdencom/skr/pkg/skill"
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
				isChanged := false
				for _, cf := range changedFiles {
					// Normalize paths
					relCf := cf
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

		// Get git metadata
		shortSHA, _ := git.GetShortSHA()
		headTags, _ := git.GetHeadTags()
		repoName, _ := cmd.Flags().GetString("repository")

		// 3. Build and Push each skill
		var errs []error
		for _, skillPath := range skills {
			skillName := filepath.Base(skillPath)

			// Construct base image name
			var baseImageName string
			if repoName != "" {
				// ghcr.io/owner/repo.skill
				baseImageName = fmt.Sprintf("%s/%s/%s.%s", registryHost, namespace, repoName, skillName)
			} else {
				// ghcr.io/owner/skill
				baseImageName = fmt.Sprintf("%s/%s/%s", registryHost, namespace, skillName)
			}

			// Determine tags
			var tags []string
			tags = append(tags, baseImageName+":latest")
			if shortSHA != "" {
				tags = append(tags, baseImageName+":"+shortSHA)
			}
			for _, t := range headTags {
				tags = append(tags, baseImageName+":"+t)
			}

			fmt.Printf("\nProcessing %s -> %v\n", skillName, tags)

			absPath, _ := filepath.Abs(skillPath)

			// Load skill metadata for annotations
			s, err := skill.Load(skillPath)
			if err != nil {
				fmt.Printf("Failed to load skill metadata for %s: %v\n", skillName, err)
				errs = append(errs, fmt.Errorf("failed to load skill %s: %w", skillName, err))
				continue
			}

			// Build annotations
			annotations := map[string]string{
				"org.opencontainers.image.title": s.Name,
			}
			if s.Description != "" {
				annotations["org.opencontainers.image.description"] = s.Description
				annotations["com.skr.description"] = s.Description
			}
			if s.Metadata.Author != "" {
				annotations["com.skr.author"] = s.Metadata.Author
			}
			if s.Metadata.Version != "" {
				annotations["com.skr.version"] = s.Metadata.Version
			}

			for _, tag := range tags {
				// Build (idempotent content-wise, just updates tag reference)
				if err := st.Build(ctx, absPath, tag, annotations); err != nil {
					fmt.Printf("Build failure for %s: %v\n", tag, err)
					errs = append(errs, fmt.Errorf("build failed for %s: %w", tag, err))
					continue
				}

				// Push
				if err := registry.Push(ctx, st, tag); err != nil {
					fmt.Printf("Push failure for %s: %v\n", tag, err)
					errs = append(errs, fmt.Errorf("push failed for %s: %w", tag, err))
					continue
				}
				fmt.Printf("Published %s\n", tag)
			}
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
	batchPublishCmd.Flags().String("repository", "", "Repository name (optional, enables repo.skill naming)")
	batchPublishCmd.MarkFlagRequired("registry")
	batchPublishCmd.MarkFlagRequired("namespace")
}
