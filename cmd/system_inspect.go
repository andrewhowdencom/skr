package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/andrewhowdencom/skr/pkg/store"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect [ref]",
	Short: "Inspect an Agent Skill",
	Long: `Inspect the metadata or content of an Agent Skill.

Can be used on local artifacts to see details like dependencies, version, and manifest.`,
	Args: cobra.ExactArgs(1),
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

		fmt.Printf("Reference: %s\n", ref)
		fmt.Printf("Digest: %s\n", desc.Digest)
		fmt.Printf("Size: %d bytes\n", desc.Size)
		fmt.Printf("MediaType: %s\n", desc.MediaType)

		// 2. Fetch Manifest
		rc, err := st.Fetch(ctx, desc)
		if err != nil {
			return fmt.Errorf("failed to fetch manifest: %w", err)
		}
		defer rc.Close()

		manifestBytes, err := io.ReadAll(rc)
		if err != nil {
			return fmt.Errorf("failed to read manifest: %w", err)
		}

		var manifest ocispec.Manifest
		if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
			return fmt.Errorf("failed to parse manifest: %w", err)
		}

		// Print Annotations
		if len(manifest.Annotations) > 0 {
			fmt.Println("\nAnnotations:")
			for k, v := range manifest.Annotations {
				fmt.Printf("  %s: %s\n", k, v)
			}
		}

		// 3. Fetch Config (if available)
		fmt.Println("\nConfig:")
		fmt.Printf("  Digest: %s\n", manifest.Config.Digest)
		fmt.Printf("  MediaType: %s\n", manifest.Config.MediaType)

		// Optionally fetch config content to show creation time etc.
		configRc, err := st.Fetch(ctx, manifest.Config)
		if err == nil {
			defer configRc.Close()
			configBytes, _ := io.ReadAll(configRc)
			var configMap map[string]interface{}
			if err := json.Unmarshal(configBytes, &configMap); err == nil {
				if created, ok := configMap["created"]; ok {
					fmt.Printf("  Created: %v\n", created)
				}
			}
		}

		fmt.Printf("\nLayers: %d\n", len(manifest.Layers))
		for i, layer := range manifest.Layers {
			fmt.Printf("  [%d] %s (%d bytes)\n", i, layer.Digest, layer.Size)
		}

		return nil
	},
}

func init() {
	systemCmd.AddCommand(inspectCmd)
}
