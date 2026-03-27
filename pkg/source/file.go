package source

import (
	"context"
	"fmt"
	"os"

	"github.com/andrewhowdencom/skr/pkg/skill"
	"github.com/andrewhowdencom/skr/pkg/store"
)

type FileFetcher struct{}

func (f *FileFetcher) Fetch(ctx context.Context, st *store.Store, ref Reference) (Reference, error) {
	// Validate it exists
	info, err := os.Stat(ref.Path)
	if err != nil {
		return Reference{}, fmt.Errorf("invalid file path %s: %w", ref.Path, err)
	}
	if !info.IsDir() {
		return Reference{}, fmt.Errorf("file source path %s must be a directory", ref.Path)
	}

	// Read SKILL.md to build annotations
	s, err := skill.LoadUnverified(ref.Path)
	if err != nil {
		return Reference{}, fmt.Errorf("failed to load skill from file path %s: %w", ref.Path, err)
	}

	tag := fmt.Sprintf("%s:latest", s.Name)
	if s.Metadata.Author != "" && s.Metadata.Version != "" {
		tag = fmt.Sprintf("%s/%s:%s", s.Metadata.Author, s.Name, s.Metadata.Version)
	}

	// Prepare Annotations map
	annotations := buildAnnotationsFromSkill(s)

	if err := st.Build(ctx, ref.Path, tag, annotations); err != nil {
		return Reference{}, fmt.Errorf("failed to build local file skill %s: %w", ref.Path, err)
	}

	// Returning the newly created OCI reference
	return ParseReference("oci://" + tag), nil
}
