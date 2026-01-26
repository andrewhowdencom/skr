package cmd

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/andrewhowdencom/skr/pkg/skill"
	"github.com/andrewhowdencom/skr/pkg/store"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install an Agent Skill",
	Long: `Install an Agent Skill to the local agent.

This command pulls the skill artifact from a registry (if not local) and
extracts it to the agent's skill directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("requires skill reference (e.g. tag or digest)")
		}
		ref := args[0]

		// 1. Discover .agent/skills directory
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}

		installDir, err := findInstallDir(cwd)
		if err != nil {
			return err
		}

		fmt.Printf("Installing to %s\n", installDir)

		// 2. Initialize Store
		ctx := cmd.Context()
		st, err := store.New("")
		if err != nil {
			return fmt.Errorf("failed to initialize store: %w", err)
		}

		// 3. Resolve Reference
		desc, err := st.Resolve(ctx, ref)
		if err != nil {
			return fmt.Errorf("failed to resolve skill '%s': %w", ref, err)
		}

		// 4. Fetch Manifest
		manifestReader, err := st.Get(ctx, desc)
		if err != nil {
			return fmt.Errorf("failed to fetch manifest: %w", err)
		}
		defer manifestReader.Close()

		manifestBytes, err := io.ReadAll(manifestReader)
		if err != nil {
			return fmt.Errorf("failed to read manifest: %w", err)
		}

		var manifest ocispec.Manifest
		if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
			return fmt.Errorf("failed to parse manifest: %w", err)
		}

		if len(manifest.Layers) != 1 {
			return fmt.Errorf("expected exactly 1 layer, got %d", len(manifest.Layers))
		}

		layerDesc := manifest.Layers[0]

		// 5. Fetch Layer
		layerReader, err := st.Get(ctx, layerDesc)
		if err != nil {
			return fmt.Errorf("failed to fetch layer: %w", err)
		}
		defer layerReader.Close()

		// 6. Unpack Layer
		// We'll unpack into installDir/<skill-name>
		// Wait, we need the skill name. Ideally it's in the config or we infer from tag.
		// For now, let's use the tag name if possible, or a temp dir?
		// Better yet, unpack to a temp dir, read SKILL.md for name, then move.

		// For simplicity in this step, let's infer name from ref (if tag) or use digest.
		// Actually, let's unpack to a temp dir first.

		tempDir, err := os.MkdirTemp("", "skr-install-*")
		if err != nil {
			return fmt.Errorf("failed to create temp dir: %w", err)
		}
		defer os.RemoveAll(tempDir)

		if err := unpackLayer(layerReader, tempDir); err != nil {
			return fmt.Errorf("failed to unpack layer: %w", err)
		}

		// Read SKILL.md to get the name
		s, err := skill.Load(tempDir)
		if err != nil {
			return fmt.Errorf("downloaded artifact is not a valid skill: %w", err)
		}

		targetPath := filepath.Join(installDir, s.Name)
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("failed to create parent dir: %w", err)
		}

		// Move tempDir to targetPath (replace if exists)
		if err := os.RemoveAll(targetPath); err != nil {
			return fmt.Errorf("failed to remove existing skill at %s: %w", targetPath, err)
		}

		// Rename can fail across devices, so generally copy+rm is safer but Rename is atomic-ish on same FS.
		if err := os.Rename(tempDir, targetPath); err != nil {
			// Fallback: Copy content
			if err := copyDir(tempDir, targetPath); err != nil {
				return fmt.Errorf("failed to move skill to install dir: %w", err)
			}
		}

		// Because we Deferred RemoveAll tempDir, and we just moved it, RemoveAll will do nothing (dir gone) or just fail silently?
		// os.RemoveAll returns nil on NotExist.

		fmt.Printf("Successfully installed skill '%s' to %s\n", s.Name, targetPath)

		return nil
	},
}

func findInstallDir(startDir string) (string, error) {
	dir := startDir
	for {
		target := filepath.Join(dir, ".agent", "skills")
		info, err := os.Stat(target)
		if err == nil && info.IsDir() {
			return target, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("could not find .agent/skills directory in any parent of %s. Please create it first.", startDir)
}

func unpackLayer(r io.Reader, dest string) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)

		// Sanity check for ZipSlip
		if !filepath.IsLocal(header.Name) { // Go 1.20+
			return fmt.Errorf("tar archive contains unsafe filename: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}
	return nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		return copyFile(path, targetPath, info.Mode())
	})
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	return os.Chmod(dst, mode)
}

func init() {
	rootCmd.AddCommand(installCmd)
}
