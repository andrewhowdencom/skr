package source

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/andrewhowdencom/skr/pkg/skill"
	"github.com/andrewhowdencom/skr/pkg/store"
)

// Fetcher defines how to fetch and prepare a skill source to be stored as an OCI artifact.
type Fetcher interface {
	// Fetch retrieves the skill from the reference, builds it if necessary,
	// stores it in the local store, and returns the resulting Reference (which will map to an OCI reference).
	Fetch(ctx context.Context, st *store.Store, ref Reference) (Reference, error)
}

// GetFetcher returns the appropriate Fetcher based on the parsed reference schema.
func GetFetcher(ref Reference) (Fetcher, error) {
	switch ref.Schema {
	case SchemaFile:
		return &FileFetcher{}, nil
	case SchemaGit:
		return &GitFetcher{}, nil
	case SchemaOCI:
		return &OCIFetcher{}, nil
	case SchemaHTTP:
		return nil, fmt.Errorf("HTTP(S) static hosting is not yet implemented. To install a git repository over HTTPS, use the 'git+https://' schema")
	default:
		return nil, fmt.Errorf("unknown schema %s", ref.Schema)
	}
}

func buildAnnotationsFromSkill(s *skill.Skill) map[string]string {
	annotations := make(map[string]string)
	annotations["org.opencontainers.image.title"] = s.Name
	if s.Description != "" {
		annotations["org.opencontainers.image.description"] = s.Description
		annotations["com.skr.description"] = s.Description
	}
	if s.Metadata.Author != "" {
		annotations["com.skr.author"] = s.Metadata.Author
	}
	if s.Metadata.Version != "" {
		annotations["com.skr.version"] = s.Metadata.Version
	}
	if len(s.Dependencies) > 0 {
		if depsJSON, err := json.Marshal(s.Dependencies); err == nil {
			annotations["com.skr.dependencies"] = string(depsJSON)
		}
	}
	return annotations
}
