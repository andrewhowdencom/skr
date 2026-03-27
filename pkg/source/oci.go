package source

import (
	"context"
	"fmt"

	"github.com/andrewhowdencom/skr/pkg/registry"
	"github.com/andrewhowdencom/skr/pkg/store"
)

type OCIFetcher struct{}

func (f *OCIFetcher) Fetch(ctx context.Context, st *store.Store, ref Reference) (Reference, error) {
	pullRef := ref.Path
	if ref.Spec != "" {
		pullRef += ":" + ref.Spec
	}

	if _, err := registry.Pull(ctx, st, pullRef); err != nil {
		return Reference{}, fmt.Errorf("failed to pull OCI reference %s: %w", pullRef, err)
	}

	// Returns the identical reference
	return ref, nil
}
