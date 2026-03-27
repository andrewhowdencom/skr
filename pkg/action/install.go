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

	"github.com/andrewhowdencom/skr/pkg/resolution"
	"github.com/andrewhowdencom/skr/pkg/skill"
	"github.com/andrewhowdencom/skr/pkg/source"
	"github.com/andrewhowdencom/skr/pkg/store"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// InstallSkill installs a skill and its dependencies from the store to the installDir.
// Returns the root name, manifest digest, and error.
func InstallSkill(ctx context.Context, st *store.Store, ref, installDir string, forcePull bool) (string, string, error) {
	// 1. Resolve all dependencies
	resolver := resolution.New(st)
	resolver.SetPuller(func(ctx context.Context, depRef string) (string, error) {
		fmt.Printf("Fetching missing dependency %s...\n", depRef)
		parsedRef := source.ParseReference(depRef)
		fetcher, err := source.GetFetcher(parsedRef)
		if err != nil {
			return "", fmt.Errorf("failed to get fetcher for %s: %w", depRef, err)
		}
		newRef, err := fetcher.Fetch(ctx, st, parsedRef)
		if err != nil {
			return "", err
		}
		return newRef.Path + ":" + newRef.Spec, nil
	})

	refs, err := resolver.Resolve(ctx, ref)
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve dependencies for %s: %w", ref, err)
	}

	var rootName string
	var rootDigest string

	// 2. Install each skill (sequentially for now)
	for i, r := range refs {
		name, d, err := installOne(ctx, st, r, installDir, forcePull)
		if err != nil {
			return "", "", fmt.Errorf("failed to install %s: %w", r, err)
		}

		// The first one in the resolved list is the root skill (BFS start)
		if i == 0 {
			rootName = name
			rootDigest = d
		}
	}

	return rootName, rootDigest, nil
}

func installOne(ctx context.Context, st *store.Store, refStr, installDir string, forcePull bool) (string, string, error) {
	// 1. Resolve Reference locally
	desc, err := st.Resolve(ctx, refStr)
	shouldPull := false

	if err != nil {
		shouldPull = true
	} else if forcePull {
		shouldPull = true
	} else {
		if len(refStr) > 7 && refStr[len(refStr)-7:] == ":latest" {
			shouldPull = true
		}
	}

	if shouldPull {
		fmt.Printf("Fetching %s...\n", refStr)
		parsedRef := source.ParseReference(refStr)
		fetcher, err := source.GetFetcher(parsedRef)
		if err != nil {
			return "", "", fmt.Errorf("failed to initialize fetcher for %s: %w", refStr, err)
		}

		newRef, fetchErr := fetcher.Fetch(ctx, st, parsedRef)
		if fetchErr != nil {
			if desc.Digest != "" {
				fmt.Printf("Warning: Failed to fetch latest (using local copy): %v\n", fetchErr)
			} else {
				return "", "", fmt.Errorf("failed to fetch %s: %w", refStr, fetchErr)
			}
		} else {
			// Resolve the newly produced artifact from the store
			resolveStr := newRef.Path
			if newRef.Spec != "" {
				resolveStr += ":" + newRef.Spec
			}
			newDesc, resolveErr := st.Resolve(ctx, resolveStr)
			if resolveErr != nil {
				return "", "", fmt.Errorf("failed to resolve newly fetched artifact %s: %w", resolveStr, resolveErr)
			}
			desc = newDesc
		}
	}

	// 2. Fetch Manifest
	manifestReader, err := st.Fetch(ctx, desc)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch manifest: %w", err)
	}
	defer func() { _ = manifestReader.Close() }()

	manifestBytes, err := io.ReadAll(manifestReader)
	if err != nil {
		return "", "", fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest ocispec.Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return "", "", fmt.Errorf("failed to parse manifest: %w", err)
	}

	if len(manifest.Layers) != 1 {
		return "", "", fmt.Errorf("expected exactly 1 layer, got %d", len(manifest.Layers))
	}

	layerDesc := manifest.Layers[0]

	// 3. Fetch Layer
	layerReader, err := st.Fetch(ctx, layerDesc)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch layer: %w", err)
	}
	defer func() { _ = layerReader.Close() }()

	// 4. Unpack Layer to Temp
	tempDir, err := os.MkdirTemp("", "skr-install-*")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	if err := unpackLayer(layerReader, tempDir); err != nil {
		return "", "", fmt.Errorf("failed to unpack layer: %w", err)
	}

	// 5. Read SKILL.md to get the name
	s, err := skill.LoadUnverified(tempDir)
	if err != nil {
		// If we can't even load it (missing file, invalid yaml), we still fail as we need the name.
		return "", "", fmt.Errorf("downloaded artifact is not a recognizable skill: %w", err)
	}

	// Soft Validate: check if it's strictly valid, but don't fail, just warn.
	if err := s.Validate(); err != nil {
		fmt.Printf("Warning: Installed skill '%s' has validation issues: %v\n", s.Name, err)
	}

	targetPath := filepath.Join(installDir, s.Name)
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return "", "", fmt.Errorf("failed to create parent dir: %w", err)
	}

	// 6. Move tempDir to targetPath (replace if exists)
	if err := os.RemoveAll(targetPath); err != nil {
		return "", "", fmt.Errorf("failed to remove existing skill at %s: %w", targetPath, err)
	}

	if err := os.Rename(tempDir, targetPath); err != nil {
		// Fallback: Copy content
		if err := copyDir(tempDir, targetPath); err != nil {
			return "", "", fmt.Errorf("failed to move skill to install dir: %w", err)
		}
	}

	return s.Name, string(desc.Digest), nil
}

func unpackLayer(r io.Reader, dest string) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer func() { _ = gzr.Close() }()

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
	defer func() { _ = in.Close() }()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	return os.Chmod(dst, mode)
}
