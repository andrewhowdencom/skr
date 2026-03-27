package source

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/andrewhowdencom/skr/pkg/store"
)

func TestGitFetcher_Fetch(t *testing.T) {
	// Create a temp directory for the store
	storeDir, err := os.MkdirTemp("", "skr-test-store-*")
	if err != nil {
		t.Fatalf("failed to create store temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(storeDir) }()

	st, err := store.New(storeDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// Create a mock git repository to clone from
	repoDir, err := os.MkdirTemp("", "skr-test-git-repo-*")
	if err != nil {
		t.Fatalf("failed to create repo temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(repoDir) }()

	// Initialize git repo
	if err := exec.Command("git", "init", repoDir).Run(); err != nil {
		t.Fatalf("failed to init git: %v", err)
	}
	_ = exec.Command("git", "-C", repoDir, "config", "user.email", "test@example.com").Run()
	_ = exec.Command("git", "-C", repoDir, "config", "user.name", "tester").Run()

	skillYaml := `---
name: git-test-skill
metadata:
  version: 2.0.0
  author: git-tester
---
`
	if err := os.WriteFile(filepath.Join(repoDir, "SKILL.md"), []byte(skillYaml), 0644); err != nil {
		t.Fatalf("failed to write SKILL.md: %v", err)
	}

	_ = exec.Command("git", "-C", repoDir, "add", "SKILL.md").Run()
	if err := exec.Command("git", "-C", repoDir, "commit", "-m", "init").Run(); err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	tests := []struct {
		name    string
		ref     Reference
		wantErr bool
	}{
		{
			name:    "valid local git repository",
			ref:     Reference{Original: "file://" + repoDir, Schema: SchemaGit, Path: repoDir, Spec: ""},
			wantErr: false,
		},
		{
			name:    "invalid git repository",
			ref:     Reference{Original: "https://does-not-exist.github.com/nothing", Schema: SchemaGit, Path: "https://does-not-exist.github.com/nothing", Spec: "main"},
			wantErr: true,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &GitFetcher{}
			got, err := f.Fetch(ctx, st, tt.ref)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitFetcher.Fetch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Schema != SchemaOCI {
					t.Errorf("Expected resulting reference schema to be OCI, got %s", got.Schema)
				}
				if got.Path != "git-tester/git-test-skill" {
					t.Errorf("Expected resulting reference path to be git-tester/git-test-skill, got %s", got.Path)
				}
				if got.Spec != "2.0.0" {
					t.Errorf("Expected resulting reference spec to be 2.0.0, got %s", got.Spec)
				}

				// Check if artifact is actually in store
				resolveRef := got.Path + ":" + got.Spec
				if _, resolveErr := st.Resolve(ctx, resolveRef); resolveErr != nil {
					t.Errorf("Expected artifact %s to be resolved in store, got error: %v", resolveRef, resolveErr)
				}
			}
		})
	}
}
