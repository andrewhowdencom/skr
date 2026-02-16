package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/andrewhowdencom/skr/pkg/store"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
)

// InspectResult represents the JSON output for the inspect command
type InspectResult struct {
	Ref          string            `json:"Ref"`
	Digest       string            `json:"Digest"`
	Size         int64             `json:"Size"`
	MediaType    string            `json:"MediaType"`
	RepoTags     []string          `json:"RepoTags"`
	Architecture string            `json:"Architecture,omitempty"`
	Os           string            `json:"Os,omitempty"`
	Created      *time.Time        `json:"Created,omitempty"`
	Author       string            `json:"Author,omitempty"`
	Config       *ocispec.Image    `json:"Config,omitempty"`
	Manifest     *ocispec.Manifest `json:"Manifest,omitempty"`
}

var inspectCmdRoot = &cobra.Command{
	Use:   "inspect [ref]",
	Short: "Inspect a skill artifact",
	Long:  `Return low-level information on a skill artifact in JSON format.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ref := args[0]
		ctx := cmd.Context()

		st, err := store.New("")
		if err != nil {
			return fmt.Errorf("failed to initialize store: %w", err)
		}

		// 1. Resolve Reference
		desc, err := st.Resolve(ctx, ref)
		if err != nil {
			return fmt.Errorf("reference %s not found: %w", ref, err)
		}

		result := InspectResult{
			Ref:       ref,
			Digest:    desc.Digest.String(),
			MediaType: desc.MediaType,
			// Size will be updated with accumulated size
		}

		// 2. Fetch Manifest
		manifestRc, err := st.Fetch(ctx, desc)
		if err != nil {
			return fmt.Errorf("failed to fetch manifest: %w", err)
		}
		defer manifestRc.Close()

		manifestBytes, err := io.ReadAll(manifestRc)
		if err != nil {
			return fmt.Errorf("failed to read manifest: %w", err)
		}

		var manifest ocispec.Manifest
		if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
			return fmt.Errorf("failed to parse manifest: %w", err)
		}
		result.Manifest = &manifest

		// Add manifest size
		result.Size += desc.Size

		// 3. Fetch Config
		configRc, err := st.Fetch(ctx, manifest.Config)
		if err != nil {
			// It might be that config is not found or not fetchable, we just log/ignore for now or fail?
			// Standard docker inspect fails if image is corrupted.
			return fmt.Errorf("failed to fetch config: %w", err)
		}
		defer configRc.Close()

		configBytes, err := io.ReadAll(configRc)
		if err != nil {
			return fmt.Errorf("failed to read config: %w", err)
		}

		// Add config size
		result.Size += manifest.Config.Size

		var imageConfig ocispec.Image
		if err := json.Unmarshal(configBytes, &imageConfig); err != nil {
			// Try parsing as generic map if it's not a standard image config?
			// But creating a skill creates a standard config.
			// For now, let's try standard OCI Image config.
			return fmt.Errorf("failed to parse config: %w", err)
		}
		result.Config = &imageConfig
		result.Architecture = imageConfig.Architecture
		result.Os = imageConfig.OS
		result.Created = imageConfig.Created
		result.Author = imageConfig.Author

		// Add layers size
		for _, layer := range manifest.Layers {
			result.Size += layer.Size
		}

		// 4. Find all tags pointing to this digest
		tags, err := st.List(ctx)
		if err != nil {
			return fmt.Errorf("failed to list tags: %w", err)
		}

		repoTags := []string{}
		for _, tag := range tags {
			tagDesc, err := st.Resolve(ctx, tag)
			if err != nil {
				continue
			}
			if tagDesc.Digest == desc.Digest {
				repoTags = append(repoTags, tag)
			}
		}
		result.RepoTags = repoTags

		// 5. Output JSON
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "    ")
		if err := enc.Encode([]InspectResult{result}); err != nil {
			return fmt.Errorf("failed to encode result: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(inspectCmdRoot)
}
