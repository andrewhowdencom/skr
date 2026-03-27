package source

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/andrewhowdencom/skr/pkg/skill"
	"github.com/andrewhowdencom/skr/pkg/store"
)

type GitFetcher struct{}

func (f *GitFetcher) Fetch(ctx context.Context, st *store.Store, ref Reference) (Reference, error) {
	tempDir, err := os.MkdirTemp("", "skr-git-clone-*")
	if err != nil {
		return Reference{}, fmt.Errorf("failed to create temp dir for git clone: %w", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Construct Git URL
	fetchURL := ref.Path

	cmd := exec.CommandContext(ctx, "git", "clone", fetchURL, tempDir)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return Reference{}, fmt.Errorf("failed to clone git repository %s: %s (%w)", fetchURL, strings.TrimSpace(stderr.String()), err)
	}

	// Checkout specific ref if provided
	if ref.Spec != "" && ref.Spec != "latest" {
		checkoutCmd := exec.CommandContext(ctx, "git", "-C", tempDir, "checkout", ref.Spec)
		var coStderr bytes.Buffer
		checkoutCmd.Stderr = &coStderr
		if err := checkoutCmd.Run(); err != nil {
			return Reference{}, fmt.Errorf("failed to checkout ref %s: %s (%w)", ref.Spec, strings.TrimSpace(coStderr.String()), err)
		}
	}

	// Get short SHA for tagging
	shaCmd := exec.CommandContext(ctx, "git", "-C", tempDir, "rev-parse", "--short", "HEAD")
	shaBytes, err := shaCmd.Output()
	var shortSHA string
	if err == nil && len(shaBytes) > 0 {
		shortSHA = string(bytes.TrimSpace(shaBytes))
	}

	s, err := skill.LoadUnverified(tempDir)
	if err != nil {
		return Reference{}, fmt.Errorf("failed to load skill from git repository %s: %w", ref.Path, err)
	}

	tag := fmt.Sprintf("%s:latest", s.Name)
	if s.Metadata.Author != "" && s.Metadata.Version != "" {
		tag = fmt.Sprintf("%s/%s:%s", s.Metadata.Author, s.Name, s.Metadata.Version)
	} else if shortSHA != "" {
		tag = fmt.Sprintf("%s:%s", s.Name, shortSHA)
	}

	annotations := buildAnnotationsFromSkill(s)
	annotations["org.opencontainers.image.source"] = fetchURL

	if err := st.Build(ctx, tempDir, tag, annotations); err != nil {
		return Reference{}, fmt.Errorf("failed to build local oci image from git %s: %w", ref.Path, err)
	}

	return ParseReference("oci://" + tag), nil
}
