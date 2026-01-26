package store

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	_ "crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/adrg/xdg"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/oci"
)

const (
	MediaTypeSkillLayer  = "application/vnd.agentskills.skill.layer.v1+tar+gzip"
	MediaTypeSkillConfig = "application/vnd.agentskills.skill.config.v1+json"
	StoreDirName         = "skr/store"
)

type Store struct {
	path string
	oci  *oci.Store
}

func New(path string) (*Store, error) {
	if path == "" {
		dataPath, err := xdg.DataFile(StoreDirName)
		if err != nil {
			return nil, fmt.Errorf("failed to get XDG data path: %w", err)
		}
		path = dataPath
	}

	err := os.MkdirAll(path, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create store directory: %w", err)
	}

	ociStore, err := oci.New(path)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OCI store: %w", err)
	}

	return &Store{
		path: path,
		oci:  ociStore,
	}, nil
}

func (s *Store) Build(ctx context.Context, srcDir string, tag string) error {
	// 1. Create a tarball of the directory
	buf := &bytes.Buffer{}
	gw := gzip.NewWriter(buf)
	tw := tar.NewWriter(gw)

	err := filepath.Walk(srcDir, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(srcDir, file)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !fi.IsDir() {
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			defer data.Close()
			if _, err := io.Copy(tw, data); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk source directory: %w", err)
	}

	if err := tw.Close(); err != nil {
		return err
	}
	if err := gw.Close(); err != nil {
		return err
	}

	layerBytes := buf.Bytes()
	layerDigest := digest.FromBytes(layerBytes)
	layerSize := int64(len(layerBytes))

	// 2. Push layer to store
	layerDesc := ocispec.Descriptor{
		MediaType: MediaTypeSkillLayer,
		Digest:    layerDigest,
		Size:      layerSize,
	}

	err = s.pushBlob(ctx, layerDesc, bytes.NewReader(layerBytes))
	if err != nil {
		return fmt.Errorf("failed to push layer: %w", err)
	}

	// 3. Create and push config (empty for now)
	config := map[string]string{
		"created": time.Now().UTC().Format(time.RFC3339),
	}
	configBytes, _ := json.Marshal(config)
	configDigest := digest.FromBytes(configBytes)
	configDesc := ocispec.Descriptor{
		MediaType: MediaTypeSkillConfig,
		Digest:    configDigest,
		Size:      int64(len(configBytes)),
	}
	err = s.pushBlob(ctx, configDesc, bytes.NewReader(configBytes))
	if err != nil {
		return fmt.Errorf("failed to push config: %w", err)
	}

	// 4. Create and push Manifest
	// 4. Create and push Manifest
	manifest := ocispec.Manifest{
		Config: configDesc,
		Layers: []ocispec.Descriptor{layerDesc},
	}
	manifest.SchemaVersion = 2

	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	manifestDesc := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageManifest,
		Digest:    digest.FromBytes(manifestBytes),
		Size:      int64(len(manifestBytes)),
	}

	err = s.pushBlob(ctx, manifestDesc, bytes.NewReader(manifestBytes))
	if err != nil {
		return fmt.Errorf("failed to push manifest: %w", err)
	}

	// 5. Tag the manifest
	if tag != "" {
		err = s.oci.Tag(ctx, manifestDesc, tag)
		if err != nil {
			return fmt.Errorf("failed to tag artifact: %w", err)
		}
	}

	return nil
}

// Get retrieves content by digest
func (s *Store) Get(ctx context.Context, target ocispec.Descriptor) (io.ReadCloser, error) {
	return s.oci.Fetch(ctx, target)
}

// List returns a list of all tags in the store
func (s *Store) List(ctx context.Context) ([]string, error) {
	var tags []string
	err := s.oci.Tags(ctx, "", func(tagsList []string) error {
		tags = append(tags, tagsList...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tags, nil
}

// Resolve resolves a reference (tag/digest) to a descriptor
func (s *Store) Resolve(ctx context.Context, ref string) (ocispec.Descriptor, error) {
	return s.oci.Resolve(ctx, ref)
}

// Prune removes all unreferenced blobs from the store
func (s *Store) Prune(ctx context.Context) (int, int64, error) {
	// 1. Identify all reachable blobs
	reachable := make(map[string]bool)

	err := s.oci.Tags(ctx, "", func(tags []string) error {
		for _, tag := range tags {
			// Resolve tag to manifest descriptor
			desc, err := s.oci.Resolve(ctx, tag)
			if err != nil {
				return err
			}
			reachable[desc.Digest.String()] = true

			// Fetch and parse manifest to find children (config + layers)
			rc, err := s.oci.Fetch(ctx, desc)
			if err != nil {
				return err
			}
			manifestBytes, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return err
			}

			var manifest ocispec.Manifest
			if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
				return err
			}

			// Mark config as reachable
			reachable[manifest.Config.Digest.String()] = true

			// Mark layers as reachable
			for _, layer := range manifest.Layers {
				reachable[layer.Digest.String()] = true
			}
		}
		return nil
	})
	if err != nil {
		return 0, 0, fmt.Errorf("failed to traverse tags: %w", err)
	}

	// 2. Iterate through all blobs in storage and delete unreachable ones
	blobsDir := filepath.Join(s.path, "blobs", "sha256")
	entries, err := os.ReadDir(blobsDir)
	if os.IsNotExist(err) {
		return 0, 0, nil // Nothing to prune
	}
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read blobs directory: %w", err)
	}

	deletedCount := 0
	var deletedSize int64

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		digestStr := "sha256:" + entry.Name()
		if !reachable[digestStr] {
			info, err := entry.Info()
			if err == nil {
				deletedSize += info.Size()
			}

			path := filepath.Join(blobsDir, entry.Name())
			if err := os.Remove(path); err != nil {
				return deletedCount, deletedSize, fmt.Errorf("failed to remove blob %s: %w", path, err)
			}
			deletedCount++
		}
	}

	return deletedCount, deletedSize, nil
}

// interface guard
var _ content.Storage = &oci.Store{}

// pushBlob pushes content if it doesn't already exist
func (s *Store) pushBlob(ctx context.Context, desc ocispec.Descriptor, r io.Reader) error {
	exists, err := s.oci.Exists(ctx, desc)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return s.oci.Push(ctx, desc, r)
}
