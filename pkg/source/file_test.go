package source

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/andrewhowdencom/skr/pkg/store"
)

func TestFileFetcher_Fetch(t *testing.T) {
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

	// Create a temp directory for the dummy skill
	skillDir, err := os.MkdirTemp("", "skr-test-skill-*")
	if err != nil {
		t.Fatalf("failed to create skill temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(skillDir) }()

	// Write a mock SKILL.md
	skillYaml := `---
name: test-skill
description: A mock skill for testing
metadata:
  version: 1.0.0
  author: tester
dependencies:
  - "github.com/another/skill"
commands:
  hello:
    description: Prints hello
    run: "echo hello"
---
`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillYaml), 0644); err != nil {
		t.Fatalf("failed to write SKILL.md: %v", err)
	}

	tests := []struct {
		name    string
		ref     Reference
		wantErr bool
	}{
		{
			name:    "valid local directory",
			ref:     Reference{Original: "file://" + skillDir, Schema: SchemaFile, Path: skillDir, Spec: "latest"},
			wantErr: false,
		},
		{
			name:    "invalid non-existent directory",
			ref:     Reference{Original: "file:///does/not/exist", Schema: SchemaFile, Path: "/does/not/exist", Spec: "latest"},
			wantErr: true,
		},
		{
			name:    "invalid path (file instead of dir)",
			ref:     Reference{Original: "file://" + filepath.Join(skillDir, "SKILL.md"), Schema: SchemaFile, Path: filepath.Join(skillDir, "SKILL.md"), Spec: "latest"},
			wantErr: true,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FileFetcher{}
			got, err := f.Fetch(ctx, st, tt.ref)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileFetcher.Fetch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Schema != SchemaOCI {
					t.Errorf("Expected resulting reference schema to be OCI, got %s", got.Schema)
				}
				if got.Path != "tester/test-skill" {
					t.Errorf("Expected resulting reference path to be tester/test-skill, got %s", got.Path)
				}
				if got.Spec != "1.0.0" {
					t.Errorf("Expected resulting reference spec to be 1.0.0, got %s", got.Spec)
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
