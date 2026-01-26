package registry

import (
	"context"
	"fmt"

	skrauth "github.com/andrewhowdencom/skr/pkg/auth"
	"github.com/andrewhowdencom/skr/pkg/store"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/credentials"
	"oras.land/oras-go/v2/registry/remote/retry"
)

// Push uploads a skill artifact from the local store to a remote registry.
func Push(ctx context.Context, st *store.Store, ref string) error {
	repo, err := remote.NewRepository(ref)
	if err != nil {
		return fmt.Errorf("invalid reference %s: %w", ref, err)
	}

	// Find credentials for the registry
	// ORAS client automatically uses the credential store helper if configured.
	// We inject our custom store backed by keyring.
	repo.Client = &auth.Client{
		Client:     retry.DefaultClient,
		Cache:      auth.DefaultCache,
		Credential: credentials.Credential(skrauth.NewStore()), // Wraps Store into CredentialFunc
	}

	// 2. Resolve Local Artifact
	_, err = st.Resolve(ctx, ref)
	if err != nil {
		return fmt.Errorf("reference %s not found in local store: %w", ref, err)
	}

	// 3. Copy from Local Store to Remote Repo
	_, err = oras.Copy(ctx, st, ref, repo, ref, oras.DefaultCopyOptions)
	if err != nil {
		return fmt.Errorf("failed to push %s: %w", ref, err)
	}

	return nil
}

// Pull downloads a skill artifact from a remote registry to the local store.
func Pull(ctx context.Context, st *store.Store, ref string) error {
	repo, err := remote.NewRepository(ref)
	if err != nil {
		return fmt.Errorf("invalid reference %s: %w", ref, err)
	}

	repo.Client = &auth.Client{
		Client:     retry.DefaultClient,
		Cache:      auth.DefaultCache,
		Credential: credentials.Credential(skrauth.NewStore()),
	}

	// 2. Copy from Remote Repo to Local Store
	// We copy the tagged reference.
	_, err = oras.Copy(ctx, repo, ref, st, ref, oras.DefaultCopyOptions)
	if err != nil {
		return fmt.Errorf("failed to pull %s: %w", ref, err)
	}

	return nil
}
