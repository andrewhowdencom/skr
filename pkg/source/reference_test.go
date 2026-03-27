package source

import (
	"testing"
)

func TestParseReference(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want Reference
	}{
		{
			name: "oci with schema and tag",
			raw:  "oci://ghcr.io/user/skill:v1.0.0",
			want: Reference{Original: "oci://ghcr.io/user/skill:v1.0.0", Schema: SchemaOCI, Path: "ghcr.io/user/skill", Spec: "v1.0.0"},
		},
		{
			name: "oci implicit without tag",
			raw:  "ghcr.io/user/skill",
			want: Reference{Original: "ghcr.io/user/skill", Schema: SchemaOCI, Path: "ghcr.io/user/skill", Spec: "latest"},
		},
		{
			name: "oci implicit with tag",
			raw:  "ghcr.io/user/skill:latest",
			want: Reference{Original: "ghcr.io/user/skill:latest", Schema: SchemaOCI, Path: "ghcr.io/user/skill", Spec: "latest"},
		},
		{
			name: "oci implicit with port and without tag",
			raw:  "localhost:5000/user/skill",
			want: Reference{Original: "localhost:5000/user/skill", Schema: SchemaOCI, Path: "localhost:5000/user/skill", Spec: "latest"},
		},
		{
			name: "oci implicit with port and tag",
			raw:  "localhost:5000/user/skill:v2",
			want: Reference{Original: "localhost:5000/user/skill:v2", Schema: SchemaOCI, Path: "localhost:5000/user/skill", Spec: "v2"},
		},
		{
			name: "file with schema",
			raw:  "file:///path/to/skill",
			want: Reference{Original: "file:///path/to/skill", Schema: SchemaFile, Path: "/path/to/skill", Spec: "latest"},
		},
		{
			name: "git with schema and refspec",
			raw:  "git://github.com/user/skill.git#main",
			want: Reference{Original: "git://github.com/user/skill.git#main", Schema: SchemaGit, Path: "git://github.com/user/skill.git", Spec: "main"},
		},
		{
			name: "git https with schema and refspec",
			raw:  "git+https://github.com/user/skill.git#v1.2.3",
			want: Reference{Original: "git+https://github.com/user/skill.git#v1.2.3", Schema: SchemaGit, Path: "https://github.com/user/skill.git", Spec: "v1.2.3"},
		},
		{
			name: "git ssh without refspec",
			raw:  "git@github.com:user/skill.git",
			want: Reference{Original: "git@github.com:user/skill.git", Schema: SchemaGit, Path: "git@github.com:user/skill.git", Spec: ""},
		},
		{
			name: "http pure defaults to http schema",
			raw:  "https://github.com/user/skill.git",
			want: Reference{Original: "https://github.com/user/skill.git", Schema: SchemaHTTP, Path: "https://github.com/user/skill.git", Spec: "latest"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseReference(tt.raw)
			if got.Schema != tt.want.Schema {
				t.Errorf("ParseReference().Schema = %v, want %v", got.Schema, tt.want.Schema)
			}
			if got.Path != tt.want.Path {
				t.Errorf("ParseReference().Path = %v, want %v", got.Path, tt.want.Path)
			}
			if got.Spec != tt.want.Spec {
				t.Errorf("ParseReference().Spec = %v, want %v", got.Spec, tt.want.Spec)
			}
		})
	}
}
