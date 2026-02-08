package resolution

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/andrewhowdencom/skr/pkg/store"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// PullFunc is a function that pulls a reference into the store.
type PullFunc func(context.Context, string) error

// Resolver handles dependency resolution for skills.
type Resolver struct {
	store  *store.Store
	puller PullFunc
}

// New creates a new Resolver.
func New(st *store.Store) *Resolver {
	return &Resolver{store: st}
}

// SetPuller sets the function to call when an artifact is missing from the store.
func (r *Resolver) SetPuller(puller PullFunc) {
	r.puller = puller
}

// Resolve resolves the full list of artifacts required for the given root reference.
// It returns a list of all unique artifacts (including dependencies) that need to be installed.
// It uses BFS traversal and detects circular dependencies.
func (r *Resolver) Resolve(ctx context.Context, rootRef string) ([]string, error) {
	queue := []string{rootRef}
	visited := make(map[string]bool)
	var resolved []string

	for len(queue) > 0 {
		currentRef := queue[0]
		queue = queue[1:]

		if visited[currentRef] {
			continue
		}
		visited[currentRef] = true
		resolved = append(resolved, currentRef)

		// Fetch Manifest to get dependencies from annotations
		desc, err := r.store.Resolve(ctx, currentRef)
		if err != nil {
			// Try pulling if configured
			if r.puller != nil {
				if pullErr := r.puller(ctx, currentRef); pullErr == nil {
					// Retry resolve after pull
					desc, err = r.store.Resolve(ctx, currentRef)
				} else {
					// Return original error wrapped with pull error context
					return nil, fmt.Errorf("failed to resolve %s locally and pull failed: %v", currentRef, pullErr)
				}
			}

			if err != nil {
				return nil, fmt.Errorf("failed to resolve %s: %w", currentRef, err)
			}
		}

		manifestReader, err := r.store.Fetch(ctx, desc)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch manifest for %s: %w", currentRef, err)
		}
		defer manifestReader.Close()

		var manifest ocispec.Manifest
		if err := json.NewDecoder(manifestReader).Decode(&manifest); err != nil {
			return nil, fmt.Errorf("failed to decode manifest for %s: %w", currentRef, err)
		}

		// Parse Dependencies from Annotation
		if depsJSON, ok := manifest.Annotations["com.skr.dependencies"]; ok {
			var deps []string
			if err := json.Unmarshal([]byte(depsJSON), &deps); err != nil {
				return nil, fmt.Errorf("failed to parse dependencies for %s: %w", currentRef, err)
			}

			// Add unseen deps to queue
			for _, dep := range deps {
				// TODO: Better cycle detection / version conflict warning here?
				// For now, naive unique string check.
				if !visited[dep] {
					queue = append(queue, dep)
				}
			}
		}
	}

	return resolved, nil
}
