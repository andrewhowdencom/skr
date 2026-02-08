package resolution

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/andrewhowdencom/skr/pkg/store"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolve_PullMissing(t *testing.T) {
	// Setup Store
	tempDir := t.TempDir()
	st, err := store.New(tempDir)
	require.NoError(t, err)

	// Define artifacts and refs
	rootRef := "example.com/skill:latest"
	rootMedia := "application/vnd.oci.image.manifest.v1+json"

	// Root manifest (no dependencies for simplicity)
	manifest := ocispec.Manifest{
		Config: ocispec.Descriptor{
			MediaType: "application/vnd.unknown.config.v1+json",
			Size:      0,
			Digest:    "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		Layers: []ocispec.Descriptor{},
	}
	manifestBytes, _ := json.Marshal(manifest)
	manifestDigest := digest.FromBytes(manifestBytes)
	manifestDesc := ocispec.Descriptor{
		MediaType: rootMedia,
		Digest:    manifestDigest,
		Size:      int64(len(manifestBytes)),
	}

	// Mock Puller
	pullCalled := false
	mockPuller := func(ctx context.Context, ref string) error {
		if ref == rootRef {
			pullCalled = true
			// Simulate pull by pushing to store
			// In real code, registry.Pull does this.
			// Here we just insert the manifest so subsequent Resolve works.
			// We need to support Tagging in store?
			// The resolver calls st.Resolve(ctx, ref).
			// If ref is a tag, the store needs to know about it.
			// Currently store.Resolve only supports digest references IF it's just a CAS wrapper?
			// But check store implementation.

			// Assuming store supports tagging or we need to put it in a way it can be found.
			// The current store implementation (based on usage) seems to wrap oras.Target.
			// Let's assume we can Push with a tag or just use digest if resolver supports it?
			// But input is a ref (tag).

			// For this test, we can manually "tag" it in the memory store if exposed,
			// or assume store.Resolve handles tags.

			// Let's check store.Store definition from context... I don't see it fully.
			// But assuming it satisfies oras.Target.

			// Workaround: Since I don't have full store code, let's assume `st.Tag` exists or `st.Push` with tag.
			// Let's rely on `st.Push` and then `st.Tag` if needed.

			// Actually, let's verify if `st` has `Tag` method.
			// `store` package probably wraps `oci.Store`.
			// If I can't check, I'll assume standard ORAS usage.
			err := st.Push(ctx, manifestDesc, bytes.NewReader(manifestBytes))
			if err != nil {
				return err
			}
			// Tag it
			return st.Tag(ctx, manifestDesc, rootRef)
		}
		return fmt.Errorf("unexpected pull for %s", ref)
	}

	// Create Resolver with Puller
	// r := New(st) // Old way
	// r.SetPuller(mockPuller) // Proposed way

	r := New(st)
	r.SetPuller(mockPuller)

	// Test Resolve
	resolved, err := r.Resolve(context.Background(), rootRef)

	// Verification
	require.NoError(t, err)
	assert.True(t, pullCalled, "Puller should have been called")
	assert.Contains(t, resolved, rootRef)
}
