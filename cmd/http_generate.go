package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/andrewhowdencom/skr/pkg/store"
	"github.com/andrewhowdencom/skr/pkg/ui"
	"github.com/spf13/cobra"
)

var outputDir string

var httpGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate the static UI site",
	Long:  `Generate a static HTML website with a static OCI registry file structure, enabling the UI to run without a backend.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		// 1. Initialize Store
		st, err := store.New(ociPath)
		if err != nil {
			return fmt.Errorf("failed to initialize store: %w", err)
		}

		fmt.Println("Scanning OCI store...")
		tags, err := st.List(ctx)
		if err != nil {
			return fmt.Errorf("failed to list skills from store: %w", err)
		}

		// 2. Prepare Output Directory
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		// 3. Generate OCI v2 Structure
		v2Dir := filepath.Join(outputDir, "v2")
		if err := os.MkdirAll(v2Dir, 0755); err != nil {
			return fmt.Errorf("failed to create v2 directory: %w", err)
		}

		// 3.1 _catalog
		repoSet := make(map[string]bool)
		repoTagsMap := make(map[string][]string)

		for _, t := range tags {
			// t is "repo:tag" or "repo"
			var repo, tag string
			if lastIdx := strings.LastIndex(t, ":"); lastIdx != -1 {
				repo = t[:lastIdx]
				tag = t[lastIdx+1:]
			} else {
				repo = t
				tag = "latest" // Default if just repo? Or skip?
			}
			repoSet[repo] = true
			if tag != "" {
				repoTagsMap[repo] = append(repoTagsMap[repo], tag)
			}
		}

		var repos []string
		for r := range repoSet {
			repos = append(repos, r)
		}
		sort.Strings(repos)

		catalogPath := filepath.Join(v2Dir, "_catalog")
		catalogData := map[string][]string{"repositories": repos}
		if err := writeJSON(catalogPath, catalogData); err != nil {
			return err
		}
		fmt.Printf("Generated %s\n", catalogPath)

		// 3.2 Per-Repo Tags and Manifests
		for _, repo := range repos {
			// Create repo structure
			// v2/<repo>/tags/list
			// v2/<repo>/manifests/<tag>

			// Handle nested repos (e.g. library/ubuntu)
			repoDir := filepath.Join(v2Dir, repo)
			if err := os.MkdirAll(repoDir, 0755); err != nil {
				return err
			}

			// Tags List
			tagsList := repoTagsMap[repo]
			sort.Strings(tagsList)

			tagsDir := filepath.Join(repoDir, "tags")
			if err := os.MkdirAll(tagsDir, 0755); err != nil {
				return err
			}

			tagsListPath := filepath.Join(tagsDir, "list")
			tagsResp := map[string]interface{}{
				"name": repo,
				"tags": tagsList,
			}
			if err := writeJSON(tagsListPath, tagsResp); err != nil {
				return err
			}

			// Manifests
			manifestsDir := filepath.Join(repoDir, "manifests")
			if err := os.MkdirAll(manifestsDir, 0755); err != nil {
				return err
			}

			for _, tag := range tagsList {
				// Fetch manifest content from store
				ref := repo + ":" + tag
				desc, err := st.Resolve(ctx, ref)
				if err != nil {
					fmt.Printf("Warning: could not resolve %s: %v\n", ref, err)
					continue
				}

				rc, err := st.Fetch(ctx, desc)
				if err != nil {
					fmt.Printf("Warning: could not fetch %s: %v\n", ref, err)
					continue
				}

				manifestPath := filepath.Join(manifestsDir, tag)
				manifestFile, err := os.Create(manifestPath)
				if err != nil {
					rc.Close()
					return err
				}

				_, err = io.Copy(manifestFile, rc)
				rc.Close()
				manifestFile.Close()
				if err != nil {
					return err
				}
			}
		}

		// 4. Copy UI Assets
		assets, err := ui.Assets()
		if err != nil {
			return fmt.Errorf("failed to load embedded assets: %w", err)
		}

		filesToCopy := []string{"index.html", "style.css", "app.js"}
		for _, fileName := range filesToCopy {
			if err := copyAsset(assets, fileName, filepath.Join(outputDir, fileName)); err != nil {
				return err
			}
			fmt.Printf("Copied %s\n", fileName)
		}

		fmt.Println("Static site generation complete.")
		return nil
	},
}

func init() {
	httpCmd.AddCommand(httpGenerateCmd)
	httpGenerateCmd.Flags().StringVarP(&outputDir, "output", "o", "build/http", "Directory to output the generated site")
}

func writeJSON(path string, data interface{}) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(data)
}

func copyAsset(assets fs.FS, src, dest string) error {
	f, err := assets.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	d, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer d.Close()

	_, err = io.Copy(d, f)
	return err
}
