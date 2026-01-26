package action

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/andrewhowdencom/skr/pkg/skill"
	"github.com/andrewhowdencom/skr/pkg/store"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// InstallSkill installs a skill from the store to the installDir.
// Returns the name of the installed skill.
func InstallSkill(ctx context.Context, st *store.Store, ref, installDir string) (string, error) {
	// 1. Resolve Reference
	desc, err := st.Resolve(ctx, ref)
	if err != nil {
		return "", fmt.Errorf("failed to resolve skill '%s': %w", ref, err)
	}

	// 2. Fetch Manifest
	manifestReader, err := st.Get(ctx, desc)
	if err != nil {
		return "", fmt.Errorf("failed to fetch manifest: %w", err)
	}
	defer manifestReader.Close()

	manifestBytes, err := io.ReadAll(manifestReader)
	if err != nil {
		return "", fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest ocispec.Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return "", fmt.Errorf("failed to parse manifest: %w", err)
	}

	if len(manifest.Layers) != 1 {
		return "", fmt.Errorf("expected exactly 1 layer, got %d", len(manifest.Layers))
	}

	layerDesc := manifest.Layers[0]

	// 3. Fetch Layer
	layerReader, err := st.Get(ctx, layerDesc)
	if err != nil {
		return "", fmt.Errorf("failed to fetch layer: %w", err)
	}
	defer layerReader.Close()

	// 4. Unpack Layer to Temp
	tempDir, err := os.MkdirTemp("", "skr-install-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	if err := unpackLayer(layerReader, tempDir); err != nil {
		return "", fmt.Errorf("failed to unpack layer: %w", err)
	}

	// 5. Read SKILL.md to get the name
	s, err := skill.Load(tempDir)
	if err != nil {
		return "", fmt.Errorf("downloaded artifact is not a valid skill: %w", err)
	}

	targetPath := filepath.Join(installDir, s.Name)
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create parent dir: %w", err)
	}

	// 6. Move tempDir to targetPath (replace if exists)
	if err := os.RemoveAll(targetPath); err != nil {
		return "", fmt.Errorf("failed to remove existing skill at %s: %w", targetPath, err)
	}

	if err := os.Rename(tempDir, targetPath); err != nil {
		// Fallback: Copy content
		if err := copyDir(tempDir, targetPath); err != nil {
			return "", fmt.Errorf("failed to move skill to install dir: %w", err)
		}
	}

	return s.Name, nil
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

		if !filepath.IsLocal(header.Name) {
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
